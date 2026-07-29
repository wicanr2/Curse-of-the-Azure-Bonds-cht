// Package borlanddebug reads legacy 16-bit Borland debug information appended
// after a DOS MZ load image.
package borlanddebug

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	legacyHeaderSize = 0x30
	legacySymbolSize = 9
	legacyModuleSize = 16
)

// Header is the 0x30-byte debug header used by the Turbo Pascal executable
// examined by this project. Counts are 16-bit in this legacy format.
type Header struct {
	FileOffset       int
	Version          uint16
	NamePoolSize     uint32
	NameCount        uint16
	TypeCount        uint16
	MemberCount      uint16
	SymbolCount      uint16
	GlobalCount      uint16
	ModuleCount      uint16
	LocalCount       uint16
	ScopeCount       uint16
	LineCount        uint16
	SourceCount      uint16
	SegmentCount     uint16
	CorrelationCount uint16
	DebuggerOffset   uint16
	DebuggerSegment  uint16
	ProgramFlags     byte
	DataPoolSize     uint16
	ExtensionSize    uint16
}

// Symbol is one legacy 10-byte symbol record.
type Symbol struct {
	Index     int
	NameIndex uint16
	TypeIndex uint16
	Offset    uint16
	Segment   uint16
	Flags     byte
	Name      string
}

// Module is one legacy 16-byte compiler-unit record.
type Module struct {
	Index            int
	NameIndex        uint16
	Language         byte
	ModelFlags       byte
	SymbolIndex      uint16
	SymbolCount      uint16
	SourceIndex      uint16
	SourceCount      uint16
	CorrelationIndex uint16
	CorrelationCount uint16
	Name             string
}

// Table contains the validated header, symbols, and Pascal-string name pool.
type Table struct {
	Header  Header
	Names   []string
	Symbols []Symbol
	Modules []Module
}

// MZLoadImageSize returns the file boundary declared by an MZ header.
func MZLoadImageSize(executable []byte) (int, error) {
	if len(executable) < 0x1c || string(executable[:2]) != "MZ" {
		return 0, errors.New("not a DOS MZ executable")
	}
	lastPageBytes := int(binary.LittleEndian.Uint16(executable[2:4]))
	pages := int(binary.LittleEndian.Uint16(executable[4:6]))
	if pages == 0 {
		return 0, errors.New("MZ page count is zero")
	}
	size := pages * 512
	if lastPageBytes != 0 {
		size = (pages-1)*512 + lastPageBytes
	}
	if size > len(executable) {
		return 0, fmt.Errorf("declared MZ image %d exceeds file size %d", size, len(executable))
	}
	return size, nil
}

