# 1095 — ECL 主迴圈與 `COMBAT` 的續跑：戰鬥是同步巢狀呼叫，`COMBAT → text` 之間沒有提交邊界

- 狀態：`READY`
- 證據等級：`exact`（DOS 主迴圈 40 條、dispatcher 分支、COMBAT handler 203 條逐條讀完；
  PC-98 對側 205 條逐條讀完並逐項對照）
- 作法見 spec 783；助憶碼表見 spec 1083；畫面模式見 spec 892／1045
- 對應 P0-RE-1（`coab-re-coverage-matrix.md` 全域 P0）的第一個 `COMBAT → text` 候選

## 範圍

| 角色 | DOS（主檔） | PC-98（語意參考） |
|---|---|---|
| ECL 主迴圈 | `overlay-02:03621h`（`retf 2`） | 對側同結構 |
| opcode dispatcher | `overlay-02:03377h` | `overlay-02:0373Eh` |
| `24h` `COMBAT` handler | `overlay-02:0179Ah`（`retf`） | `overlay-02:01820h`（`retf`） |
| lifecycle 驅動器 | `overlay-02:03691h` | 對側同結構 |

實例：`ECL2.DAX` block `0x02` 的 `0x04BC`..`0x053A`
（catalog 候選 `ECL2.DAX/0x02/0x04BC-0x053A`，盜賊公會首領戰的戰後段）。

## ★★★ 一、ECL 主迴圈只有 40 條指令

```pascal
procedure 跑ECL(起點PC: word);           { overlay-02:03621h，retf 2 }
begin
    DS:4FB4h := 起點PC;                  { ★ ECL 程式計數器 }
    DS:47E0h := 0;
    while (DS:47E0h = 0) and (DS:4FC7h = 0) do begin
        DS:75FEh := DS:75FFh;                                  { 上一個 opcode }
        DS:75FFh := byte[DS:4FA5h^ + DS:4FB4h − 8000h];        { ★ 取目前 opcode }
        if DS:4FC1h <> 0 then begin                            { 追蹤開關 }
            <overlay-34 entry#3>(DS:75FFh, @緩衝, @DS:8B76h);  { spec 1083 的助憶碼表 }
            <0A54:173Bh>(0);  <0A54:1673h>();                  { 印出來 }
        end;
        <dispatcher 3377h>();                                  { 執行一條 }
    end;
    DS:47E0h := 0;
end;
```

> ★★★ **迴圈本身不推進 PC。** `DS:4FB4h` 只在進入時被設成起點，之後
> **由每個 opcode handler 自己遞增**——`COMBAT` handler 的第一條就是
> `inc word ptr ds:4FB4h`（單 byte 指令 ＋1），帶 operand 的 handler 則自己跳過 operand。
> ⇒ remake 不能在迴圈層統一 `pc += 指令長度`；長度是 handler 的責任。

> ★★★ **ECL 位址空間的基底是 `8000h`。** 取指令算式 `DS:4FA5h^ + PC − 8000h`
> 解釋了 catalog 的 `code_address` 為何一律是 `0x8000 ＋ offset`。
> ★ 而 `DS:4FA5h` 指向的是 **decoded block ＋ 2**：實測 `ECL2.DAX` block `0x02`
> 解出的位元組裡 `0x04BE` 才是 `24h`，catalog 的 offset 是 `0x04BC`。
> 前 2 個位元組（`88 13`）不屬於程式碼，其後 20 個位元組是 5 個 lifecycle entry
> （每筆 4 bytes ＝ `01 01` ＋ word），word 依序是 `8093h`／`8133h`／`80DBh`／
> `8122h`／`8014h`，與 catalog 的 5 個 `lifecycle_entries` 一一對上。
> ⚠ 本規格不宣稱 `88 13` 這 2 個位元組的用途。

★ **`DS:4FC1h` 是原作留下的指令追蹤開關**，開啟時每執行一條就用
`overlay-34 entry#3`（spec 1083 的 65 路助憶碼表）把指令名印出來。
⇒ **spec 1083 那張表的使用點就在這裡**，它是原作自己的除錯輸出，不是註解殘留。

★ 兩個停止條件分工不同：`DS:47E0h` 由本次迴圈自己清零與檢查（單次 lifecycle 的離開），
`DS:4FC7h` 是跨迴圈的全域停止。
⚠ 本規格不宣稱這兩格分別由誰寫入。

## 二、lifecycle 驅動器連跑三段，每段之間檢查中止

`overlay-02:03691h` 依序做：

