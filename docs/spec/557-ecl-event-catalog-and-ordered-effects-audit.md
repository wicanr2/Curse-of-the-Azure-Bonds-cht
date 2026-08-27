# 第 557 輪：ECL 全事件靜態清冊與有序副作用稽核

狀態：`READY`（限 corpus inventory／靜態候選）；ordered runtime semantics 仍為
`DRAFT`，不可據此宣稱 ECL 副作用完整。

## 問題與範圍

既有 corpus gate 證明 CoAB 六個 ECL DAX 的靜態可達 opcode 都有 command metadata，
但沒有留下每個 block／lifecycle entry／instruction／edge 的版本化清冊，也沒有系統性
列出跨類型副作用可能交錯的區段。這會讓後續工作在正常路徑卡住時才逐點補規格，
並容易把 parser coverage 誤寫成 runtime completion。

本輪只建立：

1. 可由原始 archive 重生的全 corpus JSON 清冊。
2. 人類可讀 block／entry 摘要。
3. 需要回到 bytes、IDA 與 runtime 驗證的跨 effect-kind 直線候選。

本輪不實作 ordered event runtime，不判定 IF／選單的真實分支，也不宣稱靜態候選
就是完整動態事件集合。

## R1：原始定位

| 輸入 | SHA-256 | 平台 | 工具／位址空間 |
|---|---|---|---|
| `curseoftheazurebonds.zip` | `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` | DOS | Go 1.24.13；DAX decoded block payload offset，ECL code address=`0x8000+offset` |

六個 member 的 SHA-256、25 個 block ID、decoded size 與五個 lifecycle entry address
全部保存在 [`ecl-event-catalog.json`](../audit/ecl-event-catalog.json)；摘要見
[`ecl-event-catalog.md`](../audit/ecl-event-catalog.md)。packed text operand 只保存
長度與 SHA-256，不複製原文 payload。

## R2：靜態證據邊界

`ecl.TraceGraph` 對每個 entry 分開執行，只追：

- 直接 `GOTO／GOSUB` target；
- 可解碼的 sequential fallthrough；
- `EXIT／RETURN／GOTO` 等已知靜態終點。

因此以下內容可標為 `exact`：

- 六個 DAX、25 個 block、125 個 lifecycle entry 身分。
- 14,177 個以 member／block／payload offset 去重的靜態可達 instruction。
- command framing、operand bytes、code address 與直接 graph edge。

以下只標為 `hypothesis／audit candidate`：

- effect kind 是為稽核建立的類別，不是原版資料欄位。
- 701 個候選只表示同一個靜態連續區段出現至少兩種 effect kind。
- `TraceGraph` discovery order 不是 runtime order；IF、`ON GOTO／ON GOSUB`、menu、
  CALL consumer、memory predicate、戰鬥結果與 resume 都沒有在這份清冊中執行。

## P0 ordered-effects 候選

> 候選層的審查結論有 33 筆（spec 1106 擴大可達性後，31 筆依效果序列沿用、
> 其中一組同時對到兩個候選）。目前 701 個候選中大多數尚未審查，
> 逐筆結論在 [`ecl-ordered-effect-reviews.json`](../audit/ecl-ordered-effect-reviews.json)。
> 閉合的方式不是逐候選，而是逐 opcode——次序是 dispatcher 與各 handler 的性質，
> 見 [spec 1104](1104-ecl-opcode-ordered-effect-phases.md)。

現在的待辦不在候選層，而在 opcode 層：55 個 corpus opcode 中 30 支 handler 尚未讀，
逐支狀態在 [`ecl-opcode-effect-phases.md`](../audit/ecl-opcode-effect-phases.md)。


## R3：清冊契約

工具：`cmd/ecl-event-catalog`

```sh
go run ./cmd/ecl-event-catalog \
  -output docs/audit/ecl-event-catalog.json \
  -summary-output docs/audit/ecl-event-catalog.md

go run ./cmd/ecl-event-catalog \
  -check docs/audit/ecl-event-catalog.json \
  -check-summary docs/audit/ecl-event-catalog.md
```

輸出契約：

- deterministic JSON，`format_version=2`；
- archive／member SHA-256；
- member→block→五個 lifecycle entry；
- 去重 instruction、operand metadata、effect-kind candidate、直接 graph edge；
- 每筆 instruction／edge／candidate 保存 `reachable_from`；
- generated Markdown 不手工修改，drift 由 `-check` gate 阻擋。
- candidate 使用 member／block／range 組成穩定 ID；人工審查只由獨立 review ledger
  附加，未知／漂移 ID 失敗即關閉。

工具與清冊不依賴中文顯示文字作斷言；JSON 不保存 packed text 原文。

## 驗證

Docker 內以 Go 1.24.13、暫存 `modfile` 將鎖版 private engine dependency 指向本機
唯讀 nested repo；正式 `go.mod/go.sum` 未修改。

通過：

- `go test ./internal/eclcatalog ./internal/ecl ./cmd/ecl-event-catalog`
- 連續生成與 `-check／-check-summary` byte-for-byte 相符。
- 統計固定為：6 members、25 blocks、125 entries、**14,177** instructions、**701** candidates。
  2026-08-27 已依 spec 1219 消除過期產物與重生閘門；要分母請用
  `cmd/ecl-effect-coverage`。
  （`20h NEWECL` 與立即值 3／9 的 `38h PROGRAM` 是終止指令，不切在它們後面會併出假候選；
  `IF` 條件不成立時跳過下一條，不走那條路會漏掉三分之二的程式碼——見 spec 1104 §九與 spec 1106。）

## 尚未閉合

| 缺口 | 類型 | 是否阻塞玩家 | 下一步 |
|---|---|---:|---|
| effect 的全域有序 transaction model | 局部 | 否 | phase 台帳已可重生且完整覆蓋 61 支可達 opcode；已讀 27 支、**34 支仍是 `unknown`**。台帳完整性已閉合，未讀 handler 的 commit phase 仍是證據限制，見 spec 1219。 |
| 動態 `ON GOTO／ON GOSUB` 與 menu branches | 待逆向 | 是 | 把 runtime branch trace 合併回 catalog 的動態 edge 層，不覆寫靜態證據。 |
| external CALL registry | ✅ 已完成 | — | 七支分派的位址（spec 561）與逐支語意（spec 1150）都已解出；可達 168 條、四個運算元。剩下的只有 `2E10h` 的髒旗標模型。 |
| cell／terrain／劇情名稱對應 | 待研究／資料整合 | 視事件而定 | 由 GEO、ECL predicate 與正常路徑回填，不從攻略直接命名。 |
