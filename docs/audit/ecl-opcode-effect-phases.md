# ECL opcode 有序副作用記錄（commit phase）

> 本檔由 `cmd/ecl-event-catalog` 由 `internal/eclcatalog/phases.go` 的表產生；請勿手動編輯。逐支反組譯依據見 `docs/spec/1104-ecl-opcode-ordered-effect-phases.md`。

## 通則

每個 handler 的第一個動作都是推進 ECL 程式計數器 `ds:4FB4h`：帶 operand 的呼叫 overlay-07 entry#2 一次吃掉 opcode 與全部 operand，不帶 operand 的用單一 `inc`。**效果一律發生在 PC 推進之後**，所以任何停下再續跑都從下一條指令開始，同一條指令不會被執行第二次。

- 本表的位址與結論來自 DOS `START.EXE`＋`GAME.OVR`。PC-98 對側位址見 `ecl-opcode-dispatch.md`；只有通則（operand 解碼推進 PC）與 `20h` 兩項逐條比對過，其餘未比對，不宣稱兩平台一致。
- `unknown` 表示該 handler 本輪未讀，不是「沒有副作用」。它是預設值，所以漏列一個 opcode 會被覆蓋率測試擋下，不會靜默沿用鄰居的分類。
- `remake_boundary` 描述 remake 怎麼建模，不是原作停在哪裡：原作同步巢狀呼叫的地方，remake 因為不可重入而做成可續跑交易。

## 逐 opcode

