package repository

import "testing"

func TestNormalizeJSONMap(t *testing.T) {
	t.Run("returns empty map for nil", func(t *testing.T) {
		got := normalizeJSONMap(nil)
		if got == nil {
			t.Fatal("expected empty map, got nil")
		}
		if len(got) != 0 {
			t.Fatalf("expected empty map, got %v", got)
		}
	})

	t.Run("preserves existing map", func(t *testing.T) {
		input := map[string]any{"text": "hello"}
		got := normalizeJSONMap(input)
		if got["text"] != "hello" {
			t.Fatalf("expected text to be preserved, got %v", got["text"])
		}
	})
}
