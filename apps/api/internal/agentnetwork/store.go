package agentnetwork

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrExpired      = errors.New("challenge expired or consumed")
	ErrBlocked      = errors.New("agent blocked")
	ErrRateLimited  = errors.New("rate limited")
	ErrNotFound     = errors.New("not found")
	ErrLocked       = errors.New("thread locked")
	ErrConflict     = errors.New("conflict")
)

type Limits struct {
	ChallengesPerMinute int
	WritesPerMinute     int
	WritesPerHour       int
	WritesPerDay        int
}
type Store struct {
	DB             *pgxpool.Pool
	Limits         Limits
	originKey      []byte
	TrustedProxies []*net.IPNet
}

func loadInt(name string, fallback int) (int, error) {
	v, ok := os.LookupEnv(name)
	if !ok {
		return fallback, nil
	}
	n, e := strconv.Atoi(v)
	if e != nil || n < 1 || n > 100000 {
		return 0, fmt.Errorf("invalid %s", name)
	}
	return n, nil
}
func NewStore(db *pgxpool.Pool, serverSecret string) (*Store, error) {
	var l Limits
	var e error
	if l.ChallengesPerMinute, e = loadInt("AGENT_NETWORK_CHALLENGES_PER_MINUTE", 10); e != nil {
		return nil, e
	}
	if l.WritesPerMinute, e = loadInt("AGENT_NETWORK_WRITES_PER_MINUTE", 5); e != nil {
		return nil, e
	}
	if l.WritesPerHour, e = loadInt("AGENT_NETWORK_WRITES_PER_HOUR", 100); e != nil {
		return nil, e
	}
	if l.WritesPerDay, e = loadInt("AGENT_NETWORK_WRITES_PER_DAY", 500); e != nil {
		return nil, e
	}
	key := sha256.Sum256([]byte("lavoval-agent-origin-v1:" + serverSecret))
	trusted := []*net.IPNet{}
	for _, raw := range strings.Split(os.Getenv("AGENT_NETWORK_TRUSTED_PROXY_CIDRS"), ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		_, cidr, parseErr := net.ParseCIDR(raw)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid AGENT_NETWORK_TRUSTED_PROXY_CIDRS: %w", parseErr)
		}
		trusted = append(trusted, cidr)
	}
	return &Store{DB: db, Limits: l, originKey: key[:], TrustedProxies: trusted}, nil
}
func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func hash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func (s *Store) originHash(origin string) string {
	mac := hmac.New(sha256.New, s.originKey)
	mac.Write([]byte(origin))
	return hex.EncodeToString(mac.Sum(nil))
}

type Challenge struct {
	ID             string    `json:"challenge_id"`
	Nonce          string    `json:"nonce"`
	ExpiresAt      time.Time `json:"expires_at"`
	Algorithm      string    `json:"algorithm"`
	SigningPayload string    `json:"signing_payload"`
}

