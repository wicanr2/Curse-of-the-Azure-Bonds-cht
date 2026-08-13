# 《青色枷的詛咒》全遊戲逆向工程完整度矩陣

狀態日期：2026-08-13

## 結論

目前知識庫足以支援多個真實資料垂直切片，但**尚不足以作為整部遊戲的完整重建
規格**。後續先補齊本矩陣的玩家可見系統，再擴張 remake；不再以 opcode 能解析、
地圖已宣告、固定 fixture 通過或單張畫面完成，代替整個系統的語意閉合。

本矩陣只要求重建玩家會觀察到的規則、流程、資料、畫面、聲音與存檔結果。與玩家
結果無關的編譯器樣板、配置器、錯誤訊息或 raw work address 不必逐函式深挖。

## 閉合等級

每個系統都要分開記錄五層；不能從較低層直接跳成「完成」。

| 代號 | 層級 | 完成條件 |
|---|---|---|
| `R1` | 原始定位 | 輸入雜湊、平台、工具版本、位址空間、raw bytes／record 已保存。 |
| `R2` | 原版語意 | writer→projection→consumer 或等價 runtime trace 已閉合，且有推論等級。 |
| `R3` | 重建規格 | READY spec 定義資料型別、時序、副作用、錯誤行為及未知邊界。 |
| `R4` | engine＋資料 | 作品中立機制與 CoAB JSON 分離，stable ID、schema 與回歸測試完成。 |
| `R5` | 玩家驗證 | 正常輸入抵達，含 continuation、save/load、離開重訪及必要原版對照。 |

狀態只使用：

- `閉合`：本列限定範圍已達 R1–R5。
- `局部`：只有列出的子集合閉合。
- `待逆向`：缺 R1／R2，不能先猜進正式規則。
- `待規格`：原版證據大致足夠，但尚未整理成完整 R3。
- `待實作`：R1–R3 足夠，缺 R4／R5。
- `待動態驗證`：主要缺原版 runtime 或正常玩家 R5。
- `不阻塞`：不影響玩家可見結果，暫不深入。

## 全域 P0

### P0-RE-1：ECL 有序副作用與 exactly-once

目前 `RunResult` 依類型保存 CALL、SPELL、DAMAGE、NPC、CLOCK、TREASURE 等結果，
State 再按固定類別順序套用。這不能保證重建原始 opcode 的跨類型先後順序，也是
「單點測試通過、正常流程仍卡住」的首要嫌疑。

第 558 輪已排除一個錯誤 blocker：三組 `TREASURE → COMBAT` 早已有 pending
treasure→combat／service→victory→loot→resume 的第 255／257／258 輪 READY
transaction，本輪以 PC-98 IDA 與三段 DOS 真實 DAX continuation 補強並在清冊標成
`covered/exact`。這不消除全域缺口；只是下一個 probe 應改查真正未審查的
`COMBAT → text`，而不是為已閉合案例重寫 State。

閉合要求：

1. 從原始 ECL trace 建立 opcode 當下的 ordered effect record。
2. 每筆 effect 標明 immediate mutation、pause-before-commit、deferred service 或
   resume-only。
3. `PICTURE／TEXT／CALL／NPC／DUMP／CLOCK／TREASURE／COMBAT／NEWECL／PROGRAM`
   的交錯順序不因分類而改變。
4. pause／resume、戰後續跑、save/load 與重訪不重播已提交副作用。
5. 以至少一條真實的多類型 ECL 正常路徑建立原版與 remake trace diff。

目前證據入口：
[`gold-box-ecl-command-set.md`](gold-box-ecl-command-set.md)、
[`../spec/540-ecl-map-combat-audio-corpus-closure.md`](../spec/540-ecl-map-combat-audio-corpus-closure.md)。

### P0-RE-2：全遊戲事件清冊

由 ECL1–ECL6 的 25 個 block／125 個 lifecycle entry 產生版本化清冊。每個靜態與
動態可達事件都要列出 entry、PC 範圍、座標／terrain、條件旗標、選項、外部 routine、
戰鬥／寶物／Journal、寫入欄位、resume PC、重訪結果與目前 R1–R5。corpus parser
gate 只能作輸入，不等於此清冊已完成。

