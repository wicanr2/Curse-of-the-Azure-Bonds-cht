package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

// eclCodeState 起一個帶 session 的 State，位元組碼是可辨識的樣式。
func eclCodeState(t *testing.T) (*State, []byte) {
	t.Helper()
	payload := make([]byte, 2+64)
	for index := range payload[2:] {
		payload[index+2] = uint8(index%251) + 1
	}
	session, err := ecl.NewBlockSession(map[uint8][]byte{0x07: payload}, 0x07)
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(testCatalog())
	state.session = session
	return &state, payload
}

// ★★ 區 3 是 **ECL 位元組碼本身**。先前匯出時只是把匯入的那一份原封抄過去
// ——新開的遊戲沒有匯入來源，於是匯出的存檔那一塊**整塊是 0**（spec 1176）。
func TestSAVGAMCarriesTheLiveCodeWindow(t *testing.T) {
	state, payload := eclCodeState(t)
	container, err := state.savgamContainerForSave()
	if err != nil {
		t.Fatal(err)
	}
	if len(container.ECL) != partySave.SAVGAMECLMemorySize {
		t.Fatalf("第四塊 %d bytes，want %d", len(container.ECL), partySave.SAVGAMECLMemorySize)
	}
	for index, want := range payload[2:] {
		if container.ECL[index] != want {
			t.Fatalf("位址 %04Xh ＝ %02X，want %02X",
				partySave.ECLCodeLow+index, container.ECL[index], want)
		}
	}
	nonZero := 0
	for _, value := range container.ECL {
		if value != 0 {
			nonZero++
		}
	}
	if nonZero == 0 {
		t.Fatal("整塊是 0——那正是先前的行為")
	}
}

// ⚠ 位元組碼裡的 `0` 是真的指令（`00h` ＝ `EXIT`），所以這一塊**不能跳過 0**
// ——前三塊的 word 編碼刻意跳過 0，照抄會把程式打洞（spec 1176）。
func TestECLCodeCodecKeepsZeroBytes(t *testing.T) {
	record := make([]byte, partySave.SAVGAMECLMemorySize)
	record[0] = 0x11
	record[2] = 0x22
	code, err := partySave.DecodeECLCode(record)
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != partySave.SAVGAMECLMemorySize {
		t.Fatalf("解出來 %d 格，want %d（0 也要收）",
			len(code), partySave.SAVGAMECLMemorySize)
	}
	if value, ok := code[partySave.ECLCodeLow+1]; !ok || value != 0 {
		t.Fatalf("位址 %04Xh 的 0 被跳過了", partySave.ECLCodeLow+1)
	}
	round, err := partySave.EncodeECLCode(code, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(round) != string(record) {
		t.Fatal("往返之後位元組不一樣")
	}
}

// 匯入時要把存檔那一份換回去——腳本改過的位元組才回得來。
func TestSAVGAMCodeWindowIsRestoredOnImport(t *testing.T) {
	state, payload := eclCodeState(t)
	container := partySave.SAVGAMContainer{
		ECL: make([]byte, partySave.SAVGAMECLMemorySize),
	}
	copy(container.ECL, payload[2:])
	// 存檔裡的第 5 個位元組與遊戲檔不一樣——那就是「腳本改過」的形狀。
	container.ECL[4] = 0xAB
	if err := state.seedECLBanksFrom(container); err != nil {
		t.Fatal(err)
	}
	got, ok := state.session.MemoryValue(partySave.ECLCodeLow + 4)
	if !ok || got != 0xAB {
		t.Fatalf("位址 %04Xh ＝ %02X（ok=%v），want AB",
			partySave.ECLCodeLow+4, got, ok)
	}
}
