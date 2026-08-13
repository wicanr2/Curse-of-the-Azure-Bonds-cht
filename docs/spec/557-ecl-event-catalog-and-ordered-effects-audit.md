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

本輪不實作 ordered event runtime，不判定 IF／選單的真實分支，也不宣稱 33 個候選
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
- 1,355 個以 member／block／payload offset 去重的靜態可達 instruction。
- command framing、operand bytes、code address 與直接 graph edge。

以下只標為 `hypothesis／audit candidate`：

- effect kind 是為稽核建立的類別，不是原版資料欄位。
- 33 個候選只表示同一個靜態連續區段出現至少兩種 effect kind。
- `TraceGraph` discovery order 不是 runtime order；IF、`ON GOTO／ON GOSUB`、menu、
  CALL consumer、memory predicate、戰鬥結果與 resume 都沒有在這份清冊中執行。

## P0 ordered-effects 候選

> 第 558 輪勘誤：原列 P0-A 的三組 `TREASURE → COMBAT` 已由第 255／257／258
> 輪 READY contract 覆蓋，並補上 PC-98 IDA 與真實 DAX continuation 回歸；它們
> 現在由 review ledger 標成 `covered/exact`，不再是 blocker。下表保留原始排序
> 形成原因，但目前執行從 P0-B 開始。

33 個候選中，先依玩家結果風險分成下列工作類別：

| 優先級 | 代表候選 | 為何先驗證 |
|---|---|---|
| P0-A（已覆蓋） | ECL3 block `0x15 +050A..+0578`、ECL4 block `0x25 +1271..+12A7`、ECL6 block `0x45 +04F6..+0575` | 第 558 輪證明 pending treasure／battle／victory／resume 已由現行 transaction 閉合；不是新增 ordered runtime 的理由。 |
| P0-B | ECL2 block `0x02 +04BC..+053A` | 靜態上是 `COMBAT → text`，可檢查戰後文字是否依同一 runtime resume，而不是戰前一次套用。 |
| P0-C | ECL2 block `0x02 +02CB..+0325` | `text → PICTURE → CALL → combat setup`，可驗證畫面 snapshot、外部 routine 與戰鬥資料的 commit phase。 |
| P0-D | ECL4 block `0x25 +021F..+023B`、ECL5 block `0x30 +0086..+00B0` | `inventory → NEWECL` 與 `NEWECL → LOAD CHARACTER`，可檢查跨 block transaction 是否遺失前後副作用。 |
| P1 | ECL6 block `0x42／0x43` initial | `LOAD FILES／LOAD PIECES → CALL／text`，適合閉合 resource load、map adapter 與外部 routine 次序。 |
| P1 | ECL1 block `0x52` initial | 含人物加入、文字、CALL、戰鬥與 `PROGRAM` 的長序列；價值高但分支與篇幅大，應在小候選 contract 成熟後處理。 |

這些排序只決定下一個 probe，不把候選升格成正式語意。

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
- 統計固定為：6 members、25 blocks、125 entries、1,355 instructions、33 candidates。

## 尚未閉合

| 缺口 | 類型 | 是否阻塞玩家 | 下一步 |
|---|---|---:|---|
| effect 的全域有序 transaction model | 待逆向／待規格 | 是 | 三組 `TREASURE → COMBAT` 已由第 558 輪覆蓋；下一步閉合 P0-B `COMBAT → text` 與其餘未審查候選。 |
| 動態 `ON GOTO／ON GOSUB` 與 menu branches | 待逆向 | 是 | 把 runtime branch trace 合併回 catalog 的動態 edge 層，不覆寫靜態證據。 |
| external CALL registry | 待逆向 | 是 | 從 23 個靜態可達 CALL instruction 擷取 operand，逐址閉合 consumer。 |
| cell／terrain／劇情名稱對應 | 待研究／資料整合 | 視事件而定 | 由 GEO、ECL predicate 與正常路徑回填，不從攻略直接命名。 |