第 557 輪已完成第一層靜態清冊：6 個 DAX、25 個 block、125 個 lifecycle entry、
1,355 個不重複靜態可達 instruction 與 33 個跨 effect-kind 直線候選均可由原始 archive
重生。仍缺動態 branch、座標／terrain、條件旗標、consumer、resume 與 R1–R5 回填，
所以 P0-RE-2 維持 `局部`，不能標成全事件閉合。

### P0-RE-3：文件證據可重現性

- 所有現行 spec 必須有明確 `READY／DRAFT／SUPERSEDED` 與限定範圍。
- IDA 腳本必須由至少一份規格引用，並說明輸入、輸出及成功判定。
- `/tmp` 報告若是結論必要證據，改成可重生的版本化摘要；不提交 `.i64` 暫存資料庫。
- `gold-box-state.md` 與逐輪 RE worklist 降為歷史／主題筆記，不再擔任完成度真相來源。

## 系統完整度矩陣

| 系統 | 目前層級 | 判定 | 仍缺的重建知識／證據 | 下一份閉合產物 |
|---|---|---|---|---|
| 原始檔與平台 inventory | DOS 主檔多有 hash；PC-98 VFD 有缺 sector | 局部 | 建立 DOS／PC-98 executable、overlay、DAX、GEO、save、音訊與手冊的單一 manifest；標示 pristine／derived | `coab-source-manifest` |
| DAX container／壓縮 | 多種真實資產已可抽取 | 局部 | 對所有實際成員補 record count、bounds、round-trip 與 malformed gate；區分不同 DAX payload | `dax-corpus-matrix` |
| ECL framing／控制流 | 第 557／558 輪已版本化 6 DAX／25 block／125 entry／1,355 instruction 靜態清冊與穩定 candidate ID | 局部 | 33 個候選中 3 個已審查；其餘 30 個與動態 branch、間接 dispatch、外部 boundary、錯誤路徑仍缺；graph 不代表 runtime | `ecl-event-catalog` 靜態層已完成；續建動態 edge／事件 metadata |
| ECL 副作用／時序 | typed signal、多個 adapter；3 個 `TREASURE → COMBAT` 候選已 `covered/exact` | **待逆向／待規格** | 從 `COMBAT → text` 開始閉合其餘 30 個候選，再完成全域 ordered event log、commit phase、exactly-once、跨 boundary continuation | `ecl-ordered-effects` READY spec＋trace corpus |
| External `CALL` | `2E10／C01E／B200` 等有局部證據 | 待逆向 | 實際使用地址全集；每址的 caller、operand、state projection、consumer、返回與未知 fallback | `external-call-registry` |
| `NEWECL／PROGRAM` | boundary ID 與部分 context 已知 | 局部 | 全 context 的 area/resource/map/save/ending 副作用與 resume ownership | `program-newecl-context-matrix` |
| GEO 幾何／四平面 | 16 個原始 block 已宣告；loader／部分 plane consumer 有證據 | 局部 | 所有 plane 欄位、wall/door/roof/terrain interaction、wrapped edge、視覺 consumer | `geo-block-and-cell-schema` |
| 地城移動／門／SEARCH／LOOK | 正常切片與 wall=09 候選已接通 | 局部 | 普通門、鎖門、bash/knock、秘密通路、成功率、方向、時間與重訪的原版 transaction | `dungeon-movement-action-matrix` |
| AREA map／地圖顯示 | 有資料與局部畫面 | 待逆向／待實作 | player marker、探索狀態、秘密區、Journal 59 圖、縮放／色盤與 save state | `area-map-contract` |
| 世界旅行 | 14 點 arrival 與有向可達基線 | 局部 | 所有路線選單、途中遭遇、TRAIL/WILDERNESS/EXIT、取消、回返與 persistent state | `overland-route-event-matrix` |
| 城市／場所服務 | inn/store/bar/temple/training 等有切片 | 局部 | 每城可用服務、價格／治療／訓練／謠言／時間／旗標與離開後狀態 | `city-service-matrix` |
| 劇情與全地圖事件 | 多條正常主線及大量固定 fixture | **待逆向** | 每區逐格／逐事件 producer、條件、分支、副作用、重訪；完整正常 session 到結局 | `area-event-coverage` |
| Journal／Tavern Tales | 59 則手冊已重建；31 則有資料綁定 | 局部 | 剩餘 producer、解鎖條件、原圖、重讀、存檔與不提前劇透 | `journal-producer-matrix` |
| 角色 record／創角 | 核心欄位、年齡、部分職業／法術已知 | 局部 | 全種族職業、multi/dual class、能力修正、年齡效果、HP、alignment、icon 與原版 RNG | `character-rule-record-matrix` |
| DOS save bundle | raw-preserving parser/writer 與部分 sidecar | 待逆向 | 完整 `SAVGAM`、`.SAV/.GUY/.FX/.SWG` schema、角色刪除重排、未知 byte consumer、round-trip | `dos-save-bundle-schema` |
| remake save | 目前 session/PRNG/edge 可保存 | 局部 | 所有 pending ECL/combat/audio/UI transaction、版本遷移、晚期 save/reload | `remake-save-state-matrix` |
| 跨遊戲角色轉移 | 手冊證明產品功能；raw helper 局部 | 待逆向 | source selector、record conversion、裝備／法術／等級限制、round-trip | `move-party-transfer-contract`；不阻塞 CoAB 單作通關 |
| AD&D 共通規則 | 部分 AC/THAC0/save/level/spell 公式已接通 | 局部 | 完整職業／種族限制、能力修正、升級、休息、時間、負重、狀態與 item special consumer | `coab-rules-coverage` |
| 戰鬥 scheduler／initiative | 有部分 typed core 與 PC-98 研究 | 待逆向 | 原版 round/segment、held/delayed action、surprise、flee/guard/quick、死亡與戰後 handoff | `combat-turn-lifecycle` |
| 戰鬥地圖／LOS／placement | tile、footprint、扇區、掃描順序有局部 exact | 局部 | 全地形阻擋、門、尺寸、遮擋、移動代價、free attack、所有方向與出生規則 | `combat-map-rule-matrix` |
| 近戰／弓箭／投射物 | 命中核心與基本箭矢路徑存在 | 待逆向／待動態驗證 | 武器速度／射程／彈藥／多攻、逐幀 projectile、sound、impact、death、continuation | `physical-attack-matrix` |
| 玩家法術 | 12 個 handler；5 個有 visual binding | 局部 | 全法術表、target、range/area、save、duration、stack/dispel、visual/sound/death/continuation | `player-spell-matrix` |
| 敵方 AI／怪物能力 | 選敵與少量特殊能力有研究 | **待逆向** | 移動、目標優先、施法、逃跑、群體、抗性、免疫、毒素、凝視、每種特殊能力 | `monster-ai-and-specials-matrix` |
| 戰利品／商店／物品 special | inventory/treasure 有多個切片 | 局部 | TAKE/SHARE、負重、金錢換算、random item、識別、詛咒、使用次數與所有特殊 consumer | `item-treasure-service-matrix` |
| 戰鬥動畫／效果 | SPRIT／CPIC 解碼與部分 effect timeline | 局部 | 近戰、箭、12 法術、死亡、area persistent effect 的原版逐幀／phase oracle | `combat-presentation-matrix` |
| DOS 音效 | 9 WAV 與部分 selector intent | 局部 | 全 caller、缺 selector、場景、優先權、重疊、停止、存檔與原版時序 | `dos-sound-cue-matrix` |
| PC-98 音樂／音效 | YM2203/S98/driver 研究深入 | 待動態驗證／待實作 | 缺 sector 的合法來源、真實 SFX caller、save/resume、reload phase、gain 與全場景 cue | `pc98-audio-lifecycle-matrix` |
| 圖像 asset codec | PIC/BIGPIC/HEAD/BODY/CPIC/SPRIT 多數可解 | 局部 | 全資產 manifest、selector producer、palette cycle、動畫 identity 與 malformed gate | `visual-asset-corpus` |
| UI／renderer fidelity | 640×480、石框、16×15、部分舞台已完成 | 待實作／待動態驗證 | 所有 adventure/combat/map/dialog/roster/shop/spell/save/ending 狀態逐張 contract | `screen-state-fidelity-matrix` |
| 繁體中文 | 字型與部分 stable locale 已接通 | 待實作 | 全 ECL、UI、物品、法術、怪物、場所、Journal、攻略與詞彙校對 coverage | `zh-tw-content-coverage` |
| 開場到結局 | 正常新遊戲 session 已到散提爾堡世界選單；另有終戰 fixture | **待逆向／待驗證** | 後續章節、所有必要事件、最終戰、`PROGRAM 8`、結局、存檔與重開 | `campaign-spine-to-ending` |

