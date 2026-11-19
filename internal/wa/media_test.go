package wa

import (
	"testing"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"

	"github.com/thenavidm/whatsapp-mcp/internal/store"
)

func TestMediaDownloadableHasMediaType(t *testing.T) {
	d := &mediaDownloadable{directPath: "/v/x.enc?ccb=1", mediaType: whatsmeow.MediaImage}
	if got := whatsmeow.GetMediaType(d); got != whatsmeow.MediaImage {
		t.Fatalf("GetMediaType = %q, want %q", got, whatsmeow.MediaImage)
	}
}

func TestDirectPathKeepsQuery(t *testing.T) {
	cases := map[string]string{
		"https://mmg.whatsapp.net/v/t62.7118-24/1_2_n.enc?ccb=11-4&oh=abc&oe=def&_nc_sid=5e03e0": "/v/t62.7118-24/1_2_n.enc?ccb=11-4&oh=abc&oe=def&_nc_sid=5e03e0",
		"https://mmg.whatsapp.net/v/t62.7118-24/1_2_n.enc":                                       "/v/t62.7118-24/1_2_n.enc?",
		"/v/t62.7118-24/1_2_n.enc?ccb=11-4":                                                      "/v/t62.7118-24/1_2_n.enc?ccb=11-4",
	}
	for in, want := range cases {
		if got := directPath(in); got != want {
			t.Errorf("directPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFillMediaUnwrapsAndUsesDirectPath(t *testing.T) {
	img := &waE2E.ImageMessage{
		DirectPath: proto.String("/v/t62.7118-24/1_2_n.enc?ccb=11-4&oh=abc"),
		MediaKey:   []byte{1},
	}
	wrapped := []*waE2E.Message{
		{DeviceSentMessage: &waE2E.DeviceSentMessage{Message: &waE2E.Message{ImageMessage: img}}},
		{EphemeralMessage: &waE2E.FutureProofMessage{Message: &waE2E.Message{ImageMessage: img}}},
		{ViewOnceMessageV2: &waE2E.FutureProofMessage{Message: &waE2E.Message{ImageMessage: img}}},
	}
	for i, m := range wrapped {
		var sm store.StoredMessage
		fillMedia(&sm, m)
		if sm.MediaType != "image" {
			t.Errorf("case %d: media type %q, want image", i, sm.MediaType)
		}
		if got := directPath(sm.MediaURL); got != img.GetDirectPath() {
			t.Errorf("case %d: direct path %q, want %q", i, got, img.GetDirectPath())
		}
		if messageType(m) != "image" {
			t.Errorf("case %d: messageType %q, want image", i, messageType(m))
		}
	}
}
