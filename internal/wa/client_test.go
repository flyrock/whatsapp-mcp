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

	plain := buildTextMessage("hello", nil, "111000111000111@lid", chat)
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
	reply := buildTextMessage("Test reply via bot", quoted, "111000111000111@lid", chat)
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
