# 1155 — `2Dh CALL 2E10h` 的座標分歧：擋住它的不是模型，是 `4BF0`／`4BF1` 沒有來源

- 證據等級：`exact`（125 處可達的 `CALL 2E10h` 逐處分類；分歧的 23 處逐處列出）
  ＋ 一次已還原的對照實驗
- 原作模型見 [spec 1150](1150-ecl-call-external-routines.md) §`2E10h`
- `4BF0h`／`4BF1h` 的未決狀態見 [spec 515](515-fire-knife-map-position-transition.md)／
  [spec 516](516-fire-knife-external-map-handoff-audit.md)

## 兩件事被合在一起

原作把「座標何時生效」與「何時重繪」分開：

- `STOREVALUE` 收到 `C04B`／`C04C`／`C04D` 就**當場**寫
  `720Fh`／`7210h`／`7211h`，沒有任何條件（`overlay-07:0E89h`..`0F09h`）。
- `CALL 2E10h` 只看五個髒旗標決定要不要重繪。

remake 沒有髒旗標，改成「`CALL` 當下回頭掃同 block 的 `SaveWrites`，而且**必須**
有新寫的 `C04D` 才投影座標」。所以它同時扛下了兩件事，而那個 `C04D` 條件在原作
裡不存在。

## 分歧有多大：125 處裡的 23 處

沿 `ecl.TraceGraph` 走完，可達的 `CALL 2E10h` 共 125 處。依 `CALL` 之前六條指令
裡對 `C04B`／`C04C`／`C04D` 的寫入分類：

| 類別 | 處數 | remake 的行為 |
|---|---:|---|
| 寫了 `C04D` | 67 | 投影座標（與原作一致）|
| **只寫 `C04B`／`C04C`** | **23** | 已投影（`C04D` 條件已拿掉，spec 1158）|
| 只有 `PICTURE` | 3 | 原作靠 `8B62h` 髒旗標重繪 |
| 前六條沒有座標寫入 | 32 | 一致 |

⚠ 這張表是**只認 `09h SAVE` 當寫入路徑**時量的。把 `GETTABLE` 等其他
`STOREVALUE` 路徑一起認之後，最後兩列變成 **27／28**（多出來的四處是假門的
表驅動傳送）——逐處清單與可重生的量法見
[spec 1159](1159-storevalue-is-the-only-write-path.md)。

那 23 處裡玩家看得見的例子：

| 位置 | 前一條 | 情境 |
|---|---|---|
| `ECL2/0x01:0C7Eh` | `SUBTRACT 01 C04B C04B` | 往西走一格 |
| `ECL2/0x01:0DB5h` | `SAVE 0E C04C`＋`SAVE 07 C04B` | 「大夥喝倒了」被搬到門外 |
| `ECL2/0x01:1073h` | `ADD 01 C04C C04C` | 「你走出大廳」 |

## ★★★ 其中 15 處是同一個慣用法：還原存起來的座標

```text
SAVE 4BF0 C04B      { C04B := 4BF0 }
SAVE 4BF1 C04C      { C04C := 4BF1 }
CALL 2E10h
```

★ `SAVE <來源> <目的地>`——目的地是**第二個**運算元（`internal/ecl` 的
`memory[Operands[1].Word] = value`），所以這是**還原**，不是儲存。

⚠ **而 `4BF0`／`4BF1` 在整個 ECL corpus 裡只被常數寫過兩次**：
`ECL4/0x20:1C0Ah` 寫 `(7, 2)`、`ECL5/0x32:0056h` 寫 `(0Fh, 5)`。
另外 18／19 處全是把它們讀進 `C04B`／`C04C`（另有兩處 `COMPARE`）。
⇒ **絕大多數區域的值不是 ECL 寫的。** remake 這一側沒有任何 producer
（原始碼裡一次都沒出現這兩個位址）。

⚠ 這個「只有兩處」**不是走訪造成的假零**：`cmd/ecl-cell-refs` 用
`ecl.EntryPoints(data, 5)` 的**五個**進入點一起種 `TraceGraph`，初始化 entry
也在裡面。這個專案在 spec 1141／599／610 各踩過一次「只走一個入口就宣告 0 處」，
所以這一條先確認過才寫。

