package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

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
	{
		Kind:        EmailTemplateVerification,
		DisplayName: "Email verification",
		Description: "Confirm a new account email address before first sign-in.",
		Implemented: true,
	},
	{
		Kind:        EmailTemplatePasswordReset,
		DisplayName: "Password reset",
		Description: "Restore account access after a forgotten password request.",
		Implemented: false,
	},
	{
		Kind:        EmailTemplateWelcome,
		DisplayName: "Welcome",
		Description: "Introduce the workspace after successful account activation.",
		Implemented: false,
	},
	{
		Kind:        EmailTemplateSecurityNotice,
		DisplayName: "Security notice",
		Description: "Alert members about sensitive account events and confirmations.",
		Implemented: false,
	},
	{
		Kind:        EmailTemplateRunCompleted,
		DisplayName: "Run completed",
		Description: "Summarize a finished skill run and link back to results.",
		Implemented: false,
	},
}

type RenderedEmail struct {
	Subject  string
	TextBody string
	HTMLBody string
}

type verificationEmailTemplateData struct {
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

var brandedEmailHTMLTemplate = template.Must(template.New("branded-email").Parse(`<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width,initial-scale=1" />
  </head>
  <body style="margin:0; padding:0; background:{{.Brand.PageBackground}};">
    <span style="display:none!important; visibility:hidden; opacity:0; color:transparent; height:0; width:0; overflow:hidden;">
      {{.Preheader}}
    </span>
    <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="100%">
      <tr>
        <td align="center" style="padding:28px 12px;">
          <table role="presentation" cellpadding="0" cellspacing="0" border="0" width="600" style="width:600px; max-width:600px;">
            <tr>
              <td style="padding:0 0 14px 0;">
                <table role="presentation" cellpadding="0" cellspacing="0" border="0">
                  <tr>
                    <td style="font-size:28px; font-weight:700; letter-spacing:-0.03em; line-height:1.05; color:{{.Brand.Title}};">
                      {{.AppName}}
                    </td>
                  </tr>
                  <tr>
                    <td style="padding-top:6px; font-size:14px; line-height:20px; color:{{.Brand.Muted}};">
                      Human skill exchange for the AI era
                    </td>
                  </tr>
                </table>
              </td>
            </tr>
            <tr>
              <td style="background:{{.Brand.CardBackground}}; border:1px solid {{.Brand.CardBorder}}; border-radius:22px; padding:24px; box-shadow:0 14px 36px rgba(45,30,16,0.10);">
                <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                  <tr>
                    <td style="padding:0 0 18px 0;">
                      <img
                        src="{{.HeroImageURL}}"
                        alt="{{.AppName}}"
                        width="552"
                        style="display:block; width:100%; max-width:552px; height:auto; border-radius:18px; border:1px solid {{.Brand.CardBorder}};"
                      />
                    </td>
                  </tr>
                  <tr>
                    <td style="padding:0 0 16px 0;">
                      <span style="display:inline-block; padding:7px 12px; border-radius:999px; background:{{.Brand.AccentSoft}}; border:1px solid {{.Brand.AccentBorder}}; color:{{.Brand.AccentDark}}; font-size:12px; font-weight:700; letter-spacing:0.06em; text-transform:uppercase;">
                        {{.Badge}}
                      </span>
                    </td>
                  </tr>
                  <tr>
                    <td style="color:{{.Brand.Title}}; font-size:30px; line-height:36px; font-weight:700; padding:0 0 12px 0;">
                      {{.Heading}}
                    </td>
                  </tr>
                  <tr>
                    <td style="color:{{.Brand.Text}}; font-size:16px; line-height:24px; font-weight:700; padding:0 0 10px 0;">
                      {{.Greeting}}
                    </td>
                  </tr>
                  <tr>
                    <td style="color:{{.Brand.Text}}; font-size:16px; line-height:24px; padding:0 0 10px 0;">
                      {{.IntroLine}}
                    </td>
                  </tr>
                  <tr>
                    <td style="color:{{.Brand.Text}}; font-size:15px; line-height:24px; padding:0 0 18px 0;">
                      {{.DetailLine}}
                    </td>
                  </tr>
                  <tr>
                    <td style="padding:0 0 18px 0;">
                      <a
                        href="{{.ActionURL}}"
                        target="_blank"
                        rel="noopener noreferrer"
                        style="display:inline-block; padding:14px 18px; background:{{.Brand.AccentPrimary}}; border-radius:12px; color:#ffffff; text-decoration:none; font-weight:700; font-size:15px;"
                      >
                        {{.ActionLabel}}
                      </a>
                    </td>
                  </tr>
                  <tr>
                    <td style="background:{{.Brand.AccentSoft}}; border:1px solid {{.Brand.AccentBorder}}; border-radius:16px; padding:16px;">
                      <table role="presentation" width="100%" cellpadding="0" cellspacing="0" border="0">
                        <tr>
                          <td style="color:{{.Brand.Title}}; font-size:15px; line-height:21px; font-weight:700; padding:0 0 10px 0;">
                            {{.InfoTitle}}
                          </td>
                        </tr>
                        {{range .InfoLines}}
                        <tr>
                          <td style="color:{{$.Brand.Text}}; font-size:14px; line-height:22px; padding:0 0 8px 0;">
                            • {{.}}
                          </td>
                        </tr>
                        {{end}}
                      </table>
                    </td>
                  </tr>
                  <tr>
                    <td style="padding:18px 0 0 0; color:{{.Brand.Text}}; font-size:14px; line-height:22px;">
                      {{.ClosingLine}}
                    </td>
                  </tr>
                  <tr>
                    <td style="padding:14px 0 0 0; color:{{.Brand.Muted}}; font-size:13px; line-height:20px; word-break:break-word;">
                      {{.ActionHint}}
                    </td>
                  </tr>
                  <tr>
                    <td style="padding:16px 0 0 0; color:{{.Brand.Muted}}; font-size:13px; line-height:19px;">
                      {{.FooterNote}}
                    </td>
                  </tr>
                  <tr>
                    <td style="padding:16px 0 0 0;">
                      <div style="height:1px; background:{{.Brand.CardBorder}}; width:100%;"></div>
                    </td>
                  </tr>
                  <tr>
                    <td style="padding:14px 0 0 0; font-size:13px; line-height:18px;">
                      <a href="{{.WebsiteURL}}" target="_blank" rel="noopener noreferrer" style="color:{{.Brand.AccentPrimary}}; text-decoration:none; font-weight:600;">
                        <span style="opacity:0.7;">{{.AppName}}:</span> {{.WebsiteLabel}}
                      </a>
                    </td>
                  </tr>
                </table>
              </td>
            </tr>
          </table>
        </td>
      </tr>
    </table>
  </body>
</html>`))

func renderVerificationEmail(email VerificationEmail, appURL string) (RenderedEmail, error) {
	websiteURL, websiteLabel := buildWebsiteLink(appURL)
	heroImageURL := buildBrandImageURL(appURL)
	display := displayName(email.ToName)
	subject := fmt.Sprintf("%s: confirm your email", email.ProductName)

	data := verificationEmailTemplateData{
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

	var htmlBody bytes.Buffer
	if err := brandedEmailHTMLTemplate.Execute(&htmlBody, data); err != nil {
		return RenderedEmail{}, fmt.Errorf("render verification html: %w", err)
	}

	textBody := strings.Join([]string{
		fmt.Sprintf("Email confirmation | %s", data.Heading),
		"",
		data.Greeting,
		data.IntroLine,
		data.DetailLine,
		"",
		fmt.Sprintf("%s: %s", data.ActionLabel, data.ActionURL),
		"",
		data.InfoTitle,
		"- " + strings.Join(data.InfoLines, "\n- "),
		"",
		data.ClosingLine,
		data.ActionHint,
		"",
		data.FooterNote,
		"",
		fmt.Sprintf("%s: %s", data.AppName, data.WebsiteURL),
	}, "\n")

	return RenderedEmail{
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody.String(),
	}, nil
}

type PasswordResetEmail struct {
	ToEmail     string
	ToName      string
	ResetURL    string
	ProductName string
}

type PasswordChangedEmail struct {
	ToEmail     string
	ToName      string
	SignInURL   string
	ProductName string
}

func renderPasswordResetEmail(email PasswordResetEmail, appURL string) (RenderedEmail, error) {
	websiteURL, websiteLabel := buildWebsiteLink(appURL)
	heroImageURL := buildBrandImageURL(appURL)
	display := displayName(email.ToName)
	subject := fmt.Sprintf("%s: reset your password", email.ProductName)

	data := verificationEmailTemplateData{
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

	var htmlBody bytes.Buffer
	if err := brandedEmailHTMLTemplate.Execute(&htmlBody, data); err != nil {
		return RenderedEmail{}, fmt.Errorf("render password reset html: %w", err)
	}

	textBody := strings.Join([]string{
		fmt.Sprintf("Password recovery | %s", data.Heading),
		"",
		data.Greeting,
		data.IntroLine,
		data.DetailLine,
		"",
		fmt.Sprintf("%s: %s", data.ActionLabel, data.ActionURL),
		"",
		data.InfoTitle,
		"- " + strings.Join(data.InfoLines, "\n- "),
		"",
		data.ClosingLine,
		data.ActionHint,
		"",
		data.FooterNote,
		"",
		fmt.Sprintf("%s: %s", data.AppName, data.WebsiteURL),
	}, "\n")

	return RenderedEmail{
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody.String(),
	}, nil
}

func renderPasswordChangedEmail(email PasswordChangedEmail, appURL string) (RenderedEmail, error) {
	websiteURL, websiteLabel := buildWebsiteLink(appURL)
	heroImageURL := buildBrandImageURL(appURL)
	display := displayName(email.ToName)
	subject := fmt.Sprintf("%s: your password was changed", email.ProductName)

	data := verificationEmailTemplateData{
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

	var htmlBody bytes.Buffer
	if err := brandedEmailHTMLTemplate.Execute(&htmlBody, data); err != nil {
		return RenderedEmail{}, fmt.Errorf("render password changed html: %w", err)
	}

	textBody := strings.Join([]string{
		fmt.Sprintf("Security notice | %s", data.Heading),
		"",
		data.Greeting,
		data.IntroLine,
		data.DetailLine,
		"",
		fmt.Sprintf("%s: %s", data.ActionLabel, data.ActionURL),
		"",
		data.InfoTitle,
		"- " + strings.Join(data.InfoLines, "\n- "),
		"",
		data.ClosingLine,
		data.ActionHint,
		"",
		data.FooterNote,
		"",
		fmt.Sprintf("%s: %s", data.AppName, data.WebsiteURL),
	}, "\n")

	return RenderedEmail{
		Subject:  subject,
		TextBody: textBody,
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
		return "http://localhost:3000/og-image.png"
	}

	imageURL := parsed.ResolveReference(&url.URL{Path: "/og-image.png"})
	return imageURL.String()
}
