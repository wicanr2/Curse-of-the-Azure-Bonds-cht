package game

import (
	"fmt"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/audiomap"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
	partySave "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/save"
)

// MusicEvent 是規則層送給平台 adapter 的音樂意圖。TrackID 來自 game-pack，
// 因此規則與 ECL runtime 不需要知道檔案格式、解碼器或音訊裝置。
type MusicEvent struct {
	Action  string
	TrackID string
}

func (s *State) requestMusicForCurrentBlock(context string) {
	if s.dataPack == nil || s.session == nil {
		return
	}
	// 原作的派曲常式（PC-98 `sub_18AA7`）**第一件事**就是看音樂開關：
	//
	//	cmp MUSICSW, 1      ; 1 ＝ 玩家把音樂關掉了
	//	jnz 照常派曲
	//	mov MUSICNUM, 0FFh  ; 沒有曲子
	//	call 停止            ; sub_18A8E
	//	ret
	//
	// 關掉之後 `MUSICNUM` 是 255，所以再開的時候一定和新曲號不同 → 會重放。
	// 這裡用「把 activeMusicTrackID 清掉」達到同一件事。
	if s.musicSwitchOff {
		s.stopMusic()
		return
	}
	binding, found := s.dataPack.FindMusicBinding(s.session.CurrentBlockID(), context)
	if !found {
		return
	}
	event, changed := musicEventForTrack(s.activeMusicTrackID, binding.TrackID)
	if !changed {
		return
	}
	s.activeMusicTrackID = binding.TrackID
	s.musicPlaybackSnapshot = nil
	s.pendingMusicEvents = append(s.pendingMusicEvents, event)
}

// musicEventForTrack 決定「從這一首換到那一首」要發什麼。
//
// ★ 三條規則都照原作的選曲常式（PC-98 `sub_18A44`）：
//
//	cmp MUSICNUM, 曲號 ; 已經在放同一首就什麼都不做
//	jz  結束
//	mov MUSICNUM, 曲號 ; 換：先停再放
//
// 「同一首不重發」不是最佳化，是原作行為：少了它，每次重新派曲都會把曲子從頭
// 播起，走幾步就聽得出來。
//
// ★ 目標是空字串 ＝ **停**，不是放一首叫「」的曲子。原作寫 `MUSICNUM := 255`
// （沒有曲子），呼叫端是音樂開關（`sub_18AA7` 第一關）。
func musicEventForTrack(previous, next string) (MusicEvent, bool) {
	if next == previous {
		return MusicEvent{}, false
	}
	if next == "" {
		return MusicEvent{Action: "stop"}, true
	}
	return MusicEvent{Action: "play", TrackID: next}, true
}

// stopMusic 停掉正在放的那一首。**本來就沒在放就什麼都不發**——原作在這條路上
// 一律呼叫停止常式，但重複停一次在 remake 這一側只會多一個看不出差別的事件。
func (s *State) stopMusic() {
	event, changed := musicEventForTrack(s.activeMusicTrackID, "")
	if !changed {
		return
	}
	s.activeMusicTrackID = ""
	s.musicPlaybackSnapshot = nil
	s.pendingMusicEvents = append(s.pendingMusicEvents, event)
}

// MusicSwitchOff 回答「玩家有沒有把音樂關掉」（原作的 `MUSICSW`）。
func (s *State) MusicSwitchOff() bool { return s.musicSwitchOff }

// ToggleMusicSwitch 是原作的音樂開關（PC-98 全域按鍵處理 `sub_18036`）：
//
//	cmp [bp+key], 0Fh   ; Ctrl+O
//	jnz ...
//	mov al, MUSICSW
//	xor al, 1           ; 翻轉
//	mov MUSICSW, al
//	call sub_18AA7      ; 立刻重新派曲
//
// ★ 翻完**馬上重新派曲**，不是等下一次換場景：關掉要立刻安靜，打開要立刻放回
// 這個場景該放的那一首。
func (s *State) ToggleMusicSwitch() {
	s.musicSwitchOff = !s.musicSwitchOff
	s.requestMusicForCurrentBlock("")
}

// SoundSwitchOff 回答「玩家有沒有把音效關掉」（原作的 `SOUNDTYPE == 2`）。
func (s *State) SoundSwitchOff() bool { return s.soundSwitchOff }

// ToggleSoundSwitch 是原作的音效開關。參考卡寫得很清楚：
//
//	CTRL S : Toggles sound on and off (may be used at any time).
//
// 原作把目前的音源存進 `OLDSOUND`、`SOUNDTYPE := 2`（靜音），再開的時候若
// `OLDSOUND` 還是合法音源（集合 `{0,1}`）就換回去；而 `SOUNDFX`（`sub_18930`）
// 開頭就是 `cmp SOUNDTYPE, 2 / jz 返回`。
//
// ⚠ **音效與音樂是兩個開關**，互不影響：BGM 走的是 INT 7Eh（`sub_18BDB`），
// 完全不看 `SOUNDTYPE`。把兩者綁在一起會讓 Ctrl+S 連音樂一起關掉——原作不是
// 這樣（spec 1192）。
func (s *State) ToggleSoundSwitch() {
	s.soundSwitchOff = !s.soundSwitchOff
}