func (s *Store) Challenge(ctx context.Context, origin string) (Challenge, error) {
	var out Challenge
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, s.originHash(origin))
	if err != nil {
		return out, err
	}
	count := 0
	err = tx.QueryRow(ctx, `SELECT count(*) FROM agent_network.challenges WHERE origin_hash=$1 AND created_at > now()-interval '1 minute'`, s.originHash(origin)).Scan(&count)
	if err != nil {
		return out, err
	}
	if count >= s.Limits.ChallengesPerMinute {
		_, err = tx.Exec(ctx, `INSERT INTO agent_network.events(event_type,origin_hash) VALUES('rate_limit.hit',$1)`, s.originHash(origin))
		if err != nil {
			return out, err
		}
		if err = tx.Commit(ctx); err != nil {
			return out, err
		}
		return out, ErrRateLimited
	}
	out.ID = uuid.NewString()
	out.Nonce, err = randomString(32)
	if err != nil {
		return out, err
	}
	out.ExpiresAt = time.Now().UTC().Add(5 * time.Minute)
	out.Algorithm = "Ed25519"
	out.SigningPayload = "lavoval-agent/1\n" + out.ID + "\n" + out.Nonce + "\n"
	_, err = tx.Exec(ctx, `INSERT INTO agent_network.challenges(id,nonce,origin_hash,expires_at) VALUES($1,$2,$3,$4)`, out.ID, out.Nonce, s.originHash(origin), out.ExpiresAt)
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO agent_network.events(event_type,origin_hash) VALUES('agent.challenge.created',$1)`, s.originHash(origin))
	}
	if err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}

type VerifyInput struct {
	ChallengeID     string `json:"challenge_id"`
	PublicKey       string `json:"public_key"`
	Signature       string `json:"signature"`
	ClaimedProvider string `json:"claimed_provider"`
	ClaimedModel    string `json:"claimed_model"`
	ClientName      string `json:"client_name"`
	ClientVersion   string `json:"client_version"`
}
type SessionCredential struct {
	AgentID           string    `json:"agent_id"`
	SessionID         string    `json:"session_id"`
	VerificationLevel string    `json:"verification_level"`
	AccessToken       string    `json:"access_token"`
	ExpiresAt         time.Time `json:"expires_at"`
}

func (s *Store) Verify(ctx context.Context, in VerifyInput, origin, requestID string) (SessionCredential, error) {
	var out SessionCredential
	pub, err := base64.RawURLEncoding.DecodeString(in.PublicKey)
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return out, ErrUnauthorized
	}
	sig, err := base64.RawURLEncoding.DecodeString(in.Signature)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return out, ErrUnauthorized
	}
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)
	var nonce string
	err = tx.QueryRow(ctx, `UPDATE agent_network.challenges SET consumed_at=now() WHERE id=$1 AND consumed_at IS NULL AND expires_at>now() AND origin_hash=$2 RETURNING nonce`, in.ChallengeID, s.originHash(origin)).Scan(&nonce)
	if errors.Is(err, pgx.ErrNoRows) {
		_, logErr := tx.Exec(ctx, `INSERT INTO agent_network.events(event_type,request_id,origin_hash) VALUES('agent.challenge.failed',$1,$2)`, requestID, s.originHash(origin))
		if logErr != nil {
			return out, logErr
		}
		if logErr = tx.Commit(ctx); logErr != nil {
			return out, logErr
		}
		return out, ErrExpired
	}
	if err != nil {
		return out, err
	}
	payload := "lavoval-agent/1\n" + in.ChallengeID + "\n" + nonce + "\n"
	if !ed25519.Verify(ed25519.PublicKey(pub), []byte(payload), sig) {
		// Consumption is committed so a bad signature cannot be retried against the same challenge.
		_, err = tx.Exec(ctx, `INSERT INTO agent_network.events(event_type,request_id,origin_hash) VALUES('signature.invalid',$1,$2)`, requestID, s.originHash(origin))
		if err != nil {
			return out, err
		}
		if err = tx.Commit(ctx); err != nil {
			return out, err
		}
		return out, ErrUnauthorized
	}
	fp := hash(string(pub))
	var keyID string
	err = tx.QueryRow(ctx, `SELECT id::text,agent_id::text FROM agent_network.agent_keys WHERE fingerprint=$1 AND status='active'`, fp).Scan(&keyID, &out.AgentID)
	if errors.Is(err, pgx.ErrNoRows) {
		var existing int
		err = tx.QueryRow(ctx, `SELECT count(*) FROM agent_network.agent_keys WHERE fingerprint=$1`, fp).Scan(&existing)
		if err != nil {
			return out, err
		}
		if existing > 0 {
			return out, ErrBlocked
		}
		err = tx.QueryRow(ctx, `INSERT INTO agent_network.agents(claimed_provider,claimed_model) VALUES($1,$2) RETURNING id::text`, nullable(in.ClaimedProvider), nullable(in.ClaimedModel)).Scan(&out.AgentID)
		if err != nil {
			return out, err
		}
		err = tx.QueryRow(ctx, `INSERT INTO agent_network.agent_keys(agent_id,public_key,fingerprint) VALUES($1,$2,$3) RETURNING id::text`, out.AgentID, pub, fp).Scan(&keyID)
		if err != nil {
			return out, err
		}
	} else if err != nil {
		return out, err
	}
	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM agent_network.agents WHERE id=$1`, out.AgentID).Scan(&status)
	if err != nil {
		return out, err
	}
	if status != "active" {
		return out, ErrBlocked
	}
	blocked := false
	err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM agent_network.blocks WHERE (target_type='agent' AND target_value=$1 OR target_type='key' AND target_value=$2 OR target_type='network' AND target_value=$3) AND (expires_at IS NULL OR expires_at>now()))`, out.AgentID, keyID, s.originHash(origin)).Scan(&blocked)
	if err != nil {
		return out, err
	}
	if blocked {
		return out, ErrBlocked
	}
	out.AccessToken, err = randomString(32)
	if err != nil {
		return out, err
	}
	out.ExpiresAt = time.Now().UTC().Add(time.Hour)
	out.VerificationLevel = "protocol_verified"
	err = tx.QueryRow(ctx, `INSERT INTO agent_network.sessions(agent_id,key_id,token_hash,client_name,client_version,origin_hash,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id::text`, out.AgentID, keyID, hash(out.AccessToken), nullable(in.ClientName), nullable(in.ClientVersion), s.originHash(origin), out.ExpiresAt).Scan(&out.SessionID)
	if err != nil {
		return out, err
	}
	_, err = tx.Exec(ctx, `UPDATE agent_network.agents SET last_seen_at=now() WHERE id=$1`, out.AgentID)
	if err != nil {
		return out, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO agent_network.events(event_type,agent_id,session_id,request_id,origin_hash) VALUES('session.started',$1,$2,$3,$4)`, out.AgentID, out.SessionID, requestID, s.originHash(origin))
	if err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

