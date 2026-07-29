package borlanddebug

import (
	"encoding/binary"
	"testing"
)

func TestParseLegacy(t *testing.T) {
	const imageSize = 0x40
	names := []byte{'F', 'O', 'O', 0, 'B', 'A', 'R', 0}
	executable := make([]byte, imageSize+legacyHeaderSize+legacySymbolSize+legacyModuleSize+len(names))
	copy(executable, "MZ")
	binary.LittleEndian.PutUint16(executable[2:4], imageSize)
	binary.LittleEndian.PutUint16(executable[4:6], 1)
	header := executable[imageSize : imageSize+legacyHeaderSize]
	binary.LittleEndian.PutUint16(header[0:2], 0x52fb)
	binary.LittleEndian.PutUint16(header[2:4], 0x0208)
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(names)))
	binary.LittleEndian.PutUint16(header[8:10], 2)
	binary.LittleEndian.PutUint16(header[14:16], 1)
	binary.LittleEndian.PutUint16(header[18:20], 1)
	record := executable[imageSize+legacyHeaderSize : imageSize+legacyHeaderSize+legacySymbolSize]
	binary.LittleEndian.PutUint16(record[0:2], 2)
	binary.LittleEndian.PutUint16(record[2:4], 7)
	binary.LittleEndian.PutUint16(record[4:6], 0x1234)
	binary.LittleEndian.PutUint16(record[6:8], 0x5678)
	record[8] = 3
	module := executable[imageSize+legacyHeaderSize+legacySymbolSize : imageSize+legacyHeaderSize+legacySymbolSize+legacyModuleSize]
	binary.LittleEndian.PutUint16(module[0:2], 1)
	module[2], module[3] = 2, 4
	copy(executable[len(executable)-len(names):], names)

	table, err := ParseLegacy(executable)
	if err != nil {
		t.Fatal(err)
	}
	if table.Header.Version != 0x0208 || len(table.Names) != 2 || len(table.Symbols) != 1 || len(table.Modules) != 1 {
		t.Fatalf("unexpected table: %+v", table)
	}
	got := table.Symbols[0]
	if got.Name != "BAR" || got.Offset != 0x1234 || got.Segment != 0x5678 {
		t.Fatalf("unexpected symbol: %+v", got)
	}
	if got := table.Modules[0]; got.Name != "FOO" || got.Language != 2 || got.ModelFlags != 4 {
		t.Fatalf("unexpected module: %+v", got)
	}
}

func TestRejectsTrailingNameBytes(t *testing.T) {
	executable := make([]byte, 0x40+legacyHeaderSize+2)
	copy(executable, "MZ")
	binary.LittleEndian.PutUint16(executable[2:4], 0x40)
	binary.LittleEndian.PutUint16(executable[4:6], 1)
	header := executable[0x40 : 0x40+legacyHeaderSize]
	binary.LittleEndian.PutUint16(header[0:2], 0x52fb)
	binary.LittleEndian.PutUint32(header[4:8], 2)
	executable[len(executable)-2] = 0
	if _, err := ParseLegacy(executable); err == nil {
		t.Fatal("expected invalid trailing name bytes")
	}
}