```pascal
DS:47E1h := 0;
跑ECL(DS:4FB2h);   if DS:47E1h <> 0 then 離開;     { 第一段 }
…畫面模式相關的條件重繪…
DS:47E1h := 0;
跑ECL(DS:4FAAh);   if DS:47E1h <> 0 then 離開;     { 第二段 }
跑ECL(DS:4FACh);   if DS:47E1h <> 0 then 離開;     { 第三段 }
```

★ `DS:47E1h` 是**跨段的中止旗標**，與迴圈內的 `DS:47E0h` 是兩格不同的東西。
⚠ 本規格不宣稱 `DS:4FB2h`／`4FAAh`／`4FACh` 三個 PC 分別對應哪一個 lifecycle 名稱。

## ★★★ 三、`24h COMBAT` 是同步巢狀呼叫

DOS `0179Ah` 與 PC-98 `01820h` 逐項對應：

| 動作 | DOS | PC-98 |
|---|---|---|
| 遞增 ECL PC | `inc ds:4FB4h` | `inc ds:7F21h` |
| 前置旗標（非 0 就跳過整段服務分派） | `ds:8B69h`、`ds:8B56h` | `ds:0BDFBh`、`ds:0BDE8h` |
| bank 指標 | `ds:4F9Dh` | `ds:7F09h` |
| 另一 bank | `ds:4F99h` | `ds:7F05h` |
| 畫面模式 | `ds:4FBAh` | `ds:7F27h` |
| 收尾清旗標 | `ds:8B62h := 0` | `ds:0BDF4h := 0` |
| 收尾重繪 | `overlay-24 entry#37` | 同 |

### 三之一、先做服務分派，再打

```pascal
inc DS:4FB4h;                                  { PC 已指向下一條 }
if (DS:8B69h <> 0) or (DS:8B56h <> 0) then goto 打;

if bank1^[6D8h] = 1 then begin                 { 商店 }
    bank1^[6D8h] := 0;
    <overlay-06 entry#3>();  <overlay-06 entry#1 @06F6h>();   { 商店版營地選單 }
    …依 bank0^[1CCh] 呼叫 overlay-27 entry#6 或 overlay-30 entry#11…
    goto 收尾;
end;
if bank1^[5C4h] = 1 then begin                 { 營地 }
    bank1^[5C4h] := 0;
    <overlay-04 entry#2>();  <overlay-04 entry#1 @0F42h>();   { 營地主選單 }
    …同一組重繪…
    goto 收尾;
end;
```

> ★★★ **`24h` 不是只有「打一場」——它同時是服務分派點。**
> 兩個旗標各自被消費一次（讀到 1 就立刻寫 0），對應的服務跑完就直接跳到收尾，
> **這一回合不進戰鬥**。⇒ 這是 exactly-once 的原生機制：旗標由 handler 自己清。

### 三之二、戰鬥路徑

```pascal
打:
    <overlay-08 entry#2>();  <overlay-10 entry#2>();  <overlay-09 entry#2>();
    <overlay-13 entry#34>();  <overlay-32 entry#23>();  <overlay-31 entry#7>();   { 載入 }
    n := <overlay-07 entry#6>(DS:0A2A9h, DS:0A2AAh, DS:0A2ABh);
    if n > bank1^[582h] then bank1^[582h] := n;      { ★ 取最大值 }
    <overlay-08 entry#1 @0143h>();
    <overlay-05 entry#2>();                          { 初始化 }
    <overlay-05 entry#1 @1775h>();                   { ★ 戰鬥主體 }
    …依 bank0^[1CCh] 呼叫 overlay-27 entry#6 或 overlay-30 entry#11…
    if DS:0A325h <> 'P' then <overlay-29 entry#9>('y');
```

★ `overlay-05:01736h` 是設 `DS:4FBAh := 6` 的那一支，與戰鬥主體 `01775h` 相鄰
⇒ 戰鬥期間畫面模式是 **6**。
⚠ 本規格不宣稱 `bank1^[582h]`（取最大值的那格）與 `DS:0A2A9h`..`0A2ABh` 三個參數的語意。
⚠ 本規格不宣稱 `DS:0A325h = 'P'` 的判斷用意（形狀上與 spec 1083 的
`'~COMBAT ~WAIT ~FLEE ~PARLAY'` 有關，但未追到寫入端）。

### ★★★ 三之三、收尾把畫面模式還原成地城的 3 或 4

