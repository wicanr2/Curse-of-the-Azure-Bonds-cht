package save

import (
	"encoding/binary"
	"fmt"
)

// SAVGAM 的前四塊就是 ECL 位址空間的前四區（spec 1163）。大小、順序與寬度
// 全部對得上：
//
//	Area1   0x800 = 1,024 word  ->  區 0  4B00h..4EFFh
//	Area2   0x800 = 1,024 word  ->  區 1  7C00h..7FFFh
//	Runtime 0x400 =   512 word  ->  區 2  7A00h..7BFFh
//	ECL    0x1E00 = 7,680 byte  ->  區 3  8000h..9DFFh（ECL 程式碼本身）
//
// bank 內位移是 `(位址 − 基底) × 2`（spec 1096），所以 `internal/area` 那幾個
// 具名位移其實就是 ECL 位址：`0x1E0`／`0x1E2`／`0x1E4` 是 `4BF0h`／`4BF1h`／
// `4BF2h`（移動前座標快照與上一段 ECL 編號），`0x624` 是 `7F12h`。
const (
	ECLBank0Low  = 0x4B00
	ECLBank0High = 0x4EFF
	ECLBank1Low  = 0x7C00
	ECLBank1High = 0x7FFF
	ECLBank2Low  = 0x7A00
	ECLBank2High = 0x7BFF
)

// DecodeECLBank 把一塊 SAVGAM 記錄讀成 ECL 位址 → 值。**值是 0 的格子不收**：
// 原作分不出「寫過 0」與「沒寫過」，收進來只會讓 `MemoryValue` 的第二個回傳值
// 失去意義。
func DecodeECLBank(record []byte, low, high uint16) (map[uint16]uint16, error) {
	want := (int(high) - int(low) + 1) * 2
	if len(record) != want {
		return nil, fmt.Errorf("ECL bank %04X..%04X record has %d bytes, want %d", low, high, len(record), want)
	}
	memory := make(map[uint16]uint16)
	for address := int(low); address <= int(high); address++ {
		offset := (address - int(low)) * 2
		value := binary.LittleEndian.Uint16(record[offset:])
		if value == 0 {
			continue
		}
		memory[uint16(address)] = value
	}
	return memory, nil
}

// EncodeECLBank 把 session 記憶體寫回一塊 SAVGAM 記錄。**只動有鍵的位址**——
// 沒被 VM 碰過的格子維持 original 帶進來的位元組，remake 還沒解讀的欄位才不會
// 被零蓋掉。original 是 nil 就給一塊全零的。
func EncodeECLBank(memory map[uint16]uint16, low, high uint16, original []byte) ([]byte, error) {
	size := (int(high) - int(low) + 1) * 2
	out := make([]byte, size)
	if original != nil {
		if len(original) != size {
			return nil, fmt.Errorf("ECL bank %04X..%04X record has %d bytes, want %d", low, high, len(original), size)
		}
		copy(out, original)
	}
	for address, value := range memory {
		if address < low || address > high {
			continue
		}
		binary.LittleEndian.PutUint16(out[(int(address)-int(low))*2:], value)
	}
	return out, nil
}
