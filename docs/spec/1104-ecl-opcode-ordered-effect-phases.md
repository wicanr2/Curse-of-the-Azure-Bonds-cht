# 1104 — ECL opcode 的有序副作用：PC 一律先推進，畫面提交點只有一個

- 狀態：`READY`
- 證據等級：DOS `exact`（23 支 handler ＋ 操作元解碼器 ＋ lifecycle 驅動器逐條讀完，
  另以全 overlay 位元組掃描列出每個旗標的寫入點）；PC-98 只逐條比對了兩項（見 §八）
- 主迴圈與 `24h` 見 spec 1095；助憶碼表見 spec 1083；dispatcher 對照見
  [`ecl-opcode-dispatch.md`](../audit/ecl-opcode-dispatch.md)
- 對應 `RE-01`（全域 P0-RE-1）要求的「opcode 當下的 ordered effect record」

## 範圍

spec 1095 讀完了主迴圈與 `24h COMBAT`，證明戰鬥前後沒有提交邊界。剩下的 29 個
ordered-effect 候選如果一個一個當作獨立問題處理，會重複讀同一批 handler：33 個
候選只由 20 個不同的 opcode 組成，而**次序是 dispatcher 與各 handler 的性質，
不是某一段直線區域的性質**。所以本輪改成逐 opcode 閉合。

本規格宣稱的範圍：DOS `START.EXE` ＋ `GAME.OVR` 的 `overlay-02`（`INTERPET`）
與 `overlay-07`。不宣稱 PC-98 逐支相同，也不宣稱 `unknown` 的 21 支沒有副作用。

## 一、通則：PC 在效果之前推進

每個 handler 的第一個動作都是推進 ECL 程式計數器 `ds:4FB4h`：

| 形狀 | 作法 | 例 |
|---|---|---|
| 帶 operand | `overlay-07 entry#2(n)`，一次吃掉 opcode 與 n 個 operand | `0Eh PICTURE` `0841h` |
| 不帶 operand | 單一 `inc word ptr ds:4FB4h` | `3Ah DELAY` `28F3h` |
| 分支 | 解完 operand 後直接把目的位址寫進 `4FB4h` | `01h GOTO` `00E8h` |

⇒ **效果一律發生在 PC 推進之後。** 任何停下再續跑都從下一條指令開始，
同一條指令不會被執行第二次。這就是原生的 exactly-once 機制，不需要另設
「已處理」旗標。

★ 反過來說，remake 若在指令邊界暫停並保存 PC，保存的必須是**下一條**指令；
保存指令本身只有在「重新解碼沒有副作用」時才等價。目前 `2Bh` 走的就是後者
（見 §六）。

## 二、operand 的編碼與定址（`overlay-07` entry#1／#2）

`entry#2(n)`（overlay-07 局部 `0034h`）逐個 operand 讀兩個位元組並存進三張
平行表，最後統一 `PC := PC + 1` 跳過 opcode 自己：

| 表 | 位址 | 內容 |
|---|---|---|
| code | `ds:7685h + i` | 定址模式 |
| low | `ds:7705h + i` | 低位元組 |
| high | `ds:76C5h + i` | 高位元組（只有部分模式會再讀一個位元組） |
| 字串槽 | `ds:7648h + slot×100h` | 每槽 256 bytes |

`entry#1(k)`（局部 `0173h`）把第 k 個 operand 解成值：

| code | 語意 | 值 |
|---|---|---|
| `00h` | 立即位元組 | `low` |
| `01h`／`03h`／`80h` | 記憶體讀取 | `mem[(high<<8)｜low]` |
| `02h`／`81h` | 位址字面值 | `(high<<8)｜low`，不解參照 |

⚠ 本規格不宣稱 `80h`／`81h` 兩種字串模式的完整打包格式（另見既有 text 規格）。

## 三、`20h NEWECL` 是終止指令，不是 fallthrough

`0BBBh` 依序做：

```pascal
entry#2(1);                              { 推進 PC、解出新 ECL 編號 }
bank0^[1E4h] := ds:8B5Eh;                { 記住舊 ECL 編號 }
ds:8B5Eh := entry#1(1);                  { 換成新編號 }
entry#4(ds:8B5Eh);                       { 載入新 block }
entry#3();                               { 初始化 }
ds:47E0h := 1;  ds:47E1h := 1;           { ★ 兩個停止旗標 }
FillChar(ds:8B48h, 2, 0);
```

