package agentnetwork

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// readableText allows prose and source-code examples, but not terminal control
// sequences or NUL bytes. This is not a command or prompt-injection detector:
// accepted content must still never be dispatched to a shell, tool, or template.
func readableText(s string) bool {
	if !utf8.ValidString(s) {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return false
		}
	}
	return true
}

// validBoardJSON checks the wire representation before decoding into structs
// or JSONB can discard duplicate keys. Request envelope fields are lowercase
// and case-sensitive; arbitrary JSON content keeps its own case-sensitive keys.
func validBoardJSON(raw []byte, maxDepth int, envelope bool) bool {
	if !utf8.Valid(raw) || !json.Valid(raw) {
		return false
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	var value func(int, bool) bool
	value = func(level int, schema bool) bool {
		if level > maxDepth {
			return false
		}
		token, err := d.Token()
		if err != nil {
			return false
		}
		switch t := token.(type) {
		case string:
			return readableText(t)
		case json.Delim:
			switch t {
			case '{':
				seen := make(map[string]bool)
				for d.More() {
					keyToken, err := d.Token()
					key, ok := keyToken.(string)
					if err != nil || !ok || !readableText(key) || seen[key] {
						return false
					}
					if schema && key != strings.ToLower(key) {
						return false
					}
					seen[key] = true
					if !value(level+1, envelope && level == 0 && key == "content") {
						return false
					}
				}
				end, err := d.Token()
				return err == nil && end == json.Delim('}')
			case '[':
				for d.More() {
					if !value(level+1, false) {
						return false
					}
				}
				end, err := d.Token()
				return err == nil && end == json.Delim(']')
			default:
				return false
			}
		}
		return true
	}
	if !value(0, envelope) {
		return false
	}
	_, err := d.Token()
	return err == io.EOF
}
