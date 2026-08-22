# 1172 — `2E10h` 的髒旗標建進 VM：重畫不再靠「回頭掃寫入」猜

- 證據等級：`exact`（模型逐條取自 spec 1150 讀完的原作）＋
  `derived`（兩個 remake 這一側的補丁，理由與代價寫在下面）
- 上游：spec 1150（`2Dh CALL` 七路分派與 `2E10h` 的形狀）、spec 1155～1161
  （這條路上歷次的座標判準）
- 產物：`internal/ecl/view_mirror.go`、`internal/ecl/runtime.go`、
  `internal/game/state.go`

## 先前是怎麼猜的

原作的 `2Dh CALL 2E10h` **不是「重畫」而是「髒了才重畫」**，而座標與朝向不是
在那一刻才生效——`STOREVALUE` 收到 `C04B`／`C04C`／`C04D` 就**當場**寫
`720Fh`／`7210h`／`7211h` 並立 `8B68h`（spec 1150）。

remake 先前沒有那三格，也沒有旗標，改用一個視窗啟發式：

> `CALL` 當下回頭掃**同一 block、執行序在前、上一次重畫之後**的 `SaveWrites`。

它在同一次執行內幾乎等價，但視窗跨不過執行，也擋不住「同一次執行裡更早、與
這條 `CALL` 無關的座標寫入」。

## 現在的模型

```
STOREVALUE 寫 C04B/C04C/C04D  →  鏡射三格 ＋ 立 8B68h
STOREVALUE 寫 C059/C05F       →  立 8B67h
STOREVALUE 寫 4BFD/4BFE       →  立 8B6Ah
0Eh PICTURE 開圖               →  立 8B62h；0FFh 關圖清 8B62h/8B65h
31h SPRITE OFF                 →  清 8B65h
2Dh CALL C01Eh                 →  三格**當場**走一步（不弄髒）
2Dh CALL 2E10h                 →  快照三格 ＋ 五個旗標，然後把旗標清掉
```

鏡射掛在 shared `RuntimeState` 上，所以跨 block 也保留——原作那幾格就是全域。
每一條 `CallRequest` 帶著**它執行那一刻**的快照，呼叫端不必再排序。

## ★ `CALL C01Eh` 要在 VM 裡走

原作的 `MoveForward` 當場改地圖暫存器，所以同一次執行裡排在後面的 `2E10h`
看到的是走完之後的位置。remake 把 `CALL` 的副作用留到執行結束後才套用——
鏡射如果不在 VM 裡跟著走，中間那些 `2E10h` 的快照會停在走之前，下一次投影
就把隊伍拉回起點。

症狀很具體：盜賊公會的走位動畫（`ECL2/0x02:0D44h`，往東一步再往南三步）
會少掉第一步，停在 `(8,3)` 而不是 `(9,3)`。

## ⚠ 兩個 remake 這一側的補丁

原作沒有這兩樣。它們存在的原因是同一個：**remake 改變隊伍位置的路徑不只一條，
而且不是每一條都回寫 ECL 暫存器**；原作靠引擎每走一步都寫 `720Fh` 保持三格與
真實位置同步。

| 補丁 | 為什麼需要 | 拿掉的條件 |
|---|---|---|
| `ViewMirror.Block`：記「三格是哪個 block 的腳本寫的」，投影時要求與目前 block 相同 | 換 block 的進場放置在原作是引擎做的、會覆蓋三格；remake 那一半是 game pack 宣告的 spawn（`applyDeclaredDungeonSpawn`），沒有回寫暫存器。少了它，舊 block 留下的座標會贏過新地圖的錨點——火刀據點的入口會落在 `(6,0)` 而不是 `(6,1)` | 讓進場放置也回寫暫存器 |
| `ViewMirror.Written`：記「腳本這一次執行寫過三格中的哪幾格」，只投影寫過的 | 見下面「`Written` 擋的到底是什麼」 | 讓投影的時機對上原作：座標在**寫入那一刻**生效，不是等到重畫 |

兩者都寫成**可以拿掉的條件**，不是「就是這樣」。

## `Written` 擋的到底是什麼

第 660 輪把每一條位置變動都補上了 `SetMemoryValue`（正常移動、轉向、
`CALL C01Eh`、進場放置、開新遊戲、原版存檔匯入、remake 存檔讀回、
`syncDungeonStateFromECLRegisters`），**但 `Written` 還是拿不掉**。實際拆下來
之後剩的不是同步問題，是**時機問題**：

