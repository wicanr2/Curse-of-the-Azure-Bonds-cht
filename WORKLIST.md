# 《青色枷的詛咒》目前工作清單

更新日期：2026-08-10（第 537 輪後盤點）

本檔是 compact、交接與每輪開工時的目前摘要入口。詳細的反組譯證據仍在
[`docs/knowledge/golden-box-reverse-engineering-worklist.md`](docs/knowledge/golden-box-reverse-engineering-worklist.md)；
可驗證的歷史與每輪規格見 [`docs/project-status.md`](docs/project-status.md)。
本檔只保留目前有效的工作，不把歷史輪次的舊 blocker 當成現況。

## 一句話結論

重製尚未完成整作通關。現在已經有多條可重播的正常玩家垂直切片，並完成
`SEARCH`／`LOOK`、wall=09 候選橋接、E2、火刀 E1、戰後世界地圖與 save/load
的 engine＋JSON 接線；還缺火刀據點完整逐房間路徑、全 ECL／全地圖、完整戰鬥、
音樂音效、全量繁中校對、完整存檔相容與三平台發行。

## 狀態與證據規則

- `已完成（remake contract）`：重製程式、JSON、測試與玩家路徑已閉合；不代表
  原版每個 byte 已逐一證明。
- `exact`：原始 bytes 與 consumer／runtime trace 已閉合。
- `strong inference`：多項證據一致，但仍少一段原版資料流或 runtime oracle。
- `待實作`：目前玩家路徑或產品功能仍未完成。
- `待研究`：只有在要支援該功能或原版 fidelity 時才逆向；不可先把假說寫入
  正式規則。

## 第 537 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| engine＋JSON 的 `SEARCH`／`LOOK` 分離 | `S` 持續切換 `DungeonSearchEnabled`，`L` 是一次性 `LookDungeonLocation`；地圖 edge、external exit、locale 與 save v12 均有資料契約。這是重製契約已完成；原版 Search 成功率與 wall writer 仍未 exact。 |
| 下水道至 E2 | 從 `(13,10)` 正常逐格移動，經 wall=09 候選橋接抵達 `(8,15,S)`，再由 E2 進 ECL2 block 4；未用直接設定座標完成這條重製路徑。 |
| 火刀 E1 回返 | block 4 北側 E1 候選可由正常移動越界，回到下水道 `(10,15,N)`；三個 E1 座標仍屬 `strong inference`。 |
| 戰後 handoff 與存檔 | 首領勝利後的 ECL 夢境、Tilverton 世界地圖選單與 Search／edge 狀態 save/load 已有回歸；首領測試的固定戰鬥 fixture 不等於從入口逐房間打到首領。 |
| 既有資料分層 | 開場選項、玩家法術入口與多個事件已使用 stable ID、locale JSON 與 engine resolver；不能因此宣稱全量中文化。 |
| UI 基線 | 640×480、原版裂紋石框、倚天粗體 16×15、人物 HEAD／BODY 分層與多張對照截圖已建立；目前多為 `layout-reconstructed`，不是所有狀態逐像素 exact。 |

權威規格：[`docs/spec/537-search-look-e2-fire-knife-normal-route.md`](docs/spec/537-search-look-e2-fire-knife-normal-route.md)。

## 剩餘工作總表

### P0：先讓主線繼續走，不再用座標輔助掩蓋缺口

| 工作 | 現況 | 下一個可驗收成果 |
|---|---|---|
| 火刀據點完整正常路徑 | 已能由 E2 進入 block 4，並能驗證部分安靜／刀刃事件與 E1 回返；入口到首領仍不是完整逐房間正常路徑。 | 從 `(6,1,S)` 的正式 ECL continuation 出發，逐房間以 `MoveDungeon`／互動／戰鬥抵達首領；每段保存事件旗標、戰後 continuation、寶物與重訪結果。 |
| 火刀據點出口、返回世界地圖與重訪 | 世界地圖與戰後 save/load 的切片已通；完整出口條件、移除 Bond 後的封閉／開放狀態、重訪與失敗分支仍需接齊。 | 新建一條不注入座標的「入口→首領→出口→世界地圖→存檔→重載→重訪」測試，並以 game-pack stable ID 驗證旗標與文字。 |
| 開場到結局的正常玩家主線 | ECL1–ECL6 的許多 block 與窄事件已可讀、也有多個正常路徑切片；完整章節 handoff、外部 routine、分支與結局尚未串完。 | 依章節建立可中斷／可存檔的 vertical slice，最後由新隊伍或正式角色檔從開場走到結局；未支援的 boundary 必須明確 fail-closed。 |

### P1：補齊可玩規則、資料與原版行為

