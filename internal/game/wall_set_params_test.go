package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

// TestLoadPiecesWritesWallSetParams 釘住 `37h LOAD PIECES` 的存檔副作用：
// 載得進去的槽記 `{片, 槽}`，運算元是 `0FFh` 的槽記 `{0FFFFh, 0FFFFh}`。
// remake 先前完全沒有寫這三格。
func TestLoadPiecesWritesWallSetParams(t *testing.T) {
	tests := []struct {
		name   string
		pieces [3]uint16
		want   [3]partySave.SAVGAMSetBlock
	}{{
		// `ECL6.DAX/0x40:0045h`：三個槽都有片。
		name: "三個槽都載得進去", pieces: [3]uint16{0x11, 0x12, 0x10},
		want: [3]partySave.SAVGAMSetBlock{
			{BlockID: 0x11, SetID: 1}, {BlockID: 0x12, SetID: 2}, {BlockID: 0x10, SetID: 3}},
	}, {
		// `ECL5.DAX/0x31:001Bh`：只有槽 1 有片。
		name: "0FFh 的槽寫成 0FFFFh", pieces: [3]uint16{0x0C, 0xFF, 0xFF},
		want: [3]partySave.SAVGAMSetBlock{
			{BlockID: 0x0C, SetID: 1}, {BlockID: 0xFFFF, SetID: 0xFFFF}, {BlockID: 0xFFFF, SetID: 0xFFFF}},
	}, {
		// `ECL4.DAX/0x22:003Bh`：只有槽 3 是 0FFh。
		name: "只有最後一槽是 0FFh", pieces: [3]uint16{0x03, 0x08, 0xFF},
		want: [3]partySave.SAVGAMSetBlock{
			{BlockID: 0x03, SetID: 1}, {BlockID: 0x08, SetID: 2}, {BlockID: 0xFFFF, SetID: 0xFFFF}},
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(testCatalog())
			state.applyLoadPieces(ecl.RunResult{LoadPiecesRequested: true, LoadPieces: test.pieces})
			if got := state.WallSetParams(); got != test.want {
				t.Fatalf("牆面參數 ＝ %+v，預期 %+v", got, test.want)
			}
		})
	}
}

// TestLoadPiecesWallSetParamsReachTheSave 釘住它真的走進 SAVGAM 容器——
// 只更新 State 上的欄位而沒有進存檔，玩家看不出差別，但拿原版當 oracle 就對不上。
func TestLoadPiecesWallSetParamsReachTheSave(t *testing.T) {
	state := NewState(testCatalog())
	state.applyLoadPieces(ecl.RunResult{
		LoadPiecesRequested: true, LoadPieces: [3]uint16{0x0E, 0x0F, 0xFF}})
	container, err := state.savgamContainerForSave()
	if err != nil {
		t.Fatal(err)
	}
	want := [3]partySave.SAVGAMSetBlock{
		{BlockID: 0x0E, SetID: 1}, {BlockID: 0x0F, SetID: 2}, {BlockID: 0xFFFF, SetID: 0xFFFF}}
	if container.SetBlocks != want {
		t.Fatalf("存檔裡的牆面參數 ＝ %+v，預期 %+v", container.SetBlocks, want)
	}
}

// TestSaveKeepsImportedWallSetParamsWithoutLoadPieces 釘住另一個方向：
// 這一局沒跑過 `LOAD PIECES` 時，匯入原版存檔帶進來的那三組不能被零蓋掉。
func TestSaveKeepsImportedWallSetParamsWithoutLoadPieces(t *testing.T) {
	state := NewState(testCatalog())
	imported := [3]partySave.SAVGAMSetBlock{
		{BlockID: 5, SetID: 1}, {BlockID: 6, SetID: 2}, {BlockID: 7, SetID: 3}}
	state.savgamPrefix = &partySave.SAVGAMContainer{SetBlocks: imported}
	container, err := state.savgamContainerForSave()
	if err != nil {
		t.Fatal(err)
	}
	if container.SetBlocks != imported {
		t.Fatalf("匯入的牆面參數被蓋掉了：%+v", container.SetBlocks)
	}
}
