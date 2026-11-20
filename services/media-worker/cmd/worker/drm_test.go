package main

import (
	"strings"
	"testing"
)

func TestPackagerArgs_EncryptedHLSLayout(t *testing.T) {
	ladder := []RenditionProfile{
		{Name: "360p", Width: 640, Height: 360, VideoBitrate: "800k", AudioBitrate: "96k"},
		{Name: "480p", Width: 848, Height: 480, VideoBitrate: "1400k", AudioBitrate: "128k"},
	}
	mp4s := map[string]string{"360p": "mp4/360p.mp4", "480p": "mp4/480p.mp4"}
	args := packagerArgs(mp4s, mp4s["360p"], ladder, &DrmParams{KeyID: "id1", Key: "key1"})
	joined := strings.Join(args, " ")
	for _, want := range []string{
		"stream=audio", "playlist_name=audio/index.m3u8", "hls_group_id=audio",
		"stream=video", "playlist_name=360p/index.m3u8", "playlist_name=480p/index.m3u8",
		"segment_$Number%05d$.m4s", "init_segment=360p/init.mp4",
		"--enable_raw_key_encryption", "key_id=id1:key=key1",
		"Widevine,FairPlay", "cbcs", "master.m3u8",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %v", want, args)
		}
	}
}

func TestRungMP4Args_AudioOnlyFirst(t *testing.T) {
	withAudio := rungMP4Args("in.mp4", "out.mp4", RenditionProfile{Width: 640, Height: 360, VideoBitrate: "800k", AudioBitrate: "96k"}, true)
	joined := strings.Join(withAudio, " ")
	if !strings.Contains(joined, "-c:a aac") || !strings.Contains(joined, "+faststart") {
		t.Fatalf("audio mp4 must carry aac + faststart: %v", withAudio)
	}
	videoOnly := rungMP4Args("in.mp4", "out.mp4", RenditionProfile{Width: 640, Height: 360, VideoBitrate: "800k", AudioBitrate: "96k"}, false)
	if !strings.Contains(strings.Join(videoOnly, " "), "-an") {
		t.Fatalf("secondary rungs must drop audio: %v", videoOnly)
	}
}