## 區域事件 coverage

這張表只描述正常玩家可達性與事件覆蓋，不能用 map declaration 取代。

| 區域 | 目前證據 | 判定 | 必補內容 |
|---|---|---|---|
| 提爾佛頓／下水道／火刀據點 | 主線與多個正常地城切片 | 局部 | 所有可選房間、失敗分支、寶物、秘密通路、重訪與存檔重載 |
| 阿沙本福德／立石群 | 同一正常主線有局部城市與事件 | 局部 | 全場所、全部談話／謠言、替代路線與重訪 |
| 艾森布拉 | 正常主線到城外；城內多為 fixture | 待逆向 | 城內全場所、事件 producer、離城及後續旗標 |
| Hap／熔岩洞／巫師塔 | 正常主線局部已通 | 局部 | 全村民、可選房間、所有戰鬥／談判／失敗與重訪 |
| 希爾斯法 | 多為固定 fixture | 待逆向 | 正常進出城、碼頭、伏擊、場所、所有分支與世界 handoff |
| 尤拉什／摩安德之坑 | 固定與局部正常測試 | 待逆向 | 全樓層、NPC、坑內事件、戰鬥、出口、重訪與 save |
| 散提爾堡／暗神殿／眼魔洞穴 | 正常主線已到洞穴離場 | 局部 | 全城市／洞穴事件、隨機事件、替代分支與重訪 |
| Myth Drannor | Burial Glen 等多個 vertical slice；終戰 fixture | 待逆向 | 正常世界入口、全三區房間、主線旗標、最終戰與結局同 session |

