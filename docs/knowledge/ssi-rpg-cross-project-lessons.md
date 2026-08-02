# SSI RPG 跨專案重製經驗：Gold Box、冬之魔與 Wasteland

狀態：初步比較，2026-08-02；只記錄目前本機 repository 可證明的事項。

## 結論

《冬之魔》值得作為工程與驗證方法的參考，但目前不能作為 Gold Box engine 的
第二個格式 consumer。《青色枷的詛咒》有 ECL bytecode、DAX／GEO 與 Gold Box
戰鬥／角色資料；《冬之魔》的 repository 已由 MZ header、3,807 筆 relocation
與 DOS `EXEC` 證明 `DEMON.INT` 是原生 8086 executable，不是 SSI 共用 VM。
它的 `DATA*.TXT` 是固定欄位事件資料，但條件與大量規則仍寫在 native code。

因此可共用的是「重製方法與低階作品中立能力」，不是直接共用故事 VM、戰鬥、
地圖索引或存檔 schema。

## 已可沿用的方法

| 經驗 | 對 CoAB／後續作品的用途 | 邊界 |
|---|---|---|
| `dwstrings uicheck` 的零硬編碼中文 gate | CoAB 應建立同類稽核，區分作品文字與作品中立 UI fallback | 不能只掃 Unicode；要允許測試、註解與 stable fallback 的明確白名單 |
| 500／500 資料文字 coverage | 每款 game-pack 都應量化 source identity、en／zh-TW parity、orphan／missing keys | 數量相等不證明翻譯正確，也不證明原始文本已全抽出 |
| 原版 EGA／CGA／Modern theme 分離 | CoAB 忠實 DOS／PC-98 oracle與未來美化 theme 必須完整分組、可切換且不改規則 | 不可把冬之魔 atlas index 或尺寸帶入 Gold Box |
| sampled A6／高風險 smoke | 日常採代表性玩家路徑，發行前才做完整長時間通關 | 抽樣不能支撐「完整可通關」聲明 |
| 發行 deny-list 與自備原版資料 | engine／中文資料可發布，原版資料與倚天字型不得夾帶 | 各作品仍需獨立授權與檔案清單 |
| deterministic replay、save overlay、唯讀 originals | 適合 Gold Box、冬之魔與未來 Wasteland | 存檔 layout 與 RNG consumer 不能跨作品猜測 |

第 452 輪已把第一項落地為 CoAB `cmd/coab-audit`：它不採寬鬆白名單，而以
AST＋hash exact baseline 先凍結 1,315 次現存債務，之後每批遷移只能同步更新
下降後基線。操作與分類限制見 [`原始碼資料分離稽核`](../audit/README.md)。
第 453 輪再以訓練場驗證此方法能落地：規則表只留 stable spell ID／locale key，
測試從正式資料解析期望顯示，基線下降至 1,251；不能用英文 fallback 或測試內
複製譯文假裝完成資料分離。
第 454 輪神殿又證明帶價格與效果的服務表也應只保存 typed rule＋stable key；
顯示名稱、格式與選項留在 locale，基線進一步下降至 1,223。
第 455 輪酒館則保留原英文 token／fragment 作來源身分，只讓映射結果指向
stable ID；這能同時維持逆向證據、ECL continuation 與單一繁中真相來源。

## 目前不可共用的內容

- ECL VM 與 `DEMON.INT`：一個是 bytecode interpreter input，一個是 native MZ code。
- DAX／GEO、`SUM.MAP`／`MAPn.MAP`：檔名、尺寸、索引、consumer 都不同。
- Gold Box AD&D 戰鬥與《冬之魔》的職業、符文、海戰、神祇及經濟規則。
- SAVGAM／角色 record 與 `PARTY.DAT`／事件資料 schema。
- 音訊：CoAB 有 DOS／PC-98 路徑；《冬之魔》原版沒有背景配樂，只有 PC speaker
  音效。不能把 remake 新配樂稱為原版引擎能力。

## 對獨立 Golden Box engine 的影響

`golden-box-remake-engine` 繼續只收已由 Gold Box 證據支持的 ECL／DAX／GEO／
combat contracts。冬之魔可拿來壓測更低層的 renderer、bitmap font、storage、
grid、deterministic random interface 與 data-source 邊界；但必須先做最小 adapter，
且不能出現產品名稱分支，才算真正第二 consumer。

## Wasteland 後續中文化入口

Wasteland 應建立獨立 repository、game-pack／adapter 與原始資料 inventory：

1. 先確認平台、executables、資料檔、手冊、存檔及合法來源雜湊。
2. 以 IDA／raw bytes／runtime 判定 native code、overlay、資料 interpreter 或 VM，
   不因同為 SSI 先套 Gold Box／冬之魔架構。
3. 先完成一個 source bytes→typed parser→繁中 UI→save round-trip vertical slice。
4. 再用相容性 checklist 判定 renderer、font、storage、grid、RNG 或格式 codec
   哪些能升入共用 engine／skill。
5. 同樣要求遊戲內手札／長文可重讀，不叫華文玩家離開遊戲查英文紙本。

## 後續知識庫與 skill 收斂

CoAB 發行前應把目前數百份 READY spec 收斂成可路由的繁中主題入口：

- executable／IDA 非破壞性證據與 confidence；
- ECL、DAX、GEO、save、角色與戰鬥；
- DOS／PC-98 視覺 oracle、640×480 CJK layout、倚天 16×15；
- 音樂／音效與時間軸；
- JSON data separation、翻譯 coverage、遊戲內手札；
- 玩家路徑抽樣、截圖、README 與發行驗證。

skill 只保存可重用流程、檢查表與工具入口；作品劇情、原文、地址與專有資料仍留
各 game repository。冬之魔的進一步權威細節位於其本機 `AGENTS.md` 與
`docs/design/engine-extraction-study.md`，本文件不複製會過期的 checkpoint。
