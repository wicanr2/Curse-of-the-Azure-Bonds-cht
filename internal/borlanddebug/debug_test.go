package borlanddebug

import (
	"encoding/binary"
	"testing"
)

func TestParseLegacy(t *testing.T) {
	const imageSize = 0x40
	names := []byte{'F', 'O', 'O', 0, 'B', 'A', 'R', 0}
	executable := make([]byte, imageSize+legacyHeaderSize+legacySymbolSize+legacyModuleSize+legacyTypeSize+len(names))
	copy(executable, "MZ")
	binary.LittleEndian.PutUint16(executable[2:4], imageSize)
	binary.LittleEndian.PutUint16(executable[4:6], 1)
	header := executable[imageSize : imageSize+legacyHeaderSize]
	binary.LittleEndian.PutUint16(header[0:2], 0x52fb)
	binary.LittleEndian.PutUint16(header[2:4], 0x0208)
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(names)))
	binary.LittleEndian.PutUint16(header[8:10], 2)
	binary.LittleEndian.PutUint16(header[10:12], 1)
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
	typeRecord := executable[imageSize+legacyHeaderSize+legacySymbolSize+legacyModuleSize : imageSize+legacyHeaderSize+legacySymbolSize+legacyModuleSize+legacyTypeSize]
	typeRecord[0] = 0x16
	binary.LittleEndian.PutUint16(typeRecord[1:3], 1)
	binary.LittleEndian.PutUint16(typeRecord[3:5], 4)
	binary.LittleEndian.PutUint16(typeRecord[6:8], 7)
	copy(executable[len(executable)-len(names):], names)

	table, err := ParseLegacy(executable)
	if err != nil {
		t.Fatal(err)
	}
	if table.Header.Version != 0x0208 || len(table.Names) != 2 || len(table.Symbols) != 1 ||
		len(table.Modules) != 1 || len(table.Types) != 1 {
		t.Fatalf("unexpected table: %+v", table)
	}
	got := table.Symbols[0]
	if got.Name != "BAR" || got.Offset != 0x1234 || got.Segment != 0x5678 {
		t.Fatalf("unexpected symbol: %+v", got)
	}
	if got := table.Modules[0]; got.Name != "FOO" || got.Language != 2 || got.ModelFlags != 4 {
		t.Fatalf("unexpected module: %+v", got)
	}
	if got := table.Types[0]; got.Index != 1 || got.ID != 0x16 ||
		got.Name != "FOO" || got.Size != 4 || got.Detail != [3]byte{0, 7, 0} {
		t.Fatalf("unexpected type: %+v", got)
	}
	if table.MemberTableSize != 0 || table.DataPoolOffset != len(executable)-len(names) {
		t.Fatalf("unexpected table boundaries: %+v", table)
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

// RecordLayout 的連法：record 型別記錄的 Detail 後兩個位元組是**第一個成員的
// 1-based 索引**，成員依序排列、沒有對齊填塞，加到記錄大小為止（spec 1164）。
func TestRecordLayoutFollowsTheMemberChain(t *testing.T) {
	table := Table{
		Types: []Type{
			{Index: 1, ID: 8, Size: 1},                // byte
			{Index: 2, ID: 9, Size: 2},                // word
			{Index: 3, ID: 28, Size: 4, Name: "QUAD"}, // array
			{Index: 4, ID: RecordTypeID, Size: 7, Name: "REC", Detail: [3]byte{0, 2, 0}},
		},
		Members: []Member{
			{Index: 0, Name: "BEFORE", TypeIndex: 1},
			{Index: 1, Name: "FIRST", TypeIndex: 1},
			{Index: 2, Name: "SECOND", TypeIndex: 2},
			{Index: 3, Name: "THIRD", TypeIndex: 3},
			{Index: 4, Name: "AFTER", TypeIndex: 1},
		},
	}
	fields, err := table.RecordLayout("REC")
	if err != nil {
		t.Fatal(err)
	}
	want := []RecordField{
		{Offset: 0, Size: 1, Name: "FIRST", TypeID: 8, TypeIndex: 1},
		{Offset: 1, Size: 2, Name: "SECOND", TypeID: 9, TypeIndex: 2},
		{Offset: 3, Size: 4, Name: "THIRD", TypeName: "QUAD", TypeID: 28, TypeIndex: 3},
	}
	if len(fields) != len(want) {
		t.Fatalf("fields=%+v", fields)
	}
	for index := range want {
		if fields[index] != want[index] {
			t.Fatalf("field %d = %+v, want %+v", index, fields[index], want[index])
		}
	}

	// ⚠ 成員數不在型別記錄裡，是「加到記錄大小為止」——所以**記錄大小變大會
	// 悄悄多吃一個成員**，只有跨過邊界才擋得住。這條測試釘的是後者。
	table.Types[3].Size = 6
	if _, err := table.RecordLayout("REC"); err == nil {
		t.Fatal("欄位跨過記錄邊界時應該報錯")
	}
	table.Types[3].Size = 7
	table.Types[3].Detail = [3]byte{0, 0, 0}
	if _, err := table.RecordLayout("REC"); err == nil {
		t.Fatal("第一個成員索引是 0 時應該報錯")
	}
}
