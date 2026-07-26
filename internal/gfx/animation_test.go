package gfx

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestParseAnimationFramesAndMask(t *testing.T) {
	data := []byte{2}
	appendFrame := func(delay uint32, height, width, x, y int16, packed []byte) {
		header := make([]byte, 21)
		binary.LittleEndian.PutUint32(header[0:], delay)
		binary.LittleEndian.PutUint16(header[4:], uint16(height))
		binary.LittleEndian.PutUint16(header[6:], uint16(width))
		binary.LittleEndian.PutUint16(header[8:], uint16(x))
		binary.LittleEndian.PutUint16(header[10:], uint16(y))
		header[12] = 1
		copy(header[13:], []byte("metadata"))
		data = append(data, header...)
		data = append(data, packed...)
	}
	// 1 row × 1 width-unit is four packed bytes (8 pixels).
	appendFrame(7, 1, 1, -2, 3, []byte{0x01, 0x23, 0x45, 0x67})
	appendFrame(9, 1, 1, 4, 5, []byte{0x80, 0x9A, 0xBC, 0xDE})
	animation, err := ParseAnimation(data, true, 0)
	if err != nil || len(animation.Frames) != 2 {
		t.Fatalf("animation=%#v err=%v", animation, err)
	}
	first := animation.Frames[0]
	if first.Delay != 7 || first.Picture.Width() != 8 || first.Picture.X != -2 || first.Picture.Y != 3 || first.Picture.Metadata[0] != 'm' {
		t.Fatalf("first frame=%#v", first)
	}
	if pixel, _ := first.Picture.Pixel(0, 0, 0); pixel != 16 {
		t.Fatalf("masked pixel=%d, want transparent sentinel 16", pixel)
	}
	if second := animation.Frames[1]; second.Delay != 9 || second.Picture.X != 4 {
		t.Fatalf("second frame=%#v", second)
	}
}

func TestAnimationFrameIndexUsesTenthSecondDelay(t *testing.T) {
	delays := []uint32{1, 3}
	if got := AnimationFrameIndex(delays, 50*time.Millisecond); got != 0 {
		t.Fatalf("frame at 50ms=%d, want 0", got)
	}
	if got := AnimationFrameIndex(delays, 150*time.Millisecond); got != 1 {
		t.Fatalf("frame at 150ms=%d, want 1", got)
	}
	if got := AnimationFrameIndex(delays, 450*time.Millisecond); got != 0 {
		t.Fatalf("frame at 450ms=%d, want 0 after loop", got)
	}
}

func TestParseAnimationRejectsTruncatedFrame(t *testing.T) {
	if _, err := ParseAnimation([]byte{1, 0, 0, 0}, false, 0); err == nil {
		t.Fatal("truncated animation accepted")
	}
}
