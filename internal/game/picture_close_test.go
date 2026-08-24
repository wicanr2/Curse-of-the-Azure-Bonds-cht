package game

import (
	"testing"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/ecl"
)

// `0Eh PICTURE 0FFh` 的關閉訊號要真的把畫面上的圖收掉（spec 1148）。
// 先前 VM 發了 `PictureCloseRequested`、game 這一側沒有任何消費端——
// 腳本關圖之後，remake 的圖停在畫面上直到玩家自己按下一頁。
func TestApplyPictureCloseClearsTheOpenPicture(t *testing.T) {
	state := NewState(testCatalog())
	state.PictureRequested = true
	state.PictureBlock = 5
	state.BigPictureRequested = true
	state.SceneCharacterRequested = true
	state.SceneBodyBlock = 5

	state.applyPictureClose(ecl.RunResult{PictureCloseRequested: true})

	if state.PictureRequested || state.PictureBlock != 0 || state.BigPictureRequested ||
		state.SceneCharacterRequested || state.SceneBodyBlock != 0 {
		t.Errorf("關閉訊號沒有清掉畫面上的圖：requested=%v block=%d big=%v char=%v body=%d",
			state.PictureRequested, state.PictureBlock, state.BigPictureRequested,
			state.SceneCharacterRequested, state.SceneBodyBlock)
	}
}

// 同一次執行「先關後開」時收尾狀態是開的——關閉不可以把後來開的圖清掉。
// （VM 與 session 聚合層已把這種情況收成 close=false，這裡守 helper 自己的閘。）
func TestApplyPictureCloseYieldsToAReopenedPicture(t *testing.T) {
	state := NewState(testCatalog())
	state.PictureRequested = true
	state.PictureBlock = 7

	state.applyPictureClose(ecl.RunResult{
		PictureCloseRequested: true,
		PictureRequested:      true,
		PictureBlock:          9,
	})

	if !state.PictureRequested || state.PictureBlock != 7 {
		t.Errorf("close+open 的結果不該動既有狀態（開圖由開圖那一段接手）：requested=%v block=%d",
			state.PictureRequested, state.PictureBlock)
	}
}
