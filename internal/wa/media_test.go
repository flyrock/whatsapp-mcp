package wa

import (
	"testing"

	"go.mau.fi/whatsmeow"
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