每個區域最終必須產生逐事件記錄：`map/block/cell/facing/terrain → ECL entry/PC →
preconditions → ordered effects → continuation → flags/items/journal/combat → leave/re-enter →
save/reload`。

## 文件稽核發現

2026-08-12 唯讀盤點得到 535 份 `docs/spec/*.md` 與 63 支 `scripts/ida/` 腳本。
這代表研究量龐大，但不能轉換成完成百分比。現存風險包括：

- 大量早期 spec 沒有一致的機器可讀狀態與限定範圍。
- 部分必要報告只指向 `/tmp`；應保存可重生摘要與產生方式。
- 下列腳本目前未被文件按檔名引用，需判定為補引用、合併或 archived：
  `dos_map_workcell_raw_audit.idc`、`pc98_overlay14_pre_move_audit.idc`、
  `pc98_overlay24_generic_audit.idc`、`dos_overlay07_movement_audit.idc`、
  `pc98_tickclock_effect_audit.idc`、`pc98_overlay7_specials_audit.idc`、
  `pc98_scan_terrain_audit.idc`。
- `gold-box-state.md` 與 `golden-box-reverse-engineering-worklist.md` 是高價值歷史筆記，
  但篇幅與時間序列使它們不適合作為目前完成度入口。

## 第一批執行順序

1. 靜態 `ecl-event-catalog` 已完成；三組 `TREASURE → COMBAT` 已由第 558 輪閉合並
   寫入 fail-closed review ledger。下一個 probe 是 ECL2 block `0x02` 的
   `COMBAT → text` 戰後續跑。
2. 將已驗證的動態 edge、條件與 continuation 回填 catalog，逐步完成
   `ecl-ordered-effects` 規格；未審查候選維持 unknown，不因相似序列批次升格。
3. 建立 `external-call-registry`，從 23 個靜態可達 CALL 只追玩家可見副作用的
   producer→consumer。
4. 建立 `area-event-coverage`，把清冊與 GEO cell／terrain／正常路徑合併；先盤點，
   不立刻補 Go 特判。
5. 再依矩陣順序閉合戰鬥、存檔、音訊與畫面；只有 R1–R3 足夠的項目才交給
   engine／JSON 實作。

每項工作使用
[`re-closure-record-template.md`](re-closure-record-template.md)；同一 milestone
必須更新本矩陣、`WORKLIST.md` 與被推翻的 spec。矩陣不採用模糊百分比，只有
R1–R5 的逐層證據可以提升狀態。
