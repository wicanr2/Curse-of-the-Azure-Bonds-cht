package area

import (
	"encoding/binary"
	"fmt"
)

// SnapshotSize is the size of both original Area1 and Area2 records.
const SnapshotSize = 0x800

const (
	area1MapBlock   = 0x18A
	area1Dungeon    = 0x1CC
	area1LastX      = 0x1E0
	area1LastY      = 0x1E2
	area1LastECL    = 0x1E4
	area1OutdoorSky = 0x1FA
	area1IndoorSky  = 0x1FC
	area1City       = 0x342
	area1GameTime   = 0x18C
	area2GameArea   = 0x624
	area2HeadBlock  = 0x5C2
)

func checkedRecord(data []byte, name string) ([]byte, error) {
	if len(data) != SnapshotSize {
		return nil, fmt.Errorf("%s record has size %d, want %#x", name, len(data), SnapshotSize)
	}
	return data, nil
}

// DecodeArea1 reads the fields currently understood by the remake. Unknown
// bytes remain available to callers that retain the original record.
func DecodeArea1(data []byte) (State, error) {
	data, err := checkedRecord(data, "Area1")
	if err != nil {
		return State{}, err
	}
	state := State{
		Current3DMapBlockID: data[area1MapBlock],
		InDungeon:           int16(binary.LittleEndian.Uint16(data[area1Dungeon:])) != 0,
		LastXPos:            int16(binary.LittleEndian.Uint16(data[area1LastX:])),
		LastYPos:            int16(binary.LittleEndian.Uint16(data[area1LastY:])),
		LastECLBlockID:      binary.LittleEndian.Uint16(data[area1LastECL:]),
		CurrentCity:         data[area1City],
		OutdoorSkyColor:     binary.LittleEndian.Uint16(data[area1OutdoorSky:]),
		IndoorSkyColor:      binary.LittleEndian.Uint16(data[area1IndoorSky:]),
	}
	for index := range state.GameTime {
		offset := area1GameTime + index*2
		state.GameTime[index] = binary.LittleEndian.Uint16(data[offset:])
	}
	return state, nil
}

// EncodeArea1 updates known fields in a copy of original. Passing nil creates
// a zero-filled record; all bytes not covered by the known layout are kept.
func EncodeArea1(state State, original []byte) ([]byte, error) {
	var out []byte
	if original == nil {
		out = make([]byte, SnapshotSize)
	} else {
		if _, err := checkedRecord(original, "Area1"); err != nil {
			return nil, err
		}
		out = append([]byte(nil), original...)
	}
	out[area1MapBlock] = state.Current3DMapBlockID
	if state.InDungeon {
		binary.LittleEndian.PutUint16(out[area1Dungeon:], 1)
	} else {
		binary.LittleEndian.PutUint16(out[area1Dungeon:], 0)
	}
	binary.LittleEndian.PutUint16(out[area1LastX:], uint16(state.LastXPos))
	binary.LittleEndian.PutUint16(out[area1LastY:], uint16(state.LastYPos))
	binary.LittleEndian.PutUint16(out[area1LastECL:], state.LastECLBlockID)
	binary.LittleEndian.PutUint16(out[area1OutdoorSky:], state.OutdoorSkyColor)
	binary.LittleEndian.PutUint16(out[area1IndoorSky:], state.IndoorSkyColor)
	out[area1City] = state.CurrentCity
	for index, value := range state.GameTime {
		offset := area1GameTime + index*2
		binary.LittleEndian.PutUint16(out[offset:], value)
	}
	return out, nil
}

// DecodeArea2 reads the global game-area selector from an original Area2
// record. Other Area2 fields are intentionally preserved but not interpreted.
func DecodeArea2(data []byte) (State, error) {
	data, err := checkedRecord(data, "Area2")
	if err != nil {
		return State{}, err
	}
	return State{GameArea: data[area2GameArea], HeadBlockID: data[area2HeadBlock]}, nil
}

// EncodeArea2 updates only the understood game-area byte and preserves the
// remaining original Area2 record verbatim.
func EncodeArea2(state State, original []byte) ([]byte, error) {
	if original == nil {
		original = make([]byte, SnapshotSize)
	} else if _, err := checkedRecord(original, "Area2"); err != nil {
		return nil, err
	} else {
		original = append([]byte(nil), original...)
	}
	original[area2GameArea] = state.GameArea
	original[area2HeadBlock] = state.HeadBlockID
	return original, nil
}