★ `4BF0` 落在 `7C00h` **之下**，所以它不是 spec 624 那批接到角色欄位的偽變數，
而是持久 ECL 記憶體（`DS:4FA5h^` 那 `1E00h` bytes，存檔第 5 欄）裡的普通工作變數
——會跟著存檔一起帶走。

## 對照實驗（已還原）

把那個 `C04D` 條件改成原作的模型（任一格被寫過就代表座標生效）之後，
四個測試立刻失敗，症狀一致：

```text
normal Fire Knife hideout leader route step 4: dungeon step from (0,1) toward 6 is blocked
```

隊伍被丟到 `(0,1)`——正是「還原一組沒有來源的座標」會有的結果。

⇒ **`C04D` 條件不是把 `2E10h` 讀錯了，它是在補 `4BF0`／`4BF1` 沒有 producer
這個洞。** 先前把 `2Dh` 的 `partial` 理由寫成「remake 沒有那五個髒旗標」，
方向對但擋路的不是它：就算把五個髒旗標全部實作出來，那 15 處還原仍然會把
隊伍丟到錯的格子。

## ★★★ `4BF0`／`4BF1` 的來源：地城主迴圈每走一步之前的座標

用 [spec 1098](1098-ecl-read-path-and-character-projection.md) 的區 0 換算
`bank0 + (位址 − 4B00h) × 2` 把兩個格子對回原生位址：

```text
ECL 4BF0h → bank0^[1E0h]
ECL 4BF1h → bank0^[1E2h]
```

而 [spec 1045](1045-dungeon-main-loop-dos.md) 早就逐條讀過那兩格的寫入端
——地城主迴圈 `overlay-02:3772h`，在 `3984h`：

```pascal
if DS:4FC7h = 0 then begin
    bank0^[1E0h] := DS:720Fh;      { 移動前的 X }
    bank0^[1E2h] := DS:7210h;      { 移動前的 Y }
    MOVEPARTY();                    { overlay-14 entry#1 }
    DRAWWINDOW();                   { overlay-28 entry#1 }
    if (bank0^[1E0h] <> DS:720Fh) or (bank0^[1E2h] <> DS:7210h) then
        <far 0713:0020h>();         { 座標被改過才叫 }
end;
```

⇒ **`4BF0`／`4BF1` ＝ 這一步移動之前的隊伍座標。** 主迴圈先存下來，
`MOVEPARTY` 之後再比對「隊伍到底有沒有移動」。所以腳本的
`SAVE 4BF0 C04B; SAVE 4BF1 C04C` 是**把隊伍退回上一格**，
而不是還原某個特殊地點。

★ 缺的從來不是逆向：spec 1045 記著「記下座標」，spec 1098 記著換算公式，
兩份都是 `exact`。缺的是把它們兜起來的那一句。⚠ 而 `overlay-02:3984h`
那一段 **IDA 沒有認成程式碼**，所以位元組掃描才找得到；只查函式清單會落空。

## remake

- `MoveDungeon` 現在在每一步之前寫 `4BF0`／`4BF1`
  （`snapshotPreMoveCoordinates`），與原作同一個時機。
- **`C04D` 條件已經拿掉**：只寫 `C04B`／`C04C` 的重畫就是真實移動，朝向不變。
  那 23 處因此全部生效，四條照著舊行為長出來的路線測試逐條重新推導過
  （[spec 1158](1158-hap-village-extent-and-refused-edges.md)）。

## 明確不宣稱

- 沒有宣稱 `bank0^[1E0h]`／`[1E2h]` 除了主迴圈以外還有沒有寫入端
  （位元組掃描另外命中 `overlay-09`／`12`／`17`／`22`／`34` 各一處，本輪沒有逐處讀）。
- 沒有宣稱那 23 處在 remake 現況下玩家一定會卡住——只宣稱**投影沒有發生**，
  而原作在同一點座標已經生效。
- 沒有宣稱那 32 處「前六條沒有座標寫入」的更前面沒有寫入——窗口只有六條。