```pascal
收尾:
    if DS:8B56h <> 0 then DS:8B56h := 0;
    if bank0^[1CCh] <> 0 then DS:4FBAh := 4 else DS:4FBAh := 3;
    bank1^[594h] := bank1^[594h] and 1;
    FillChar(DS:8B48h, 2, 0);
    DS:8B62h := 0;
    <overlay-24 entry#37>();                        { 依新模式重繪 }
    retf;
```

> ★★★ **這條算式與 spec 1045 的地城主迴圈一字不差**
> （`DS:4FBAh := 4; if bank0^[1CCh] = 0 then DS:4FBAh := 3;`）。
> ⇒ `COMBAT` 收尾做的是**把畫面模式從戰鬥的 6 還原成地城的 3／4**，
> 並用 `overlay-24 entry#37` 依新模式重繪。
> ★ `DS:4FBAh` 是既有結論的**畫面模式**（spec 892 起）：
> 1 ＝ 營地／商店、2 ＝ 存讀檔、3／4 ＝ 地城兩種、5 ＝ 角色檢視分支、6 ＝ 戰鬥。

## ★★★ 四、`COMBAT → text` 的續跑：沒有提交邊界

以實例 `ECL2.DAX` block `0x02` 為例（catalog offset，`+2` 才是解出位元組的位移）：

| offset | opcode | 指令 | 內容 |
|---|---|---|---|
| `0x046B` | `09h` | `SAVE` | `0Ah` → `7F79h`（★ 進入條件的第一步） |
| `0x0477` | `0Ah` | `LOAD CHARACTER` | `7F79h` ⇒ 把**第 10 個角色**設成目前角色 |
| `0x047B` | `03h` | `COMPARE` | `7D00h`, `0` ⇒ 讀該角色的 `+196h`（角色欄位投影，spec 1098） |
| `0x0481` | `16h` | `IF =` | |
| `0x0482` | `01h` | `GOTO` | `0x84BC` ⇒ **成立才跳到 `COMBAT`** |
| `0x04BC` | `24h` | `COMBAT` | 單 byte |
| `0x04BD` | `12h` | `PRINTCLEAR` | 40 bytes：`THE GUILDMASTER GASPS, 'ON BALANCE, I'D RATHER BE IN ` |
| `0x04E8` | `11h` | `PRINT` | 21 bytes：`YULASH,' AND THEN HE DIES. ` |
| `0x0500` | `11h` | `PRINT` | 42 bytes：`YOU FIND INFORMATION ON HIS BODY AND LOG IT AS JOURNAL ` |
| `0x052D` | `11h` | `PRINT` | 6 bytes：`ENTRY 4.` |
| `0x0536` | `02h` | `GOSUB` | → `0x8D04` |
| `0x053A` | `01h` | `GOTO` | → `0x8CCB` |

★ **`COMBAT` 不是無條件執行的**：先載入第 10 個角色、檢查 `+196h = 0` 才跳過去
（`7D00h` 是角色欄位投影，不是 bank1 記憶體——見 spec 1098 §三）。

執行序列：

1. `COMBAT` handler 開頭把 PC 從 `0x84BC` 推到 `0x84BD`。
2. handler 內部**同步跑完整場戰鬥**（`overlay-05 entry#1`），期間畫面模式是 6。
3. handler 收尾把畫面模式還原成 3／4 並重繪，然後 `retf`。
4. dispatcher `retf`，回到主迴圈。
5. 主迴圈檢查 `DS:47E0h` 與 `DS:4FC7h`——**兩者都沒有被 `COMBAT` handler 寫過**——
   繼續取 `PC = 0x84BD` 的 opcode `12h`，執行 `PRINTCLEAR`。

> ★★★ **`COMBAT` 與其後的四條文字指令是同一個 `while` 迴圈裡的連續五次迭代，
> 中間沒有任何 pause、commit 或 resume 邊界。** 戰鬥是巢狀的同步呼叫，
> 打完就從 handler 返回，PC 早在戰鬥開始前就已經指向下一條。
> ⇒ 原作不需要「戰後 resume PC」這種機制——**PC 從未離開過**。

### 對 remake 的意義

remake 跑在 Ebiten 事件迴圈上，無法在 opcode handler 內阻塞整場戰鬥，
所以必須用 continuation 模擬。**語意等價的要求是**：

1. 戰鬥請求發出時，PC 必須**已經**指向下一條指令（不是戰鬥結束後才推進）。
2. 戰鬥結束後恢復執行，必須從**同一個迴圈狀態**繼續，不得重跑 `COMBAT` 那一條。
3. `COMBAT` 之後的 `PRINTCLEAR`／`PRINT` 與 `COMBAT` 本身**屬於同一批有序副作用**，
   不因為中間插入了一場戰鬥就分成兩個交易——原作根本沒有分。
