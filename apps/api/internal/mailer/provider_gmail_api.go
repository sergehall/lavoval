package mailer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

type GmailAPIProvider struct {
	cfg         config.Config
	client      *http.Client
	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time
}

type gmailSendRequest struct {
	Raw string `json:"raw"`
}

type gmailSendResponse struct {
	ID string `json:"id"`
}

func NewGmailAPIProvider(cfg config.Config) *GmailAPIProvider {
	return &GmailAPIProvider{
		cfg:         cfg,
		client:      &http.Client{},
		accessToken: strings.TrimSpace(cfg.GmailAPIAccessToken),
	}
}

func (p *GmailAPIProvider) Send(ctx context.Context, msg Message) (SendResult, error) {
	if strings.TrimSpace(p.cfg.SMTPFromEmail) == "" {
		return SendResult{}, permanentDeliveryError("gmail_api_from_email_missing", fmt.Errorf("smtp from email is required for gmail api delivery"))
	}

	token, err := p.accessTokenForSend(ctx)
	if err != nil {
		return SendResult{}, err
	}

	fromName := p.cfg.SMTPFromName
	if fromName == "" {
		fromName = p.cfg.AppName
	}

	rawMessage, err := buildMultipartMessage(fromName, p.cfg.SMTPFromEmail, msg.RecipientEmail, RenderedEmail{
		Subject:  msg.Subject,
		TextBody: msg.TextBody,
		HTMLBody: msg.HTMLBody,
	}, msg.Headers)
	if err != nil {
		return SendResult{}, permanentDeliveryError("mime_build_failed", fmt.Errorf("build gmail api message: %w", err))
	}

	requestBody, err := json.Marshal(gmailSendRequest{
		Raw: base64.RawURLEncoding.EncodeToString(rawMessage),
	})
	if err != nil {
		return SendResult{}, permanentDeliveryError("gmail_api_encode_failed", fmt.Errorf("marshal gmail api request: %w", err))
	}

	user := strings.TrimSpace(p.cfg.GmailAPIUser)
	if user == "" {
		user = "me"
	}

	baseURL := strings.TrimRight(strings.TrimSpace(p.cfg.GmailAPIBaseURL), "/")
	if baseURL == "" {
		baseURL = "https://gmail.googleapis.com/gmail/v1"
	}

	sendURL := fmt.Sprintf("%s/users/%s/messages/send", baseURL, user)
	result, err := p.sendWithToken(ctx, sendURL, requestBody, token)
	if err == nil {
		return result, nil
	}

	var deliveryErr *DeliveryError
	if !errors.As(err, &deliveryErr) || deliveryErr.Code != "gmail_api_401" || strings.TrimSpace(p.cfg.GmailAPIRefreshToken) == "" {
		return SendResult{}, err
	}

	refreshedToken, refreshErr := p.refreshAccessToken(ctx)
	if refreshErr != nil {
		return SendResult{}, refreshErr
	}

	return p.sendWithToken(ctx, sendURL, requestBody, refreshedToken)
}

func (p *GmailAPIProvider) sendWithToken(ctx context.Context, sendURL string, requestBody []byte, token string) (SendResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sendURL, bytes.NewReader(requestBody))
	if err != nil {
		return SendResult{}, permanentDeliveryError("gmail_api_request_failed", fmt.Errorf("build gmail api request: %w", err))
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return SendResult{}, temporaryDeliveryError("gmail_api_request_failed", fmt.Errorf("execute gmail api request: %w", err))
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if readErr != nil {
		return SendResult{}, temporaryDeliveryError("gmail_api_read_failed", fmt.Errorf("read gmail api response: %w", readErr))
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return SendResult{}, temporaryDeliveryError("gmail_api_401", fmt.Errorf("gmail api unauthorized: %s", strings.TrimSpace(string(body))))
	case resp.StatusCode == http.StatusTooManyRequests:
		return SendResult{}, temporaryDeliveryError("gmail_api_429", fmt.Errorf("gmail api throttled: %s", strings.TrimSpace(string(body))))
	case resp.StatusCode >= 500:
		return SendResult{}, temporaryDeliveryError(fmt.Sprintf("gmail_api_%d", resp.StatusCode), fmt.Errorf("gmail api server error: %s", strings.TrimSpace(string(body))))
	case resp.StatusCode >= 400:
		return SendResult{}, permanentDeliveryError(fmt.Sprintf("gmail_api_%d", resp.StatusCode), fmt.Errorf("gmail api client error: %s", strings.TrimSpace(string(body))))
	}

	var sendResp gmailSendResponse
	if err := json.Unmarshal(body, &sendResp); err != nil {
		return SendResult{}, temporaryDeliveryError("gmail_api_decode_failed", fmt.Errorf("decode gmail api response: %w", err))
	}

	return SendResult{
		Provider:          "gmail_api",
		ProviderMessageID: sendResp.ID,
	}, nil
}

