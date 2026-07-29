// Package pc98ovr reads the Turbo Pascal overlay container used by the
// PC-9801 version of Curse of the Azure Bonds.
package pc98ovr

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

const (
	magic             = "TPOV"
	controlRecordSize = 0x30
)

// Control describes one Turbo Pascal overlay control record embedded in the
// resident executable. FileOffset points into the TPOV file. CodeSize bytes
// of executable code are followed by RelocationSize bytes of fixups.
type Control struct {
	ExecutableOffset int
	FileOffset       uint32
	CodeSize         uint16
	RelocationSize   uint16
	EntryCount       uint16
}

// Overlay is one validated code/fixup span in the overlay file.
type Overlay struct {
	Control
	Code       []byte
	Relocation []byte
}

// ParseControls finds structurally valid Turbo Pascal control records in a
// resident MZ image. Validation is based on TPOV bounds and record chaining;
// an isolated CD 3F byte sequence is not accepted as evidence.
func ParseControls(executable, overlayFile []byte) ([]Control, error) {
	if len(overlayFile) < len(magic) || string(overlayFile[:len(magic)]) != magic {
		return nil, errors.New("overlay file does not start with TPOV")
	}

	var candidates []Control
	for offset := 0; offset+controlRecordSize <= len(executable); offset++ {
		if executable[offset] != 0xcd || executable[offset+1] != 0x3f {
			continue
		}
		fileOffset := binary.LittleEndian.Uint32(executable[offset+4 : offset+8])
		codeSize := binary.LittleEndian.Uint16(executable[offset+8 : offset+10])
		relocationSize := binary.LittleEndian.Uint16(executable[offset+10 : offset+12])
		end := uint64(fileOffset) + uint64(codeSize) + uint64(relocationSize)
		if fileOffset < uint32(len(magic)) || codeSize == 0 || end > uint64(len(overlayFile)) {
			continue
		}
		candidates = append(candidates, Control{
			ExecutableOffset: offset,
			FileOffset:       fileOffset,
			CodeSize:         codeSize,
			RelocationSize:   relocationSize,
			EntryCount:       binary.LittleEndian.Uint16(executable[offset+12 : offset+14]),
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].FileOffset != candidates[j].FileOffset {
			return candidates[i].FileOffset < candidates[j].FileOffset
		}
		return candidates[i].ExecutableOffset < candidates[j].ExecutableOffset
	})

	// A real control table describes a chain beginning immediately after TPOV.
	// Keep only that chain so coincidental CD 3F sequences in resident code do
	// not become false overlay records.
	next := uint32(len(magic))
	controls := make([]Control, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate.FileOffset < next {
			continue
		}
		if candidate.FileOffset > next {
			break
		}
		controls = append(controls, candidate)
		next += uint32(candidate.CodeSize) + uint32(candidate.RelocationSize)
	}
	if len(controls) == 0 {
		return nil, errors.New("no chained Turbo Pascal overlay controls found")
	}
	return controls, nil
}

// Decode validates controls and slices executable code separately from fixups.
func Decode(executable, overlayFile []byte) ([]Overlay, error) {
	controls, err := ParseControls(executable, overlayFile)
	if err != nil {
		return nil, err
	}
	decoded := make([]Overlay, 0, len(controls))
	for _, control := range controls {
		codeStart := uint64(control.FileOffset)
		codeEnd := codeStart + uint64(control.CodeSize)
		relocationEnd := codeEnd + uint64(control.RelocationSize)
		if relocationEnd > uint64(len(overlayFile)) {
			return nil, fmt.Errorf("overlay at 0x%X exceeds TPOV bounds", control.FileOffset)
		}
		decoded = append(decoded, Overlay{
			Control:    control,
			Code:       overlayFile[codeStart:codeEnd],
			Relocation: overlayFile[codeEnd:relocationEnd],
		})
	}
	return decoded, nil
}

// InterruptOffsets returns offsets of literal INT interrupt instructions in
// code. Callers must not use it on relocation tables or arbitrary file bytes.
func InterruptOffsets(code []byte, interrupt byte) []int {
	var offsets []int
	for offset := 0; offset+1 < len(code); offset++ {
		if code[offset] == 0xcd && code[offset+1] == interrupt {
			offsets = append(offsets, offset)
		}
	}
	return offsets
}
