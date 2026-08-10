# 第五百三十三輪：PC-98 `MOVEPARTY` helper gate 與結果欄位邊界

狀態：`READY`（靜態 raw bytes／連續控制流；不宣稱秘密門或正常路徑）

日期：2026-08-10

## 本輪目的

第 528 輪已閉合 `MOVEPARTY` 的 action token、`AL=1／2／3` 分流與共同續跑點，
但尚未檢查 B／P／K helper 的完整範圍。本輪只追查這些 helper 是否有可回查的
map writer、角色／物件欄位 writer 或其他結果條件，讓下一個 runtime boundary
不會把 helper 名稱誤當成秘密門語意。

## 輸入與工具

| 證據／工具 | SHA-256／版本 | 位址基準 |
|---|---|---|
| `workplace/ida406/pc98-overlays/overlay-14.bin` | `a8e03ba9a5381c3a9f7ab411ced3262b21e0b65b948160d614386d677610e7b9` | overlay-local code offset |
| `scripts/ida/pc98_overlay14_action_helpers_audit.idc` | `24b022ed516001fe40522e974087ccac9ed21a72b5f9fd6ad90a65d117d00499` | IDA disposable raw database |
| IDA Pro | `ida-pro-9.4-ver3:latest`／IDA Pro 9.4 | 8086 overlay-local database；原始資料庫唯讀 |

原始 overlay 與基準 `.i64` 沒有被改寫；IDA 只在 Docker 內的 disposable database
上建立指令項目。range label 是導覽標籤，不是函式改名或已證實語意。

## Exact 前置 gate 與 result flow

`MOVEPARTY` `00C9:0BCCh` 的連續 bytes 顯示：

```text
0x0BDB  mov AL, [DS:7F27h]
0x0BDE  cmp AL, 04h
0x0BE2  不相等時跳至 0x0D77
0x0BE5  les DI, [DS:7F09h]
0x0BE9  cmp word ES:[DI+592h], 00FFh
0x0BF0  小於時繼續，否則跳至 0x0D6Ch
0x0BF5  mov byte [DS:0A66Ch], 01h
0x0C06  call far 017C:0039h
0x0C0B  AL=01h 時寫入 local success byte=01h
0x0C16  AL=02h／03h 時進入共同後續流程
```

上述是 raw operation／branch 的 `exact` 證據。`DS:7F27h` 的正式模式名稱、
`DS:7F09h` 指標所指 record、`+592h`／`0A66Ch` 的欄位語意，以及
`017C:0039h` 的產品語意仍是 `unknown`；不能把 `+592h` 直接命名成搜索次數、
秘密門狀態或角色欄位。

當 `ES:[DI+592h] >= FFh` 時，`0x0D6Ch` 會以 `DS:7F09h` 取得同一 far pointer，
將 `ES:[DI+592h]` 清成零，再進入後續 far-call／資料清理。這證明清除操作，
不證明它是「找不到門」或「搜索完成」。

## B／P／K helper 的可回查操作

### B helper：local `0x02F5`

- 讀取目前候選 far pointer 的 `ES:[DI+111h]`，必要時呼叫 `15Dh:0052h`。
- 以 `ES:[DI+18Ah]` 載入下一個 far pointer，保留原始 pointer chain。
- 使用 `DS:0A2A9h／0A2AAh／0A2ABh` 作為呼叫參數，呼叫 `017C:0039h`，再依
  `AL` 比較值更新 local success byte／暫存旗標。
- 在 local `0x0566` 與 `0x05A4` 直接呼叫 set writer `0x014C`。

### P helper：local `0x05B4`

- 由 `DS:9598h` 取得候選 far pointer，讀取候選 record 的 `+0EBh`、`+196h`。
- 使用 `13Eh:004Dh` 與 `017C:0039h` 相關呼叫結果判斷 local success byte；
  這些 far-call 的正式名稱仍未由本輪證明。
- 以 `ES:[DI+18Ah]` 延續 pointer chain。
- 在 local `0x062F` 與 `0x066Dh` 直接呼叫 set writer `0x014C`；第二次呼叫
  前的座標／方向算術是 exact raw operation，但不命名成另一側門格。

### K helper：local `0x0714`

- 呼叫 local `0x067Dh` 取得候選 far pointer。
- 當返回值不是 `FFh` 時，將該候選 record 的 `ES:[DI+1Eh]` 清成零，並將
  local success byte 設為一；否則設為零。
- 這是 exact field write／return branch；`+1Eh` 的正式欄位名稱仍是 `unknown`。

### 共同結果 helper：local `0x078Ch` 與 `0x0807`

local `0x078Ch` 會再次呼叫 far `017C:0039h`。若結果非零，依候選 record 的
`+489Eh／+48A7h` 加到暫存座標，將 `DS:0A2A9h／0A2AAh` 限制在 `0..0Fh`，並在
`DS:7F09h` 指向的 record 寫入 `ES:[DI+5AAh]=1`。這些是 exact writer／clamp
操作；`+489Eh`、`+48A7h`、`+5AAh` 與非零結果的正式語意仍未知。

local `0x0807` 則進入另一段 far-call／角色或物件資料處理；本輪只保留其原始
位址，不把它命名成地圖重繪、開門或事件續跑。

## 與第三平面 writer 的關係

本輪確認 B／P helper 的四個 set-writer caller 仍是：

```text
0x0566 → 0x014C
0x05A4 → 0x014C
0x062F → 0x014C
0x066D → 0x014C
```

`0x014C` 對 `THE3DMAP+300h` selected 2-bit field 寫入 raw `01`，而
`0x003E` 的 clear writer 仍由 movement-result 分支呼叫。這只閉合
helper→writer 的靜態 call-site；沒有同版本 runtime 的 before／after bytes，
不能把 raw `01` 命名成 detail `1`、開門、另一側同步或可通行。

## 對 remake 的限制與 probe 結果

- 本輪沒有新增 `secret_door`、`search`、movement predicate 或 ECL flag JSON。
- 本輪沒有把 `(13,10)`→`(8,15)` 寫入正常玩家路徑。
- 在 remake 已正常抵達 `(13,10)` 並完成騎士事件後，曾以一次性 probe 呼叫
  現有 `SearchDungeonLocation`；Docker focused test 通過且沒有產生出口事件。
  這只表示目前 remake ECL／Search 介面沒有現成秘密門行為，不是 DOS／PC-98
  原版 runtime 證據；probe 已還原，沒有留在 regression。

下一個真正能升級到 engine／JSON 的證據仍是：同版本 DOS／PC-98 oracle 中，逐一
記錄 action token、`017C:0039h`／helper return、`THE3DMAP+300h` 前後 bytes、
`BLOCKCODE` 結果、重訪／save-load 與 ECL continuation。未取得前保持 fail-closed。

## 第五百三十四輪勘誤

中文說明書印刷頁 3–4 明列 `Move characters where?` 與四個 SSI Gold Box
跨遊戲角色轉移方向，印刷頁 12 的 `ADD CHARACTER` 又列出 `CURSE／POOL／
HILLSFAR` 來源；完整證據見
[`docs/spec/534-chinese-manual-moveparty-character-transfer.md`](534-chinese-manual-moveparty-character-transfer.md)。
因此本規格的 raw bytes、call-site、欄位寫入與 `unknown` 等級仍有效，但把
`MOVEPARTY` 放入秘密門／地圖 P0-2 的語意路由已 `SUPERSEDED`：目前應優先把
它視為角色資料轉移工具候選。騎士事件後 `(13,10)`→`(8,15)` 的地圖 handoff
仍需另找 wall interaction／external map consumer，不能用本規格的 helper 取代。
