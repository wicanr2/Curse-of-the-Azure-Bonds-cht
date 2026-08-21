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
| `ViewMirror.Written`：記「腳本這一次執行寫過三格中的哪幾格」，只投影寫過的 | 三格沒有被每一條位置變動同步，整組蓋回去會把沒被腳本碰過的那一格拉回舊值 | 讓每一條位置變動都經過 `SetMemoryValue` |

兩者都寫成**可以拿掉的條件**，不是「就是這樣」。

## 測試

- `internal/ecl/view_mirror_test.go` 跑的是**真的 bytecode**：寫座標 →
  `CALL 2E10h` → 再寫一輪 → 再 `CALL`，兩次快照必須各自停在自己那一刻。
  ★ 用 bytecode 而不是手寫 `RunResult`——這一條要擋的正是「快照在什麼時候取」，
  手寫 fixture 會把答案直接寫進輸入裡。
- `internal/game` 那幾支 fixture 改用 `scriptView()`，它走的是 VM 同一支
  `ViewMirror.Store`。**測試裡不手寫鏡射的欄位**，否則改了 VM 也不會紅。

## 明確不宣稱

- `8B65h` 的生產者（`ECL2` entry#8／PC-98 `GODRAWWINDOW`）沒有接。它只影響
  「要不要重畫」，而 remake 收到 `2E10h` 一律重畫，所以目前沒有可觀察差異。
- 沒有宣稱 `DS:47E3h` 是什麼（spec 1150 已列為未定項，它是重畫的前置條件）。
- 沒有宣稱 `8B62h`／`8B67h`／`8B6Ah` 被立起來之後，除了重畫還有誰讀。
