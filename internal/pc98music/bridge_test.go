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
}

func TestAuditBridgeRejectsUnidentifiedInputs(t *testing.T) {
	if _, err := AuditBridge([]byte("not GAME.EXE"), []byte("not MSCDRV.EXE")); err == nil ||
		!strings.Contains(err.Error(), "GAME.EXE SHA-256") {
		t.Fatalf("unidentified input err=%v", err)
	}
}
