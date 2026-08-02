package pc98music

import (
	"reflect"
	"testing"
)

func syntheticPlayback(t *testing.T) (*TrackPlayback, []MusicEvent) {
	t.Helper()
	data := make([]byte, 0x500)
	copy(data[0x210:], []byte{
		0xB5, 0x00, 0xC0, 0x00, 0xCC, 0x00, 0xD8, 0x00,
		0xE5, 0x00, 0xF2, 0x00, 0x01, 0x01, 0x10, 0x01,
		0x20, 0x01, 0x31, 0x01, 0x43, 0x01, 0x57, 0x01,
	})
	copy(data[0x228:], []byte{0xAE, 0x06, 0x4E, 0x06})
	data[0x10], data[0x11] = 1, 0
	data[0x12], data[0x13] = 0xFF, 0xFF

	streams := [][]byte{
		{0x85, 0x02, 0x8A, 0x64, 0x03, 0x02},
		{0x80, 0x02},
		{0x80, 0x02},
		{0x02, 0x02},
		{0x80, 0x02},
		{0xB0, 0x07, 0x38, 0x02, 0x02},
		{0x20, 0x02},
	}
	track := TrackDescriptor{HeaderWords: [4]int{0, 0xC0, 0, 0}}
	for channel, stream := range streams {
		offset := 0x300 + channel*0x20
		copy(data[offset:], stream)
		raw1, raw2 := 0, 10
		if channel < 3 {
			raw1, raw2 = 1, 100
		}
		track.Channels = append(track.Channels, TrackChannel{
			Channel: channel, SequenceOffset: offset,
			SequenceLength: len(stream), RawParameter1: raw1, RawParameter2: raw2,
		})
	}
	playback, initial, err := newTrackPlayback(data, track)
	if err != nil {
		t.Fatal(err)
	}
	return playback, initial
}

func TestTrackPlaybackInitializesTempoFMAndPSG(t *testing.T) {
	playback, initial := syntheticPlayback(t)
	if len(initial) != 14 {
		t.Fatalf("initial event count=%d events=%+v", len(initial), initial)
	}
	if initial[0].Kind != EventRegisterWrite ||
		initial[0].Register != 0x26 || initial[0].Value != 0xC0 {
		t.Fatalf("tempo event=%+v", initial[0])
	}
	if initial[1].Kind != EventSetParameterBlock ||
		initial[2].Kind != EventSetVolume ||
		initial[3].Register != 0x28 {
		t.Fatalf("FM init=%+v", initial[1:4])
	}
	if playback.registers[0x08] != 10 ||
		playback.registers[0x09] != 10 ||
		playback.registers[0x0A] != 10 ||
		playback.registers[0x07] != 0xB8 {
		t.Fatalf("PSG/mixer registers=%v", playback.registers[7:11])
	}
}

func TestTrackPlaybackEmitsCommandSideEffectsAndSustain(t *testing.T) {
	playback, _ := syntheticPlayback(t)
	events, err := playback.Tick(64)
	if err != nil {
		t.Fatal(err)
	}
	if playback.registers[0xA0] != 0xB5 ||
		playback.registers[0xA4] != 0x00 {
		t.Fatalf(
			"FM registers A0=%02X A4=%02X",
			playback.registers[0xA0], playback.registers[0xA4],
		)
	}
	if playback.registers[0x00] != 0xAE ||
		playback.registers[0x01] != 0x06 ||
		playback.registers[0x08] != 11 {
		t.Fatalf("PSG registers=%v", playback.registers[:11])
	}
	if playback.registers[0x07] != 0x38 {
		t.Fatalf("PSG B0 direct register write=%02X", playback.registers[0x07])
	}
	foundParameter, foundVolume, foundKeyOn := false, false, false
	for _, event := range events {
		foundParameter = foundParameter ||
			(event.Kind == EventSetParameterBlock && event.Parameter == 2)
		foundVolume = foundVolume ||
			(event.Kind == EventSetVolume && event.Value == 100)
		foundKeyOn = foundKeyOn ||
			(event.Kind == EventRegisterWrite && event.Register == 0x28 &&
				event.Value == 0xF0)
	}
	if !foundParameter || !foundVolume || !foundKeyOn {
		t.Fatalf("events lack FM parameter side effects: %+v", events)
	}

	if _, err := playback.Tick(64); err != nil {
		t.Fatal(err)
	}
	if playback.registers[0x08] != 10 {
		t.Fatalf("PSG envelope decrement=%d, want 10", playback.registers[0x08])
	}
}

func TestTrackPlaybackRejectsPSGPeriodOutsideTable(t *testing.T) {
	playback, _ := syntheticPlayback(t)
	state := &playback.channels[3]
	var events []MusicEvent
	err := playback.startPSG(3, state, StreamCommand{
		Opcode: 0x01, Name: "note", Operands: []byte{1},
	}, &events)
	if err == nil {
		t.Fatal("PSG note below table was accepted")
	}
}

func TestTrackPlaybackSnapshotRestoresExactEventContinuation(t *testing.T) {
	original, _ := syntheticPlayback(t)
	for tick := 0; tick < 1; tick++ {
		if _, err := original.Tick(4096); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := original.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	var want []MusicEvent
	for tick := 0; tick < 1; tick++ {
		events, err := original.Tick(4096)
		if err != nil {
			t.Fatal(err)
		}
		want = append(want, events...)
	}
	restored, _ := syntheticPlayback(t)
	if err := restored.Restore(snapshot); err != nil {
		t.Fatal(err)
	}
	var got []MusicEvent
	for tick := 0; tick < 1; tick++ {
		events, err := restored.Tick(4096)
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, events...)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("restored track event continuation differs")
	}
	bad := snapshot
	bad.Channels[0].Machine.PC = bad.Channels[0].Machine.End + 1
	if err := restored.Restore(bad); err == nil {
		t.Fatal("out-of-range sequence PC was accepted")
	}
}

func TestPlaybackAuditSeparatesCalledAndAudibleParameters(t *testing.T) {
	playback, _ := syntheticPlayback(t)
	track := TrackDescriptor{Selector: 99}
	for channel := range playback.channels {
		state := playback.channels[channel]
		track.Channels = append(track.Channels, TrackChannel{
			Channel: channel, SequenceOffset: state.machine.Start,
			SequenceLength: state.machine.End - state.machine.Start,
			RawParameter1:  1, RawParameter2: 100,
		})
	}
	audit, err := auditTrackPlayback(playback.driverData, track, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.ParameterIndices) != 2 || audit.ParameterIndices[0] != 1 ||
		audit.ParameterIndices[1] != 2 ||
		len(audit.AudibleParameterIndices) != 1 || audit.AudibleParameterIndices[0] != 2 ||
		!audit.AudibleParametersComplete {
		t.Fatalf("playback audit = %+v", audit)
	}
}
