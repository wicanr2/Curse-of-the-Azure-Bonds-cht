package monster

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	ItemRecordSize   = 0x3F
	AffectRecordSize = 9
)

type ItemRecord struct {
	Name            string
	Type            uint8
	NameNumbers     [3]uint8
	Plus            int
	PlusSave        uint8
	Readied         bool
	HiddenNameFlags uint8
	Cursed          bool
	Weight          int16
	Count           uint8
	Value           int16
	Affects         [3]uint8
}

type AffectRecord struct {
	Kind     uint8
	Value    uint16
	Duration uint8
	Active   bool
	Data     [4]byte
}

// ChineseName covers item types observed in the real MON1 encounter record.
// The complete item-name table remains data work; unknown types deliberately
// return a stable diagnostic instead of a fabricated translation.
func ChineseName(item ItemRecord) string {
	base, ok := map[uint8]string{
		28: "弩矢", // Quarrel
		35: "闊劍", // Broad Sword
		46: "輕弩", // Light Crossbow
		55: "鏈甲", // Chain Mail
		59: "盾牌", // Shield
	}[item.Type]
	if !ok {
		return fmt.Sprintf("未翻譯物品(0x%02X)", item.Type)
	}
	if item.Plus > 0 {
		base = fmt.Sprintf("+%d %s", item.Plus, base)
	}
	if item.Cursed {
		base += "（詛咒）"
	}
	if item.Count > 1 && item.Type == 28 {
		base = fmt.Sprintf("%s ×%d", base, item.Count)
	}
	return base
}

func ChineseAffectName(affect AffectRecord) string {
	if name, ok := map[uint8]string{
		0x18: "偵測隱形",
		0x5A: "酸液吐息",
	}[affect.Kind]; ok {
		return name
	}
	return fmt.Sprintf("未翻譯效果(0x%02X)", affect.Kind)
}

func ParseItems(data []byte) ([]ItemRecord, error) {
	if len(data)%ItemRecordSize != 0 {
		return nil, fmt.Errorf("item data is %d bytes, not a multiple of %d", len(data), ItemRecordSize)
	}
	items := make([]ItemRecord, 0, len(data)/ItemRecordSize)
	for offset := 0; offset < len(data); offset += ItemRecordSize {
		items = append(items, parseItem(data[offset:offset+ItemRecordSize]))
	}
	return items, nil
}

func parseItem(data []byte) ItemRecord {
	name := strings.TrimRight(string(data[:0x2A]), "\x00 ")
	return ItemRecord{
		Name: name, Type: data[0x2E],
		NameNumbers: [3]uint8{data[0x2F], data[0x30], data[0x31]},
		Plus:        int(int8(data[0x32])), PlusSave: data[0x33],
		Readied: data[0x34] != 0, HiddenNameFlags: data[0x35], Cursed: data[0x36] != 0,
		Weight: int16(binary.LittleEndian.Uint16(data[0x37:0x39])), Count: data[0x39],
		Value:   int16(binary.LittleEndian.Uint16(data[0x3A:0x3C])),
		Affects: [3]uint8{data[0x3C], data[0x3D], data[0x3E]},
	}
}

func ParseAffects(data []byte) ([]AffectRecord, error) {
	if len(data)%AffectRecordSize != 0 {
		return nil, fmt.Errorf("affect data is %d bytes, not a multiple of %d", len(data), AffectRecordSize)
	}
	affects := make([]AffectRecord, 0, len(data)/AffectRecordSize)
	for offset := 0; offset < len(data); offset += AffectRecordSize {
		record := AffectRecord{
			Kind: data[offset], Value: binary.LittleEndian.Uint16(data[offset+1 : offset+3]),
			Duration: data[offset+3], Active: data[offset+4] != 0,
		}
		copy(record.Data[:], data[offset+5:offset+9])
		affects = append(affects, record)
	}
	return affects, nil
}