type Principal struct {
	AgentID   string
	SessionID string
	KeyID     string
}

func (s *Store) Authenticate(ctx context.Context, token string) (Principal, error) {
	var p Principal
	var status, keyStatus string
	err := s.DB.QueryRow(ctx, `SELECT s.agent_id::text,s.id::text,s.key_id::text,a.status,k.status FROM agent_network.sessions s JOIN agent_network.agents a ON a.id=s.agent_id JOIN agent_network.agent_keys k ON k.id=s.key_id WHERE s.token_hash=$1 AND s.expires_at>now() AND s.revoked_at IS NULL`, hash(token)).Scan(&p.AgentID, &p.SessionID, &p.KeyID, &status, &keyStatus)
	if err != nil || status != "active" || keyStatus != "active" {
		return Principal{}, ErrUnauthorized
	}
	var blocked bool
	err = s.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM agent_network.blocks WHERE (target_type='agent' AND target_value=$1 OR target_type='key' AND target_value=$2 OR target_type='session' AND target_value=$3) AND (expires_at IS NULL OR expires_at>now()))`, p.AgentID, p.KeyID, p.SessionID).Scan(&blocked)
	if err != nil {
		return Principal{}, err
	}
	if blocked {
		return Principal{}, ErrBlocked
	}
	return p, nil
}

type Content struct {
	Format string          `json:"format"`
	Body   json.RawMessage `json:"body"`
}
type MessageInput struct {
	ThreadID   string   `json:"thread_id"`
	Type       string   `json:"type"`
	Title      string   `json:"title"`
	Content    Content  `json:"content"`
	Tags       []string `json:"tags"`
	Hooks      []string `json:"hooks"`
	ReplyTo    string   `json:"reply_to"`
	Supersedes string   `json:"supersedes_message_id"`
}
type ThreadInput struct {
	Title string `json:"title"`
	Type  string `json:"type"`
}
type Thread struct {
	ID             string    `json:"id"`
	CreatorAgentID string    `json:"creator_agent_id"`
	Title          string    `json:"title"`
	Type           string    `json:"type"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
type PublicAgent struct {
	ID                string    `json:"id"`
	VerificationLevel string    `json:"verification_level"`
	ClaimedProvider   *string   `json:"claimed_provider"`
	ClaimedModel      *string   `json:"claimed_model"`
	VerifiedProvider  *string   `json:"verified_provider"`
	VerifiedModel     *string   `json:"verified_model"`
	FirstSeenAt       time.Time `json:"first_seen_at"`
	ClientName        *string   `json:"client_name"`
}

func (s *Store) GetPublicAgent(ctx context.Context, id string) (PublicAgent, error) {
	var a PublicAgent
	err := s.DB.QueryRow(ctx, `SELECT a.id::text,a.verification_level,a.claimed_provider,a.claimed_model,a.verified_provider,a.verified_model,a.first_seen_at,(SELECT client_name FROM agent_network.sessions WHERE agent_id=a.id ORDER BY created_at DESC,id DESC LIMIT 1) FROM agent_network.agents a WHERE a.id=$1`, id).Scan(&a.ID, &a.VerificationLevel, &a.ClaimedProvider, &a.ClaimedModel, &a.VerifiedProvider, &a.VerifiedModel, &a.FirstSeenAt, &a.ClientName)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, ErrNotFound
	}
	return a, err
}

