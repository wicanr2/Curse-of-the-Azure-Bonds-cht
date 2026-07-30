package main

import (
	"encoding/binary"
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/borlanddebug"
)

func TestSoundFXSelectorTableAndAddressMapping(t *testing.T) {
	t.Parallel()

	table := soundFXSelectorTable()
	if len(table) != 34 {
		t.Fatalf("table bytes=%d, want 34", len(table))
	}
	for index, want := range []int{
		255, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15,
	} {
		raw := int(binary.LittleEndian.Uint16(table[index*2:]))
		if raw != want {
			t.Fatalf("table[%d]=%d, want %d", index, raw, want)
		}
		address := uint16(0x4838 + index*2)
		got, ok := soundFXSelector(address)
		if !ok || got != want {
			t.Fatalf("soundFXSelector(%04X)=(%d,%v), want (%d,true)", address, got, ok, want)
		}
	}
	for _, address := range []uint16{0x4837, 0x4839, 0x485A} {
		if _, ok := soundFXSelector(address); ok {
			t.Fatalf("accepted out-of-table DS:%04X", address)
		}
	}
}

func TestSoundFXNamesUseBorlandModuleAndSymbols(t *testing.T) {
	t.Parallel()

	table := borlanddebug.Table{
		Modules: []borlanddebug.Module{
			{Index: 0, Name: "PROGRAM"},
			{Index: 1, Name: "INTRO"},
			{Index: 2, Name: "INTERPET"},
			{Index: 3, Name: "PROTECT"},
			{Index: 4, Name: "COMBAT"},
			{Index: 5, Name: "COMPTACT"},
			{Index: 6, Name: "COMPREP"},
			{Index: 7, Name: "TREASURE"},
			{Index: 8, Name: "ENCAMP"},
			{Index: 9, Name: "MENU"},
			{Index: 10, Name: "DISPLAY"},
			{Index: 11, Name: "CHARACT"},
			{Index: 12, Name: "MONSTER"},
			{Index: 13, Name: "ITEM"},
			{Index: 14, Name: "COMSTUFF"},
		},
		Symbols: []borlanddebug.Symbol{
			{Name: "LOADCOMSTUFF", Segment: 0x00B8, Offset: 0},
			{Name: "REALMOVE", Segment: 0x00B8, Offset: 0x078E},
			{Name: "CHECKPARTINGBLOWS", Segment: 0x00B8, Offset: 0x0999},
			{Name: "PADFX", Segment: 0x0C29, Offset: 0x484E},
		},
	}
	if got := soundFXModuleName(table, 13); got != "COMSTUFF" {
		t.Fatalf("module=%q", got)
	}
	if got := soundFXFunctionName(table, 13, 0x095D); got != "REALMOVE" {
		t.Fatalf("function=%q", got)
	}
	if got := soundFXSelectorName(table, 0x484E); got != "PADFX" {
		t.Fatalf("selector symbol=%q", got)
	}
}