4. 服務分派（商店／營地）與戰鬥是 `24h` 的**三選一**，且旗標讀到就清
   ⇒ exactly-once 由清旗標保證，重入時旗標已是 0，自然走戰鬥路徑。
5. 畫面模式的還原（6 → 3／4）與重繪是 `COMBAT` 的一部分，
   必須在文字輸出**之前**完成，否則 `PRINTCLEAR` 會畫在戰鬥畫面上。

⚠ **目前 remake 的落差**（`internal/ecl/runtime.go:960`）：
`0x24` 的三選一是 `memory[0x7F6C]` → Shop、`memory[0x7EE2]` → Temple。
原作 DOS 的兩格是 `bank1^[6D8h]` → **商店**、`bank1^[5C4h]` → **營地（Camp）**，
第二格呼叫的是 `overlay-04 entry#1`（營地主選單，spec 1030），不是神殿。
★★★ **映射已由 spec 1096 解出**，兩格精確對上：

| remake | 算式 | 原作 | `24h` handler 的分支 |
|---|---|---|---|
| `memory[0x7F6C]` | `(7F6Ch−7C00h)×2 = 6D8h` | `bank1^[6D8h]` | 商店 ✓ |
| `memory[0x7EE2]` | `(7EE2h−7C00h)×2 = 5C4h` | `bank1^[5C4h]` | **營地（Camp）**，不是神殿 |

> ★★★ **remake 選的兩個位址與原作映射一致；但第二格的語意標成 Temple 與原作不符**
> ——`bank1^[5C4h]` 那一支呼叫的是 `overlay-04 entry#1`（營地主選單，spec 1030）。
>
> ★★★ **這兩格在 1,355 條 ECL 指令裡沒有任何一條寫過**（已全掃確認）。
> ⇒ 它們是**引擎寫入、ECL 讀取**的共用格子，正是 spec 1096 §五第 2 點指出的
> 最高風險類別：map 寫入永遠成功，對不上時不會有任何錯誤訊息。
>
> ★★★ **查完了：語意有等價物，那一格則是一條沒有 producer 的死路。**
>
> - 「營地跑完 ECL 再回來」在 remake 走的是**另一條路**：`EnterDungeonCamp` 跑
>   lifecycle entry 2（`pre_camp`）、把結果交給 `applyDungeonLifecycleResult`，
>   再開營地選單；中斷走 entry 3（`RunCampInterrupted`）。機制與原作不同
>   （生命週期入口 vs `24h` 旗標），**效果等價**。
> - `0x7EE2` 在 remake 的正式程式碼裡**沒有任何一處寫入**（ECL 那一側本來就沒寫過）。
>   所以 `TempleRequested` 目前走不到，語意標成 Temple **不會產生錯誤行為**。
>
> ⚠ 它仍然是個會誤導下一個人的名字。`TestCampRequestFlagHasNoProducer` 釘住
> 「沒有 producer」這件事：哪天有人補上寫入端，就必須先決定那一格是營地還是神殿。

## 明確不宣稱

- 沒有宣稱 decoded block 前 2 個位元組（`88 13`）的用途。
- 沒有宣稱 `DS:47E0h`／`DS:4FC7h`／`DS:47E1h` 分別由誰寫入。
- 沒有宣稱 `DS:4FB2h`／`4FAAh`／`4FACh` 各對應哪一個 lifecycle 名稱。
- 沒有宣稱 `bank1^[582h]`（取最大值）與 `bank1^[594h]`（`and 1`）的語意。
- 沒有宣稱 `DS:0A2A9h`..`0A2ABh` 三個參數與 `overlay-07 entry#6` 的介面。
- 沒有宣稱 `DS:0A325h = 'P'` 的判斷用意。
- 沒有宣稱 `overlay-05 entry#1 @1775h`（戰鬥主體）的內部行為——本規格只確認
  它是同步呼叫、返回後 PC 不變。
- 沒有宣稱戰鬥中全隊死亡或逃跑時是否有別的路徑寫入 `47E0h`／`4FC7h`
  ——handler 本身沒有寫這兩格。
- 沒有宣稱 `memory[0x7EE2]`（remake）對應原作哪一格。
- 沒有宣稱 `PRINT` 之後 `GOSUB 0x8D04`／`GOTO 0x8CCB` 的目標內容
  （Journal entry 4 的實際登錄端未追）。