// ParseLegacy parses the legacy 16-bit 0x52FB table. The ASCIIZ name pool is
// required to occupy the final NamePoolSize bytes and contain exactly
// NameCount strings; these checks prevent applying this layout to newer
// formats.
func ParseLegacy(executable []byte) (Table, error) {
	debugOffset, err := MZLoadImageSize(executable)
	if err != nil {
		return Table{}, err
	}
	if debugOffset+legacyHeaderSize > len(executable) ||
		binary.LittleEndian.Uint16(executable[debugOffset:debugOffset+2]) != 0x52fb {
		return Table{}, errors.New("no legacy Borland 0x52FB header at MZ image boundary")
	}
	raw := executable[debugOffset : debugOffset+legacyHeaderSize]
	header := Header{
		FileOffset:       debugOffset,
		Version:          binary.LittleEndian.Uint16(raw[2:4]),
		NamePoolSize:     binary.LittleEndian.Uint32(raw[4:8]),
		NameCount:        binary.LittleEndian.Uint16(raw[8:10]),
		TypeCount:        binary.LittleEndian.Uint16(raw[10:12]),
		MemberCount:      binary.LittleEndian.Uint16(raw[12:14]),
		SymbolCount:      binary.LittleEndian.Uint16(raw[14:16]),
		GlobalCount:      binary.LittleEndian.Uint16(raw[16:18]),
		ModuleCount:      binary.LittleEndian.Uint16(raw[18:20]),
		LocalCount:       binary.LittleEndian.Uint16(raw[20:22]),
		ScopeCount:       binary.LittleEndian.Uint16(raw[22:24]),
		LineCount:        binary.LittleEndian.Uint16(raw[24:26]),
		SourceCount:      binary.LittleEndian.Uint16(raw[26:28]),
		SegmentCount:     binary.LittleEndian.Uint16(raw[28:30]),
		CorrelationCount: binary.LittleEndian.Uint16(raw[30:32]),
		DebuggerOffset:   binary.LittleEndian.Uint16(raw[36:38]),
		DebuggerSegment:  binary.LittleEndian.Uint16(raw[38:40]),
		ProgramFlags:     raw[40],
		DataPoolSize:     binary.LittleEndian.Uint16(raw[43:45]),
		ExtensionSize:    binary.LittleEndian.Uint16(raw[46:48]),
	}
	symbolStart := debugOffset + legacyHeaderSize
	symbolEnd := symbolStart + int(header.SymbolCount)*legacySymbolSize
	moduleEnd := symbolEnd + int(header.ModuleCount)*legacyModuleSize
	nameStart := len(executable) - int(header.NamePoolSize)
	if moduleEnd > nameStart || nameStart < debugOffset+legacyHeaderSize {
		return Table{}, errors.New("legacy Borland table spans overlap or exceed file")
	}
	names, err := parseASCIIZNames(executable[nameStart:], int(header.NameCount))
	if err != nil {
		return Table{}, err
	}
	symbols := make([]Symbol, int(header.SymbolCount))
	for index := range symbols {
		record := executable[symbolStart+index*legacySymbolSize : symbolStart+(index+1)*legacySymbolSize]
		symbol := Symbol{
			Index:     index,
			NameIndex: binary.LittleEndian.Uint16(record[0:2]),
			TypeIndex: binary.LittleEndian.Uint16(record[2:4]),
			Offset:    binary.LittleEndian.Uint16(record[4:6]),
			Segment:   binary.LittleEndian.Uint16(record[6:8]),
			Flags:     record[8],
		}
		if symbol.NameIndex > 0 && int(symbol.NameIndex) <= len(names) {
			symbol.Name = names[symbol.NameIndex-1]
		}
		symbols[index] = symbol
	}
	modules := make([]Module, int(header.ModuleCount))
	for index := range modules {
		record := executable[symbolEnd+index*legacyModuleSize : symbolEnd+(index+1)*legacyModuleSize]
		module := Module{
			Index:            index,
			NameIndex:        binary.LittleEndian.Uint16(record[0:2]),
			Language:         record[2],
			ModelFlags:       record[3],
			SymbolIndex:      binary.LittleEndian.Uint16(record[4:6]),
			SymbolCount:      binary.LittleEndian.Uint16(record[6:8]),
			SourceIndex:      binary.LittleEndian.Uint16(record[8:10]),
			SourceCount:      binary.LittleEndian.Uint16(record[10:12]),
			CorrelationIndex: binary.LittleEndian.Uint16(record[12:14]),
			CorrelationCount: binary.LittleEndian.Uint16(record[14:16]),
		}
		if module.NameIndex > 0 && int(module.NameIndex) <= len(names) {
			module.Name = names[module.NameIndex-1]
		}
		modules[index] = module
	}
	return Table{Header: header, Names: names, Symbols: symbols, Modules: modules}, nil
}

func parseASCIIZNames(pool []byte, count int) ([]string, error) {
	names := make([]string, 0, count)
	offset := 0
	for len(names) < count {
		if offset >= len(pool) {
			return nil, fmt.Errorf("name pool ended after %d of %d names", len(names), count)
		}
		end := offset
		for end < len(pool) && pool[end] != 0 {
			end++
		}
		if end == len(pool) {
			return nil, fmt.Errorf("name %d lacks ASCIIZ terminator", len(names)+1)
		}
		names = append(names, string(pool[offset:end]))
		offset = end + 1
	}
	if offset != len(pool) {
		return nil, fmt.Errorf("name pool has %d trailing bytes", len(pool)-offset)
	}
	return names, nil
}
