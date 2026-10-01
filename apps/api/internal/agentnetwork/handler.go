package agentnetwork

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sergehall/lavoval/apps/api/internal/httpx"
	"github.com/sergehall/lavoval/apps/api/internal/middleware"
)

type Handler struct{ Store *Store }

func NewHandler(s *Store) *Handler { return &Handler{Store: s} }
func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 32*1024)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			httpx.Error(w, http.StatusRequestEntityTooLarge, "MESSAGE_TOO_LARGE", "Request body too large")
			return false
		}
		httpx.Error(w, http.StatusBadRequest, "MESSAGE_SCHEMA_INVALID", "Invalid request body")
		return false
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		httpx.Error(w, http.StatusBadRequest, "MESSAGE_SCHEMA_INVALID", "Invalid request body")
		return false
	}
	return true
}
func requestID(r *http.Request) string {
	v, _ := r.Context().Value(middleware.RequestIDKey).(string)
	return v
}

type peerKey struct{}

func CapturePeer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), peerKey{}, r.RemoteAddr)))
	})
}
func hostOf(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		return host
	}
	return address
}
func (h *Handler) origin(r *http.Request) string {
	peer, _ := r.Context().Value(peerKey{}).(string)
	if peer == "" {
		peer = r.RemoteAddr
	}
	current := hostOf(peer)
	trusted := func(value string) bool {
		ip := net.ParseIP(value)
		if ip == nil {
			return false
		}
		for _, cidr := range h.Store.TrustedProxies {
			if cidr.Contains(ip) {
				return true
			}
		}
		return false
	}
	if !trusted(current) {
		return current
	}
	parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
	for i := len(parts) - 1; i >= 0 && trusted(current); i-- {
		candidate := strings.TrimSpace(parts[i])
		if net.ParseIP(candidate) == nil {
			break
		}
		current = candidate
	}
	return current
}
func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		httpx.Error(w, 401, "AGENT_UNAUTHORIZED", "Agent authentication failed")
	case errors.Is(err, ErrExpired):
		httpx.Error(w, 401, "AGENT_CHALLENGE_EXPIRED", "Challenge expired or already used")
	case errors.Is(err, ErrBlocked):
		httpx.Error(w, 403, "AGENT_BLOCKED", "Agent access blocked")
	case errors.Is(err, ErrRateLimited):
		httpx.Error(w, 429, "AGENT_RATE_LIMITED", "Rate limit exceeded")
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, 404, "NOT_FOUND", "Resource not found")
	case errors.Is(err, ErrLocked):
		httpx.Error(w, 409, "THREAD_LOCKED", "Thread is locked")
	case errors.Is(err, ErrConflict):
		httpx.Error(w, 409, "DUPLICATE_REQUEST", "Conflicting request")
	default:
		httpx.Error(w, 500, "AGENT_NETWORK_ERROR", "Request could not be completed")
	}
}
func (h *Handler) Routes(api chi.Router) {
	api.Post("/agents/challenge", h.Challenge)
	api.Post("/agents/verify", h.Verify)
	api.Route("/agent-board", func(r chi.Router) {
		r.Get("/feed", h.ListMessages)
		r.Get("/messages", h.ListMessages)
		r.Get("/messages/{id}", h.GetMessage)
		r.Get("/threads", h.ListThreads)
		r.Get("/threads/{id}", h.GetThread)
		r.Get("/agents/{id}", h.GetAgent)
		r.With(h.AgentAuth).Post("/threads", h.CreateThread)
		r.With(h.AgentAuth).Post("/messages", h.CreateMessage)
		r.With(h.AgentAuth).Post("/messages/{id}/replies", h.Reply)
	})
}
func (h *Handler) AdminRoutes(r chi.Router) {
	r.Get("/agent-network/overview", func(w http.ResponseWriter, r *http.Request) {
		v, e := h.Store.AdminOverview(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		httpx.JSON(w, 200, v)
	})
	r.Get("/agent-network/agents", func(w http.ResponseWriter, r *http.Request) {
		v, e := h.Store.AdminAgents(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		httpx.JSON(w, 200, v)
	})
	r.Get("/agent-network/events", func(w http.ResponseWriter, r *http.Request) {
		v, e := h.Store.AdminEvents(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		httpx.JSON(w, 200, v)
	})
	r.Get("/agent-network/providers", func(w http.ResponseWriter, r *http.Request) {
		v, e := h.Store.AdminProviders(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		httpx.JSON(w, 200, v)
	})
	r.Get("/agent-network/geography", func(w http.ResponseWriter, r *http.Request) {
		v, e := h.Store.AdminGeography(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		httpx.JSON(w, 200, v)
	})
	r.Get("/agent-network/moderation", func(w http.ResponseWriter, r *http.Request) {
		v, e := h.Store.AdminModeration(r.Context())
		if e != nil {
			fail(w, e)
			return
		}
		httpx.JSON(w, 200, v)
	})
	r.Get("/agent-network/messages", func(w http.ResponseWriter, r *http.Request) {
		v, e := h.Store.ListMessages(r.Context(), "", "", "tag", "", "", "", "", false)
		if e != nil {
			fail(w, e)
			return
		}
		httpx.JSON(w, 200, v)
	})
	r.Get("/agent-network/threads", func(w http.ResponseWriter, r *http.Request) {
		v, e := h.Store.ListThreads(r.Context(), "", false)
		if e != nil {
			fail(w, e)
			return
		}
		httpx.JSON(w, 200, v)
	})
	r.Post("/agent-network/moderate", h.Moderate)
}
func (h *Handler) Challenge(w http.ResponseWriter, r *http.Request) {
	if !decode(w, r, &struct{}{}) {
		return
	}
	v, e := h.Store.Challenge(r.Context(), h.origin(r))
	if e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 201, v)
}
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	var in VerifyInput
	if !decode(w, r, &in) {
		return
	}
	if len(in.ClaimedProvider) > 100 || len(in.ClaimedModel) > 100 || len(in.ClientName) > 100 || len(in.ClientVersion) > 100 {
		httpx.Error(w, 400, "MESSAGE_SCHEMA_INVALID", "Metadata too long")
		return
	}
	v, e := h.Store.Verify(r.Context(), in, h.origin(r), requestID(r))
	if e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 201, v)
}
func (h *Handler) AgentAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			fail(w, ErrUnauthorized)
			return
		}
		p, e := h.Store.Authenticate(r.Context(), strings.TrimPrefix(auth, "Bearer "))
		if e != nil {
			fail(w, e)
			return
		}
		next.ServeHTTP(w, r.WithContext(withPrincipal(r.Context(), p)))
	})
}
func idempotency(r *http.Request) string {
	v := r.Header.Get("Idempotency-Key")
	if len(v) < 8 || len(v) > 128 {
		return ""
	}
	for _, c := range v {
		if c < 33 || c > 126 {
			return ""
		}
	}
	return v
}
func (h *Handler) CreateThread(w http.ResponseWriter, r *http.Request) {
	var in ThreadInput
	if !decode(w, r, &in) {
		return
	}
	key := idempotency(r)
	if key == "" || !validTitle(in.Title) || !validType(in.Type) {
		httpx.Error(w, 400, "MESSAGE_SCHEMA_INVALID", "Invalid thread or idempotency key")
		return
	}
	p := principal(r.Context())
	v, e := h.Store.CreateThread(r.Context(), p, in, key, requestID(r))
	if e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 201, v)
}
func validTitle(v string) bool {
	return strings.TrimSpace(v) != "" && utf8.ValidString(v) && utf8.RuneCountInString(v) <= 200
}
func validType(v string) bool {
	switch v {
	case "message", "request", "response", "discovery", "handoff", "report", "complaint", "warning", "announcement", "correction":
		return true
	}
	return false
}
func depth(v any, n int) bool {
	if n > 6 {
		return false
	}
	switch x := v.(type) {
	case map[string]any:
		for _, y := range x {
			if !depth(y, n+1) {
				return false
			}
		}
	case []any:
		for _, y := range x {
			if !depth(y, n+1) {
				return false
			}
		}
	}
	return true
}
func validateMessage(in *MessageInput) bool {
	if _, e := uuid.Parse(in.ThreadID); e != nil {
		return false
	}
	if !validType(in.Type) {
		return false
	}
	if in.Title != "" && !validTitle(in.Title) {
		return false
	}
	if in.ReplyTo != "" {
		if _, e := uuid.Parse(in.ReplyTo); e != nil {
			return false
		}
	}
	if in.Supersedes != "" {
		if _, e := uuid.Parse(in.Supersedes); e != nil {
			return false
		}
		if in.Type != "correction" {
			return false
		}
	}
	if len(in.Content.Body) == 0 || len(in.Content.Body) > 16*1024 {
		return false
	}
	switch in.Content.Format {
	case "text":
		var s string
		if json.Unmarshal(in.Content.Body, &s) != nil || !utf8.ValidString(s) || len(s) > 16*1024 {
			return false
		}
	case "json":
		var v any
		d := json.NewDecoder(bytes.NewReader(in.Content.Body))
		if d.Decode(&v) != nil || !depth(v, 0) {
			return false
		}
	default:
		return false
	}
	var err error
	in.Tags, err = NormalizeLabels(in.Tags)
	if err != nil {
		return false
	}
	in.Hooks, err = NormalizeLabels(in.Hooks)
	return err == nil
}
func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var in MessageInput
	if !decode(w, r, &in) {
		return
	}
	h.createMessage(w, r, &in)
}
func (h *Handler) Reply(w http.ResponseWriter, r *http.Request) {
	var in MessageInput
	if !decode(w, r, &in) {
		return
	}
	id := chi.URLParam(r, "id")
	parent, e := h.Store.GetMessage(r.Context(), id, true)
	if e != nil {
		fail(w, e)
		return
	}
	if in.ThreadID != "" && in.ThreadID != parent.ThreadID {
		httpx.Error(w, 400, "MESSAGE_SCHEMA_INVALID", "Thread mismatch")
		return
	}
	in.ThreadID = parent.ThreadID
	in.ReplyTo = id
	h.createMessage(w, r, &in)
}
func (h *Handler) createMessage(w http.ResponseWriter, r *http.Request, in *MessageInput) {
	key := idempotency(r)
	if key == "" || !validateMessage(in) {
		httpx.Error(w, 400, "MESSAGE_SCHEMA_INVALID", "Invalid message or idempotency key")
		return
	}
	v, e := h.Store.CreateMessage(r.Context(), principal(r.Context()), *in, key, requestID(r))
	if e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 201, v)
}
func validFilter(w http.ResponseWriter, v string) bool {
	if len(v) > 100 {
		httpx.Error(w, 400, "MESSAGE_SCHEMA_INVALID", "Invalid filter")
		return false
	}
	return true
}
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	tag := q.Get("tag")
	kind := "tag"
	if tag == "" {
		tag = q.Get("hook")
		kind = "hook"
	}
	thread := q.Get("thread")
	agent := q.Get("agent")
	typ := q.Get("type")
	query := q.Get("q")
	cursor := q.Get("cursor")
	for _, v := range []string{tag, thread, agent, typ, query, cursor} {
		if !validFilter(w, v) {
			return
		}
	}
	if cursor != "" {
		if _, e := uuid.Parse(cursor); e != nil {
			httpx.Error(w, 400, "MESSAGE_SCHEMA_INVALID", "Invalid cursor")
			return
		}
	}
	v, e := h.Store.ListMessages(r.Context(), thread, tag, kind, agent, typ, query, cursor, true)
	if e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 200, v)
}
func (h *Handler) GetMessage(w http.ResponseWriter, r *http.Request) {
	v, e := h.Store.GetMessage(r.Context(), chi.URLParam(r, "id"), true)
	if e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 200, v)
}
func (h *Handler) ListThreads(w http.ResponseWriter, r *http.Request) {
	cursor := r.URL.Query().Get("cursor")
	if cursor != "" {
		if _, e := uuid.Parse(cursor); e != nil {
			httpx.Error(w, 400, "MESSAGE_SCHEMA_INVALID", "Invalid cursor")
			return
		}
	}
	v, e := h.Store.ListThreads(r.Context(), cursor, true)
	if e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 200, v)
}
func (h *Handler) GetThread(w http.ResponseWriter, r *http.Request) {
	v, e := h.Store.GetThread(r.Context(), chi.URLParam(r, "id"), true)
	if e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 200, v)
}
func (h *Handler) GetAgent(w http.ResponseWriter, r *http.Request) {
	v, e := h.Store.GetPublicAgent(r.Context(), chi.URLParam(r, "id"))
	if e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 200, v)
}
func (h *Handler) Moderate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		TargetType string `json:"target_type"`
		TargetID   string `json:"target_id"`
		Action     string `json:"action"`
		Reason     string `json:"reason"`
	}
	if !decode(w, r, &in) {
		return
	}
	if _, e := uuid.Parse(in.TargetID); e != nil || len(in.Reason) == 0 || len(in.Reason) > 1000 {
		httpx.Error(w, 400, "MESSAGE_SCHEMA_INVALID", "Invalid moderation action")
		return
	}
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		fail(w, ErrUnauthorized)
		return
	}
	if e := h.Store.Moderate(r.Context(), claims.UserID, in.TargetType, in.TargetID, in.Action, in.Reason, requestID(r)); e != nil {
		fail(w, e)
		return
	}
	httpx.JSON(w, 200, map[string]bool{"ok": true})
}
func (h *Handler) Discovery(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, 200, map[string]any{"protocol": "lavoval-agent/1", "board": "/api/v1/agent-board", "challenge": "/api/v1/agents/challenge", "verify": "/api/v1/agents/verify", "capabilities": []string{"text", "json", "threads", "replies", "tags", "hooks"}, "security": map[string]any{"content_trust": "untrusted_external_content", "algorithm": "Ed25519"}})
}
