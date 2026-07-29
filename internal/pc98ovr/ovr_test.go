package pc98ovr

import (
	"encoding/binary"
	"testing"
)

func TestDecodeChainedControls(t *testing.T) {
	t.Parallel()

	ovr := append([]byte("TPOV"), []byte{0x90, 0xcd, 0xd2, 0xc3, 0x01, 0x02}...)
	ovr = append(ovr, []byte{0xcd, 0x21, 0xc3, 0x03}...)
	exe := make([]byte, 0x90)
	writeControl := func(at int, fileOffset uint32, codeSize, relocationSize, entryCount uint16) {
		exe[at], exe[at+1] = 0xcd, 0x3f
		binary.LittleEndian.PutUint32(exe[at+4:at+8], fileOffset)
		binary.LittleEndian.PutUint16(exe[at+8:at+10], codeSize)
		binary.LittleEndian.PutUint16(exe[at+10:at+12], relocationSize)
		binary.LittleEndian.PutUint16(exe[at+12:at+14], entryCount)
	}
	writeControl(0x10, 4, 4, 2, 3)
	writeControl(0x50, 10, 3, 1, 5)

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
