package pc98music

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/wicanr2/golden-box-remake-engine/audio/pc98soundbios"
	"github.com/wicanr2/golden-box-remake-engine/audio/ym2203"
)

const (
	fmParameterTableFile  = 0x45A2
	fmParameterBlockBytes = 100
	fmParameterWordCount  = 50
	fmEmbeddedBlockCount  = 20
)

// FMParameterBlock is the NEC Sound BIOS AH=16h WORD-format tone block.
// Field order follows PC-9800 Technical Data Book BIOS (1992), pp. 370-372.
// Raw is retained so later Sound BIOS register-trace work can audit every
// field without reconstructing commercial source bytes from this type.
type FMParameterBlock struct {
	Raw                 [fmParameterWordCount]uint16
	FeedbackAlgorithm   byte
	AttackRate          [4]byte
	OperatorMask        byte
	DecayRate           [4]byte
	LFOWaveform         byte
	SustainRate         [4]byte
	LFOSyncDelay        byte
	ReleaseRate         [4]byte
	LFOSpeed            uint16
	SustainLevel        [4]byte
	LFOPitchDepth       int8
	OutputLevel         [4]byte
	LFOAmplitudeDepth   int8
	KeyScale            [4]byte
	LFOPitchDepthCoarse byte
	Multiple            [4]byte
	Reserved40          byte
	// Detune preserves the signed WORD interpretation. The embedded corpus
	// contains both sign-extended negatives and 3-bit values 4..7, so their
	// normalization to the NEC documented -4..3 range remains unresolved.
	Detune                  [4]int16
	Reserved45              byte
	LFOAmplitudeDepthCoarse [4]byte
}

// YM2203ToneSignature is the stable, volume-independent FM timbre state
// written by NEC Sound BIOS. Operator order follows the physical YM2203
// register slots (1, 3, 2, 4). Total-level values are excluded because the
// track volume modifies carrier levels after loading a parameter block.
type YM2203ToneSignature [21]byte

var ym2203OperatorOrder = [...]int{0, 2, 1, 3}
var ym2203PhysicalOffsets = [...]byte{0x0C, 0x04, 0x08, 0x00}

// YM2203RegisterWrite is one address/data write to the title's single OPN.
type YM2203RegisterWrite struct {
	Register byte
	Value    byte
}

// YM2203Signature projects the NEC WORD-format parameters to the register
// values verified against Hoot S98 traces. NEC rate and sustain-level fields
// are attenuation/delay parameters, so Sound BIOS complements them before
// writing the chip. Signed DETUNE is shifted as an 8-bit value; this preserves
// bit 7 for negative inputs exactly as the runtime trace does.
func (block FMParameterBlock) YM2203Signature() YM2203ToneSignature {
	var result YM2203ToneSignature
	result[0] = block.FeedbackAlgorithm
	for physical, operator := range ym2203OperatorOrder {
		result[1+physical] = byte(block.Detune[operator])<<4 | block.Multiple[operator]
		result[5+physical] = block.KeyScale[operator]<<6 | (0x1f - block.AttackRate[operator])
		result[9+physical] = 0x1f - block.DecayRate[operator]
		result[13+physical] = 0x1f - block.SustainRate[operator]
		result[17+physical] = (0x0f-block.SustainLevel[operator])<<4 |
			(0x0f - block.ReleaseRate[operator])
	}
	return result
}

// YM2203LevelSequence reproduces NEC Sound BIOS AH=16h SETPARABLOCK followed
// by AH=1Fh SETVOLUME. The first item is the parameter block's four output
// levels in physical register order. Each following item is one complete
// re-render after replacing the next carrier's output level with volume.
func (block FMParameterBlock) YM2203LevelSequence(volume byte) ([][4]byte, error) {
	var levels [4]byte
	for physical, operator := range ym2203OperatorOrder {
		levels[physical] = 0x7f - block.OutputLevel[operator]
	}
	result := [][4]byte{levels}
	carriers, err := ym2203.CarrierOperators(block.FeedbackAlgorithm & 7)
	if err != nil {
		return nil, err
	}
	for _, operator := range carriers {
		physical, err := ym2203.PhysicalOperatorIndex(operator)
		if err != nil {
			return nil, err
		}
		levels[physical] = 0x7f - volume
		result = append(result, levels)
	}
	return result, nil
}

// YM2203ToneWrites expands one complete NEC Sound BIOS tone redraw in the
// register order proven by the original S98 traces.
func (block FMParameterBlock) YM2203ToneWrites(
	channel int, levels [4]byte,
) ([]YM2203RegisterWrite, error) {
	if channel < 0 || channel >= 3 {
		return nil, fmt.Errorf("YM2203 FM channel %d is outside 0..2", channel)
	}
	signature := block.YM2203Signature()
	result := make([]YM2203RegisterWrite, 0, len(signature)+4)
	result = append(result, YM2203RegisterWrite{
		Register: byte(0xB0 + channel),
		Value:    signature[0],
	})
	signatureOffset := 5
	for group, base := range []byte{0x50, 0x60, 0x70, 0x80, 0x40, 0x30} {
		for physical, operatorOffset := range ym2203PhysicalOffsets {
			value := levels[physical]
			if group != 4 {
				if group == 5 {
					value = signature[1+physical]
				} else {
					value = signature[signatureOffset+physical]
				}
			}
			result = append(result, YM2203RegisterWrite{
				Register: base + operatorOffset + byte(channel),
				Value:    value,
			})
		}
		if group < 4 {
			signatureOffset += 4
		}
	}
	return result, nil
}

