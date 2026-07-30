package pc98music

import "testing"

func TestYM2203EventRendererExpandsToneAndCarrierVolume(t *testing.T) {
	block := FMParameterBlock{
		FeedbackAlgorithm: 4,
		OperatorMask:      0x0F,
		OutputLevel:       [4]byte{10, 20, 30, 40},
	}
	renderer := newYM2203EventRenderer([]FMParameterBlock{block})
	writes, err := renderer.Render([]MusicEvent{
		{Channel: 1, Kind: EventSetParameterBlock, Parameter: 0},
		{Channel: 1, Kind: EventSetVolume, Value: 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	// Algorithm 4 has two carriers: one parameter load plus two complete
	// carrier redraws, 25 register writes each.
	if got, want := len(writes), 75; got != want {
		t.Fatalf("writes=%d, want %d", got, want)
	}
	for _, index := range []int{0, 25, 50} {
		if writes[index].Register != 0xB1 || writes[index].Value != 4 {
			t.Fatalf("tone header[%d]=%+v", index, writes[index])
		}
	}
	// Physical TL order is 1,3,2,4 at offsets 0C,04,08,00.
	base := writes[17:21]
	if base[0].Register != 0x4D || base[0].Value != 117 ||
		base[1].Register != 0x45 || base[1].Value != 97 ||
		base[2].Register != 0x49 || base[2].Value != 107 ||
		base[3].Register != 0x41 || base[3].Value != 87 {
		t.Fatalf("base TL writes=%+v", base)
	}
}

func TestYM2203EventRendererSkipsUnavailableStartupTone(t *testing.T) {
	renderer := newYM2203EventRenderer([]FMParameterBlock{{}})
	writes, err := renderer.Render([]MusicEvent{
		{Channel: 0, Kind: EventSetParameterBlock, Parameter: 58},
		{Channel: 0, Kind: EventSetVolume, Value: 80},
		{Channel: 0, Kind: EventRegisterWrite, Register: 0x28, Value: 0},
		{Channel: 0, Kind: EventSetParameterBlock, Parameter: 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(writes), 26; got != want {
		t.Fatalf("writes=%d, want %d", got, want)
	}
	if writes[0] != (YM2203RegisterWrite{Register: 0x28, Value: 0}) {
		t.Fatalf("direct write=%+v", writes[0])
	}
}
