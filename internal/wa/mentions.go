package wa

import (
	"fmt"
	"strings"
	"unicode"

	"go.mau.fi/whatsmeow/types"
)

// NormalizeMention turns a bare phone number or a JID into a non-device JID.
// Bare numbers become <digits>@s.whatsapp.net. @lid JIDs are accepted as-is
// (device suffixes are stripped). The message text is never modified.
func NormalizeMention(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "@")
	if raw == "" {
		return "", fmt.Errorf("empty mention")
	}

	if strings.Contains(raw, "@") {
		jid, err := types.ParseJID(raw)
		if err != nil {
			return "", fmt.Errorf("%q is not a valid mention id: %w", raw, err)
		}
		if jid.Server != types.DefaultUserServer && jid.Server != types.HiddenUserServer {
			return "", fmt.Errorf("%q is not a person JID (want @s.whatsapp.net or @lid)", raw)
		}
		return jid.ToNonAD().String(), nil
	}

	var digits strings.Builder
	for _, r := range raw {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}
	if digits.Len() < 8 {
		return "", fmt.Errorf("%q is not a valid phone number or JID", raw)
	}
	return digits.String() + "@" + types.DefaultUserServer, nil
}

// NormalizeMentions normalizes a list of mentions, dropping blanks and
// de-duplicating while preserving order.
func NormalizeMentions(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, r := range raw {
		if strings.TrimSpace(r) == "" {
			continue
		}
		jid, err := NormalizeMention(r)
		if err != nil {
			return nil, err
		}
		if seen[jid] {
			continue
		}
		seen[jid] = true
		out = append(out, jid)
	}
	return out, nil
}

// MentionToken is the "@<user>" fragment that must appear in the message text
// for WhatsApp to highlight and notify the mentioned person.
func MentionToken(jid string) string {
	parsed, err := types.ParseJID(jid)
	if err != nil || parsed.User == "" {
		return ""
	}
	return "@" + parsed.User
}

// ValidateMentionText checks that text already contains "@<user>" for every
// mentioned JID. It never rewrites the text.
func ValidateMentionText(text string, mentions []string) error {
	for _, jid := range mentions {
		token := MentionToken(jid)
		if token == "" {
			return fmt.Errorf("cannot derive an @token for mention %s", jid)
		}
		if !strings.Contains(text, token) {
			return fmt.Errorf("message text must contain %q to mention %s; the text is not rewritten automatically", token, jid)
		}
	}
	return nil
}
