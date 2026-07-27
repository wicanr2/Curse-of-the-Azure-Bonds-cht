package ecl

import (
	"reflect"
	"testing"
)

func TestEntryPointsReadsWordCommandSets(t *testing.T) {
	block := []byte{0x88, 0x13,
		0xFF, 0x01, 0x34, 0x80,
		0xFF, 0x02, 0x78, 0x80,
		0xFF, 0x03, 0xBC, 0x80,
		0xFF, 0x01, 0xDE, 0x80,
		0xFF, 0x01, 0xF0, 0x80,
	}
	got, next, err := EntryPoints(block, 5)
	if err != nil {
		t.Fatal(err)
	}
	want := []uint16{0x8034, 0x8078, 0x80BC, 0x80DE, 0x80F0}
	if !reflect.DeepEqual(got, want) || next != 20 {
		t.Fatalf("EntryPoints() = %#v, next %d; want %#v, next 20", got, next, want)
	}
}

func TestEntryPointsRejectsNonWordHeader(t *testing.T) {
	_, _, err := EntryPoints([]byte{0, 0, 0, 0x80, 0}, 1)
	if err == nil {
		t.Fatal("EntryPoints() accepted a non-word command-set header")
	}
}

func TestSmokeInitializationEntriesKeepsPerEntryResults(t *testing.T) {
	block := []byte{0, 0}
	for index := 0; index < 5; index++ {
		block = append(block, 0xFF, 0x01, byte(0x14+index), 0x80)
	}
	block = append(block, 0, 0, 0, 0, 0)
	reports, err := SmokeInitializationEntries(block, 5, 4, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 5 {
		t.Fatalf("reports=%d, want 5", len(reports))
	}
	for index, report := range reports {
		if report.Index != index || report.Start != 0x14+index || report.Err != nil || report.Result.Steps != 1 {
			t.Fatalf("report[%d]=%+v", index, report)
		}
	}
}
