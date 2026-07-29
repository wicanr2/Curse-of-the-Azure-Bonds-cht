package pc98music

import (
	"encoding/binary"
	"reflect"
	"strings"
	"testing"
)

func validFMParameterFixture() []byte {
	words := [fmParameterWordCount]uint16{}
	signedWord := func(value int16) uint16 { return uint16(value) }
	words[0] = 0x3A
	for operator := 0; operator < 4; operator++ {
		words[1+operator] = uint16(10 + operator)
		words[6+operator] = uint16(20 + operator)
		words[11+operator] = uint16(15 + operator)
		words[16+operator] = uint16(8 + operator)
		words[21+operator] = uint16(operator)
		words[26+operator] = uint16(64 + operator)
		words[31+operator] = uint16(operator % 4)
		words[36+operator] = uint16(2 + operator)
		words[41+operator] = signedWord(int16(-1 + operator))
		words[46+operator] = uint16(operator)
	}
	words[5] = 0x0F
	words[10] = 2
	words[15] = 12
	words[20] = 3000
	words[25] = signedWord(-12)
	words[30] = 9
	words[35] = 4

	raw := make([]byte, fmParameterBlockBytes)
	for index, value := range words {
		binary.LittleEndian.PutUint16(raw[index*2:], value)
	}
	return raw
}

func TestParseFMParameterBlockMapsOfficialWordFields(t *testing.T) {
	block, err := parseFMParameterBlock(validFMParameterFixture())
	if err != nil {
		t.Fatal(err)
	}
	if block.FeedbackAlgorithm != 0x3A ||
		block.OperatorMask != 0x0F ||
		block.AttackRate != [4]byte{10, 11, 12, 13} ||
		block.OutputLevel != [4]byte{64, 65, 66, 67} ||
		block.Detune != [4]int16{-1, 0, 1, 2} ||
		block.LFOSpeed != 3000 ||
		block.LFOPitchDepth != -12 {
		t.Fatalf("parsed block=%+v", block)
	}
}

func TestYM2203SignatureAppliesSoundBIOSTransforms(t *testing.T) {
	block, err := parseFMParameterBlock(validFMParameterFixture())
	if err != nil {
		t.Fatal(err)
	}
	got := block.YM2203Signature()
	if got[0] != block.FeedbackAlgorithm ||
		got[1] != byte(block.Detune[0])<<4|block.Multiple[0] ||
		got[2] != byte(block.Detune[2])<<4|block.Multiple[2] ||
		got[5] != block.KeyScale[0]<<6|(0x1f-block.AttackRate[0]) ||
		got[9] != 0x1f-block.DecayRate[0] ||
		got[13] != 0x1f-block.SustainRate[0] ||
		got[17] != (0x0f-block.SustainLevel[0])<<4|(0x0f-block.ReleaseRate[0]) {
		t.Fatalf("unexpected YM2203 signature: %x", got)
	}
}

func TestYM2203LevelSequenceUsesAlgorithmCarriers(t *testing.T) {
	block := FMParameterBlock{
		FeedbackAlgorithm: 4,
		OutputLevel:       [4]byte{102, 107, 102, 107},
	}
	got, err := block.YM2203LevelSequence(105)
	if err != nil {
		t.Fatal(err)
	}
	want := [][4]byte{
		{25, 25, 20, 20},
		{25, 25, 20, 22},
		{25, 25, 22, 22},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("level sequence=%v, want %v", got, want)
	}
}

func TestYM2203ModulationUsesLogicalOperatorDepths(t *testing.T) {
	block := FMParameterBlock{
		LFOPitchDepth:           30,
		LFOPitchDepthCoarse:     1,
		LFOAmplitudeDepth:       12,
		LFOAmplitudeDepthCoarse: [4]byte{15, 0, 5, 10},
	}
	pitch, levels := block.YM2203Modulation(
		0x4000, 0x2000, [4]byte{40, 50, 60, 70},
	)
	if pitch != 0x2003 {
		t.Fatalf("pitch=%04x, want 2003", pitch)
	}
	// Physical order is logical 1,3,2,4.
	if levels != [4]byte{41, 50, 60, 72} {
		t.Fatalf("levels=%v, want [41 50 60 72]", levels)
	}
}

func TestParseFMParameterBlockRejectsInvalidFieldAndLength(t *testing.T) {
	if _, err := parseFMParameterBlock(make([]byte, 99)); err == nil ||
		!strings.Contains(err.Error(), "99 bytes") {
		t.Fatalf("length error=%v", err)
	}
	raw := validFMParameterFixture()
	binary.LittleEndian.PutUint16(raw[1*2:], 0x20)
	if _, err := parseFMParameterBlock(raw); err == nil ||
		!strings.Contains(err.Error(), "ATTACK_RATE") {
		t.Fatalf("attack-rate error=%v", err)
	}
}

func TestEmbeddedFMParameterTableDoesNotOverlapMissingSector(t *testing.T) {
	if fmParameterTableFile < DriverMissingEnd {
		t.Fatalf(
			"parameter table 0x%X overlaps missing range ending 0x%X",
			fmParameterTableFile, DriverMissingEnd,
		)
	}
	if got := fmEmbeddedBlockCount * fmParameterBlockBytes; got != 2000 {
		t.Fatalf("embedded bank bytes=%d, want 2000", got)
	}
}

func TestParametersWithinEmbeddedBank(t *testing.T) {
	if !parametersWithinEmbeddedBank([]int{0, 7, 19}) {
		t.Fatal("available embedded parameter indices reported incomplete")
	}
	if parametersWithinEmbeddedBank([]int{19, 20}) {
		t.Fatal("index 20 reported available in twenty-block bank")
	}
	if _, err := EmbeddedFMParameterBlocks([]byte("wrong driver")); err == nil ||
		!strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("wrong-driver error=%v", err)
	}
}