| opcode | 名稱 | DOS handler | PC 推進 | commit phase | 推論等級 | 規格 | 說明 |
|---|---|---|---|---|---|---|---|
| `0x00` | EXIT | `0052h` | `inc` | `terminal` | `exact` | `1104` | 設 `47E0h:=1` 結束迴圈，釋放整條 GOSUB 框架鏈，並消費 `47E2h` 還原角色指標。 |
| `0x01` | GOTO | `00E8h` | `assign` | `control_flow` | `exact` | `1104` | 解 1 個 operand 後把 `high:low` 合成的位址寫進 `4FB4h`。 |
| `0x02` | GOSUB | `0107h` | `assign` | `control_flow` | `exact` | `1104` | 解 1 個 operand 後交給 overlay-07 entry#24 推框架並跳轉。 |
| `0x03` | COMPARE | `011Eh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x04` | ADD | `019Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x05` | SUBTRAT | `019Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（與 `04h` 同一支）。 |
| `0x06` | DIVIDE | `019Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（與 `04h` 同一支）。 |
| `0x07` | MULTIPLY | `019Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（與 `04h` 同一支）。 |
| `0x08` | RANDOM | `0244h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x09` | SAVE | `02A2h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x0A` | LOAD CHARACTER | `02F0h` | `decode_operands` | `immediate` | `exact` | `1104`、`1098` | 先寫 `47E2h:=1` 再推進 PC；走 `650Ah` 名冊鏈選定角色寫 `6506h`，`8B6Dh` 記錄有沒有找到。`47E2h` 由 `00h`／`38h`／驅動器消費，屬另一條 deferred 線。 |
| `0x0B` | LOAD MONSTER | `0466h` | `decode_operands` | `immediate` | `exact` | `1104` | 配置怪物節點掛上 `650Ah` 鏈；`47E6h` 是本場已放置數，達 `3Fh` 就整支跳過。 |
| `0x0C` | SETUP MONSTER | `03CAh` | `decode_operands` | `immediate` | `exact` | `1104` | 寫 bank1 `580h`／`582h`，`582h` 取兩者最大值。 |
| `0x0D` | APPROACH | `0801h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x0E` | PICTURE | `0841h` | `decode_operands` | `deferred` | `exact` | `1104` | operand `0FFh` 走關閉分支並就地 flush；其餘只設 `8B49h`／`8B62h`／`8B63h`，畫面提交點在 `2Dh` 的 `2E10h`。 |
| `0x10` | INPUT STRING | `0972h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀；`INPUT STRING` 是可輸入、可退格、可確認的互動，不是預先列出答案的選單。 |
| `0x11` | PRINT | `09EAh` | `decode_operands` | `immediate` | `exact` | `1104` | 與 `12h` 同一支，靠重讀 `ds:75FFh` 分流；`4FC9h` 在輸出期間為 1。 |
| `0x12` | PRINTCLEAR | `09EAh` | `decode_operands` | `immediate` | `exact` | `1104` | 比 `11h` 多寫 `65A1h:=11h`、`65A0h:=1`，且輸出常式最後一個引數是 1 不是 0。 |
| `0x13` | RETURN | `0A86h` | `assign` | `control_flow` | `exact` | `1104` | 彈出 GOSUB 框架；框架鏈為空時直接呼叫 `00h` 的 handler。 |
| `0x14` | COMPARE AND | `0AD6h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x15` | VERTICAL MENU | `0EBDh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x16` | IF = | `0B3Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（`16h`..`1Bh` 同一支）。 |
| `0x17` | IF <> | `0B3Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（`16h`..`1Bh` 同一支）。 |
| `0x18` | IF < | `0B3Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（`16h`..`1Bh` 同一支）。 |
| `0x19` | IF > | `0B3Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（`16h`..`1Bh` 同一支）。 |
| `0x1A` | IF <= | `0B3Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（`16h`..`1Bh` 同一支）。 |
| `0x1B` | IF >= | `0B3Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（`16h`..`1Bh` 同一支）。 |
| `0x1C` | CLEARMONSTERS | `120Eh` | `inc` | `immediate` | `exact` | `1104` | `47E6h:=0`、`8B69h:=0`、`7603h:=8`，並沿 `6F8Ch` 鏈逐節點釋放。 |
| `0x1D` | PARTYSTRENGTH | `1271h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x20` | NEWECL | `0BBBh` | `decode_operands` | `terminal` | `exact` | `1104` | 把舊 ECL 編號存進 bank0 `1E4h`，載入新 block（overlay-07 entry#4）並初始化（entry#3，PC 重設為 `8000h`、重讀五個 lifecycle entry），最後 `47E0h:=1`、`47E1h:=1`；驅動器看到 `47E1h` 就從 lifecycle 頂端重跑。⇒ 後一個位元組永遠不會被執行。 |
| `0x21` | LOAD FILES | `0C15h` | `decode_operands` | `deferred` | `exact` | `1104` | 與 `37h` 同一支，靠重讀 `ds:75FFh` 分流；設 `47E3h:=1`、`47E5h:=1`。重繪要等 `47E4h` 與 `47E5h` 都非零。 |
| `0x24` | COMBAT | `179Ah` | `inc` | `pause_before_commit` | `exact` | `1095` | 原作是同步巢狀呼叫，也是商店／營地的服務分派點；remake 因為不可重入而做成可續跑交易。 |
| `0x25` | ON GOTO | `1A9Bh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（與 `26h` 同一支）。 |
| `0x27` | TREASURE | `1B53h` | `decode_operands` | `pause_before_commit` | `exact` | `255`、`257`、`258`、`558` | 寶物服務邊界，戰後 continuation 已閉合。 |
| `0x28` | ROB | `1F46h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x29` | ENCOUNTER MENU | `2086h` | `decode_operands` | `unknown` | `unknown` | `1083` | handler 尚未讀；`29h` 是 ENCOUNTER MENU 不是 PARLAY。 |
| `0x2A` | GETTABLE | `0E13h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x2B` | HORIZONTAL MENU | `1082h` | `decode_operands` | `pause_before_commit` | `exact` | `1104` | 先解 2 個固定 operand 取得目的位址與項數，`dec 4FB4h` 退一格後改以項數重解一次；選擇結果在 handler 內寫回目的位址。 |
| `0x2D` | CALL | `2F02h` | `decode_operands` | `commit_point` | `exact` | `1104` | operand 值減 `7FFFh` 後走七路 switch；`2E10h` 分支是 `8B62h`／`8B65h`／`8B67h`／`8B68h`／`8B6Ah` 五個髒旗標的唯一提交點。未列入 switch 的目標靜默 no-op。 |
| `0x2E` | DAMAGE | `2942h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀；運算元順序（旗標, 骰數, 面數, 加值, 豁免旗標）由公開參考支持，尚未逐條反組譯。 |
| `0x2F` | AND | `0DA4h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀（與 `30h` 同一支）。 |
| `0x31` | SPRITE OFF | `2C8Fh` | `inc` | `immediate` | `exact` | `1104` | 只在 `8B65h` 非零時做事，做完把 `8B65h`、`8B62h` 清零——旗標自清就是 exactly-once。 |
| `0x32` | FIND ITEM | `2847h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x33` | PRINT RETURN | `2CEAh` | `inc` | `immediate` | `exact` | `1104` | 兩條分支都寫 `65A0h:=1` 與 `inc 65A1h`，差別只在有沒有清 `8B61h`。 |
| `0x35` | SAVE TABLE | `0E71h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x36` | ADD NPC | `2DA9h` | `decode_operands` | `immediate` | `exact` | `1104` | 第二個 operand 除以 2 後或上 `80h` 寫進角色 `+0F7h`，再重繪名冊。 |
| `0x37` | LOAD PIECES | `0C15h` | `decode_operands` | `deferred` | `exact` | `1104` | 與 `21h` 同一支；設 `47E3h:=1`、`47E4h:=1`。兩個閂鎖都到齊才重繪，且兩者在進入新 block 時（overlay-02:3772h）一起被清掉。 |
| `0x38` | PROGRAM | `30DDh` | `decode_operands` | `terminal` | `exact` | `1104`、`1087` | 先消費 `47E2h` 才推進 PC。值 0 是即時重繪；值 8 是通關序列（問 Y／N，答 Y 才離開程式）；值 9 與值 3 都轉呼叫 `00h` 的 handler，值 3 另外設 `4FC7h:=1` 的全域停止。⇒ 只有值 3 與值 9 無條件終止。 |
| `0x39` | WHO | `2D5Eh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x3A` | DELAY | `28F3h` | `inc` | `immediate` | `exact` | `1104` | 推進 PC 後呼叫 resident 的延遲常式，沒有 ECL 記憶體副作用。 |
| `0x3C` | PROTECTION | `321Fh` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x3D` | CLEAR BOX | `2D15h` | `inc` | `immediate` | `exact` | `1104` | 就地清框並重畫，收尾 `8B6Eh:=0`、`8B60h:=1`。 |
| `0x3E` | DUMP | `3251h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x3F` | FIND SPECIAL | `3284h` | `decode_operands` | `unknown` | `unknown` | — | handler 尚未讀。 |
| `0x40` | DESTROY ITEMS | `32D8h` | `decode_operands` | `immediate` | `exact` | `1104` | 逐角色走 `+14Dh` 物品鏈，`+2Eh` 等於 operand 就移除，最後重繪該角色。 |

已讀 25／55 支，尚未讀 30 支。
