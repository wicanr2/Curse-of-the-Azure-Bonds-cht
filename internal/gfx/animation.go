package gfx

import (
	"encoding/binary"
	"fmt"
	"time"
)

// AnimationFrame is one frame from the custom PIC/SPRIT animation stream.
// Delay is the original tenth-of-a-second timing value and Picture contains
// the decoded indexed pixels plus its original placement metadata.
type AnimationFrame struct {
	Delay   uint32
	Picture Picture
}

type Animation struct {
	Frames []AnimationFrame
}

// AnimationFrameIndex selects a frame using the original delay unit of one
// tenth of a second. Zero-delay frames are given one tick so malformed or
// placeholder timing cannot create a zero-duration loop.
func AnimationFrameIndex(delays []uint32, elapsed time.Duration) int {
	if len(delays) == 0 {
		return -1
	}
	total := time.Duration(0)
	for _, delay := range delays {
		if delay == 0 {
			delay = 1
		}
		total += time.Duration(delay) * 100 * time.Millisecond
	}
	position := elapsed % total
	for index, delay := range delays {
		if delay == 0 {
			delay = 1
		}
		span := time.Duration(delay) * 100 * time.Millisecond
		if position < span {
			return index
		}
		position -= span
	}
	return len(delays) - 1
}

// ParseAnimation decodes the frame stream used by SPRIT*.DAX and by the
// reference load_pic_final routine. Unlike a normal DAX picture block, the
// stream starts with a frame count and each frame has its own dimensions.
func ParseAnimation(data []byte, masked bool, maskColor uint8) (Animation, error) {
	if len(data) < 1 {
		return Animation{}, fmt.Errorf("animation is empty")
	}
	count := int(data[0])
	if count == 0 {
		return Animation{}, fmt.Errorf("animation has zero frames")
	}
	if count > 64 {
		return Animation{}, fmt.Errorf("animation has unreasonable frame count %d", count)
	}
	frames := make([]AnimationFrame, 0, count)
	offset := 1
	for index := 0; index < count; index++ {
		const headerSize = 21 // delay(4), dimensions/position(8), reserved(1), metadata(8)
		if offset+headerSize > len(data) {
			return Animation{}, fmt.Errorf("frame %d header is truncated at %d", index, offset)
		}
		delay := binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		height := int16(binary.LittleEndian.Uint16(data[offset:]))
		offset += 2
		widthUnits := int16(binary.LittleEndian.Uint16(data[offset:]))
		offset += 2
		x := int16(binary.LittleEndian.Uint16(data[offset:]))
		offset += 2
		y := int16(binary.LittleEndian.Uint16(data[offset:]))
		offset += 2
		offset++ // reference skips one byte after y_pos
		var metadata [8]byte
		copy(metadata[:], data[offset:offset+8])
		offset += 8
		if height <= 0 || widthUnits <= 0 {
			return Animation{}, fmt.Errorf("frame %d has invalid dimensions %dx%d", index, widthUnits, height)
		}
		// Width is stored in 8-pixel units. Two indexed pixels occupy one
		// source byte, hence widthUnits * height * 4 bytes.
		packedSize := int(height) * int(widthUnits) * 4
		if packedSize <= 0 || offset+packedSize > len(data) {
			return Animation{}, fmt.Errorf("frame %d pixels are truncated: need %d at %d", index, packedSize, offset)
		}
		pixels := make([]uint8, int(height)*int(widthUnits)*8)
		for pixel := 0; pixel < len(pixels); pixel += 2 {
			packed := data[offset+pixel/2]
			first, second := packed>>4, packed&0x0F
			if masked && first == maskColor {
				first = 16
			}
			if masked && second == maskColor {
				second = 16
			}
			pixels[pixel], pixels[pixel+1] = first, second
		}
		offset += packedSize
		frames = append(frames, AnimationFrame{Delay: delay, Picture: Picture{
			WidthUnits: uint16(widthUnits), HeightUnits: uint16(height), X: x, Y: y,
			ItemCount: 1, Metadata: metadata, Pixels: pixels,
		}})
	}
	if offset != len(data) {
		return Animation{}, fmt.Errorf("animation has %d trailing bytes", len(data)-offset)
	}
	return Animation{Frames: frames}, nil
}
