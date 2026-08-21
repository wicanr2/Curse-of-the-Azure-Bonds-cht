# 1150 — `2Dh CALL` 的七支分派：名字、序列游標與那個走不到的音效

- 證據等級：`exact`（DOS `overlay-02:2F02h`..`3072h` 124 條、PC-98
  `overlay-02:30C6h`..`3236h` 逐條對照，結構完全相同）
- 選擇子換算（`selector = 運算元 − 7FFFh`）與七支的位址見 spec 561
- handler 位址與條數見 [`docs/audit/ecl-opcode-handlers-dos.md`](../audit/ecl-opcode-handlers-dos.md)

## 七支各是誰

分派器認得七個選擇子，其餘一律走到 epilogue 直接返回。目標名字由 PC-98 的
Borland 符號表對回來——兩平台的 TPOV **entry 槽號**相同，所以
`far <段>:<位移>` 解出 `overlay-NN entry#K`（spec 1112）之後，查 PC-98 同一
overlay 的第 K 個 stub 就得到名字：

| 運算元 | 選擇子 | 目標 | 這一支做什麼 |
|---|---|---|---|
| `2E10h` | `AE11h` | `THREED.SPECIALCODE`／`DRAWWIN` entry#1／`GENERIC.SHOWLOCATION`／`THREED.WALLCODE` | 重算目前格的特殊碼；畫面髒了才重畫 |
| `C01Eh` | `401Fh` | `ECL2.MOVEFORWARD` | 依朝向前進一格（不做碰撞判定）|
| `B200h` | `3201h` | `SOUNDFX` | 播 10 或 11 號音效 |
| `C018h` | `4019h` | `THREED.WALLCODE` | `bank1^[1CCh] = 0` 時才重算牆面碼 |
| `8000h` | `0001h` | `ECL2.GODUEL(1)` | corpus 沒用到 |
| `8001h` | `0002h` | `ECL2.GODUEL(0)` | corpus 沒用到 |
| `6803h` | `E804h` | `PORTRAIT.SHOWPORTRAIT` | 圖片序列往前推一格 |

⚠ **這座橋是槽號，不是位址。** 同一個槽在兩平台的本地位址常常不同，也常常
一邊是 `FFFFh`（那一版沒有這支）。`overlay-29` 的 `000Ch`／`0091h`／`00B3h`／
`010Bh` 四個位址兩平台一字不差、`overlay-24` 的 `0025h`／`021Dh`／`02A7h`／
`03C5h` 也是，那是巧合不是規則：`overlay-28`（`DRAWWIN`）DOS 有六個實際 stub、
PC-98 只有三個，DOS 的 entry#1 在 `00CCh` 而 PC-98 的在 `0016h`。這一對已經
另外逐指令比對過（台帳：DOS `overlay-28:00CCh` 是完整版、PC-98 `0016h` 是砍掉
`DS:8B67h` 那條分支的簡化版），所以叫得出名字；沒比對過的槽不能只憑槽號命名。

## corpus 用到四個，不是兩個

沿 `ecl.TraceGraph` 走完 25 個 block，可達的 `2Dh` 共 **168 條**，運算元四種：

| 運算元 | 條數 |
|---|---:|
| `2E10h` | 125 |
| `B200h` | 19 |
| `C01Eh` | 13 |
| `6803h` | 11 |

⚠ 拿 `docs/audit/ecl-event-catalog.json` 數會得到 78 條——**那份 committed 產物是舊的**
（4,222 條指令、55 個 opcode），現跑 `cmd/ecl-event-catalog` 是 14,177 條、61 個
opcode，`2Dh` 正好 168。它目前重生不出來，見 `WORKLIST.md` 的 ⚠。

## `6803h`：圖片序列的下一格

```asm
3035  SHOWPORTRAIT(3, 3, 1, 序列[游標])      ; 游標 ＝ DS:722Dh，記錄在 DS:722Ch
3058  inc  byte ptr ds:722Dh                 ; 游標 ＋1
305C  cmp  al, ds:722Ch                      ; 超過張數
3063  jbe  short loc_306A
3065  mov  byte ptr ds:722Dh, 1              ; 就回到第 1 格
306A  call far ptr 542h:0B33h                ; GAMEDELAY——與 3Ah DELAY 同一支
```

**先畫再推**，所以第 k 次呼叫畫的是第 `((k−1) mod 張數) + 1` 格。收尾那個
`542h:0B33h` 就是 `3Ah DELAY` 整支唯一做的事（`28F3h`，7 條），PC-98 對應
`418h:14AAh GAMEDELAY`。

### 序列記錄的版面

`p ＝ DS:722Ch`（PC-98 `DS:0A2C6h`）：

