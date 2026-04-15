package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sergehall/lavoval/apps/api/internal/config"
)

const totpDigits = 6

type totpManager struct {
	issuer string
	period time.Duration
	key    []byte
}

func newTOTPManager(cfg config.Config) totpManager {
	secretSource := cfg.MFASecretKey
	if secretSource == "" {
		secretSource = cfg.JWTSecret + ":mfa"
	}

	sum := sha256.Sum256([]byte(secretSource))

	return totpManager{
		issuer: cfg.MFATOTPIssuer,
		period: cfg.MFATOTPPeriod,
		key:    sum[:],
	}
}

func (m totpManager) GenerateSecret() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate totp secret: %w", err)
	}

	return strings.TrimRight(base32.StdEncoding.EncodeToString(buf), "="), nil
}

func (m totpManager) ProvisioningURI(accountName string, secret string) string {
	label := fmt.Sprintf("%s:%s", m.issuer, accountName)
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", m.issuer)
	values.Set("algorithm", "SHA1")
	values.Set("digits", strconv.Itoa(totpDigits))
	values.Set("period", strconv.Itoa(int(m.period.Seconds())))

	return fmt.Sprintf("otpauth://totp/%s?%s", url.PathEscape(label), values.Encode())
}

func (m totpManager) EncryptSecret(secret string) (string, error) {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", fmt.Errorf("init mfa cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("init mfa gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate mfa nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func (m totpManager) DecryptSecret(ciphertext string) (string, error) {
	raw, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode mfa secret: %w", err)
	}

	block, err := aes.NewCipher(m.key)
	if err != nil {
		return "", fmt.Errorf("init mfa cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("init mfa gcm: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("decode mfa secret: malformed payload")
	}

	nonce, data := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt mfa secret: %w", err)
	}

	return string(plaintext), nil
}

func (m totpManager) VerifyCode(secret string, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return false
	}

	period := int64(m.period.Seconds())
	if period <= 0 {
		period = 30
	}

	for offset := int64(-1); offset <= 1; offset++ {
		counter := (now.UTC().Unix() / period) + offset
		expected, err := generateTOTPCode(secret, counter)
		if err == nil && hmac.Equal([]byte(expected), []byte(code)) {
			return true
		}
	}

	return false
}

func generateTOTPCode(secret string, counter int64) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(secret))
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(normalized)
	if err != nil {
		return "", fmt.Errorf("decode totp secret: %w", err)
	}

	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, uint64(counter))

	mac := hmac.New(sha1.New, decoded)
	if _, err := mac.Write(msg); err != nil {
		return "", fmt.Errorf("hash totp code: %w", err)
	}

	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)

	code := value % 1000000
	return fmt.Sprintf("%06d", code), nil
}