// YM2203Modulation projects one signed Sound BIOS software-LFO sample onto
// the current F-number and four physical-register total levels.
func (block FMParameterBlock) YM2203Modulation(
	sample int16, baseFNumber uint16, baseLevels [4]byte,
) (uint16, [4]byte) {
	pitchDepth := int16(block.LFOPitchDepth) * int16(block.LFOPitchDepthCoarse)
	pitch := pc98soundbios.Pitch(baseFNumber, sample, pitchDepth)
	levels := baseLevels
	for physical, operator := range ym2203OperatorOrder {
		levels[physical] = pc98soundbios.TotalLevel(
			baseLevels[physical], sample, block.LFOAmplitudeDepth,
			block.LFOAmplitudeDepthCoarse[operator],
		)
	}
	return pitch, levels
}

// SoundBIOSModulationConfig maps only the NEC fields consumed by the
// game-neutral Timer B scheduler. The proprietary parameter bank and its
// offsets remain in this CoAB adapter.
func (block FMParameterBlock) SoundBIOSModulationConfig() pc98soundbios.ModulationConfig {
	return pc98soundbios.ModulationConfig{
		Waveform:  block.LFOWaveform,
		SyncDelay: block.LFOSyncDelay,
		Speed:     block.LFOSpeed,
	}
}

// FMParameterBankAudit reports only provenance, hashes and index coverage.
// The proprietary tone values stay in the user's local media.
type FMParameterBankAudit struct {
	FileOffset         int    `json:"file_offset"`
	BlockBytes         int    `json:"block_bytes"`
	CompleteBlocks     int    `json:"complete_blocks"`
	SHA256             string `json:"sha256"`
	UsedIndices        []int  `json:"used_indices"`
	UnavailableIndices []int  `json:"unavailable_indices"`
}

func parameterWord(raw []byte, index int) uint16 {
	offset := index * 2
	return uint16(raw[offset]) | uint16(raw[offset+1])<<8
}

func parameterByte(value uint16, field string, limit uint16) (byte, error) {
	if value > limit {
		return 0, fmt.Errorf("%s value 0x%04X exceeds 0x%02X", field, value, limit)
	}
	return byte(value), nil
}

func parameterSignedByte(value uint16, field string, minimum, maximum int16) (int8, error) {
	signed := int16(value)
	if signed < minimum || signed > maximum {
		return 0, fmt.Errorf("%s value %d is outside %d..%d", field, signed, minimum, maximum)
	}
	return int8(signed), nil
}

