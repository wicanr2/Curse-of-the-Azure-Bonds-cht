// Package pc98ovr reads the Turbo Pascal overlay container used by the
// PC-9801 version of Curse of the Azure Bonds.
package pc98ovr

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

const (
	magic             = "TPOV"
	controlRecordSize = 0x30
	entryTableOffset  = 0x20
	entryRecordSize   = 5
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
	Code              []byte
	Relocation        []byte
	Entries           []Entry
	RelocationOffsets []uint16
}

// Entry describes one resident five-byte overlay dispatch stub. StubOffset is
// relative to the control segment; CodeOffset is relative to the overlay code
// span. The on-disk form is CD 3F, code-offset word, flags byte.
type Entry struct {
	Index            int
	ExecutableOffset int
	StubOffset       uint16
	CodeOffset       uint16
	Flags            uint8
}

// ResolveStub maps a resident far-pointer offset to its overlay entry. The
// caller must separately prove that the pointer segment is this control's
// segment; matching a numeric offset alone is insufficient.
func (o Overlay) ResolveStub(stubOffset uint16) (Entry, bool) {
	if stubOffset < entryTableOffset {
		return Entry{}, false
	}
	delta := int(stubOffset) - entryTableOffset
	if delta%entryRecordSize != 0 {
		return Entry{}, false
	}
	index := delta / entryRecordSize
	if index < 0 || index >= len(o.Entries) {
		return Entry{}, false
	}
	return o.Entries[index], true
}

// ResolveCode returns every resident entry that dispatches to a handler-local
// code offset. Multiple stubs may legitimately share one handler, so reverse
// lookup returns a slice and preserves each original stub and entry index.
func (o Overlay) ResolveCode(codeOffset uint16) []Entry {
	entries := make([]Entry, 0, 1)
	for _, entry := range o.Entries {
		if entry.CodeOffset == codeOffset {
			entries = append(entries, entry)
		}
	}
	return entries
}

// FarCallWordArgument identifies the Turbo Pascal sequence
// PUSH word ptr DS:[address]; CALL FAR segment:offset.
type FarCallWordArgument struct {
	CallOffset      int
	ArgumentAddress uint16
	TargetOffset    uint16
	TargetSegment   uint16
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
		entries, err := parseEntries(executable, control)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.CodeOffset != 0xFFFF && int(entry.CodeOffset) >= int(control.CodeSize) {
				return nil, fmt.Errorf(
					"overlay at 0x%X entry %d code offset 0x%X exceeds code size 0x%X",
					control.FileOffset, entry.Index, entry.CodeOffset, control.CodeSize,
				)
			}
		}
		relocation := overlayFile[codeEnd:relocationEnd]
		relocationOffsets, err := parseRelocationOffsets(relocation, int(control.CodeSize))
		if err != nil {
			return nil, fmt.Errorf("overlay at 0x%X relocation: %w", control.FileOffset, err)
		}
		decoded = append(decoded, Overlay{
			Control:           control,
			Code:              overlayFile[codeStart:codeEnd],
			Relocation:        relocation,
			Entries:           entries,
			RelocationOffsets: relocationOffsets,
		})
	}
	return decoded, nil
}

func parseEntries(executable []byte, control Control) ([]Entry, error) {
	start := control.ExecutableOffset + entryTableOffset
	end := start + int(control.EntryCount)*entryRecordSize
	if start < 0 || end > len(executable) {
		return nil, fmt.Errorf("overlay control at 0x%X entry table exceeds executable", control.ExecutableOffset)
	}
	entries := make([]Entry, 0, control.EntryCount)
	for index := 0; index < int(control.EntryCount); index++ {
		offset := start + index*entryRecordSize
		if executable[offset] != 0xCD || executable[offset+1] != 0x3F {
			return nil, fmt.Errorf("overlay control at 0x%X entry %d lacks CD 3F stub", control.ExecutableOffset, index)
		}
		entries = append(entries, Entry{
			Index:            index,
			ExecutableOffset: offset,
			StubOffset:       uint16(entryTableOffset + index*entryRecordSize),
			CodeOffset:       binary.LittleEndian.Uint16(executable[offset+2 : offset+4]),
			Flags:            executable[offset+4],
		})
	}
	return entries, nil
}

func parseRelocationOffsets(relocation []byte, codeSize int) ([]uint16, error) {
	if len(relocation)%2 != 0 {
		return nil, fmt.Errorf("odd relocation byte count %d", len(relocation))
	}
	offsets := make([]uint16, 0, len(relocation)/2)
	previous := -1
	for offset := 0; offset < len(relocation); offset += 2 {
		value := binary.LittleEndian.Uint16(relocation[offset : offset+2])
		if int(value)+2 > codeSize {
			return nil, fmt.Errorf("fixup 0x%X exceeds code size 0x%X", value, codeSize)
		}
		if int(value) <= previous {
			return nil, fmt.Errorf("fixup 0x%X is not strictly increasing", value)
		}
		offsets = append(offsets, value)
		previous = int(value)
	}
	return offsets, nil
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

// WordOffsets returns offsets of a literal little-endian 16-bit value in code.
// A match is only a candidate operand; callers must disassemble the surrounding
// instruction before assigning semantics.
func WordOffsets(code []byte, value uint16) []int {
	low := byte(value)
	high := byte(value >> 8)
	var offsets []int
	for offset := 0; offset+1 < len(code); offset++ {
		if code[offset] == low && code[offset+1] == high {
			offsets = append(offsets, offset)
		}
	}
	return offsets
}

// PatternOffsets returns every overlapping occurrence of pattern in code.
func PatternOffsets(code, pattern []byte) []int {
	if len(pattern) == 0 {
		return nil
	}
	var offsets []int
	for start := 0; start+len(pattern) <= len(code); {
		next := bytes.Index(code[start:], pattern)
		if next < 0 {
			break
		}
		offset := start + next
		offsets = append(offsets, offset)
		start = offset + 1
	}
	return offsets
}

// FarCallWordArguments returns exact direct far calls whose single argument
// is loaded from a resident DS word immediately before the call.
func FarCallWordArguments(
	code []byte, targetOffset, targetSegment uint16,
) []FarCallWordArgument {
	var calls []FarCallWordArgument
	for callOffset := 4; callOffset+5 <= len(code); callOffset++ {
		if code[callOffset] != 0x9A ||
			binary.LittleEndian.Uint16(code[callOffset+1:callOffset+3]) != targetOffset ||
			binary.LittleEndian.Uint16(code[callOffset+3:callOffset+5]) != targetSegment {
			continue
		}
		pushOffset := callOffset - 4
		if code[pushOffset] != 0xFF || code[pushOffset+1] != 0x36 {
			continue
		}
		calls = append(calls, FarCallWordArgument{
			CallOffset:      callOffset,
			ArgumentAddress: binary.LittleEndian.Uint16(code[pushOffset+2 : callOffset]),
			TargetOffset:    targetOffset,
			TargetSegment:   targetSegment,
		})
	}
	return calls
}