func (s *State) requestMusicForSignal(signal string, value uint16) {
	if s.dataPack == nil || s.session == nil {
		return
	}
	cue, found := s.dataPack.FindMusicCue(s.session.CurrentBlockID(), signal, value)
	if !found {
		return
	}
	s.requestMusicForCurrentBlock(cue.Context)
}

// requestCombatMusic 是**開戰換曲**。
//
// ★ 原作在這裡不看場景：`INITCOMBAT`（COMPREP，overlay-10）把曲號**直接推給
// `MSCPLAY`**，完全不碰 `MUSICNO`（spec 1192）：
//
//	cmp byte [LOADMONNUM], 47h
//	jnz  →  mov al, 07h   ; 戰鬥
//	        mov al, 0Bh   ; 地城二
//	push ax / call MSCPLAY
//
// ⚠ 所以戰鬥曲**不是**由 ECL 段決定的，是由**載入了哪一組怪物**決定的——
// 這也是為什麼只掃 `mov byte [MUSICNO], imm` 的時候這兩首會落在外面。
//
// remake 這一側用 `context` 表達「戰鬥中」。那個 `47h` 的分岔要靠
// `monster_set` cue，而**共用 engine 目前只收 `picture` 這一種 signal**
// （`music_cues[i] has unsupported signal "monster_set"`）⇒ 分岔還表達不出來，
// 所有戰鬥都放「戰鬥」那一首。
//
// ⚠ 這個限制是**真的**，和先前那個「停音樂被 engine 擋住」不一樣：那次是把停止
// 誤當成劇情資料，原作根本沒有那種資料；這次原作明明白白寫著
// `cmp byte [LOADMONNUM], 47h`。`TestEnginePackCannotExpressMonsterSetCueYet`
// 把它釘住——鬆綁時會紅，那時候補上 cue 就好。
func (s *State) requestCombatMusic(monsterID uint8) {
	if s.dataPack == nil || s.session == nil {
		return
	}
	context := combatMusicContext
	if cue, found := s.dataPack.FindMusicCue(
		s.session.CurrentBlockID(), "monster_set", uint16(monsterID)); found {
		context = cue.Context
	}
	s.requestMusicForCurrentBlock(context)
}

// 事件驅動的換曲點各自的 context。原作那幾處都**不看 `CURRENTECL`**（各自把曲號
// 推給 `MSCPLAY` 或直接寫 `MUSICNO` 之後派曲），所以 pack 那一側是每一段都列。
const (
	// combatMusicContext ＝ 開戰（`INITCOMBAT`，COMPREP）。
	combatMusicContext = audiomap.CombatContext
	// titleMusicContext ＝ 開場（`DOINTRO`，overlay-01 `093Ch`，曲目 1 標題）。
	titleMusicContext = audiomap.TitleContext
	// creationMusicContext ＝ 角色建立（`GEN`，overlay-17 `0B08h`，曲目 2）。
	creationMusicContext = audiomap.CreationContext
	// endingMusicContext ＝ 結局過場（PC-98 overlay-18 `168Dh`，曲目 10 結局）。
	//
	// ★ 原作寫在結局文字的**正上方**：`mov byte [MUSICNO], 0Ah` 之後立刻
	// `push` 同一格再 `call far 893h:114h`（`MSCPLAY`），才開始逐頁印那四段
	// 結局文字（每段之後 `call far 418h:0E6Ah` 等鍵）。所以換曲點是「結局開始
	// 演」那一刻，不是打完最終戰、也不是回主選單。
	endingMusicContext = audiomap.EndingContext
	// partyWipeMusicContext ＝ 全滅（POSTCOM，PC-98 overlay-05 `1955h`，曲目 2）。
	//
	// ★ 原作的判準是 `PARTYDEAD && !ADUEL`（`7F34h`／`0BDE8h`，符號表直接讀出）：
	// `PARTYDEAD` 由 POSTCOM 走一次角色名冊算出來——**每一位**的 `CHARSTATUS`
	// （`+196h`）都落在 `{ANIMATED, TEMPGONE, DEAD, STONED, GONE}` 才算全滅，
	// 也就是還有人 `UNCONC`／`DYING` 就**不算**（spec 427 的 ordinal）。
	// 決鬥（`ADUEL`）輸掉不算全滅。
	//
	// ⚠ remake 這一側用 `combat.StatusEnemyWon` 當那一刻，**判準不完全相同**：
	// 那是戰鬥層的「敵方獲勝」，不是逐位查 `CHARSTATUS`。決鬥在 remake 走不到
	// （`GODUEL` 的兩個 `2Dh CALL` 選擇子 corpus 都沒用到，spec 1150），所以
	// `ADUEL` 那一半不影響結果。
	partyWipeMusicContext = audiomap.PartyWipeContext
)

