# CoAB 逆向工程閉合紀錄範本

每個會影響玩家結果的子系統、指令、外部 routine、資料欄位或事件群，必須以本範本
建立可獨立審查的紀錄。未填欄位保留 `unknown`，不可刪掉欄位來讓文件看似完成。

## 基本資料

| 欄位 | 內容 |
|---|---|
| Matrix ID／主題 | `[例如 ECL-EFFECT-ORDER]` |
| 本輪問題 | `[要排除的歧義或要重建的玩家行為]` |
| 限定範圍 | `[作品、平台、block、routine、資料表或畫面]` |
| 明確不宣稱 | `[本紀錄不能證明什麼]` |
| 文件狀態 | `DRAFT／READY／SUPERSEDED` |
| 推論等級 | `exact／strong inference／hypothesis／unknown` |

## R1：原始定位

| 輸入 | SHA-256 | 平台／版本 | 工具／版本 | 位址空間／載入基準 |
|---|---|---|---|---|
| `[檔名]` | `[hash]` | `[DOS／PC-98]` | `[IDA Pro 9.4 等]` | `[file offset／linear／segment:offset／overlay local／ECL work]` |

- 原始函式名／位址：`[保留原名與定位]`
- 連續 raw bytes／record 範圍：`[不可只列零散命中]`
- 可重生腳本與輸出：`[scripts/ida/...；成功判定與報告位置]`
- 交叉驗證：`[另一工具、原始資料或 runtime]`

## R2：原版語意閉合

### 資料流

```text
writer／producer
  → copy／projection／temporary record
  → consumer／formula／renderer／save writer
  → 玩家可見結果
```

| 階段 | 原始定位 | 已觀察行為 | 證據 | 等級 |
|---|---|---|---|---|
| writer／producer | `[位址]` | `[行為]` | `[bytes/xref/trace]` | `[等級]` |
| projection | `[位址]` | `[行為]` | `[bytes/xref/trace]` | `[等級]` |
| consumer | `[位址]` | `[行為]` | `[bytes/xref/trace]` | `[等級]` |
| 玩家結果 | `[狀態／畫面]` | `[行為]` | `[runtime/save diff]` | `[等級]` |

### 時序與副作用

- 前置狀態：`[flags、座標、party、RNG、clock]`
- 執行順序：`[ordered operations]`
- 暫停點：`[文字／選單／輸入／戰鬥／寶物／PROGRAM]`
- commit 時機：`[immediate／pause-before-commit／deferred／resume-only]`
- continuation：`[resume PC、block、owner]`
- exactly-once：`[如何避免 save/load、戰後或重訪重播]`
- 失敗／取消／戰敗分支：`[若適用]`

## R3：重建規格

- typed input／output：`[型別與範圍]`
- invariants／bounds：`[record count、stride、合法值、座標與錯誤處理]`
- 亂數與時間：`[來源、snapshot、draw count、clock transaction]`
- 未知值策略：`[fail-closed／保留 raw／unsupported boundary]`
- READY spec：`[docs/spec/...]`
- 被取代規格／斷言：`[SUPERSEDED 連結與推翻原因]`

## R4：engine＋CoAB 資料

| 層級 | 實作／資料 | 驗證 |
|---|---|---|
| 共用 engine | `[作品中立 schema/runtime]` | `[套件測試]` |
| CoAB game pack | `[stable ID、JSON、原始來源]` | `[schema/audit]` |
| State adapter | `[只作投影，不硬編劇情]` | `[integration test]` |
| renderer／audio | `[若適用]` | `[deterministic capture/trace]` |
| save | `[版本與欄位]` | `[round-trip/mutation diff]` |

產品層測試的顯示文字必須由 stable ID 與正式 JSON 解析，不可複製中文或英文
literal 作為另一份真相來源。

## R5：玩家與原版驗證

| Gate | 情境／輸入 | 證據 | 結果 |
|---|---|---|---|
| 原版 runtime | `[DOSBox／PC-98 狀態]` | `[trace/screenshot/save hash]` | `[通過／缺]` |
| 正常玩家路徑 | `[新遊戲或正式 save 起點]` | `[無 direct-entry 的測試／trace]` | `[通過／缺]` |
| continuation | `[pause/combat/treasure 後]` | `[resume PC/state]` | `[通過／缺]` |
| save/load | `[儲存點與重載後動作]` | `[state diff]` | `[通過／缺]` |
| 離開重訪 | `[區域與旗標]` | `[前後結果]` | `[通過／缺]` |
| 原版/remake 對照 | `[同狀態、座標、seed、theme]` | `[圖／逐幀／音訊]` | `[exact/nearby/layout-only]` |

## 尚未閉合與下一步

| 缺口 | 類型 | 是否阻塞玩家 | 下一個最小可重現動作 |
|---|---|---:|---|
| `[缺口]` | `待逆向／待動態驗證／待規格／待實作` | `[是／否]` | `[明確動作]` |

完成本紀錄後，必須同步：

1. 更新 `coab-re-coverage-matrix.md` 對應列。
2. 更新根目錄 `WORKLIST.md`，移除已被證據推翻或關閉的 blocker。
3. 搜尋 `docs/`、`AGENTS.md`、`CONTEXT.md`、測試名稱與程式註解中的舊斷言。
4. 將衝突規格標為 `SUPERSEDED`，保留其原始證據與訂正原因。
5. 只有 R1–R5 在限定範圍閉合後，才可把該列標為 `閉合`。
