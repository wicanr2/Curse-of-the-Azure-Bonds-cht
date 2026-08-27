# Markdown 最後整合斷言稽核（2026-08-27）

狀態：完成。本輪不是按文件年代找「舊」字句，而是從最後整合可能讀錯的 gate
反向抽查：建置／runtime 依賴、發行狀態、驗證口徑、真相來源、手抄數字與絕對式
完成聲明。歷史文件只要有清楚封存邊界就保留，不把舊觀測改寫成今日事實。

## 已修正的錯誤斷言

| 類別 | 舊斷言 | 現場證據 | 修正 |
|---|---|---|---|
| 手抄數量 | README：1,180 份 spec、471 個 Go 檔／133,595 行 | 排除 `workplace/` 與巢狀 engine repo 後：`docs/spec/*.md` 1,223 份、Go 489 檔／141,042 行 | 更新數字並明寫數量不是完成度；不再聲稱每份 spec 都由數量本身證明有狀態 |
| 發行 gate | `remake-status`：三平台發行「沒有數字」，且完整可玩前不做 | `tools/package-release.sh` 已產 Linux x86_64、Windows x86_64、macOS x86_64／arm64；Linux 與 Wine smoke 已做 | 未量測列縮窄為 Windows／macOS 真機、簽署／公證及對外授權，不再把打包寫成未實作 |
| 工作入口 | `docs/knowledge/README.md`：`WORKLIST.md` 是目前執行順序 | `WORKLIST.md` 已於同日封存；目前 frontier 在完成度自評與當期 audit | 更新台帳對應，避免 compact 後重開已完成工作 |
| 音訊狀態 | 知識庫入口：播放生命週期尚未閉合 | `audio-lifecycle` 4／4、換曲點 13／13；21 個 runtime OGG、WAV 0 | 改為技術接線已閉合，留下人耳 loop 與散布權兩個不同 gate |
| 一般強度 | 「下一缺口是戰術／loadout」容易被讀成必須讓自動路線通關 | 2,455 幀後真實全滅；手動對照也全滅；使用者已決定真實全滅合法停止 | 不再由合法敗局自動產生戰術、loadout 或強制通關待辦 |
| 圖像切換 | 「runtime 改讀 PNG」未標完成態，可能被讀成已切完 | 程式仍由 `loadTileImages`、`loadCombatTerrainImages`、`loadAreaMapSymbols`、`loadSkyImages`、`loadMapPieceSets` 讀 DAX | 明寫 780 張 sprite PNG 已切換、六類圖像仍讀 DAX；PNG 獨立化仍是 release frontier |
| 文件權威 | `AGENTS.md` 第 9 節稱大量舊 checkpoint 為「目前權威狀態」；`CONTEXT.md` 稱自己只含現況 | 兩檔實際都保留大量帶日期的歷史 milestone | 加上全節封存契約；舊段的「目前／下一步／尚待」只能按段落日期解讀 |

## 經交叉驗證後保留的目前斷言

- 音訊資產：`assets/audio/` 有 21 個 OGG（12 首音樂＋9 個音效）、WAV 0；runtime
  正式主路徑使用 OGG，但 PC-98 driver fallback 與研究輸入仍保留。
- 圖像資產：`assets/` 有 787 張 PNG，其中 `assets/sprites/` 780 張；現行程式仍有
  六類原版圖像 DAX loader，因此不能宣稱 release 已圖像自足。
- 一般強度：`key-driven-normal-strength.json` 為 2,455 幀、251 格、8 段、199 句、
  fallback 0、真實全滅 1、`won=false`；這是合法停止，不是 scheduler 失敗。
- 三平台：開發封包可重生不等於 Windows／macOS 真機驗收，更不等於已取得原版
  衍生素材的對外散布權。
- 99% 是玩家可見體驗的產品目標與風險評分口徑，不是把不同分母平均出的覆蓋率，
  也不能豁免阻塞通關、存檔毀損、落回原文或主要規則錯誤。

## 文件處置結論

2026-08-27 第二輪依「是否有獨立證據與現行用途」重新判定後，另移除三份：

- `GOLDEN_BOX_RE.md`：只有開工前 token／成本猜測，且採用後來未使用的 SDL2 假設；
  沒有可回查原始證據，也不再參與規劃。
- 根目錄 `knowledge-router.md`：已被 Codex 外部知識路由取代，並指向不存在的舊 skill；
  留著反而會造成錯誤路由。
- `docs/audit/pc98-audio-lifecycle.md`：與正式生成物 `audio-lifecycle.md` 逐位元組相同；
  產生器、spec 與狀態表都只引用後者。

`WORKLIST.md`、`coab-remake-todo.md` 與 `docs/history.md` 仍保存獨立決策、勘誤或
玩家向歷史內容，維持封存而不刪。先前刪除的 `docs/project-status.md` 也維持移除，
因其現在式摘要已由 README、CONTEXT、完成度自評與可重生 audit 涵蓋。