func parseFMParameterBlock(raw []byte) (FMParameterBlock, error) {
	if len(raw) != fmParameterBlockBytes {
		return FMParameterBlock{}, fmt.Errorf(
			"FM parameter block has %d bytes, want %d",
			len(raw), fmParameterBlockBytes,
		)
	}
	var block FMParameterBlock
	for index := range block.Raw {
		block.Raw[index] = parameterWord(raw, index)
	}
	var err error
	if block.FeedbackAlgorithm, err = parameterByte(block.Raw[0], "FB_ALG", 0x3F); err != nil {
		return FMParameterBlock{}, err
	}
	for operator := 0; operator < 4; operator++ {
		fields := []struct {
			target *byte
			word   int
			name   string
			limit  uint16
		}{
			{&block.AttackRate[operator], 1 + operator, "ATTACK_RATE", 0x1F},
			{&block.DecayRate[operator], 6 + operator, "DECAY_RATE", 0x1F},
			{&block.SustainRate[operator], 11 + operator, "SUSTAIN_RATE", 0x1F},
			{&block.ReleaseRate[operator], 16 + operator, "RELEASE_RATE", 0x0F},
			{&block.SustainLevel[operator], 21 + operator, "SUSTAIN_LEVEL", 0x0F},
			{&block.OutputLevel[operator], 26 + operator, "OUTPUT_LEVEL", 0x7F},
			{&block.KeyScale[operator], 31 + operator, "KEY_SCALE", 0x03},
			{&block.Multiple[operator], 36 + operator, "MULTIPLE", 0x0F},
			{&block.LFOAmplitudeDepthCoarse[operator], 46 + operator, "LFO_AMPLITUDE_COARSE", 0x0F},
		}
		for _, field := range fields {
			*field.target, err = parameterByte(block.Raw[field.word], field.name, field.limit)
			if err != nil {
				return FMParameterBlock{}, fmt.Errorf("operator %d: %w", operator+1, err)
			}
		}
		block.Detune[operator] = int16(block.Raw[41+operator])
	}
	if block.OperatorMask, err = parameterByte(block.Raw[5], "OPERATOR_MASK", 0x0F); err != nil {
		return FMParameterBlock{}, err
	}
	if block.LFOWaveform, err = parameterByte(block.Raw[10], "LFO_WAVEFORM", 0x03); err != nil {
		return FMParameterBlock{}, err
	}
	if block.LFOSyncDelay, err = parameterByte(block.Raw[15], "LFO_SYNC_DELAY", 0xFF); err != nil {
		return FMParameterBlock{}, err
	}
	block.LFOSpeed = block.Raw[20]
	if block.LFOSpeed > 0x3FFF {
		return FMParameterBlock{}, fmt.Errorf("LFO_SPEED value 0x%04X exceeds 0x3FFF", block.LFOSpeed)
	}
	if block.LFOPitchDepth, err = parameterSignedByte(
		block.Raw[25], "LFO_PITCH_DEPTH", -128, 127,
	); err != nil {
		return FMParameterBlock{}, err
	}
	if block.LFOAmplitudeDepth, err = parameterSignedByte(
		block.Raw[30], "LFO_AMPLITUDE_DEPTH", -128, 127,
	); err != nil {
		return FMParameterBlock{}, err
	}
	if block.LFOPitchDepthCoarse, err = parameterByte(
		block.Raw[35], "LFO_PITCH_COARSE", 0x0F,
	); err != nil {
		return FMParameterBlock{}, err
	}
	if block.Reserved40, err = parameterByte(block.Raw[40], "RESERVED_40", 0xFF); err != nil {
		return FMParameterBlock{}, err
	}
	if block.Reserved45, err = parameterByte(block.Raw[45], "RESERVED_45", 0xFF); err != nil {
		return FMParameterBlock{}, err
	}
	return block, nil
}

func parseEmbeddedFMParameterBlocks(driver []byte) ([]FMParameterBlock, string, error) {
	end := fmParameterTableFile + fmEmbeddedBlockCount*fmParameterBlockBytes
	if fmParameterTableFile < DriverMissingEnd {
		return nil, "", fmt.Errorf("FM parameter table unexpectedly overlaps missing sector")
	}
	if end > len(driver) {
		return nil, "", fmt.Errorf(
			"embedded FM parameter table 0x%X..0x%X exceeds %d-byte driver",
			fmParameterTableFile, end, len(driver),
		)
	}
	rawBank := driver[fmParameterTableFile:end]
	blocks := make([]FMParameterBlock, 0, fmEmbeddedBlockCount)
	for index := 0; index < fmEmbeddedBlockCount; index++ {
		start := index * fmParameterBlockBytes
		block, err := parseFMParameterBlock(rawBank[start : start+fmParameterBlockBytes])
		if err != nil {
			return nil, "", fmt.Errorf("FM parameter block %d: %w", index, err)
		}
		blocks = append(blocks, block)
	}
	sum := sha256.Sum256(rawBank)
	return blocks, hex.EncodeToString(sum[:]), nil
}

func auditFMParameterBank(driver []byte, playback []PlaybackAudit) (FMParameterBankAudit, error) {
	blocks, digest, err := parseEmbeddedFMParameterBlocks(driver)
	if err != nil {
		return FMParameterBankAudit{}, err
	}
	usedSet := make(map[int]bool)
	for _, track := range playback {
		for _, index := range track.ParameterIndices {
			usedSet[index] = true
		}
	}
	used := make([]int, 0, len(usedSet))
	var unavailable []int
	for index := range usedSet {
		used = append(used, index)
		if index >= len(blocks) {
			unavailable = append(unavailable, index)
		}
	}
	sort.Ints(used)
	sort.Ints(unavailable)
	return FMParameterBankAudit{
		FileOffset:         fmParameterTableFile,
		BlockBytes:         fmParameterBlockBytes,
		CompleteBlocks:     len(blocks),
		SHA256:             digest,
		UsedIndices:        used,
		UnavailableIndices: unavailable,
	}, nil
}

func parametersWithinEmbeddedBank(indices []int) bool {
	for _, index := range indices {
		if index < 0 || index >= fmEmbeddedBlockCount {
			return false
		}
	}
	return true
}

// EmbeddedFMParameterBlocks parses the twenty complete tone blocks present in
// the identified user-media driver. It deliberately does not fabricate the
// additional indices referenced by eight tracks.
func EmbeddedFMParameterBlocks(driver []byte) ([]FMParameterBlock, error) {
	if hash := fileSHA256(driver); hash != DriverSHA256 {
		return nil, fmt.Errorf("MSCDRV.EXE SHA-256 %s, want %s", hash, DriverSHA256)
	}
	blocks, _, err := parseEmbeddedFMParameterBlocks(driver)
	return blocks, err
}
