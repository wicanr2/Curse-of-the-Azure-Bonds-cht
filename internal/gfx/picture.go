// Package gfx decodes the indexed-picture payload used by SSI DAX image
// blocks. It is deliberately renderer-neutral: Ebiten can turn Pixels into
// RGBA later, while tests and tools can inspect the original 4-bit values.
package gfx

import "fmt"

type Picture struct {
	WidthUnits  uint16
	HeightUnits uint16
	X           int16
	Y           int16
	ItemCount   uint8
	Metadata    [8]byte
	Pixels      []uint8
}

// MergePictures overlays source onto destination using the masked-icon rule
// used by the Gold Box combat renderer. Transparent pixels (index 16) reveal
// the other layer; overlapping opaque pixels are combined with bitwise OR.
// The source is top-left aligned, matching the original MergeIcon routine.
func MergePictures(destination, source Picture) (Picture, error) {
	if destination.ItemCount == 0 || source.ItemCount == 0 {
		return Picture{}, fmt.Errorf("cannot merge empty picture items")
	}
	if destination.Width() != source.Width() || source.Height() > destination.Height() {
		return Picture{}, fmt.Errorf("cannot merge %dx%d source onto %dx%d destination", source.Width(), source.Height(), destination.Width(), destination.Height())
	}
	result := destination
	result.Pixels = append([]uint8(nil), destination.Pixels...)
	for item := 0; item < int(source.ItemCount); item++ {
		if item >= int(destination.ItemCount) {
			break
		}
		for y := 0; y < source.Height(); y++ {
			for x := 0; x < source.Width(); x++ {
				sourceValue := source.Pixels[item*source.ItemSize()+y*source.Width()+x]
				destinationIndex := item*destination.ItemSize() + y*destination.Width() + x
				destinationValue := result.Pixels[destinationIndex]
				switch {
				case destinationValue == 16:
					result.Pixels[destinationIndex] = sourceValue
				case sourceValue == 16:
					// Keep the destination pixel.
				default:
					result.Pixels[destinationIndex] = destinationValue | sourceValue
				}
			}
		}
	}
	return result, nil
}

func (p Picture) Width() int  { return int(p.WidthUnits) * 8 }
func (p Picture) Height() int { return int(p.HeightUnits) }

func (p Picture) ItemSize() int { return p.Width() * p.Height() }

func (p Picture) Pixel(item, x, y int) (uint8, bool) {
	if item < 0 || item >= int(p.ItemCount) || x < 0 || x >= p.Width() || y < 0 || y >= p.Height() {
		return 0, false
	}
	return p.Pixels[item*p.ItemSize()+y*p.Width()+x], true
}

// ParsePicture decodes the 17-byte SSI picture header followed by packed
// nibbles. Each source byte contains two indexed pixels, high nibble first.
// When masked is true, maskColor is exposed as palette index 16, matching the
// reference engine's transparent sentinel.
func ParsePicture(data []byte, masked bool, maskColor uint8) (Picture, error) {
	if len(data) < 17 {
		return Picture{}, fmt.Errorf("picture is %d bytes, shorter than header", len(data))
	}
	picture := Picture{
		WidthUnits:  uint16(data[2]) | uint16(data[3])<<8,
		HeightUnits: uint16(data[0]) | uint16(data[1])<<8,
		X:           int16(uint16(data[4]) | uint16(data[5])<<8),
		Y:           int16(uint16(data[6]) | uint16(data[7])<<8),
		ItemCount:   data[8],
	}
	copy(picture.Metadata[:], data[9:17])
	if picture.WidthUnits == 0 || picture.HeightUnits == 0 || picture.ItemCount == 0 {
		return Picture{}, fmt.Errorf("picture has invalid dimensions %dx%d items=%d", picture.WidthUnits, picture.HeightUnits, picture.ItemCount)
	}
	pixelsPerItem := picture.Width() * picture.Height()
	packedSize := picture.ItemSize() / 2
	expected := 17 + int(picture.ItemCount)*packedSize
	if len(data) != expected {
		return Picture{}, fmt.Errorf("picture payload is %d bytes, want %d for %dx%d items=%d", len(data), expected, picture.Width(), picture.Height(), picture.ItemCount)
	}
	picture.Pixels = make([]uint8, int(picture.ItemCount)*pixelsPerItem)
	pos := 17
	for item := 0; item < int(picture.ItemCount); item++ {
		base := item * pixelsPerItem
		for pixel := 0; pixel < pixelsPerItem; pixel += 2 {
			packed := data[pos]
			pos++
			first, second := packed>>4, packed&0x0F
			if masked && first == maskColor {
				first = 16
			}
			if masked && second == maskColor {
				second = 16
			}
			picture.Pixels[base+pixel] = first
			picture.Pixels[base+pixel+1] = second
		}
	}
	return picture, nil
}

type WallDef struct {
	Rows [5][156]uint8
}

func (w WallDef) ID(row, column int) (uint8, bool) {
	if row < 0 || row >= len(w.Rows) || column < 0 || column >= len(w.Rows[row]) {
		return 0, false
	}
	return w.Rows[row][column], true
}

// ParseWallDef decodes one 5x156 WALLDEF block.
func ParseWallDef(data []byte) (WallDef, error) {
	const size = 5 * 156
	if len(data) != size {
		return WallDef{}, fmt.Errorf("wall definition is %d bytes, want %d", len(data), size)
	}
	var result WallDef
	for row := range result.Rows {
		copy(result.Rows[row][:], data[row*156:(row+1)*156])
	}
	return result, nil
}

// ParseWallDefs decodes one or more concatenated 5x156 wall definition
// records. Some DAX blocks contain two records, as in the original loader's
// blockCount calculation.
func ParseWallDefs(data []byte) ([]WallDef, error) {
	const size = 5 * 156
	if len(data) == 0 || len(data)%size != 0 {
		return nil, fmt.Errorf("wall definitions are %d bytes, not a multiple of %d", len(data), size)
	}
	result := make([]WallDef, 0, len(data)/size)
	for offset := 0; offset < len(data); offset += size {
		wall, err := ParseWallDef(data[offset : offset+size])
		if err != nil {
			return nil, err
		}
		result = append(result, wall)
	}
	return result, nil
}
