# 1192 — 原機音訊的播放生命週期：從分母到覆蓋率

狀態：`READY`（動作分類、玩家開關、選擇子常數的身分都已釘住並接上）

## 問題

`cmd/dseg-writers` 已經盤出「誰決定什麼時候該響」——六個音訊位址的 17 處寫入。
但那是**分母不是覆蓋率**：它說原作在幾個地方改音訊狀態，沒說 remake 有沒有
對應的動作。

## 三格狀態、三個常數

掃的六個位址不是同一種東西，而**分錯類會讓報表反過來說話**：

| 位址 | 符號 | 是什麼 | 種類 |
|---|---|---|---|
| `8BE1h` | `MUSICNUM` | 驅動程式手上的曲目編號（`255` ＝ 沒有曲子）| 狀態 |
| `8BE3h` | `MUSICSW` | 音樂開關（`1` ＝ 玩家關掉了）| 狀態 |
| `8BF3h` | `MUSICNO` | 這個場景該放第幾首 | 狀態 |
| `4838h` | `SOUNDHALT` | `SOUNDFX` 的選擇子 `255` | **常數** |
| `483Ah` | `SOUNDOFF` | `SOUNDFX` 的選擇子 `0` | **常數** |
| `483Ch` | `SOUNDON` | `SOUNDFX` 的選擇子 `1` | **常數** |

後三個和 `CASTFX`…`CRASHFX` 排在資料段同一張表裡，值在檔案裡就寫死：

```
4838h  ff 00  00 00  01 00  02 00  03 00  04 00 …
       255    0      1      2      3      4
       HALT   OFF    ON     CAST   MISS   SPELLHIT …
```

所以它們的**寫入次數永遠是零**，有意義的是**讀取**——讀它們的地方就是 `SOUNDFX`
的呼叫點。選擇子的身分與解碼歸 `internal/pc98sfx`。

> ⚠ **三個相鄰符號同時掃出零，是模型錯了的訊號，不是結論。** 這一支本來把它們
> 當狀態格掃，掃出 0 處寫入，再照「原作沒寫到就不算待辦」的規則印成 ✅ ——
> 三個假零被當成三項做完的工作。

## 停止不是劇情資料

派曲常式是 PC-98 `sub_18AA7`（linear `18AA7h`）：

```
cmp MUSICSW, 1          ; 玩家關掉音樂了嗎
jnz 照常派曲
mov MUSICNUM, 0FFh      ; 沒有曲子
call 停止驅動            ; sub_18A8E
ret
照常派曲:
cmp WHEREAMI, 5／1／7   ; 這三種狀態不換曲
jz  ret
al := CURRENTECL         ; ← 場景碼，就是 ECL 段
  1／31h                     -> MUSICNO := 3
  11h 12h 21h 22h 23h 15h 43h 45h -> MUSICNO := 4
  50h／51h                    -> WLDTWN 非零 ? 6 : 5
  20h／23h／40h／42h          -> MUSICNO := 8   ⚠ 23h 上面就被吃掉了，這裡到不了
  2／10h／5／35h              -> MUSICNO := 9
  3／4／25h／32h／33h         -> MUSICNO := 0Ch
  30h                         -> 不改 MUSICNO，但仍重新派一次
  其餘                        -> **ret，音樂繼續放**
selectTrack(MUSICNO)     ; sub_18A44
```

★ **default 是 `ret` 不是停止。** 原作沒有「這一段不放音樂」這種資料——查不到
就維持現況。真正會停音樂的**只有一處**：玩家把音樂關掉。

`CURRENTECL` 就是 ECL 段編號，所以 game-pack 用 `ecl_blocks` 當 binding 的鍵
是對的；`WLDTWN` 對應 `context: "pc98-town-services-menu"` 那條。

### 選曲常式的兩條規則

`sub_18A44`（`selectTrack`）：

```
cmp MUSICSW, 1      ; 關著就什麼都不做
jz  ret
dec 曲號             ; 曲號 1 起算，內部 0 起算
cmp MUSICNUM, 曲號   ; 已經在放同一首就什麼都不做
jz  ret
mov MUSICNUM, 曲號
call 停止 → 延遲 → 開始
```

「同一首不重發」是**原作行為不是最佳化**：少了它，每次重新派曲都會把曲子從頭
播起。`musicEventForTrack` 照這條寫。

## 兩個玩家開關，互不影響

全域按鍵處理是 `sub_18036`（讀一個鍵，處理全域熱鍵）。**那支常式一共只有四個
分支**，音訊佔兩個；完整的四個與另外兩個是什麼，見 spec 1194。

| 原作鍵 | 內部碼 | 做什麼 | remake |
|---|---|---|---|
| Ctrl+S | `13h` | `OLDSOUND := SOUNDTYPE`；`SOUNDTYPE := 2`（靜音），再按換回來 | `State.ToggleSoundSwitch` |
| Ctrl+O | `0Fh` | `MUSICSW ^= 1`，然後**立刻**重新派曲 | `State.ToggleMusicSwitch` |

Ctrl+S 在 DOS 參考卡上就寫著：