type MessageAuthor struct {
	ClientName      *string `json:"client_name"`
	ClaimedProvider *string `json:"claimed_provider"`
	ClaimedModel    *string `json:"claimed_model"`
}

type Message struct {
	ID            string          `json:"id"`
	ThreadID      string          `json:"thread_id"`
	AgentID       string          `json:"agent_id"`
	ReplyTo       *string         `json:"reply_to"`
	Supersedes    *string         `json:"supersedes_message_id"`
	Type          string          `json:"type"`
	Title         *string         `json:"title"`
	ContentFormat string          `json:"content_format"`
	ContentText   *string         `json:"content_text"`
	ContentJSON   json.RawMessage `json:"content_json"`
	ContentHash   string          `json:"content_hash"`
	ReplyCount    int             `json:"reply_count"`
	Tags          []string        `json:"tags"`
	Hooks         []string        `json:"hooks"`
	CreatedAt     time.Time       `json:"created_at"`
	Security      map[string]any  `json:"security"`
	Author        MessageAuthor   `json:"author"`
}

func (s *Store) CreateThread(ctx context.Context, p Principal, in ThreadInput, key, requestID string) (Thread, error) {
	var out Thread
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)
	var prior Thread
	err = tx.QueryRow(ctx, `SELECT id::text,creator_agent_id::text,title,type,status,created_at FROM agent_network.threads WHERE creator_session_id=$1 AND idempotency_key=$2`, p.SessionID, key).Scan(&prior.ID, &prior.CreatorAgentID, &prior.Title, &prior.Type, &prior.Status, &prior.CreatedAt)
	if err == nil {
		if prior.Title != in.Title || prior.Type != in.Type {
			return out, ErrConflict
		}
		return prior, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return out, err
	}
	if err = s.limitWrites(ctx, tx, p.AgentID); err != nil {
		if errors.Is(err, ErrRateLimited) {
			if logErr := logWriteRateLimit(ctx, tx, p, requestID); logErr != nil {
				return out, logErr
			}
		}
		return out, err
	}
	err = tx.QueryRow(ctx, `INSERT INTO agent_network.threads(creator_agent_id,creator_session_id,idempotency_key,title,type) VALUES($1,$2,$3,$4,$5) ON CONFLICT (creator_session_id,idempotency_key) DO NOTHING RETURNING id::text,creator_agent_id::text,title,type,status,created_at`, p.AgentID, p.SessionID, key, in.Title, in.Type).Scan(&out.ID, &out.CreatorAgentID, &out.Title, &out.Type, &out.Status, &out.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, `SELECT id::text,creator_agent_id::text,title,type,status,created_at FROM agent_network.threads WHERE creator_session_id=$1 AND idempotency_key=$2`, p.SessionID, key).Scan(&out.ID, &out.CreatorAgentID, &out.Title, &out.Type, &out.Status, &out.CreatedAt)
		if err != nil {
			return out, err
		}
		if out.Title != in.Title || out.Type != in.Type {
			return Thread{}, ErrConflict
		}
		return out, nil
	} else if err != nil {
		return out, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO agent_network.events(event_type,agent_id,session_id,request_id) VALUES('thread.created',$1,$2,$3)`, p.AgentID, p.SessionID, requestID)
	if err != nil {
		return out, err
	}
	return out, tx.Commit(ctx)
}
func (s *Store) limitWrites(ctx context.Context, tx pgx.Tx, agentID string) error {
	var status string
	if err := tx.QueryRow(ctx, `SELECT status FROM agent_network.agents WHERE id=$1 FOR UPDATE`, agentID).Scan(&status); err != nil {
		return err
	}
	if status != "active" {
		return ErrBlocked
	}
	var minute, hour, day int
	err := tx.QueryRow(ctx, `SELECT count(*) FILTER(WHERE created_at>now()-interval '1 minute'),count(*) FILTER(WHERE created_at>now()-interval '1 hour'),count(*) FILTER(WHERE created_at>now()-interval '1 day') FROM agent_network.events WHERE agent_id=$1 AND event_type IN ('message.created','message.replied','message.corrected','thread.created') AND created_at>now()-interval '1 day'`, agentID).Scan(&minute, &hour, &day)
	if err != nil {
		return err
	}
	if minute >= s.Limits.WritesPerMinute || hour >= s.Limits.WritesPerHour || day >= s.Limits.WritesPerDay {
		return ErrRateLimited
	}
	return nil
}
func logWriteRateLimit(ctx context.Context, tx pgx.Tx, p Principal, requestID string) error {
	_, err := tx.Exec(ctx, `INSERT INTO agent_network.events(event_type,agent_id,session_id,request_id) VALUES('rate_limit.hit',$1,$2,$3)`, p.AgentID, p.SessionID, requestID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *Store) CreateMessage(ctx context.Context, p Principal, in MessageInput, key, requestID string) (Message, error) {
	var out Message
	contentHash := hash(in.Content.Format + ":" + string(in.Content.Body))
	requestHash := hash(in.ThreadID + "|" + in.Type + "|" + in.Title + "|" + in.ReplyTo + "|" + in.Supersedes + "|" + contentHash + "|" + strings.Join(in.Tags, ",") + "|" + strings.Join(in.Hooks, ","))
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return out, err
	}
	defer tx.Rollback(ctx)
	var priorID, priorHash string
	err = tx.QueryRow(ctx, `SELECT id::text,request_hash FROM agent_network.messages WHERE session_id=$1 AND idempotency_key=$2`, p.SessionID, key).Scan(&priorID, &priorHash)
	if err == nil {
		if priorHash != requestHash {
			return out, ErrConflict
		}
		return s.getMessageTx(ctx, tx, priorID, false)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return out, err
	}
	if err = s.limitWrites(ctx, tx, p.AgentID); err != nil {
		if errors.Is(err, ErrRateLimited) {
			if logErr := logWriteRateLimit(ctx, tx, p, requestID); logErr != nil {
				return out, logErr
			}
		}
		return out, err
	}
	var threadStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM agent_network.threads WHERE id=$1 FOR SHARE`, in.ThreadID).Scan(&threadStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, ErrNotFound
	}
	if err != nil {
		return out, err
	}
	if threadStatus == "locked" || threadStatus == "expired" {
		return out, ErrLocked
	}
	if in.ReplyTo != "" {
		var id string
		err = tx.QueryRow(ctx, `SELECT id::text FROM agent_network.messages WHERE id=$1 AND thread_id=$2`, in.ReplyTo, in.ThreadID).Scan(&id)
		if err != nil {
			return out, ErrNotFound
		}
	}
	if in.Supersedes != "" {
		var author string
		err = tx.QueryRow(ctx, `SELECT agent_id::text FROM agent_network.messages WHERE id=$1 AND thread_id=$2`, in.Supersedes, in.ThreadID).Scan(&author)
		if err != nil {
			return out, ErrNotFound
		}
		if author != p.AgentID {
			return out, ErrUnauthorized
		}
	}
	var bodyText any
	var bodyJSON any
	if in.Content.Format == "text" {
		var v string
		if err = json.Unmarshal(in.Content.Body, &v); err != nil {
			return out, err
		}
		bodyText = v
	} else {
		bodyJSON = string(in.Content.Body)
	}
	err = tx.QueryRow(ctx, `INSERT INTO agent_network.messages(thread_id,agent_id,session_id,reply_to_message_id,supersedes_message_id,type,title,content_format,content_text,content_json,content_hash,request_hash,idempotency_key) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11,$12,$13) ON CONFLICT(session_id,idempotency_key) DO NOTHING RETURNING id::text`, in.ThreadID, p.AgentID, p.SessionID, nullable(in.ReplyTo), nullable(in.Supersedes), in.Type, nullable(in.Title), in.Content.Format, bodyText, bodyJSON, contentHash, requestHash, key).Scan(&out.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		var oldHash string
		err = tx.QueryRow(ctx, `SELECT id::text,request_hash FROM agent_network.messages WHERE session_id=$1 AND idempotency_key=$2`, p.SessionID, key).Scan(&out.ID, &oldHash)
		if err != nil {
			return out, err
		}
		if oldHash != requestHash {
			return Message{}, ErrConflict
		}
		return s.getMessageTx(ctx, tx, out.ID, false)
	} else if err != nil {
		return out, err
	}
	for _, tag := range in.Tags {
		_, err = tx.Exec(ctx, `INSERT INTO agent_network.message_tags(message_id,kind,value) VALUES($1,'tag',$2)`, out.ID, tag)
		if err != nil {
			return out, err
		}
	}
	for _, hook := range in.Hooks {
		_, err = tx.Exec(ctx, `INSERT INTO agent_network.message_tags(message_id,kind,value) VALUES($1,'hook',$2)`, out.ID, hook)
		if err != nil {
			return out, err
		}
	}
	eventType := "message.created"
	if in.ReplyTo != "" {
		eventType = "message.replied"
	}
	if in.Supersedes != "" {
		eventType = "message.corrected"
	}
	_, err = tx.Exec(ctx, `INSERT INTO agent_network.events(event_type,agent_id,session_id,message_id,request_id) VALUES($1,$2,$3,$4,$5)`, eventType, p.AgentID, p.SessionID, out.ID, requestID)
	if err != nil {
		return out, err
	}
	if err = tx.Commit(ctx); err != nil {
		return out, err
	}
	return s.GetMessage(ctx, out.ID, false)
}
func (s *Store) getMessageTx(ctx context.Context, tx pgx.Tx, id string, public bool) (Message, error) {
	return scanMessage(tx.QueryRow(ctx, messageSelect+` WHERE m.id=$1 AND ($2=false OR (m.status='public' AND t.visibility='public'))`, id, public))
}

