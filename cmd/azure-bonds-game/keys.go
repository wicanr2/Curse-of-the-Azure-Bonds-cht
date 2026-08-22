package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// keySource 是前端讀鍵盤的唯一出口。
//
// ★ 存在的理由：戰役測試直接呼叫 `state.X()` 推進劇情，**輸入那一層一次都沒被跑到**
// ——那正是「開場到結局」還沒有分母的地方。有了這個接縫，測試才能用**按鍵**驅動
// `(*app).Update()`，走的路和玩家完全一樣。
//
// ⚠ 前端**不要**再直接呼叫 `inpututil.IsKeyJustPressed`／`ebiten.IsKeyPressed`：
// 繞過去的那一行在按鍵驅動的測試裡永遠不會觸發，而且不會有人發現——
// 測試照樣綠，因為它根本走不到那裡。`TestFrontendReadsKeysOnlyThroughTheSeam`
// 會把新的直接呼叫擋下來。
type keySource interface {
	// JustPressed 對應 `inpututil.IsKeyJustPressed`：這一幀剛被按下。
	JustPressed(key ebiten.Key) bool
	// Pressed 對應 `ebiten.IsKeyPressed`：現在是按著的（修飾鍵用）。
	Pressed(key ebiten.Key) bool
}

// ebitenKeys 是正式執行時的來源，直接轉給 ebiten。
type ebitenKeys struct{}

func (ebitenKeys) JustPressed(key ebiten.Key) bool { return inpututil.IsKeyJustPressed(key) }
func (ebitenKeys) Pressed(key ebiten.Key) bool     { return ebiten.IsKeyPressed(key) }

// justPressed／keyDown 是前端內部唯一該用的兩支。`a.keys` 沒設時退回真的 ebiten，
// 這樣既有的建構路徑不必每一條都記得填。
func (a *app) justPressed(key ebiten.Key) bool {
	if a.keys == nil {
		return ebitenKeys{}.JustPressed(key)
	}
	return a.keys.JustPressed(key)
}

func (a *app) keyDown(key ebiten.Key) bool {
	if a.keys == nil {
		return ebitenKeys{}.Pressed(key)
	}
	return a.keys.Pressed(key)
}

// ctrlPressed 是全域音訊開關的修飾鍵。
func (a *app) ctrlPressed() bool {
	return a.keyDown(ebiten.KeyControlLeft) || a.keyDown(ebiten.KeyControlRight)
}

// globalAudioKeys 是原作的全域按鍵處理（PC-98 `sub_18036`）在 remake 這一側的
// 對應：**Ctrl+S 關音效、Ctrl+O 關音樂**，兩個獨立。
//
// 參考卡對前者寫得很明白：`CTRL S : Toggles sound on and off (may be used at
// any time).`——所以它要放在所有模式**前面**，連地圖預覽那種提早 return 的畫面
// 也管得到。後者是 PC-98 版自己加的（`sub_18036` 比 `0Fh` ＝ Ctrl+O，翻
// `MUSICSW` 之後立刻重新派曲）。
//
// ⚠ 回傳 true 代表**這個按鍵被吃掉了**：原作是在讀鍵的地方就攔下來，不會再傳給
// 任何模式。少了這一步，同一幀的 `S` 會被戰鬥選單當成別的指令。
func (a *app) globalAudioKeys() bool {
	if !a.ctrlPressed() {
		return false
	}
	switch {
	case a.justPressed(ebiten.KeyS):
		a.state.ToggleSoundSwitch()
		return true
	case a.justPressed(ebiten.KeyO):
		a.state.ToggleMusicSwitch()
		return true
	}
	return false
}

// combatAltPressed 是戰鬥選單的修飾鍵（Alt ＝ 那一項的「進階」版本）。
func (a *app) combatAltPressed() bool {
	return a.keyDown(ebiten.KeyAltLeft) || a.keyDown(ebiten.KeyAltRight)
}