```
CTRL S : Toggles sound on and off (may be used at any time).
```

「may be used at any time」是實作要求：這兩顆要擺在所有模式**前面**，連地圖預覽
那種提早 return 的畫面也管得到，而且**按鍵要被吃掉**——原作是在讀鍵的地方攔下來
的，不會再傳給任何模式。少了這一步，同一幀的 `S` 會被戰鬥選單當成別的指令。

⚠ **兩個開關是獨立的**：

- 音效走 `SOUNDFX`（`sub_18930`），開頭就是 `cmp SOUNDTYPE, 2 / jz ret`。
- BGM 走 INT 7Eh（`sub_18BDB`），**完全不看 `SOUNDTYPE`**。

所以 Ctrl+S 關掉音效時音樂照放，Ctrl+O 關掉音樂時音效照響。綁在一起是很容易犯
又不會被任何測試抓到的錯，`TestSoundAndMusicSwitchesAreIndependent` 擋住它。

★ 原作在**播放端**擋音效（`SOUNDFX` 開頭就返回），不是不產生事件。remake 照同一
個位置擋（`ConsumeSoundEvents`），佇列才不會在關著的時候越積越多。

## 不作聲的選擇子有五個，而且原因分兩處

`SOUNDFX` 的早退不是連號的一串：

```
18942h  cmp ax, 0／1／0Dh／0Eh／0FFh  → ret
18962h  cmp ax, 2／4／6／9            → 走公式
18995h  cmp ax, 0Fh                   → ret        ← **另一處**
1899Dh  其餘                          → 走表格
```

⚠ 兩件事很容易讀錯，而且錯了都只會安靜、不會報錯：

- 第一段最後一個是 **`0FFh` 不是 `0Fh`**。`SOUNDHALT` 根本不是選擇子。
- `0Fh` ＝ 15 ＝ `CRASHFX` 的早退在**第二處**。只讀第一段會以為 15 有聲音，
  然後把它從無聲清單拿掉——實際上表格在第 13 格之後就沒有資料了
  （第 13／14／15 格讀出來的是 37124、26755 這種不可能是頻率的值）。

`TestSilentSelectorsAreExactlyTheFiveTheOriginalReturnsEarlyOn` 把整個集合釘住。

## DOS 版沒有 BGM

這不是「還沒查」，是**已經答完**的問題，而且是列舉不是搜尋：

- PC-98 的音樂來自一支**獨立的驅動程式** `MSCDRV.EXE`（`internal/pc98music` 靠它的
  SHA-256 認人）。
- DOS image 的 **94 個成員逐一列舉**過：沒有那支驅動，也沒有任何音樂資料檔；
  執行檔只有 `START.EXE` 與 `COPYCURS.EXE`。
- DOS 段 `0713h` 沒有 `MSCPLAY` 的對應入口——最接近的候選只對上 7 處指紋裡的 2 處，
  而且一堆並列（spec 1187 的同一次比對）。

⚠ **正對照**：同一份列舉找得到音效實際住的地方（`START.EXE`／`GAME.OVR`），
所以「找不到音樂」不是因為看不見檔案。`TestDOSImageHasNoMusicData` 把這一組
（寬判準零命中 ＋ 正對照）釘住；image 裡真的多了音樂檔才會紅。

⇒ 音訊的「DOS 那一半」只有**音效**，而音效的對應格子已經認出來了（spec 1187，
描述子表 14 格逐模組相同）。

## 這個數字證明不了什麼

- **PC-98 版面**。DOS 版沒有除錯符號也沒有音樂，位址不能套（spec 1187）。
- 位元組直掃有偽陽性（剛好長得像 opcode 的立即數），所以逐處都印所屬常式。
  不走 far-call 對照表是因為表比實際少，而**下界看起來和全集一樣合理**（spec 1186）。
- 「接上了」只表示 remake 有對應的動作，**不表示在正確的時機發**——那要實機比對。

## 已被推翻的斷言

- ~~「停」接不上，卡點在共用 engine 不收 `track_id` 是空的 binding~~ ——
  那條路一開始就不該走：原作沒有「這一段不放音樂」的資料，停止來自**按鍵**。
  engine 收不收空字串與這件事無關（空的 `track_id` 是打錯字，本來就該被擋）。
- ~~`sfx-halt`／`sfx-off`／`sfx-on` 原作一處都沒寫到，所以不算待辦~~ ——
  那三個是常數不是狀態格，見上。

## 相關

- `cmd/audio-lifecycle-audit`、`docs/audit/audio-lifecycle.md`
- `internal/game/music_events.go`、`internal/game/sound_events.go`
- `cmd/azure-bonds-game/keys.go`：`globalAudioKeys`
- `internal/pc98sfx`：選擇子的身分與解碼
- `TestMusicSwitchIsTheOnlyThingThatStopsMusic`、
  `TestSoundAndMusicSwitchesAreIndependent`、
  `TestCtrlSAndCtrlOTogglePlayerAudioThroughTheKeySeam`
- spec 1194（全域熱鍵的完整四個）
- spec 1183（dseg 掃描）、1186（音效觸發語意）、1187（DOS 音效描述子）
