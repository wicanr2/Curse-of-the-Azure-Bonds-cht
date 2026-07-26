// Package dax reads the SSI Gold Box block container used by Curse of the
// Azure Bonds. It deliberately stops at the container boundary; ECL payload
// semantics belong to a later, separately specified layer.
package dax

import (
	"encoding/binary"
	"fmt"
)

const headerEntrySize = 9

type Entry struct {
	ID          byte
	Offset      uint32
	DecodedSize uint16
	PackedSize  uint16
}

type Block struct {
	Entry Entry
	Data  []byte
}

func Parse(data []byte) ([]Block, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("DAX is shorter than header")
	}
	headerMinusTwo := int(binary.LittleEndian.Uint16(data[:2]))
	dataOffset := headerMinusTwo + 2
	if dataOffset < 2 || dataOffset > len(data) || (dataOffset-2)%headerEntrySize != 0 {
		return nil, fmt.Errorf("invalid DAX header size %d", headerMinusTwo)
	}
	entries := make([]Entry, 0, (dataOffset-2)/headerEntrySize)
	for pos := 2; pos < dataOffset; pos += headerEntrySize {
		entry := Entry{
			ID:          data[pos],
			Offset:      binary.LittleEndian.Uint32(data[pos+1 : pos+5]),
			DecodedSize: binary.LittleEndian.Uint16(data[pos+5 : pos+7]),
			PackedSize:  binary.LittleEndian.Uint16(data[pos+7 : pos+9]),
		}
		start := dataOffset + int(entry.Offset)
		end := start + int(entry.PackedSize)
		if start < dataOffset || end < start || end > len(data) {
			return nil, fmt.Errorf("block %d exceeds DAX boundary", entry.ID)
		}
		entries = append(entries, entry)
	}
	blocks := make([]Block, 0, len(entries))
	for _, entry := range entries {
		start := dataOffset + int(entry.Offset)
		packed := data[start : start+int(entry.PackedSize)]
		decoded, err := decodeRLE(packed, int(entry.DecodedSize))
		if err != nil {
			return nil, fmt.Errorf("block %d: %w", entry.ID, err)
		}
		blocks = append(blocks, Block{Entry: entry, Data: decoded})
	}
	return blocks, nil
}

func decodeRLE(packed []byte, expected int) ([]byte, error) {
	decoded := make([]byte, 0, expected)
	for pos := 0; pos < len(packed) && len(decoded) < expected; {
		control := int8(packed[pos])
		pos++
		if control >= 0 {
			count := int(control) + 1
			if pos+count > len(packed) {
				return nil, fmt.Errorf("literal run exceeds packed block")
			}
			decoded = append(decoded, packed[pos:pos+count]...)
			pos += count
			continue
		}
		count := -int(control)
		if pos >= len(packed) {
			return nil, fmt.Errorf("repeat run has no value")
		}
		for i := 0; i < count; i++ {
			decoded = append(decoded, packed[pos])
		}
		pos++
	}
	if len(decoded) != expected {
		return nil, fmt.Errorf("decoded %d bytes, expected %d", len(decoded), expected)
	}
	return decoded, nil
}
