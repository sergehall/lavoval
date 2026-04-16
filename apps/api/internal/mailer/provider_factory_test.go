package mailer

import "testing"

func TestNormalizeMailProvider(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "", want: "smtp"},
		{in: "smtp", want: "smtp"},
		{in: "gmail_api", want: "gmail_api"},
		{in: "gmail-api", want: "gmail_api"},
		{in: "noop", want: "noop"},
	}

	for _, tt := range tests {
		if got := normalizeMailProvider(tt.in); got != tt.want {
			t.Fatalf("normalizeMailProvider(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