`entry#3`（overlay-07 `01FCh`）把 **`ds:4FB4h` 重設為 `8000h`**，清空 GOSUB
框架鏈，然後從新 block 的開頭連讀五個 operand，依序寫進 `4FAAh`、`4FACh`、
`4FAEh`、`4FB0h`、`4FB2h`——★ **這五格就是 lifecycle entry 的 PC**，與清冊
`lifecycle_entries` 的五筆一一對上。

`entry#4`（`0380h`）把 DAX 讀進 `4FA5h^`，來源跳過前 2 個位元組、長度減 2
——與 spec 1095「`DS:4FA5h` 指向 decoded block ＋ 2」相符。

lifecycle 驅動器（`3691h`）看到 `47E1h <> 0` 時**跳回迴圈頂端重跑整個
lifecycle**，不是離開：

```
375E:  cmp ds:47E1h, 0
3763:  jz  3768
3765:  jmp 3694        ; ★ 回到驅動器開頭
```

⇒ `NEWECL` 之後的那個位元組**永遠不會被執行**。兩項獨立證據：停止旗標，
以及 PC 被重設為新 block 的 `8000h`。

## 四、畫面的提交點只有一個：`CALL 2E10h`

五個髒旗標 `8B62h`／`8B65h`／`8B67h`／`8B68h`／`8B6Ah` 的**全部**寫入點，
由全 overlay 的 `C6 06 <addr> <imm>`／`FE 06`／`FE 0E` 位元組掃描列出：

| 旗標 | 設為 1 的位置 | 清為 0 的位置 |
|---|---|---|
| `8B62h` | `overlay-02:0868h`（`0Eh PICTURE`）、`overlay-07:066Fh` | `2Dh` 的 `2E10h` 分支、`00h EXIT`、`PICTURE` flush 分支、`31h`、`overlay-07:0202h`、overlay-11 兩處 |
| `8B65h` | `overlay-07:05FCh` | 同上組 |
| `8B67h` | `overlay-07:0220h`（`entry#3`）、`0F16h`、`0F28h` | 只有 `2Dh` 的 `2E10h` 分支 |
| `8B68h` | `overlay-07:0E9Ah`、`0EB3h`、`0F04h`、`1BCDh` | 只有 `2Dh` 的 `2E10h` 分支 |
| `8B6Ah` | `overlay-07:0D99h` | 只有 `2Dh` 的 `2E10h` 分支 |

`2Dh CALL` 的 `2E10h` 分支（`2F31h`..`2FA7h`）：

```pascal
ds:7213h := <overlay-30 entry#6>(ds:720Fh, ds:7210h);     { 無條件重取目前格 terrain }
if ds:47E3h = 0 then goto 收尾;                            { 資源還沒載入就不提交 }
if (8B62h｜8B65h｜8B67h｜8B68h｜8B6Ah) = 0 then goto 收尾;
ds:7580h := 1;
<overlay-28 entry#1>();  <overlay-24 entry#38>();          { 重繪 }
8B6Ah := 0; 8B67h := 0; 8B68h := 0; 8B62h := 0; 8B65h := 0;
ds:7212h := <overlay-30 entry#4>(ds:720Fh, ds:7210h, ds:7211h);
```

⇒ `0Eh PICTURE`（operand 不是 `0FFh` 時）只設旗標；玩家看到畫面換掉是在後面
那個 `CALL 2E10h`。這正是候選
`ECL2.DAX/0x02/0x02CB-0x0325`（P0-C）與其餘四個
`PICTURE → PRINTCLEAR → CALL` 候選的次序答案。

★ `PICTURE` 的 operand 是 `0FFh` 時走另一條路：就地 flush（若 `8B62h` 或
`8B65h` 非零則呼叫 `overlay-28 entry#1` 並清掉兩者），再清 `8B48h`／`8B49h`。
所以 `PICTURE 0FFh` 是「關掉圖框」，不是「載入第 255 號圖」。