| 工作 | 目前缺口 | 驗收方向 |
|---|---|---|
| 全 ECL 與外部 routine | opcode corpus 已有大量 `READY` 邊界，但 `CALL`、`NEWECL`、地圖服務、劇情旗標、NPC 離隊、輸入與 continuation 仍有未接 consumer。 | 只逆向會改變玩家結果的 producer→state→consumer；每個完成事件都要有 raw bytes／runtime trace、JSON contract、stable ID 測試與正常輸入路徑。 |
| 全地圖與世界旅行 | 目前只有多個 GEO／ECL 區段和世界地圖切片；仍缺全 GEO／AREA／WILDERNESS、所有入口出口、地形事件、持久 map state 與錯誤出生點稽核。 | 建立全地圖可達性與邊界表；逐區以原始資料和正常移動驗證，不把攻略座標直接寫成規則。 |
| 戰鬥規則、AI、法術效果與動畫 | 已有部分 AD&D 數值、敵方選敵、延後施法與 12 個玩家法術入口；仍缺完整敵我 AI、遠程／弓箭、法術逐項效果、saving throw、持續區域、死亡動畫、回合節奏與音效 cue。 | 對近戰、弓箭、Magic Missile、Fireball、Lightning Bolt、Stinking Cloud／Cloudkill 等分開驗收 windup、travel、impact、damage、save、death、persistent effect、聲音與 ECL handoff；影片只能證明演出，數值要回 bytes／DOSBox。 |
| 存檔、角色規則與跨遊戲轉移 | remake save v12 已保存 Search／edge；DOS／PC-98 `SAVGAM`、角色 sidecar、完整 record、年齡／職業／特殊能力、刪除／改名與 `MOVEPARTY` 跨遊戲 transfer 尚未完整 round-trip。 | 先完成版本化 parser／serializer 與 save mutation diff，再以角色檔跨 Gold Box 來源做 stable transfer contract；不能把 `MOVEPARTY` 靜態 helper 直接當秘密門。 |
| 全量繁體中文化與遊戲內手札 | 多個事件、選項、手札與攻略已資料化；仍缺全 ECL／物品／法術／怪物／地名／UI 字串、中文校對、長文分頁與所有原版需查手冊的內容整合。 | 以 stable `message_id` 做 coverage／orphan／source-drift audit；玩家可在遊戲內讀手札，不要求查外部文件才能通過事件。 |
| 音樂與音效 | YM2203、S98、PC98 sound BIOS、cycle PCM 等 engine 知識與部分合成測試已有；完整 DOS／PC-98 producer、播放生命週期、音效與戰鬥事件同步仍未完成。 | 先完成每個場景／戰鬥 cue 的資料綁定與可重播播放，再用 DOS／PC-98 runtime 對照 phase、音量、音效次序；合成器測試不能冒稱硬體 exact。 |
| UI、素材與原版 fidelity | 目前有 640×480、石框、字型、人物舞台與數張截圖；冒險／戰鬥／地圖／對話／頭像的所有狀態仍未逐張比對，palette cycle、sprite timing、左上場景填格與 PC-98 密度仍需抽樣校準。 | 每張對照標示平台、狀態、save／seed、theme 與 `exact`／`nearby`／`layout-only`；原版 theme 與美化 theme 分開，先完成原版忠實驗收。 |

### P2：完成後才做的發行工作

| 工作 | 門檻 |
|---|---|
| Windows／Linux／macOS 打包 | P0 主線、P1 規則／資料／音訊與存檔通過後，才做三平台可重現 build、資產授權檢查、存檔位置與首次啟動 smoke。 |
| README／截圖／40–60 秒推廣片 | 截圖只使用目前版本可重播狀態；推廣片只在可玩整合完成後製作。8 小時錄影不是本專案目標。 |
| 日後美化 theme 與 donate | 原版忠實 theme 永久保留；美化與 donate 只作後續／local 設定，donate 資訊不得上傳 GitHub。 |

## 仍需逆向，但不應阻塞目前 remake 路徑的項目

這些是原版 parity 或跨作品知識庫工作，不是重新打開第 537 輪已接通的路徑：

1. `wall=09` 第三平面 before／after writer、Search 成功率、原版 E1 精確座標與
   同版重訪 trace。現有 CoAB edge 是 `strong inference`，不要擴大成所有
   `wall=09` 都可走。
2. DOS `CALL 2E10h` 到 map service 的 selector producer、`DS:7206h`／
   `DS:7212／7213` consumer、`C04B..C04F` projection 與原版 redraw／位置 trace。
   這是原版 fidelity 稽核；除非某個玩家結果被阻塞，不再逐行追無關 overlay。
3. PC-98 `MOVEPARTY` 的角色轉移 selector／record／save round-trip。中文說明書
   已證明產品功能邊界，但尚未證明每個 raw helper 與 transfer record 的一對一
   runtime 對應。

## 明確不做的事情

- 不為了湊「完整反組譯」而逐行解讀與玩家結果無關的 function。
- 不把 `BDF1`、`SEARCHREC`、`MOVEPARTY`、相同十六進位數字或單一 xref 重新命名成
  秘密門、detail、年齡、旗標或地圖 owner。
- 不以 direct-entry、固定座標、注入戰鬥、測試模式或窄測試宣稱完整通關。
- 不在 JSON 尚未成為真相來源前，把劇情文字、裝備、法術或測試期望值硬編碼回 Go。
- 不在遊戲完整可玩前花時間做三平台 release 或長篇推廣影片。

## 完成聲明的共同驗收門檻

至少要通過：

1. 新隊伍／正式角色能由開場以正常輸入走到結局，包含移動、互動、裝備／使用、
   戰鬥、治療／休息、存檔、退出、重載與一個後期任務重訪。
2. 所有已宣稱支援的內容由 CoAB JSON／locale 與 engine contract 驅動；未支援
   行為明確失敗即關閉，不以 fallback 假裝完成。
3. 原版／remake 的畫面、動畫、音樂、音效、規則與存檔比較都有證據等級；近似畫面
   不標成 pixel-perfect。
4. Docker 內完成受影響套件、代表性正常玩家路徑、save round-trip、截圖／包裝 smoke；
   再集中 commit＋push 兩個 repository。

下一個最小可重現工作：從已完成的 E2 `(8,15,S)`／block 4 `(6,1,S)` session
開始，接通火刀據點第一個尚未完成的逐房間 ECL boundary；完成後再更新本檔、
`docs/project-status.md`、詳細 worklist 與 `CONTEXT.md`，不要先擴大到無關反組譯。