func (p *GmailAPIProvider) accessTokenForSend(ctx context.Context) (string, error) {
	p.mu.Lock()
	token := p.accessToken
	expiresAt := p.expiresAt
	p.mu.Unlock()

	if token != "" && (expiresAt.IsZero() || time.Until(expiresAt) > 30*time.Second) {
		return token, nil
	}

	if strings.TrimSpace(p.cfg.GmailAPIRefreshToken) == "" {
		if token != "" {
			return token, nil
		}
		return "", permanentDeliveryError("gmail_api_not_configured", fmt.Errorf("gmail api access token is not configured"))
	}

	return p.refreshAccessToken(ctx)
}

func (p *GmailAPIProvider) refreshAccessToken(ctx context.Context) (string, error) {
	if strings.TrimSpace(p.cfg.GmailAPIRefreshToken) == "" ||
		strings.TrimSpace(p.cfg.GmailAPIClientID) == "" ||
		strings.TrimSpace(p.cfg.GmailAPIClientSecret) == "" {
		return "", permanentDeliveryError("gmail_api_refresh_not_configured", fmt.Errorf("gmail api refresh token flow is not fully configured"))
	}

	form := url.Values{}
	form.Set("client_id", p.cfg.GmailAPIClientID)
	form.Set("client_secret", p.cfg.GmailAPIClientSecret)
	form.Set("refresh_token", p.cfg.GmailAPIRefreshToken)
	form.Set("grant_type", "refresh_token")

	tokenURL := strings.TrimSpace(p.cfg.GmailAPITokenURL)
	if tokenURL == "" {
		tokenURL = "https://oauth2.googleapis.com/token"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", permanentDeliveryError("gmail_api_refresh_request_failed", fmt.Errorf("build gmail refresh request: %w", err))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", temporaryDeliveryError("gmail_api_refresh_request_failed", fmt.Errorf("execute gmail refresh request: %w", err))
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if readErr != nil {
		return "", temporaryDeliveryError("gmail_api_refresh_read_failed", fmt.Errorf("read gmail refresh response: %w", readErr))
	}

	switch {
	case resp.StatusCode >= 500:
		return "", temporaryDeliveryError(fmt.Sprintf("gmail_api_refresh_%d", resp.StatusCode), fmt.Errorf("gmail token endpoint server error: %s", strings.TrimSpace(string(body))))
	case resp.StatusCode >= 400:
		return "", permanentDeliveryError(fmt.Sprintf("gmail_api_refresh_%d", resp.StatusCode), fmt.Errorf("gmail token endpoint client error: %s", strings.TrimSpace(string(body))))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", temporaryDeliveryError("gmail_api_refresh_decode_failed", fmt.Errorf("decode gmail refresh response: %w", err))
	}
	if strings.TrimSpace(tokenResp.AccessToken) == "" {
		return "", temporaryDeliveryError("gmail_api_refresh_missing_token", fmt.Errorf("gmail refresh returned no access token"))
	}

	p.mu.Lock()
	p.accessToken = tokenResp.AccessToken
	if tokenResp.ExpiresIn > 0 {
		p.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	} else {
		p.expiresAt = time.Time{}
	}
	p.mu.Unlock()

	return tokenResp.AccessToken, nil
}

var _ Provider = (*GmailAPIProvider)(nil)
