package wa

import (
	"testing"

	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"

	"github.com/thenavidm/whatsapp-mcp/internal/store"
)

func TestBuildTextMessagePlainAndQuoted(t *testing.T) {
	chat := types.JID{User: "111000111000111", Server: types.HiddenUserServer}

	plain := buildTextMessage("hello", nil, "111000111000111@lid", chat, nil)
	if plain.GetConversation() != "hello" || plain.GetExtendedTextMessage() != nil {
		t.Fatalf("unexpected plain message: %+v", plain)
	}

	quoted := &store.Message{
		ID:        "ORIG123",
		ChatJID:   "111000111000111@lid",
		SenderJID: "",
		Content:   "Test message for reply",
		IsFromMe:  true,
	}
	reply := buildTextMessage("Test reply via bot", quoted, "111000111000111@lid", chat, nil)
	ext := reply.GetExtendedTextMessage()
	if ext == nil {
		t.Fatal("expected ExtendedTextMessage for quoted reply")
	}
	if ext.GetText() != "Test reply via bot" {
		t.Fatalf("want reply text, got %q", ext.GetText())
	}
	ci := ext.GetContextInfo()
	if ci == nil {
		t.Fatal("expected ContextInfo on quoted reply")
	}
	if ci.GetStanzaID() != "ORIG123" {
		t.Fatalf("want StanzaID ORIG123, got %q", ci.GetStanzaID())
	}
	if ci.GetParticipant() != "111000111000111@lid" {
		t.Fatalf("want Participant 111000111000111@lid, got %q", ci.GetParticipant())
	}
	if ci.GetQuotedMessage().GetConversation() != "Test message for reply" {
		t.Fatalf("want QuotedMessage conversation, got %q", ci.GetQuotedMessage().GetConversation())
	}
}

func TestExtractQuotedID(t *testing.T) {
	wrapped := &waE2E.Message{
		DeviceSentMessage: &waE2E.DeviceSentMessage{
			Message: &waE2E.Message{
				ExtendedTextMessage: &waE2E.ExtendedTextMessage{
					Text: proto.String("reply body"),
					ContextInfo: &waE2E.ContextInfo{
						StanzaID: proto.String("STANZA99"),
					},
				},
			},
		},
	}
	if got := extractQuotedID(wrapped); got != "STANZA99" {
		t.Fatalf("want STANZA99, got %q", got)
	}

	plain := &waE2E.Message{Conversation: proto.String("no quote")}
	if got := extractQuotedID(plain); got != "" {
		t.Fatalf("want empty quoted_id, got %q", got)
	}
}

func TestNormalizeAndValidateMentions(t *testing.T) {
	got, err := NormalizeMention("15555550100")
	if err != nil || got != "15555550100@s.whatsapp.net" {
		t.Fatalf("bare number: got %q, %v", got, err)
	}
	got, err = NormalizeMention("+1 555 555 0100")
	if err != nil || got != "15555550100@s.whatsapp.net" {
		t.Fatalf("spaced number: got %q, %v", got, err)
	}
	got, err = NormalizeMention("111000111000111@lid")
	if err != nil || got != "111000111000111@lid" {
		t.Fatalf("lid: got %q, %v", got, err)
	}
	got, err = NormalizeMention("15555550100:85@s.whatsapp.net")
	if err != nil || got != "15555550100@s.whatsapp.net" {
		t.Fatalf("device jid: got %q, %v", got, err)
	}
	if _, err := NormalizeMention("120363000000000001@g.us"); err == nil {
		t.Fatal("group JID should be rejected")
	}

	mentions, err := NormalizeMentions([]string{"15555550100", "15555550100@s.whatsapp.net", "111000111000111@lid"})
	if err != nil {
		t.Fatal(err)
	}
	if len(mentions) != 2 {
		t.Fatalf("want 2 unique mentions, got %v", mentions)
	}

	text := "Mention test @111000111000111"
	if err := ValidateMentionText(text, []string{"111000111000111@lid"}); err != nil {
		t.Fatalf("valid text rejected: %v", err)
	}
	if err := ValidateMentionText("hello", []string{"111000111000111@lid"}); err == nil {
		t.Fatal("text missing @token should be rejected")
	}
}

func TestBuildTextMessageMentionsAndQuote(t *testing.T) {
	chat := types.JID{User: "120363000000000001", Server: types.GroupServer}
	quoted := &store.Message{
		ID: "ORIG123", ChatJID: chat.String(), SenderJID: "222000222000222@lid",
		Content: "earlier", IsFromMe: false,
	}
	mentions := []string{"111000111000111@lid", "15555550100@s.whatsapp.net"}
	msg := buildTextMessage("hi @111000111000111", quoted, "111000111000111@lid", chat, mentions)
	ext := msg.GetExtendedTextMessage()
	if ext == nil {
		t.Fatal("expected ExtendedTextMessage")
	}
	ci := ext.GetContextInfo()
	if ci.GetStanzaID() != "ORIG123" {
		t.Fatalf("quote lost: %q", ci.GetStanzaID())
	}
	if got := ci.GetMentionedJID(); len(got) != 2 || got[0] != mentions[0] || got[1] != mentions[1] {
		t.Fatalf("mentions: %v", got)
	}

	mentionOnly := buildTextMessage("Mention test @111000111000111", nil, "111000111000111@lid", chat, []string{"111000111000111@lid"})
	if mentionOnly.GetConversation() != "" {
		t.Fatal("mention-only messages must be ExtendedTextMessage, not Conversation")
	}
	if got := mentionOnly.GetExtendedTextMessage().GetContextInfo().GetMentionedJID(); len(got) != 1 || got[0] != "111000111000111@lid" {
		t.Fatalf("mention-only MentionedJID: %v", got)
	}
}
