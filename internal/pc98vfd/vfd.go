// Package pc98vfd reads the VFD1.00 container used by the supplied PC-98
// reference disks. It is intentionally read-only: original media is evidence.
package pc98vfd

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	signatureSize       = 8
	descriptorOffset    = 0xDC
	descriptorSize      = 12
	descriptorsPerTrack = 26
)

var signature = [signatureSize]byte{'V', 'F', 'D', '1', '.', '0', '0', 0}

// Sector describes one VFD track-directory entry.
type Sector struct {
	Cylinder   uint8
	Head       uint8
	Number     uint8
	SizeCode   uint8
	Flags      [4]byte
	DataOffset uint32
	Data       []byte
}

// Missing reports whether the VFD directory records a sector without payload.
func (s Sector) Missing() bool { return s.DataOffset == ^uint32(0) }

// Size returns the payload size encoded by the NEC sector size code.
func (s Sector) Size() int {
	if s.SizeCode > 7 {
		return 0
	}
	return 128 << s.SizeCode
}

// Image is a parsed, read-only VFD disk image.
type Image struct {
	Sectors []Sector
}

// Parse reads the track directory for a fixed PC-98 geometry. Geometry is
// supplied explicitly because VFD1.00 reserves 26 descriptors per track and
// does not carry a trustworthy global sector count.
func Parse(data []byte, cylinders, heads, sectorsPerTrack int) (*Image, error) {
	if len(data) < descriptorOffset {
		return nil, errors.New("VFD image is shorter than its header")
	}
	if string(data[:signatureSize]) != string(signature[:]) {
		return nil, fmt.Errorf("unsupported VFD signature %q", data[:signatureSize])
	}
	if cylinders <= 0 || heads <= 0 || sectorsPerTrack <= 0 {
		return nil, errors.New("geometry values must be positive")
	}
	if sectorsPerTrack > descriptorsPerTrack {
		return nil, fmt.Errorf(
			"%d sectors per track exceeds VFD directory capacity %d",
			sectorsPerTrack,
			descriptorsPerTrack,
		)
	}

	count := cylinders * heads * sectorsPerTrack
	result := &Image{Sectors: make([]Sector, 0, count)}
	for cylinder := 0; cylinder < cylinders; cylinder++ {
		for head := 0; head < heads; head++ {
			track := cylinder*heads + head
			for sectorIndex := 0; sectorIndex < sectorsPerTrack; sectorIndex++ {
				index := track*descriptorsPerTrack + sectorIndex
				offset := descriptorOffset + index*descriptorSize
				if offset+descriptorSize > len(data) {
					return nil, fmt.Errorf("VFD descriptor %d lies outside image", index)
				}
				entry := data[offset : offset+descriptorSize]
				sector := Sector{
					Cylinder:   entry[0],
					Head:       entry[1],
					Number:     entry[2],
					SizeCode:   entry[3],
					DataOffset: binary.LittleEndian.Uint32(entry[8:12]),
				}
				copy(sector.Flags[:], entry[4:8])
				if int(sector.Cylinder) != cylinder ||
					int(sector.Head) != head ||
					int(sector.Number) != sectorIndex+1 {
					return nil, fmt.Errorf(
						"descriptor %d has CHR %d/%d/%d, want %d/%d/%d",
						index,
						sector.Cylinder,
						sector.Head,
						sector.Number,
						cylinder,
						head,
						sectorIndex+1,
					)
				}
				size := sector.Size()
				if size == 0 {
					return nil, fmt.Errorf(
						"sector %d/%d/%d has invalid size code %d",
						cylinder,
						head,
						sectorIndex+1,
						sector.SizeCode,
					)
				}
				if !sector.Missing() {
					end := uint64(sector.DataOffset) + uint64(size)
					if end > uint64(len(data)) {
						return nil, fmt.Errorf(
							"sector %d/%d/%d payload exceeds image",
							cylinder,
							head,
							sectorIndex+1,
						)
					}
					sector.Data = data[sector.DataOffset:end]
				}
				result.Sectors = append(result.Sectors, sector)
			}
		}
	}
	return result, nil
}

// MissingSectors returns all descriptors whose VFD payload offset is
// 0xFFFFFFFF. Their CHRN identity is retained as evidence.
func (i *Image) MissingSectors() []Sector {
	if i == nil {
		return nil
	}
	var missing []Sector
	for _, sector := range i.Sectors {
		if sector.Missing() {
			missing = append(missing, sector)
		}
	}
	return missing
}