## 五、`21h LOAD FILES` ＋ `37h LOAD PIECES` 是配對閂鎖

兩者共用 `0C15h`，靠重讀 `ds:75FFh` 分流；重繪在函式尾端：

```pascal
if (ds:47E4h <> 0) and (ds:47E5h <> 0) and (ds:4FBBh = 3) then begin
    if (ds:4FBAh <> 3) and (ds:8B6Eh <> 0) then <重繪三連呼叫>;
    ds:8B6Eh := 0;
end;
```

`47E5h` 由 `21h` 設、`47E4h` 由 `37h` 設，**兩個都到齊才重繪**，任何一個單獨
出現都不會有可見效果。兩格在進入新 block 時（`overlay-02:3772h`）與
`47E2h`、`47E3h` 一起被清為 0 ⇒ 閂鎖的作用域是一個 block session。

這解釋了候選 `ECL6.DAX/0x42`、`0x43` 的
`LOAD FILES → LOAD PIECES → CALL` 為何三者必須同時出現，
以及 `ECL1.DAX/0x50`、`0x51` 的 `LOAD FILES → PICTURE` 為何看不到中間狀態。

★ **`resume_only` 這一類是空的。** 所有「要等這一輪跑完才發生」的效果都在
lifecycle 驅動器（`3691h`）與 block 轉換函式（`3772h`）裡，不在任何 opcode
handler 內。這是本輪的負面結果，值得記著：找 opcode 的 resume 語意會白費。

## 六、`2Bh HORIZONTAL MENU` 的變長解碼

`1082h` 先 `entry#2(2)` 解出目的位址與項數，接著 **`dec word ptr ds:4FB4h`**
退一格，再 `entry#2(項數)` 重解一次——第二次的結果決定最終 PC，也把選項字串
載進 `7648h + i×100h` 的槽。選擇結果在 handler 內就寫回目的位址。

remake 的 `internal/ecl/runtime.go` 用 `ParseOperands(payload, headNext-1, count)`
表達同一件事，數值上與 `dec` 相符。

⚠ 一個刻意的差異：remake 在等待選擇時把 PC 保存成**指令自身**而不是下一條。
原作沒有這個狀態（它在 handler 內阻塞）。兩者等價的前提是「重新解碼沒有副作用」
——原作解碼確實只寫 `7685h`／`7705h`／`76C5h`／`7648h` 這幾張會被下一次解碼
覆蓋的暫存表。**這是刻意的，不要當成 bug 修掉。**

## 七、`2Dh CALL` 的完整登記表（`RE-03`）

`2F02h` 把 operand 的 16-bit 值減去 `7FFFh` 之後走七路 switch：

七支的目標名字與逐支內容見 [spec 1150](./1150-ecl-call-external-routines.md)；
本節只保留分派形狀與出現次數。

| operand 值 | switch 比較值 | 動作 | corpus 可達條數 |
|---|---|---|---|
| `2E10h` | `0AE11h` | 重取 terrain ＋ 五個髒旗標的提交點（§四） | **125** |
| `8000h` | `1` | `overlay-07 entry#25(1)` ＝ `GODUEL(1)` | 0 |
| `8001h` | `2` | `overlay-07 entry#25(0)` ＝ `GODUEL(0)` | 0 |
| `B200h` | `3201h` | 依 `ds:8B4Ch`（＝ ECL 格 `03DE`）取一個 word 呼叫 `SOUNDFX` | **19** |
| `C01Eh` | `401Fh` | `overlay-07 entry#26` ＝ `MOVEFORWARD` | **13** |
| `C018h` | `4019h` | bank1 `1CCh` 為 0 時 `ds:7212h := overlay-30 entry#4(...)` | 0 |
| `6803h` | `0E804h` | 依 `ds:722Dh` 取一組指標呼叫 `overlay-29 entry#1`（`SHOWPORTRAIT`）繪製，`722Dh` 遞增並在超過 `ds:722Ch` 時繞回 1，最後呼叫與 `3Ah DELAY` 相同的 resident 延遲常式 ⇒ 推圖片序列一格 | **11** |