// requestEndingMusic 是結局過場那一首。與開戰、開場、角色建立同一類：原作
// **不看 `CURRENTECL`**，所以 pack 那一側 25 段全列。
func (s *State) requestEndingMusic() {
	s.requestMusicForCurrentBlock(endingMusicContext)
}

// requestPartyWipeMusic 是全滅畫面那一首。原作在印
// 「モンスターはパーティーを全滅させ、喜んでいる。」與等鍵提示**之後**才換曲，
// 而換完曲才把 `TTY := 1` 切回文字模式。
func (s *State) requestPartyWipeMusic() {
	s.requestMusicForCurrentBlock(partyWipeMusicContext)
}

// RequestTitleMusic 是開場那一首。由前端在**顯示標題畫面時**呼叫。
//
// ⚠ 不能在 `NewState` 就發：建構還不是「畫面上出現標題」，而且會讓所有只是造一個
// State 的測試都收到音樂事件（`TestActiveECLBlockRequestsPC98MusicSelector` 正是
// 釘住「建構不發音樂」）。
func (s *State) RequestTitleMusic() {
	s.requestMusicForCurrentBlock(titleMusicContext)
}

// restoreSceneMusic 是**戰鬥結束回到場景曲**。
//
// ★ 為什麼需要它：原作的派曲常式（`sub_18AA7`）由主迴圈在場景變動時呼叫，
// 戰鬥結束回到地城時它會依 `CURRENTECL` 重算 `MUSICNO`，於是自然換回場景曲。
// remake 這一側只有「換段」會觸發派曲，而戰鬥前後**段沒有變**——少了這一支，
// 戰鬥曲會一直放下去，而且不會有任何錯誤。
func (s *State) restoreSceneMusic() {
	s.requestMusicForCurrentBlock("")
}

func (s *State) requestMusicIfBlockChanged(previous uint8) {
	if s.session != nil && s.session.CurrentBlockID() != previous {
		// 換段的瓶頸就在這裡，所以接縫盤點也搭這班車（`block_edges.go`）。
		recordBlockEdge(previous, s.session.CurrentBlockID())
		s.captureArrival(s.session.CurrentBlockID())
		s.requestMusicForCurrentBlock("")
	}
}

// ConsumeMusicEvents 只轉交一次尚未處理的音樂意圖。平台 adapter 可在曲目素材
// 尚未解碼時忽略事件，但 deterministic runtime 仍能驗證原版選曲行為。
func (s *State) ConsumeMusicEvents() []MusicEvent {
	events := append([]MusicEvent(nil), s.pendingMusicEvents...)
	s.pendingMusicEvents = nil
	return events
}

// SetMusicPlaybackSnapshot attaches the platform player's exact synthesis
// continuation immediately before a remake JSON save.
func (s *State) SetMusicPlaybackSnapshot(snapshot *pc98music.TrackPCMStreamSnapshot) error {
	if snapshot != nil && s.activeMusicTrackID == "" {
		return fmt.Errorf("music playback snapshot has no active track ID")
	}
	if snapshot == nil {
		s.musicPlaybackSnapshot = nil
		return nil
	}
	copy := cloneTrackPCMStreamSnapshot(*snapshot)
	s.musicPlaybackSnapshot = &copy
	return nil
}

func cloneTrackPCMStreamSnapshot(snapshot pc98music.TrackPCMStreamSnapshot) pc98music.TrackPCMStreamSnapshot {
	snapshot.SynthState = append([]byte(nil), snapshot.SynthState...)
	snapshot.Pending = append([]byte(nil), snapshot.Pending...)
	for index := range snapshot.Playback.Channels {
		machine := &snapshot.Playback.Channels[index].Machine
		machine.CallStack = append([]int(nil), machine.CallStack...)
		machine.LoopStack = append([]pc98music.LoopSnapshot(nil), machine.LoopStack...)
	}
	return snapshot
}

func (s *State) musicSnapshot() *partySave.MusicSnapshot {
	if s.activeMusicTrackID == "" {
		return nil
	}
	result := &partySave.MusicSnapshot{TrackID: s.activeMusicTrackID}
	if s.musicPlaybackSnapshot != nil {
		copy := cloneTrackPCMStreamSnapshot(*s.musicPlaybackSnapshot)
		result.Stream = &copy
	}
	return result
}

// MusicPlaybackSnapshot returns the loaded continuation for the platform
// adapter. A nil stream means the stable track restarts normally when a local
// driver is available.
func (s *State) MusicPlaybackSnapshot() (string, *pc98music.TrackPCMStreamSnapshot, bool) {
	if s.activeMusicTrackID == "" {
		return "", nil, false
	}
	if s.musicPlaybackSnapshot == nil {
		return s.activeMusicTrackID, nil, true
	}
	copy := cloneTrackPCMStreamSnapshot(*s.musicPlaybackSnapshot)
	return s.activeMusicTrackID, &copy, true
}
