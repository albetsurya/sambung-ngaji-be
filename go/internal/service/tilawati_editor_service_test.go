package service

import (
	"math"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func testTilawatiEditorService() *TilawatiEditorService {
	return &TilawatiEditorService{audioRoot: filepath.Join("/tmp", "frontend", "public", "audio", "tilawati")}
}

func TestCreateMergedAudioRendersTrimmedOgg(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg unavailable")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe unavailable")
	}
	root := t.TempDir()
	source := filepath.Join(root, "source.ogg")
	if output, err := exec.Command("ffmpeg", "-v", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=1", "-c:a", "libvorbis", source).CombinedOutput(); err != nil {
		t.Fatalf("source fixture: %v %s", err, output)
	}
	service := &TilawatiEditorService{audioRoot: root}
	output := filepath.Join(root, "output.ogg")
	if err := service.createMergedAudio([]FlattenedSegment{
		{SourceAudioURL: "/audio/tilawati/source.ogg", Start: 0.2, End: 0.4},
		{SourceAudioURL: "/audio/tilawati/source.ogg", Start: 0.6, End: 0.9},
	}, output); err != nil {
		t.Fatal(err)
	}
	duration, err := service.getAudioDuration(output)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(duration-0.5) > 0.03 {
		t.Fatalf("rendered duration=%f, want 0.5s", duration)
	}
}

func TestValidateTimelineDraft(t *testing.T) {
	draft := TimelineDraft{Timeline: PublishedTimeline{Version: 2}, UpdatedAt: time.Now().UnixMilli()}
	if err := validateTimelineDraft(1, 1, draft); err != nil {
		t.Fatal(err)
	}
	for _, invalid := range []struct {
		jilid, page int
		draft       TimelineDraft
	}{
		{7, 1, draft}, {1, 45, draft}, {1, 1, TimelineDraft{}},
		{1, 1, TimelineDraft{Timeline: PublishedTimeline{Version: 2}, UpdatedAt: time.Now().Add(time.Hour).UnixMilli()}},
	} {
		if err := validateTimelineDraft(invalid.jilid, invalid.page, invalid.draft); err == nil {
			t.Fatal("expected invalid draft to be rejected")
		}
	}
}

func TestFlattenTimelineSegmentsUsesPerSegmentSource(t *testing.T) {
	service := testTilawatiEditorService()
	segments := service.flattenTimelineSegments(TimelineClip{
		SourceAudioURL: "/audio/tilawati/jilid_1/hal1/fallback.ogg",
		TrimStart:      0,
		TrimEnd:        2,
		Segments: []TimelineClipSegment{
			{SourceAudioURL: "/audio/tilawati/jilid_1/hal1/first.ogg", Start: 0, End: 1},
			{SourceAudioURL: "/audio/tilawati/jilid_1/hal1/second.ogg", Start: 2, End: 3},
		},
	})

	if len(segments) != 2 {
		t.Fatalf("expected 2 flattened segments, got %d", len(segments))
	}
	if segments[0].SourceAudioURL == segments[1].SourceAudioURL {
		t.Fatalf("expected merged timeline to preserve each segment source")
	}
}

func TestAudioURLToFileRejectsRemoteAndTraversalPaths(t *testing.T) {
	service := testTilawatiEditorService()
	for _, audioURL := range []string{
		"https://example.com/audio.ogg",
		"/audio/tilawati/../../secrets.txt",
	} {
		if _, err := service.audioURLToFile(audioURL); err == nil {
			t.Fatalf("expected %q to be rejected", audioURL)
		}
	}
}

func TestValidatePublishRequestRejectsDuplicateClipIDs(t *testing.T) {
	service := testTilawatiEditorService()
	clip := TimelineClip{
		ID:       "same",
		Segments: []TimelineClipSegment{{SourceAudioURL: "/audio/tilawati/jilid_1/hal1/a.ogg", Start: 0, End: 1}},
		TrimEnd:  1,
	}
	err := service.validatePublishRequest(PublishTimelineRequest{
		Jilid:    1,
		Page:     1,
		Timeline: PublishedTimeline{Clips: []TimelineClip{clip, clip}},
	})
	if err == nil {
		t.Fatal("expected duplicate clip IDs to be rejected")
	}
}
