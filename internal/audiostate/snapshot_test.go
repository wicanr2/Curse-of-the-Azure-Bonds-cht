package audiostate

import "testing"

func TestSnapshotValidationRejectsAmbiguousOrUnboundedRecords(t *testing.T) {
	valid := Snapshot{Version: CurrentVersion, Enabled: true, OneShots: []OneShot{{
		Backend: BackendDOSWAV, Key: "2", PositionFrames: 123,
	}}}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	duplicate := valid
	duplicate.OneShots = append(duplicate.OneShots, duplicate.OneShots[0])
	if err := duplicate.Validate(); err == nil {
		t.Fatal("duplicate one-shot identity was accepted")
	}
	nonCanonical := Clone(valid)
	nonCanonical.OneShots[0].Key = "02"
	if err := nonCanonical.Validate(); err == nil {
		t.Fatal("non-canonical DOS selector was accepted")
	}
	disabled := valid
	disabled.Enabled = false
	if err := disabled.Validate(); err == nil {
		t.Fatal("disabled snapshot with active one-shot was accepted")
	}
	oversized := Snapshot{Version: CurrentVersion, Enabled: true, OneShots: make([]OneShot, MaxActiveOneShots+1)}
	if err := oversized.Validate(); err == nil {
		t.Fatal("oversized one-shot list was accepted")
	}
}
