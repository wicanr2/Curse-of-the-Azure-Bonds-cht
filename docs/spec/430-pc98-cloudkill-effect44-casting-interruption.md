# 第 430 輪：PC-98 毒雲術 effect 44h 的施法中斷

狀態：`READY`（限毒雲術落點造成的直接死亡與 pending spell 中斷）

## 問題與結論

第 429 輪只證明 `PUTDAMAGE` 的最終正傷害會中斷施法，尚不能判定不經
`PUTDAMAGE` 的毒雲術直接死亡。PC-98 原始控制流現在證明，毒雲術會加入 raw
effect `44h`；該 effect 在戰鬥模式有獨立 consumer，會在角色仍有 pending
spell 時顯示「未能完成吟唱法術」、消耗第一個 matching memorized spell byte，
再清除 pending spell。這條路徑不依賴正傷害。

因此 remake 在毒雲術落點使 HD 0–4 自動死亡，或使 HD 5–6 豁免失敗死亡時，
必須先建立施法中斷事件，再完成死亡 handoff。HD 7+、HD 5–6 豁免成功者不受
effect `44h` 影響，不得消耗法術格。

## 非破壞性輸入

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbols、resident stubs | `exact` |
| `PC98-GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay corpus | `exact` |
| overlay 12 | `7e1e8394652b02986da00944769331e0be7db7fb8476054fbe0369dbca05b9d9` | effect table、effect `44h` handler | `exact` |
| overlay 22 | `c54729525d576c11d731d64a1b06ee2547b2562b73e3708a1beaafc535cabbe8` | 毒雲術建立與 effect writer | `exact` |
| `pc98_cloudkill_interruption_audit.idc` | `330ca8f1a67ce25b95293ecbf5b98a25ca083db7ed56286c38621723d463022a` | 原始 offset／bytes／指令 ledger | 工具 |

原始 executable、overlay 與基準資料庫全程唯讀。IDA Pro 9.4 database 與完整
ledger 只建立於 `/tmp/coab-ida-430*`；版本化 IDC 不 rename、不 patch，輸出保留
未修改的 overlay-local offset、bytes 與組合語言。下列語意都是附加註記。
乾淨重建後 overlay 12／22 ledger 分別為 5,136／11,103 行；驗收同時檢查
輸出檔非空，不能只依 IDA exit code 0 判定腳本成功。

## Exact writer → table → consumer 鏈

### 毒雲術 writer

- overlay 22 local `52CBh` 是 Pascal short string
  `は毒雲を作り出した。`，可定位毒雲術 handler。
- 同一 handler 建立以落點為中心的 3×3 tile `1Ch`；local
  `5887h..59D6h` 的格內角色 callback 檢查並加入 raw effect `44h`。
- 因此 raw `44h` 與毒雲術受影響角色的 writer 關係是 `exact`，不是按文字
  或 spell 名稱猜出的對應。

### effect table 解析

- overlay 12 `INITEFFPROX` local `323Eh..3247h` 把 resident far pointer
  `008B:016Fh` 寫入 `DS:A150h`。
- effect table base 是 `DS:A044h`，每項四 bytes；
  `(A150h-A044h)/4 = 67`，即 zero-based slot 67、one-based raw effect `44h`。
- typed TPOV resolver 將該 resident stub 解析至 overlay 12 entry 67、local
  `1621h`。這條解析同時保存 table slot、control segment、resident stub 與
  overlay-local offset，不以相同數字跨位址空間猜 handler。

### effect `44h` consumer

overlay 12 local `1621h..16BFh`：

- `165Eh..1666h` 只在 combat mode `DS:7F27h == 5` 時清 Action `+01h`；
  該 raw byte的完整 typed 語意仍是 `unknown`。
- `166Eh..1677h` 要求 Action `+00h > 0`，即角色有 pending spell。
- local `1606h` 的 Pascal short string是
  `は呪文を唱えきれなかった。`，語意為「未能完成吟唱法術」。
- `1699h..16ABh` 把 Action `+00h` 的 spell ID 與目標 Player 傳給
  `014A:0070h`。
- 第 429 輪已 exact 解析 `014A:0070h` 為 overlay 24 entry 16、local
  `1739h`：掃描 Player `+1Eh..+71h`，清除第一個 matching memorized byte。
- `16B0h..16B8h` 最後清除 Action `+00h`。

這閉合「毒雲術 writer → raw effect `44h` → effect table → pending spell
consumer → memorized slot consumer」完整資料流，推論等級為 `exact`。

## Remake contract

- 共用 engine `combat/action.InterruptSpell` 已能原子清除作品中立 spell／target
  transaction；它不負責判斷哪種作品事件觸發中斷，因此本輪不需修改 engine。
- CoAB Battle 以單一 `interruptPendingSpell` 建立 stable fighter／spell event；
  正傷害與毒雲術 direct-death writer 共用事件格式，但保留兩個原作觸發點。
- `CastCloudkill` 只對實際 `Killed` impact 呼叫中斷：HD 0–4 自動死亡，HD 5–6
  豁免失敗死亡。中斷事件必須在 `SetHitPoints(0)` 清除死亡角色 Action 前建立。
- State 仍由正式 roster 按 stable fighter／spell ID 移除第一個 matching slot，
  並由 locale stable ID `combat_spell_interrupted` 解析訊息；測試不得複製譯文。
- 原始 effect handler 未清 Action delay，但 remake 的死亡 handoff 會清整個 typed
  Action。這是既有死亡狀態正規化，不可把死亡後 delay=0 反寫成原始 raw layout。

## 驗收與未完成邊界

- Battle regression：HD 4 pending caster 被毒雲術直接殺死時建立一筆中斷；
  HD 7 pending caster 保持未受影響。
- State regression：正式 roster `[Bless, Curse, Bless]` 只移除第一個 Bless，
  顯示文字由正式 locale stable ID 取得，死亡狀態完成。
- 第 357 輪的 3×3、HD 5／6 save 與 HD 7 無效果 regression 必須保持通過。
- Docker／Xvfb、`--network none`、本機 engine replace 的正式
  `./cmd/... ./gamepack ./internal/...` gate 共 31 套件通過；指定 Cloudkill
  core、玩家施法路徑與 stable-ID 中斷測試另以 `-count=1` 通過。

本輪沒有把規則擴張到沉默、麻痺、石化、睡眠或其他 effect；它們仍須各自找
writer、effect table 與 consumer。尚未完成的範圍還包括 monster memorized
slot raw writeback、毒雲每回合重複判定、protect-magic 例外、原版中斷訊息
停留時間／音效，以及可穩定重現敵方毒雲術的完整正常玩家動態 oracle。