★ **CoAB 的腳本用到四個目標**：沿 `ecl.TraceGraph` 走完 25 個 block，可達的 `2Dh`
共 **168 條**（`2E10h` 125、`B200h` 19、`C01Eh` 13、`6803h` 11）；另外三路
（`8000h`／`8001h`／`C018h`）一次都沒出現。
⚠ 舊產物曾只數到 78 條（4,222 條指令、55 個 opcode）；2026-08-27 已依
spec 1219 重生 `docs/audit/ecl-event-catalog.json`，現為 14,177 條、61 個
opcode，`2Dh` 正好 168。
★ **不在 switch 內的目標會靜默 no-op**（直接落到收尾 `306Fh`），不會報錯。

`722Ch` 是圖片序列的記錄（`p^[0]` 張數、`p^[1]` 游標、第 i 格的遠指標在
`p + 8i − 2`），生產者是 `0Eh PICTURE` 的 `LOADSEQUENCE`；版面與另一個自走的
驅動器見 spec 1150。

## 八、PC-98 對照

本輪只逐條比對兩項，其餘只有 dispatcher 位址對照（`ecl-opcode-dispatch.md`）：

| 項目 | DOS | PC-98 | 結果 |
|---|---|---|---|
| operand 解碼推進 PC | `overlay-07 entry#2`（`0034h`） | `overlay-07 entry#2`（`008Eh`） | 每個 handler 的第一個動作相同 |
| `20h NEWECL` | `0BBBh`，`47E0h`／`47E1h` | `0C26h`，`7896h`／`7897h` | 動作序列逐條對應 |

PC-98 的 `0Ch SETUP MONSTER`（`03E0h`）比 DOS 多了在頭尾保存／還原 `ds:0A2A8h`
並在結束前呼叫 `overlay-26 entry#6`。⚠ 本規格不宣稱這兩處的語意。

## 九、對清冊的修正

`internal/eclcatalog` 的 `endsStraightLine` 原本不含 `20h`，於是把 `NEWECL`
之後的位元組併進同一段直線區域。修正後：

| 項目 | 修正前 | 修正後 |
|---|---|---|
| 不重複靜態可達 instruction | 1,355 | 1,355（不變） |
| 跨 effect-kind 直線候選 | 33 | **32** |

> 這兩個數字後來又被 spec 1106 改寫：`IF` 條件不成立會跳過下一條指令，補上那條
> 路之後可達指令變成 **4,222**、候選變成 **154**。本節保留的是當輪的差異。
| `ECL4.DAX/0x25` 候選 ID | `0x021F-0x023B` | `0x021F-0x022E` |
| `ECL5.DAX/0x30/0x0086-0x00B0` | 存在 | **消失**（`NEWECL → LOAD CHARACTER` 是假序列） |

指令數不變的原因：corpus 裡只有兩個 `NEWECL`，而兩者後面的指令都另有入口
（一個是 `pre_camp` lifecycle entry，一個是 `0x00D7` 的 `GOTO` 與 `0x0014`
的 `GOSUB` 目標），所以它們本來就會被列入，只是不該接在 `NEWECL` 後面。

`38h PROGRAM` 的終止性依 operand 值而定（值 3 與值 9 轉呼叫 `00h` 的 handler，
值 3 另外設 `4FC7h:=1`；值 0 與值 8 會正常回到迴圈）。⇒ spec 1106 已依值切斷：
corpus 的三個 `PROGRAM` 全是立即值 3 或 9。

## 九之二、`1Ch CLEARMONSTERS` 不會撤銷 `0Ch SETUP MONSTER`

`120Eh` 只寫 `47E6h`（已放置怪物數）、`8B69h`、`7603h`、`ds:6F70h` 起 28 個位元組，
再沿 `ds:6F8Ch` 鏈逐節點釋放。`SETUP MONSTER` 寫的 `ds:7601h`、`ds:7602h` 與
bank1 `580h`／`582h` **一格都沒被碰到**。

corpus 有四個位置先 `SETUP MONSTER` 再 `CLEARMONSTERS` 才開打：
`ECL3.DAX` block `0x11 +1154h`、`ECL3.DAX` block `0x12 +06C5h`、
`ECL4.DAX` block `0x21 +05BBh`、`ECL5.DAX` block `0x32 +077Ah`。
remake 原本在 `1Ch` 連 `MonsterSetup` 一起清掉，這四場的敵人因此拿不到
`AnimationBlock`。已改成只清 spawn 清單，並以
`TestRealClearMonstersKeepsEarlierMonsterSetup` 釘住四個站點。