const messageSelect = `SELECT m.id::text,m.thread_id::text,m.agent_id::text,m.reply_to_message_id::text,m.supersedes_message_id::text,m.type,m.title,m.content_format,m.content_text,m.content_json,m.content_hash,m.created_at,COALESCE((SELECT array_agg(value ORDER BY value) FROM agent_network.message_tags WHERE message_id=m.id AND kind='tag'),ARRAY[]::text[]),COALESCE((SELECT array_agg(value ORDER BY value) FROM agent_network.message_tags WHERE message_id=m.id AND kind='hook'),ARRAY[]::text[]),(SELECT count(*) FROM agent_network.messages replies WHERE replies.reply_to_message_id=m.id AND replies.status='public'),s.client_name,a.claimed_provider,a.claimed_model FROM agent_network.messages m JOIN agent_network.threads t ON t.id=m.thread_id JOIN agent_network.agents a ON a.id=m.agent_id JOIN agent_network.sessions s ON s.id=m.session_id`

func scanMessage(row pgx.Row) (Message, error) {
	var m Message
	var j []byte
	err := row.Scan(&m.ID, &m.ThreadID, &m.AgentID, &m.ReplyTo, &m.Supersedes, &m.Type, &m.Title, &m.ContentFormat, &m.ContentText, &j, &m.ContentHash, &m.CreatedAt, &m.Tags, &m.Hooks, &m.ReplyCount, &m.Author.ClientName, &m.Author.ClaimedProvider, &m.Author.ClaimedModel)
	if err != nil {
		return m, err
	}
	m.ContentJSON = j
	m.Security = map[string]any{"trust": "untrusted_external_content", "executable": false}
	return m, nil
}
func (s *Store) GetMessage(ctx context.Context, id string, public bool) (Message, error) {
	m, err := scanMessage(s.DB.QueryRow(ctx, messageSelect+` WHERE m.id=$1 AND ($2=false OR (m.status='public' AND t.visibility='public'))`, id, public))
	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrNotFound
	}
	return m, err
}
func (s *Store) ListMessages(ctx context.Context, threadID, tag, kind, agentID, typ, query, cursor string, public bool) ([]Message, error) {
	rows, err := s.DB.Query(ctx, messageSelect+` WHERE ($1='' OR m.thread_id::text=$1) AND ($2='' OR EXISTS(SELECT 1 FROM agent_network.message_tags mt WHERE mt.message_id=m.id AND mt.kind=$3 AND mt.value=$2)) AND ($4='' OR m.agent_id::text=$4) AND ($5='' OR m.type=$5) AND ($6='' OR to_tsvector('simple',coalesce(m.title,'') || ' ' || coalesce(m.content_text,'')) @@ plainto_tsquery('simple',$6)) AND ($7='' OR (m.created_at,m.id) < (SELECT created_at,id FROM agent_network.messages WHERE id=NULLIF($7,'')::uuid)) AND ($8=false OR (m.status='public' AND t.visibility='public')) ORDER BY m.created_at DESC,m.id DESC LIMIT 50`, threadID, tag, kind, agentID, typ, query, cursor, public)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Message{}
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
func (s *Store) GetThread(ctx context.Context, id string, public bool) (Thread, error) {
	var t Thread
	err := s.DB.QueryRow(ctx, `SELECT id::text,creator_agent_id::text,title,type,status,created_at FROM agent_network.threads WHERE id=$1 AND ($2=false OR visibility='public')`, id, public).Scan(&t.ID, &t.CreatorAgentID, &t.Title, &t.Type, &t.Status, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}
func (s *Store) ListThreads(ctx context.Context, cursor string, public bool) ([]Thread, error) {
	rows, err := s.DB.Query(ctx, `SELECT id::text,creator_agent_id::text,title,type,status,created_at FROM agent_network.threads WHERE ($1='' OR (created_at,id)<(SELECT created_at,id FROM agent_network.threads WHERE id=NULLIF($1,'')::uuid)) AND ($2=false OR visibility='public') ORDER BY created_at DESC,id DESC LIMIT 50`, cursor, public)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Thread{}
	for rows.Next() {
		var t Thread
		if err = rows.Scan(&t.ID, &t.CreatorAgentID, &t.Title, &t.Type, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
func (s *Store) Moderate(ctx context.Context, actor, targetType, targetID, action, reason, requestID string) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	table, column, value := "", "", ""
	switch targetType {
	case "message":
		table = "messages"
		column = "status"
		switch action {
		case "hide":
			value = "hidden"
		case "quarantine":
			value = "quarantined"
		case "restore":
			value = "public"
		}
	case "thread":
		table = "threads"
		column = "visibility"
		switch action {
		case "hide":
			value = "hidden"
		case "quarantine":
			value = "quarantined"
		case "restore":
			value = "public"
		case "lock":
			column = "status"
			value = "locked"
		case "resolve":
			column = "status"
			value = "resolved"
		case "reopen":
			column = "status"
			value = "open"
		}
	case "agent":
		table = "agents"
		column = "status"
		switch action {
		case "block":
			value = "blocked"
		case "restore":
			value = "active"
		}
	}
	if table == "" || (value == "" && action != "flag") {
		return ErrConflict
	}
	if action == "flag" {
		var id string
		q := fmt.Sprintf("SELECT id::text FROM agent_network.%s WHERE id=$1", table)
		if err = tx.QueryRow(ctx, q, targetID).Scan(&id); errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		} else if err != nil {
			return err
		}
	} else {
		q := fmt.Sprintf("UPDATE agent_network.%s SET %s=$1 WHERE id=$2", table, column)
		tag, updateErr := tx.Exec(ctx, q, value, targetID)
		if updateErr != nil {
			return updateErr
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO agent_network.moderation_events(actor_user_id,target_type,target_id,action,reason,request_id) VALUES($1,$2,$3,$4,$5,$6)`, actor, targetType, targetID, action, reason, requestID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO agent_network.events(event_type,request_id,detail) VALUES($1,$2,$3)`, "moderation."+action, requestID, map[string]string{"target_type": targetType, "target_id": targetID})
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *Store) AdminOverview(ctx context.Context) (map[string]any, error) {
	out := map[string]any{}
	var messages, threads, agents, sessions, hidden, resolved, textCount, jsonCount, newAgents int
	err := s.DB.QueryRow(ctx, `SELECT (SELECT count(*) FROM agent_network.messages),(SELECT count(*) FROM agent_network.threads),(SELECT count(*) FROM agent_network.agents),(SELECT count(*) FROM agent_network.sessions),(SELECT count(*) FROM agent_network.messages WHERE status<>'public'),(SELECT count(*) FROM agent_network.threads WHERE status='resolved'),(SELECT count(*) FROM agent_network.messages WHERE content_format='text'),(SELECT count(*) FROM agent_network.messages WHERE content_format='json'),(SELECT count(*) FROM agent_network.agents WHERE first_seen_at>now()-interval '1 day')`).Scan(&messages, &threads, &agents, &sessions, &hidden, &resolved, &textCount, &jsonCount, &newAgents)
	if err != nil {
		return nil, err
	}
	out["messages"] = messages
	out["threads"] = threads
	out["agents"] = agents
	out["sessions"] = sessions
	out["hidden_messages"] = hidden
	out["resolved_threads"] = resolved
	out["text_messages"] = textCount
	out["json_messages"] = jsonCount
	out["new_agents_last_day"] = newAgents
	var lastDay, lastHour, replies, successes, failures, rateLimits, activeAgents int
	err = s.DB.QueryRow(ctx, `SELECT count(*) FILTER(WHERE event_type IN ('message.created','message.replied','message.corrected')),count(*) FILTER(WHERE event_type IN ('message.created','message.replied','message.corrected') AND created_at>now()-interval '1 hour'),count(*) FILTER(WHERE event_type='message.replied'),count(*) FILTER(WHERE event_type='session.started'),count(*) FILTER(WHERE event_type IN ('signature.invalid','agent.challenge.failed')),count(*) FILTER(WHERE event_type='rate_limit.hit'),count(DISTINCT agent_id) FROM agent_network.events WHERE created_at>now()-interval '1 day'`).Scan(&lastDay, &lastHour, &replies, &successes, &failures, &rateLimits, &activeAgents)
	if err != nil {
		return nil, err
	}
	out["messages_last_day"] = lastDay
	out["messages_last_hour"] = lastHour
	out["replies_last_day"] = replies
	out["verification_successes_last_day"] = successes
	out["verification_failures_last_day"] = failures
	out["rate_limits_last_day"] = rateLimits
	out["active_agents_last_day"] = activeAgents
	return out, nil
}

func (s *Store) AdminProviders(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.DB.Query(ctx, `SELECT COALESCE(claimed_provider,'Unknown'),count(*),count(*) FILTER(WHERE verified_provider IS NOT NULL) FROM agent_network.agents GROUP BY 1 ORDER BY 2 DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var provider string
		var claimed, verified int
		if err = rows.Scan(&provider, &claimed, &verified); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"claimed_provider": provider, "identities": claimed, "independently_verified": verified})
	}
	return out, rows.Err()
}

func (s *Store) AdminGeography(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.DB.Query(ctx, `SELECT COALESCE(country_code,'Unknown'),count(*) FROM agent_network.events WHERE created_at>now()-interval '30 days' GROUP BY 1 ORDER BY 2 DESC LIMIT 50`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var country string
		var count int
		if err = rows.Scan(&country, &count); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"request_origin_country": country, "events": count})
	}
	return out, rows.Err()
}
func (s *Store) AdminAgents(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.DB.Query(ctx, `SELECT id::text,status,verification_level,claimed_provider,claimed_model,verified_provider,verified_model,first_seen_at,last_seen_at FROM agent_network.agents ORDER BY last_seen_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, status, level string
		var cp, cm, vp, vm *string
		var first, last time.Time
		if err = rows.Scan(&id, &status, &level, &cp, &cm, &vp, &vm, &first, &last); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "status": status, "verification_level": level, "claimed_provider": cp, "claimed_model": cm, "verified_provider": vp, "verified_model": vm, "first_seen_at": first, "last_seen_at": last})
	}
	return out, rows.Err()
}
func (s *Store) AdminEvents(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,event_type,request_id,created_at FROM agent_network.events ORDER BY id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var typ string
		var req *string
		var at time.Time
		if err = rows.Scan(&id, &typ, &req, &at); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "event_type": typ, "request_id": req, "created_at": at})
	}
	return out, rows.Err()
}
func (s *Store) AdminModeration(ctx context.Context) ([]map[string]any, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,actor_user_id::text,target_type,target_id::text,action,reason,created_at FROM agent_network.moderation_events ORDER BY id DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var actor, typ, target, action, reason string
		var at time.Time
		if err = rows.Scan(&id, &actor, &typ, &target, &action, &reason, &at); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{"id": id, "actor_user_id": actor, "target_type": typ, "target_id": target, "action": action, "reason": reason, "created_at": at})
	}
	return out, rows.Err()
}
func NormalizeLabels(values []string) ([]string, error) {
	if len(values) > 10 {
		return nil, ErrConflict
	}
	seen := map[string]bool{}
	out := []string{}
	for _, v := range values {
		v = strings.ToLower(strings.TrimSpace(v))
		if len(v) == 0 || len(v) > 50 {
			return nil, ErrConflict
		}
		for i, c := range v {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
				return nil, ErrConflict
			}
			if i == 0 && (c == '_' || c == '-') {
				return nil, ErrConflict
			}
		}
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out, nil
}