| 位移 | 內容 |
|---|---|
| `p+0` | 張數 |
| `p+1` | 目前游標（1 起算；0 ＝ 沒有序列）|
| `p + 8i − 6` | 第 i 格的延遲（4 bytes，取自檔案）|
| `p + 8i − 2` | 第 i 格影像緩衝區的遠指標 |

⚠ PC-98 的游標在 `p+2` 而不是 `p+1`——就是既有的「PC-98 記錄大 1 byte」那個差異
（`DISPOSESEQUENCE` 逐項釋放時 DOS 用 `p + 8i − 2`、PC-98 用 `p + 8i − 1`）。

### 生產者是 `0Eh PICTURE`

spec 1148 讀出的一般圖那一支，兩條指令正好是本節的另一半：

```asm
08B5  LOADSEQUENCE(圖名, n, 0, @DS:722Ch)   ; overlay-29 entry#4
08CD  SHOWPORTRAIT(3, 3, 1, 序列[1])         ; DS:7232h/7234h ＝ p + 8×1 − 2
```

`LOADSEQUENCE` 在 `overlay-29:0270h` 寫 `p^[1] := 1`，所以**換圖會把游標設回第
1 格**；`PICTURE` 自己再直接畫第 1 格。⚠ 它有一格快取（圖名與子索引都相同就
直接返回），所以連續兩次 `PICTURE` 同一張圖**不會**重設游標。corpus 裡沒有這種
用法，這個差別今天觀察不到。

### 檔案裡確實是多格

`PIC*.DAX` 的每個 block 都是序列，不是單張圖：長度一律等於 `1 + 張數 × 3893`，
每格 21 bytes 表頭加 88×88 的 4bpp 像素。remake 的 `gfx.ParseAnimationWithDelta`
早就照這個版面解，`assets/sprites/pic*-block-XX-frame-NN.png` 就是它產的。

### ★ 有兩個驅動器共用同一個游標

`6803h` 不是唯一會推格的地方。`MENUS`（DOS `overlay-26:0507h`..`05A6h`）在等
輸入時自己跑一個迴圈：游標非 0 就畫目前這一格，拿 `p + 8i − 6` 的延遲乘 10 與
時鐘比，到時間就推一格、一樣迴繞。所以

- **自走那一支**負責「站在選單前圖會動」，節奏是每格自帶的延遲；
- **`6803h`** 是腳本另外推的格，用在**沒有選單可等**的地方。

corpus 裡 `6803h` 只出現在開場（`ECL1/0x52` 位移 `01E2h`..`020Ah`，十一條連在
一起，緊接在 `PICTURE 1Dh` 與「NOWHERE IN THE REALMS IS COMPLETELY SAFE.」之後）
——那一段是定時播放的開場白，沒有選單，所以動畫要腳本自己推。

### 動畫關掉的時候

`LOADSEQUENCE` 在 `overlay-29:0279h` 有一道：

```pascal
if (DS:4FBFh = 0) and (圖名 = 'PIC' 或 'FINAL') then 張數 := 1;
```

`DS:4FBFh` 是 Alter 選單的 **Animation on/off**（spec 909）。關掉之後序列只載
一格，`6803h` 與 `MENUS` 那一支都在同一格上迴繞 ⇒ 圖片變成靜止。同一個旗標也
決定要不要印 `'Loading...Please Wait'`（開著才印，因為要載的格數多）。

## `B200h`：11 號那一支在 CoAB 走不到

```asm
2FCF  ax := ds:8B4Ch
2FD2  cmp ax, 8    → SOUNDFX(ds:25C0h ＝ 10)
2FE2  cmp ax, 0Ah  → SOUNDFX(ds:25C2h ＝ 11)
2FF2  否則           SOUNDFX(ds:25C0h ＝ 10)
```

`DS:8B4Ch` 的唯一寫入端是 `ECL2.STOREVALUE`（DOS `overlay-07:0E61h` 的
`cmp ax, 3DEh`），也就是 **ECL 格 `03DE`**。全 corpus 對 `03DE` 只有 **15 次
存取，全部是 `SAVE 05 03DE`**，沒有任何讀取——所以值永遠是 5，永遠走預設那一支。
`8B4Ch` 也沒有原生寫入端：以 `4C 8B` 逐位元組掃過 36 個 overlay 與 `START.EXE`，
另外三筆命中都是相鄰兩條指令湊出來的假命中。

⇒ **`CALL B200h` 在 CoAB 一律播 10 號音效**，remake 對到 `SoundStep` 是對的；
第 276 輪留的「值 10 → sound B 分支未實作」不是缺口，是走不到的碼。

## `2E10h`：髒了才重畫

