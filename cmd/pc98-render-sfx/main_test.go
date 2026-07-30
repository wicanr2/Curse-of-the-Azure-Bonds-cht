package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteWAVUsesStereoSigned16Format(t *testing.T) {
	path := filepath.Join(t.TempDir(), "effect.wav")
	pcm := stereoPCM([]int16{-1, 0x1234})
	if err := writeWAV(path, pcm); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		t.Fatalf("invalid WAV header: %q %q", data[:4], data[8:12])
	}
	if got := binary.LittleEndian.Uint16(data[22:24]); got != 2 {
		t.Fatalf("channels=%d, want 2", got)
	}
	if got := binary.LittleEndian.Uint32(data[40:44]); got != uint32(len(pcm)) {
		t.Fatalf("data length=%d, want %d", got, len(pcm))
	}
}
