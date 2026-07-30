package pc98music

import (
	"strings"
	"testing"
)

func TestVerifyAnchorsAcceptsExactBytesAndRejectsMismatch(t *testing.T) {
	data := []byte{0x00, 0x10, 0x20, 0x30, 0x00}
	anchors := []Anchor{{
		Binary:     "fixture",
		Label:      "exact",
		FileOffset: 1,
		Bytes:      []byte{0x10, 0x20, 0x30},
	}}
	results, err := verifyAnchors(data, anchors)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].FileOffset != 1 || results[0].Size != 3 {
		t.Fatalf("results=%+v", results)
	}
	data[2] = 0xFF
	if _, err := verifyAnchors(data, anchors); err == nil ||
		!strings.Contains(err.Error(), "byte mismatch") {
		t.Fatalf("mismatch err=%v", err)
	}
}

func TestDriverBridgeAnchorsDoNotOverlapMissingSector(t *testing.T) {
	for _, anchor := range driverAnchors {
		end := anchor.FileOffset + len(anchor.Bytes)
		if anchor.FileOffset < DriverMissingEnd && end > DriverMissingStart {
			t.Fatalf("%s overlaps missing range: 0x%X..0x%X", anchor.Label, anchor.FileOffset, end)
		}
	}
	for _, service := range soundBIOSServices {
		end := service.FileOffset + 4
		if service.FileOffset < DriverMissingEnd && end > DriverMissingStart {
			t.Fatalf("%s overlaps missing range: 0x%X..0x%X", service.Name, service.FileOffset, end)
		}
	}
}

func TestDriverAnchorsCoverTimerBInterruptOwnership(t *testing.T) {
	want := map[string]bool{
		"INSTALL_MUSIC_TIMER_INTERRUPT":   false,
		"INITIALIZE_TIMER_B_26_27":        false,
		"TIMER_B_ONLY_PLAYBACK_DISPATCH":  false,
		"TIMER_B_RESTART_AND_ACKNOWLEDGE": false,
	}
	for _, anchor := range driverAnchors {
		if _, ok := want[anchor.Label]; ok {
			want[anchor.Label] = true
		}
	}
	for label, found := range want {
		if !found {
			t.Fatalf("missing Timer B ownership anchor %s", label)
		}
	}
}

func TestSoundBIOSServiceTableMatchesObservedCommandSet(t *testing.T) {
	want := []byte{
		0x00, 0x02,
		0x10, 0x11, 0x12, 0x13, 0x14,
		0x16, 0x17, 0x18, 0x19, 0x1A, 0x1B, 0x1C, 0x1D, 0x1E, 0x1F,
	}
	if len(soundBIOSServices) != len(want) {
		t.Fatalf("services=%d, want %d", len(soundBIOSServices), len(want))
	}
	seen := make(map[byte]bool, len(soundBIOSServices))
	for _, service := range soundBIOSServices {
		seen[service.Command] = true
	}
	for _, command := range want {
		if !seen[command] {
			t.Fatalf("missing Sound BIOS command AH=%02X", command)
		}
	}
	if seen[0x15] {
		t.Fatal("MSCDRV has no observed INT D2h SETTEMPO wrapper")
	}
}

func TestAuditTrackDescriptorsRejectsOutOfRangePointer(t *testing.T) {
	data := make([]byte, driverTrackTableFile+driverPublicTracks*2)
	data[driverTrackTableFile] = 0xFF
	data[driverTrackTableFile+1] = 0x7F
	if _, err := auditTrackDescriptors(data); err == nil ||
		!strings.Contains(err.Error(), "maps outside input") {
		t.Fatalf("out-of-range descriptor err=%v", err)
	}
}

func TestRangeCompleteUsesHalfOpenMissingInterval(t *testing.T) {
	tests := []struct {
		start, end int
		want       bool
	}{
		{DriverMissingStart - 1, DriverMissingStart, true},
		{DriverMissingStart - 1, DriverMissingStart + 1, false},
		{DriverMissingEnd - 1, DriverMissingEnd, false},
		{DriverMissingEnd, DriverMissingEnd + 1, true},
	}
	for _, test := range tests {
		if got := rangeComplete(test.start, test.end); got != test.want {
			t.Fatalf("rangeComplete(0x%X,0x%X)=%v, want %v", test.start, test.end, got, test.want)
		}
	}
}

func TestExtractTrackSequencesRejectsWrongDriverAndSelector(t *testing.T) {
	if _, err := ExtractTrackSequences([]byte("wrong driver"), 1); err == nil ||
		!strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("wrong driver err=%v", err)
	}
	if _, err := ExtractTrackSequences(nil, 0); err == nil ||
		!strings.Contains(err.Error(), "outside 1..12") {
		t.Fatalf("selector bounds err=%v", err)
	}
}

func TestAuditBridgeRejectsUnidentifiedInputs(t *testing.T) {
	if _, err := AuditBridge([]byte("not GAME.EXE"), []byte("not MSCDRV.EXE")); err == nil ||
		!strings.Contains(err.Error(), "GAME.EXE SHA-256") {
		t.Fatalf("unidentified input err=%v", err)
	}
}