```pascal
DS:7213h := SPECIALCODE(DS:720Fh, DS:7210h);          { 每次都算 }
if (DS:47E3h <> 0) and (8B62h or 8B65h or 8B67h or 8B68h or 8B6Ah <> 0) then begin
    DS:7580h := 1;
    <DRAWWIN entry#1>();  SHOWLOCATION();
    8B6Ah := 0; 8B67h := 0; 8B68h := 0; 8B62h := 0; 8B65h := 0;
    DS:7212h := WALLCODE(DS:720Fh, DS:7210h, DS:7211h);
end;
```

五個是**髒旗標**，各自有生產者：

| 旗標 | 誰設 1 |
|---|---|
| `8B62h` | `0Eh PICTURE` 開圖 |
| `8B65h` | `ECL2` entry#8（PC-98 `GODRAWWINDOW`）；`31h SPRITE OFF` 會清 |
| `8B67h` | 寫 ECL 格 `C059`／`C05F` |
| `8B68h` | 寫 ECL 格 `C04B`／`C04C`／`C04D` |
| `8B6Ah` | 寫 ECL 格 `4BFD`／`4BFE` |

`720Fh`／`7210h`／`7211h` 的來源同時定案：`STOREVALUE` 收到 `C04B` 就寫
`720Fh`、`C04C` 寫 `7210h`、`C04D` 折成 0／2／4／6 寫 `7211h`，**當場寫、三者
都順手把 `8B68h` 設 1**（DOS `overlay-07:0E89h`..`0F09h`）。所以座標與朝向不是
在 `CALL` 那一刻才生效的，`CALL 2E10h` 只負責「髒了就重畫」。

這三格是 X／Y／朝向還有第二個證人：`SHOWLOCATION`（DOS `overlay-24:2BAAh`）組
狀態列時就是 `Str(DS:720Fh) + ',' + Str(DS:7210h) + ' '` 再接
`DS:2540h + DS:7211h × 3` 那張方向名表。

順帶關掉 spec 1148 的一個未定項：`4FBAh`／`4FBBh` 由 ECL 格 `4BE6` 決定——
`STOREVALUE` 發現 `4BE6` 的新值與舊值不同時，先 `4FBBh := 4FBAh`，再依新值是
0 或非 0 把 `4FBAh` 設成 3 或 4（`overlay-07:0DA0h`..`0DC9h`）。`PICTURE 0FFh`
那句「兩個都等於 4 就不重繪」＝「本來就是這種畫面、現在還是」。

## remake

- `2Dh CALL 6803h` 現在會推 `RunResult.PictureFrameAdvances`；`0Eh PICTURE`
  開圖時歸零，對上原作 `LOADSEQUENCE` 的 `p^[1] := 1`。
- `State.PictureFrameAdvances()` 是目前這張圖被腳本推了幾格；畫面把它當成
  自走相位上的位移（`(自走索引 + 推格數) mod 張數`），因為原作那兩個驅動器
  共用同一個游標。
- `B200h` 維持 `SoundStep`，理由改成「另一支走不到」而不是「還沒接」。

⚠ 仍是 `partial`，剩下的是 `2E10h`。remake 沒有那五個髒旗標，改用「`CALL`
當下回頭掃同 block、執行序在前的 `SaveWrites`」。125 處可達的 `CALL 2E10h`
現在全部照原作的模型投影——包含只寫 `C04B`／`C04C` 的那 23 處（朝向刻意不變，
[spec 1157](1157-restore-previous-cell-sites-and-shared-handlers.md)、
[spec 1158](1158-hap-village-extent-and-refused-edges.md)）；
`4BF0`／`4BF1` 的 producer 是地城主迴圈的移動前快照
（[spec 1155](1155-redraw-call-coordinate-divergence.md)）。

★ 剩下的差異是**視窗不等於髒旗標**：同一次執行裡更早、與這條 `CALL` 無關的
座標寫入仍會被算進來。要收掉得把 `720Fh`／`7210h`／`7211h` 與那五個髒旗標
建出來，讓 `CALL` 只做「髒了才重畫」。

## 明確不宣稱

- 沒有宣稱 `DS:47E3h` 是什麼（只知道 `21h LOAD FILES` 會設 1、ECL 主迴圈會設 0
  與 1，而它是 `2E10h` 重繪的前置條件）。
- 沒有宣稱 `DS:7580h`、`bank1^[3FEh]`、`bank1^[1CCh]` 以外的欄位語意。
- 沒有宣稱 `GODUEL` 做什麼——只知道 `8000h`／`8001h` 差一個常數 1／0，而且
  corpus 沒有用到。
- 沒有宣稱序列每一格那 4 bytes 的延遲單位是什麼；只知道 `MENUS` 把它乘 10 之後
  與時鐘比，而 `6803h` 這條路根本不看它。
- 沒有宣稱 DOS 的 `overlay-28` entry#1 叫 `DRAWWINDOW`（見上面的槽號警告）。
