package pc98ovr

import (
	"encoding/binary"
	"reflect"
	"testing"
)

func TestDecodeChainedControls(t *testing.T) {
	t.Parallel()

	ovr := append([]byte("TPOV"), []byte{0x90, 0xcd, 0xd2, 0xc3, 0x01, 0x00}...)
	ovr = append(ovr, []byte{0xcd, 0x21, 0xc3, 0x00, 0x00}...)
	exe := make([]byte, 0x90)
	writeControl := func(at int, fileOffset uint32, codeSize, relocationSize, entryCount uint16) {
		exe[at], exe[at+1] = 0xcd, 0x3f
		binary.LittleEndian.PutUint32(exe[at+4:at+8], fileOffset)
		binary.LittleEndian.PutUint16(exe[at+8:at+10], codeSize)
		binary.LittleEndian.PutUint16(exe[at+10:at+12], relocationSize)
		binary.LittleEndian.PutUint16(exe[at+12:at+14], entryCount)
		for index := 0; index < int(entryCount); index++ {
			entry := at + entryTableOffset + index*entryRecordSize
			exe[entry], exe[entry+1] = 0xcd, 0x3f
			codeOffset := uint16(index)
			if index >= int(codeSize) {
				codeOffset = 0xffff
			}
			binary.LittleEndian.PutUint16(exe[entry+2:entry+4], codeOffset)
		}
	}
	writeControl(0x10, 4, 4, 2, 3)
	writeControl(0x50, 10, 3, 2, 5)

	decoded, err := Decode(exe, ovr)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 2 {
		t.Fatalf("got %d overlays, want 2", len(decoded))
	}
	if got := InterruptOffsets(decoded[0].Code, 0xd2); len(got) != 1 || got[0] != 1 {
		t.Fatalf("INT D2 offsets = %v, want [1]", got)
	}
	if got := InterruptOffsets(decoded[0].Relocation, 0xd2); len(got) != 0 {
		t.Fatalf("relocation bytes leaked into code scan: %v", got)
	}
	if !reflect.DeepEqual(decoded[0].RelocationOffsets, []uint16{1}) {
		t.Fatalf("relocation offsets=%v, want [1]", decoded[0].RelocationOffsets)
	}
	entry, ok := decoded[0].ResolveStub(entryTableOffset + 2*entryRecordSize)
	if !ok || entry.Index != 2 || entry.CodeOffset != 2 {
		t.Fatalf("resolved entry=%+v ok=%v", entry, ok)
	}
	if _, ok := decoded[0].ResolveStub(entryTableOffset + 1); ok {
		t.Fatal("accepted a pointer between resident stubs")
	}
	if got := decoded[0].ResolveCode(2); len(got) != 1 || got[0] != entry {
		t.Fatalf("reverse code resolution=%+v, want %+v", got, entry)
	}
	if got := decoded[0].ResolveCode(0x1234); len(got) != 0 {
		t.Fatalf("unknown code offset resolved to %+v", got)
	}
}

func TestDecodeRejectsMalformedEntryAndRelocation(t *testing.T) {
	t.Parallel()

	build := func() ([]byte, []byte) {
		exe := make([]byte, 0x50)
		exe[0x10], exe[0x11] = 0xcd, 0x3f
		binary.LittleEndian.PutUint32(exe[0x14:0x18], 4)
		binary.LittleEndian.PutUint16(exe[0x18:0x1a], 4)
		binary.LittleEndian.PutUint16(exe[0x1a:0x1c], 2)
		binary.LittleEndian.PutUint16(exe[0x1c:0x1e], 1)
		exe[0x30], exe[0x31] = 0xcd, 0x3f
		binary.LittleEndian.PutUint16(exe[0x32:0x34], 1)
		return exe, append([]byte("TPOV"), []byte{0x90, 0x90, 0x90, 0xc3, 0x01, 0x00}...)
	}

	tests := []struct {
		name   string
		mutate func([]byte, []byte)
	}{
		{"entry signature", func(exe, _ []byte) { exe[0x30] = 0x90 }},
		{"entry code offset", func(exe, _ []byte) { binary.LittleEndian.PutUint16(exe[0x32:0x34], 4) }},
		{"fixup outside code", func(_, ovr []byte) { binary.LittleEndian.PutUint16(ovr[8:10], 3) }},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			exe, ovr := build()
			test.mutate(exe, ovr)
			if _, err := Decode(exe, ovr); err == nil {
				t.Fatal("accepted malformed overlay metadata")
			}
		})
	}
}

func TestParseControlsRejectsUnchainedCoincidence(t *testing.T) {
	t.Parallel()

	ovr := append([]byte("TPOV"), make([]byte, 32)...)
	exe := make([]byte, 0x50)
	exe[0], exe[1] = 0xcd, 0x3f
	binary.LittleEndian.PutUint32(exe[4:8], 12)
	binary.LittleEndian.PutUint16(exe[8:10], 4)

	if _, err := ParseControls(exe, ovr); err == nil {
		t.Fatal("accepted a control that does not begin the TPOV chain")
	}
}

func TestWordOffsets(t *testing.T) {
	t.Parallel()

	got := WordOffsets([]byte{0x30, 0x7c, 0x90, 0x30, 0x7c}, 0x7c30)
	if !reflect.DeepEqual(got, []int{0, 3}) {
		t.Fatalf("WordOffsets = %v, want [0 3]", got)
	}
}

func TestPatternOffsetsIncludesOverlaps(t *testing.T) {
	t.Parallel()

	got := PatternOffsets([]byte{0xaa, 0xaa, 0xaa}, []byte{0xaa, 0xaa})
	if !reflect.DeepEqual(got, []int{0, 1}) {
		t.Fatalf("PatternOffsets = %v, want [0 1]", got)
	}
}

func TestFarCallWordArgumentsRequiresAdjacentDSPush(t *testing.T) {
	t.Parallel()

	code := []byte{
		0xFF, 0x36, 0x48, 0x48,
		0x9A, 0x00, 0x00, 0x93, 0x08,
		0x90,
		0x9A, 0x00, 0x00, 0x93, 0x08,
	}
	got := FarCallWordArguments(code, 0x0000, 0x0893)
	if len(got) != 1 {
		t.Fatalf("calls=%+v", got)
	}
	if got[0].CallOffset != 4 || got[0].ArgumentAddress != 0x4848 {
		t.Fatalf("call=%+v", got[0])
	}
}
