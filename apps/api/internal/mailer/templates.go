package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/url"
	"strings"
	texttmpl "text/template"
)

//go:embed templates/layouts/*.html templates/emails/*.html templates/emails/*.txt
var emailTemplatesFS embed.FS

type EmailTemplateKind string

const (
	EmailTemplateVerification   EmailTemplateKind = "verification"
	EmailTemplatePasswordReset  EmailTemplateKind = "password_reset"
	EmailTemplateWelcome        EmailTemplateKind = "welcome"
	EmailTemplateSecurityNotice EmailTemplateKind = "security_notice"
	EmailTemplateRunCompleted   EmailTemplateKind = "run_completed"
)

type EmailTemplateCatalogEntry struct {
	Kind        EmailTemplateKind
	DisplayName string
	Description string
	Implemented bool
}

var EmailTemplateCatalog = []EmailTemplateCatalogEntry{
	{Kind: EmailTemplateVerification, DisplayName: "Email verification", Description: "Confirm a new account email address before first sign-in.", Implemented: true},
	{Kind: EmailTemplatePasswordReset, DisplayName: "Password reset", Description: "Restore account access after a forgotten password request.", Implemented: true},
	{Kind: EmailTemplateWelcome, DisplayName: "Welcome", Description: "Introduce the workspace after successful account activation.", Implemented: false},
	{Kind: EmailTemplateSecurityNotice, DisplayName: "Security notice", Description: "Alert members about sensitive account events and confirmations.", Implemented: true},
	{Kind: EmailTemplateRunCompleted, DisplayName: "Run completed", Description: "Summarize a finished skill run and link back to results.", Implemented: false},
}

type RenderedEmail struct {
	Subject  string
	TextBody string
	HTMLBody string
}

type emailTemplateData struct {
	Preheader    string
	Badge        string
	Heading      string
	Greeting     string
	IntroLine    string
	DetailLine   string
	ActionLabel  string
	ActionURL    string
	ActionHint   string
	InfoTitle    string
	InfoLines    []string
	ClosingLine  string
	FooterNote   string
	AppName      string
	WebsiteURL   string
	WebsiteLabel string
	HeroImageURL string
	Brand        emailBrandPalette
}

type emailBrandPalette struct {
	PageBackground string
	CardBackground string
	CardBorder     string
	Title          string
	Text           string
	Muted          string
	AccentSoft     string
	AccentBorder   string
	AccentPrimary  string
	AccentDark     string
}

var lavovalEmailPalette = emailBrandPalette{
	PageBackground: "#f4efe7",
	CardBackground: "#fffdf9",
	CardBorder:     "#e4d8ca",
	Title:          "#17202a",
	Text:           "#374151",
	Muted:          "#536471",
	AccentSoft:     "#f7ebe4",
	AccentBorder:   "#dfc1b1",
	AccentPrimary:  "#b44f23",
	AccentDark:     "#893a18",
}

func renderVerificationEmail(email VerificationEmail, appURL string) (RenderedEmail, error) {
	websiteURL, websiteLabel := buildWebsiteLink(appURL)
	heroImageURL := buildBrandImageURL(appURL)
	display := displayName(email.ToName)

	data := emailTemplateData{
		Preheader:   fmt.Sprintf("Confirm your %s email and activate your account.", email.ProductName),
		Badge:       "Email confirmation",
		Heading:     fmt.Sprintf("Confirm your email to activate %s", email.ProductName),
		Greeting:    fmt.Sprintf("Hi %s,", display),
		IntroLine:   fmt.Sprintf("Thanks for creating a %s account. Confirm this email address to unlock sign in and start using your workspace.", email.ProductName),
		DetailLine:  "This verification step makes sure important security notices and marketplace updates reach the right inbox.",
		ActionLabel: "Confirm email",
		ActionURL:   email.VerifyURL,
		ActionHint:  fmt.Sprintf("Secure confirmation link: %s", email.VerifyURL),
		InfoTitle:   "What happens next",
		InfoLines: []string{
			fmt.Sprintf("Email on file: %s", email.ToEmail),
			"After confirmation you can sign in and start building your Lavoval identity, offers, and runs history.",
			"If you did not create this account, you can safely ignore this email.",
		},
		ClosingLine:  "If the button does not open on your device, copy the secure link below into your browser.",
		FooterNote:   fmt.Sprintf("%s will never ask you to confirm an account by replying with a password or code.", email.ProductName),
		AppName:      email.ProductName,
		WebsiteURL:   websiteURL,
		WebsiteLabel: websiteLabel,
		HeroImageURL: heroImageURL,
		Brand:        lavovalEmailPalette,
	}

	return renderEmailTemplate("verification", fmt.Sprintf("%s: confirm your email", email.ProductName), data)
}