## 十、產物

| 產物 | 內容 |
|---|---|
| [`ecl-opcode-effect-phases.md`](../audit/ecl-opcode-effect-phases.md)／`.json` | 46 個 corpus opcode 的 commit phase 台帳，由 `internal/eclcatalog/phases.go` 產生 |
| `cmd/ecl-event-catalog -phases-output／-check-phases` | 產生與擋 drift |
| `scripts/resolve_far_calls.py` | far call `seg:off` → resident 位址或 `overlay-N entry#i` |
| `tools/ida/dump_functions_batch.py` | 一次容器啟動 dump 多支函式 |

重生指令：

```sh
./tools/go.sh run ./cmd/ecl-event-catalog \
    -output docs/audit/ecl-event-catalog.json \
    -summary-output docs/audit/ecl-event-catalog.md \
    -phases-output docs/audit/ecl-opcode-effect-phases.json \
    -phases-summary-output docs/audit/ecl-opcode-effect-phases.md
```

## 十一、明確不宣稱

- 21 支 `unknown` handler 的任何副作用或次序（`03h`..`09h`、`14h`、`16h`..`1Bh`、
  `1Dh`、`25h`、`29h`、`2Ah`、`2Fh`、`32h`、`35h`）。
- `11h`／`12h` 呼叫的 resident 輸出常式（`resident:006074h`）是否會等待按鍵；
  這不影響 ECL 次序，但也還沒證明。
- `2Bh` 的選單常式（`overlay-07 entry#20`）內部如何取鍵。
- `7686h`（`PRINT` 在 `>= 80h` 時跳過文字解碼）與 `8B61h`、`8B63h`、`8B64h`、
  `8B6Dh`、`65A0h`、`65A1h` 的完整語意。
- `47E2h` 這條 deferred 線的 remake 對應：`0Ah LOAD CHARACTER` 在進入時把它設為 1，
  `00h EXIT`、`38h PROGRAM` 與驅動器（`3691h`／`3772h`）都會把 `ds:6506h` 還原成
  `47E7h` 並清掉它。⇒ **原作在一次執行結束時會把「目前角色」還原成進入時的那一個**，
  remake 目前沒有這個還原。本輪只記錄機制，沒有改 remake，因為還沒有玩家可見的
  失敗案例把它釘住。
- `CALL` 未使用的五路分支、`722Ah` 指標陣列版面。
- PC-98 除 §八 兩項以外的任何逐條對應。

## 十二、回歸

| 測試 | 釘住什麼 |
|---|---|
| `internal/ecl.TestRealNewEclStopsBeforeTheFollowingLifecycleEntry` | 真實 ECL4 block `0x25`：`NEWECL` 回報新 block `0x50`、停在 `0x022E`、只跑 4 條指令 |
| `internal/ecl.TestRealPictureCommitCallPrecedesCombatSetup` | 真實 ECL2 block `0x02`：四步預算時還沒有 `CALL`／怪物設定，七步後 `CALL 2E10h` 在 `0x030A` 且怪物設定已出現；`PICTURE 0FFh` 不產生圖塊請求 |
| `internal/eclcatalog.TestPhaseLedgerHandlerAddressesMatchDispatchTable` | 台帳的 DOS handler 位址與 `ecl-opcode-dispatch.md` 一致 |
| `internal/eclcatalog.TestPhaseLedgerConfidenceMatchesEvidence` | 有 phase 就要有推論等級與規格引用，反之亦然 |
| `internal/eclcatalog.TestNewEclIsTerminalAndEndsStraightLine` | `20h` 同時是 `terminal` 且會切斷直線區域 |
| `internal/eclcatalog.TestExactlyOneCommitPointOpcode` | 提交點只有 `2Dh` 一個 |
| `eclcatalog.VerifyPhaseCoverage` | corpus 有的 opcode 台帳一定有列，台帳有的 opcode corpus 一定有 |