> 原作的 `STOREVALUE` 一寫 `C04B`／`C04C` 就**當場**改 `720Fh`／`7210h`，
> 隊伍在那一刻就已經在新格子上了；`2E10h` 只是把畫面重畫。
> remake 反過來——位置要等到 `2E10h` 投影才生效。

只要有一次執行「寫了座標卻沒有 `2E10h`」，兩個模型就分岔：原作的隊伍已經
搬走，remake 的還在原地，而 `Dirty` 會**跨執行留著**（只有重畫會清），
`Written` 卻在每一次頂層執行開頭歸零（`session.go` 的
`runFromSeedWithPartyContextAndInputs`）。於是下一次執行的第一個 `2E10h`
會拿到「`Dirty` 立著、`Written` 全 0、座標是上一次留下的」這種組合。

實例（`ECL3/0x10` 指揮官側門）：某一次執行寫 `C04B=2`／`C04C=5`／`C04D=1`
之後就結束、沒有重畫；下一次執行的 `2E10h` 於是帶著 `(2,5,2)` 但
`Written=000`。拿掉 `Written` 就會在那一刻把隊伍搬到 `(2,5)`。

要拿掉它，正確的做法**不是**再補一個遮罩，而是把投影移到寫入那一刻——
也就是讓 `STOREVALUE` 寫 `C04B`／`C04C`／`C04D` 時就更新 remake 的隊伍位置，
`2E10h` 只負責重畫與清旗標。那是一次獨立的模型改動，牽動
`applyECLCallSignals` 的整條投影路徑，不在本規格的範圍內。

## 第 660 輪補上的同步點

原作的引擎每動一次位置或朝向就寫 `720Fh`；remake 這一側對應的是
`State.syncDungeonECLRegisters()`。以下幾條先前沒有接上，現在都接了：

| 路徑 | 先前的洞 |
|---|---|
| `TurnDungeon`／`TurnDungeonWithGrid` | 轉向不跑 ECL，整條路徑沒碰暫存器 |
| `LoadSAVGAMPrefix` | 原版存檔的 map state 五個位元組**就是**那五格，先前只填了 `MapX`／`MapY`，`DungeonX`／`DungeonY` 停在上一次的值 |
| `LoadPartyFile` | 版本回退與範圍夾擠（`= 7, 13, 0`）會改座標，改完沒回寫 |
| `finishNewGameEntry` | `C04B`／`C04C` 缺格時退回 `(7,13)`，退回之後沒回寫 |
| `syncDungeonStateFromECLRegisters` | 同上，讀進來之後沒有把補齊的值寫回去 |

同一輪還移除了 `2Dh` 這條路上唯一一個硬寫死的座標補丁——`GO WITH GUARDS`
在 resume 前寫 `C04B=0`／`C04C=3`／`C04D=1`。`ECL3/0x10:192Ah`..`1997h` 的
查表走位迴圈本來就會把隊伍走到 `(0,3)`（見 spec 332），那個補丁是照著一個
被反組譯推翻的斷言寫的。

## 測試

- `internal/ecl/view_mirror_test.go` 跑的是**真的 bytecode**：寫座標 →
  `CALL 2E10h` → 再寫一輪 → 再 `CALL`，兩次快照必須各自停在自己那一刻。
  ★ 用 bytecode 而不是手寫 `RunResult`——這一條要擋的正是「快照在什麼時候取」，
  手寫 fixture 會把答案直接寫進輸入裡。
- `internal/game` 那幾支 fixture 改用 `scriptView()`，它走的是 VM 同一支
  `ViewMirror.Store`。**測試裡不手寫鏡射的欄位**，否則改了 VM 也不會紅。
- `internal/game/dungeon_register_sync_test.go` 釘住上面兩條最容易漏的同步點
  （轉向、讀檔）。兩支都做過正對照：把同步拿掉會紅。
- **沒有**為「移除 `GO WITH GUARDS` 補丁」加測試。理由寫在 spec 332：
  remake 每走一步就重新錨定暫存器，執行前的硬寫會被第一步抹掉，端到端沒有
  可觀察差異。加一支永遠不會紅的測試比不加更糟。

## 明確不宣稱

- `8B65h` 的生產者（`ECL2` entry#8／PC-98 `GODRAWWINDOW`）沒有接。它只影響
  「要不要重畫」，而 remake 收到 `2E10h` 一律重畫，所以目前沒有可觀察差異。
- 沒有宣稱 `DS:47E3h` 是什麼（spec 1150 已列為未定項，它是重畫的前置條件）。
- 沒有宣稱 `8B62h`／`8B67h`／`8B6Ah` 被立起來之後，除了重畫還有誰讀。