func renderPasswordResetEmail(email PasswordResetEmail, appURL string) (RenderedEmail, error) {
	websiteURL, websiteLabel := buildWebsiteLink(appURL)
	heroImageURL := buildBrandImageURL(appURL)
	display := displayName(email.ToName)

	data := emailTemplateData{
		Preheader:   fmt.Sprintf("Reset your %s password with a secure one-time link.", email.ProductName),
		Badge:       "Password recovery",
		Heading:     fmt.Sprintf("Reset your %s password", email.ProductName),
		Greeting:    fmt.Sprintf("Hi %s,", display),
		IntroLine:   "Use the secure action below to open the password reset flow and choose a new password for this inbox.",
		DetailLine:  "This link expires soon and can only be used once. If you did not request a password reset, you can ignore this email and your current password will keep working.",
		ActionLabel: "Reset password",
		ActionURL:   email.ResetURL,
		ActionHint:  fmt.Sprintf("Secure password reset link: %s", email.ResetURL),
		InfoTitle:   "What this link does",
		InfoLines: []string{
			fmt.Sprintf("Account email: %s", email.ToEmail),
			"This link opens the secure password reset flow for your account.",
			"After you set a new password, use it the next time you sign in.",
		},
		ClosingLine:  "If the button does not open on your device, copy the secure link below into your browser.",
		FooterNote:   fmt.Sprintf("%s support will never ask you to share your password or recovery link over email.", email.ProductName),
		AppName:      email.ProductName,
		WebsiteURL:   websiteURL,
		WebsiteLabel: websiteLabel,
		HeroImageURL: heroImageURL,
		Brand:        lavovalEmailPalette,
	}

	return renderEmailTemplate("password_reset", fmt.Sprintf("%s: reset your password", email.ProductName), data)
}

func renderPasswordChangedEmail(email PasswordChangedEmail, appURL string) (RenderedEmail, error) {
	websiteURL, websiteLabel := buildWebsiteLink(appURL)
	heroImageURL := buildBrandImageURL(appURL)
	display := displayName(email.ToName)

	data := emailTemplateData{
		Preheader:   fmt.Sprintf("Your %s password was updated.", email.ProductName),
		Badge:       "Security notice",
		Heading:     fmt.Sprintf("Your %s password was changed", email.ProductName),
		Greeting:    fmt.Sprintf("Hi %s,", display),
		IntroLine:   "This is a confirmation that your Lavoval password has just been updated.",
		DetailLine:  "If you made this change, you do not need to do anything else. If you did not make it, reset your password again right away and review access to your account.",
		ActionLabel: "Go to sign in",
		ActionURL:   email.SignInURL,
		ActionHint:  fmt.Sprintf("Secure sign-in link: %s", email.SignInURL),
		InfoTitle:   "Recommended next steps",
		InfoLines: []string{
			fmt.Sprintf("Account email: %s", email.ToEmail),
			"Use your new password the next time you sign in.",
			"If this change looks suspicious, request another password reset immediately.",
		},
		ClosingLine:  "If the button does not open on your device, copy the secure link below into your browser.",
		FooterNote:   fmt.Sprintf("%s support will never ask you to reply with your password or reset link.", email.ProductName),
		AppName:      email.ProductName,
		WebsiteURL:   websiteURL,
		WebsiteLabel: websiteLabel,
		HeroImageURL: heroImageURL,
		Brand:        lavovalEmailPalette,
	}

	return renderEmailTemplate("password_changed", fmt.Sprintf("%s: your password was changed", email.ProductName), data)
}

func renderEmailTemplate(name string, subject string, data emailTemplateData) (RenderedEmail, error) {
	htmlTpl, err := template.ParseFS(
		emailTemplatesFS,
		"templates/layouts/base.html",
		fmt.Sprintf("templates/emails/%s.html", name),
	)
	if err != nil {
		return RenderedEmail{}, fmt.Errorf("parse html template %s: %w", name, err)
	}

	textTpl, err := texttmpl.ParseFS(emailTemplatesFS, fmt.Sprintf("templates/emails/%s.txt", name))
	if err != nil {
		return RenderedEmail{}, fmt.Errorf("parse text template %s: %w", name, err)
	}

	var htmlBody bytes.Buffer
	if err := htmlTpl.ExecuteTemplate(&htmlBody, "base", data); err != nil {
		return RenderedEmail{}, fmt.Errorf("render html template %s: %w", name, err)
	}

	var textBody bytes.Buffer
	if err := textTpl.Execute(&textBody, data); err != nil {
		return RenderedEmail{}, fmt.Errorf("render text template %s: %w", name, err)
	}

	return RenderedEmail{
		Subject:  subject,
		TextBody: textBody.String(),
		HTMLBody: htmlBody.String(),
	}, nil
}

func buildWebsiteLink(appURL string) (string, string) {
	if appURL == "" {
		return "http://localhost:3000", "localhost:3000"
	}

	parsed, err := url.Parse(appURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "http://localhost:3000", "localhost:3000"
	}

	return parsed.Scheme + "://" + parsed.Host, parsed.Host
}

func buildBrandImageURL(appURL string) string {
	base, _ := buildWebsiteLink(appURL)
	parsed, err := url.Parse(base)
	if err != nil {
		return "http://localhost:3000/email-brand-120x40.png"
	}

	imageURL := parsed.ResolveReference(&url.URL{Path: "/email-brand-120x40.png"})
	return imageURL.String()
}

func displayName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "there"
	}
	return trimmed
}
