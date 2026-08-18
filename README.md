# 《青色枷的詛咒》繁體中文化／Remake

SSI《Curse of the Azure Bonds》的資料驅動重製專案。目標是讓玩家以繁體中文從
開場玩到結局。**目前是可執行的多垂直切片，尚未完整通關**。

本 repo 負責 CoAB 的 game pack、翻譯、原始素材轉換、攻略與整合測試；可重用的
ECL／DAX／GEO／戰鬥／存檔引擎在獨立的
[golden-box-remake-engine](https://github.com/wicanr2/golden-box-remake-engine)。
劇情、座標、物品與翻譯資料不寫死在共用 engine。

## 目前狀態（2026-08-17 實測）

| | |
|---|---|
| 反組譯覆蓋 | 2,874 個函式全部進台帳，**待解讀 0**（已解讀 2,137／不阻塞 162／邊界碎片 575）|
| 規格 | 1,110 份 `docs/spec/*.md`，各自標證據等級與「明確不宣稱」邊界 |
| 程式 | 303 個 `.go`／93,874 行；`internal/game` 沒有區域或劇情專屬檔 |
| ECL 文字 | 控制流可達 1,022 頁，**未接上 0** |
| ECL 副作用 | 可達 14,177 條指令，**只讀掉運算元 0**、部分完成 1,057 |
| 戰鬥法術 | 可宣告 **73／73** 全部有 handler、視覺與音效 |
| 存檔 | 角色記錄 422 bytes：decoded 294／documented 99／unknown 29 |
| UI 詞條 | game pack `zh-TW` 1,363 條、`assets/locale/zh-TW.json` 917 條 |
| 全套測試 | `./tools/go.sh test ./...` 全綠 |

### 已接通的正常玩家路徑

同一個新遊戲 session 可從開場走到：提爾佛頓開場 → 城市設施 → 城門／盜賊公會 →
下水道 → 火刀據點至首領 → 戰後世界地圖 → 阿沙本福德 → 立石群 → 艾森布拉城外 →
Hap／熔岩洞／巫師塔 → 散提爾堡 → **眼魔洞穴（手札 59、Dexam 雙戰、東門離場）**。

路徑之後的區域仍會遇到未完成的功能。

### 角色建立

原版的四段選單（種族 → 性別 → 職業 → 陣營）已完整重現，資料全部取自原作
資料段並轉成 JSON：

- 六個可選種族——原作的顯示迴圈沒有收半獸人的分支，所以建角選不到它，
  但編號仍維持原作的 1..7 不重新連號。
- 職業選項由種族查表；陣營選項再由職業組合查表（聖騎士只有守序善、
  遊俠只有三個善）。
- 屬性擲點是 `3d6 + 1`、每個屬性擲六次取最大，加上七個種族各自的屬性調整。
- 多職角色的 HP 是八個職業槽各擲一次生命骰、逐槽體質加值後除以職業數。
- 名字支援中文（退格按字元不按位元組），`Save <名字>?` 答 `N` 角色不留。

另有一套範本流程供快速開局，兩者並存。

## 畫面預覽

以下是目前 remake 的代表畫面，皆在 Docker／Xvfb 產生並逐張檢視。石框與 88×88
場景是原始素材證據；640×480 的中文延伸與完整戰鬥畫面仍是 `layout-reconstructed`，
不能據此宣稱整作完成。

![旅店人物事件：DOS 石框、HEAD／BODY 舞台與繁中敘事](docs/screenshots/gold-box-layout-adventure.png)

![原版四段建角的名字輸入段](docs/screenshots/guided-creation-name.png)

![繁中戰鬥：兩隊各在視野一側，地形與戰鬥員走同一條相機座標](docs/screenshots/gold-box-layout-combat.png)

![提爾佛頓第一人稱：朝北的走廊，原版 88×88 場景內框與右側隊伍欄](docs/screenshots/tilverton-first-person-remake.png)

![Burial Glen 紅網戰鬥：四隻巨蛛的 CPIC 圖示、原版地城素材與戰鬥 footer](docs/screenshots/burial-glen-red-web-spiders.png)

### 法術演出的四種通道

原作不是「一支法術一套圖」。下面每一格都是從遊戲檔案直接取出來的原始素材，
由左到右是四種通道：

![法術演出的四種通道：共用投射物、閃電電弧、魔法命中、兩種雲、睡眠、弓箭](docs/reference/spell-visual-channels.png)

| 通道 | 圖 | 說明 |
|---|---|---|
| 共用施法投射物 | `COMSPR 05`／`85` | 在分派到各支 handler **之前**就播，所以每一支法術都有 |
| 閃電電弧 | `COMSPR 06`／`86` | 沿線逐格播，全表唯一 |
| 魔法傷害命中 | `COMSPR 0A`／`8A` | 對每個受傷的目標各播一次 |
| 持續區域格 | `RANDCOM 04`（綠＝惡臭之雲）／`02`（藍白＝致命毒雲）| 寫進地圖格，效果活著就一直畫 |
| 產生器／投射物 | `COMSPR 09`（睡眠）／`00`（弓箭）| 睡眠由參數合成；弓箭有八個方向 |

判讀見 [spec 1126](docs/spec/1126-spell-visual-slots.md) 與
[spec 1128](docs/spec/1128-cloud-areas-are-the-obstacle-terrain.md)。
兩種雲同時是戰術地圖的**障礙格**（地形碼 `1Eh`／`1Ch`）——低階角色繞開毒雲、
七級以上的老手硬闖。

這五張的產生指令與雜湊在
[`docs/screenshots/manifest.json`](docs/screenshots/manifest.json)，由
`cmd/screenshot-audit` 驗；上一輪修掉的四個圖層對齊錯誤見
[spec 1130](docs/spec/1130-screenshot-layer-alignment.md)。

更多地圖、人物舞台、戰鬥時間軸與素材圖在[截圖目錄](docs/screenshots/)；原版忠實
theme 與日後美化 theme 分開維護。

## 玩家文件

- [繁體中文攻略入口](docs/guide/README.md)：目前可玩的主線、分區攻略與無雷提示。
- [繁中遊玩手冊](docs/manual/curse-of-the-azure-bonds-zh-TW.md)：按鍵、戰鬥、紮營、
  手札與遊玩說明。
- [給中文玩家的 Gold Box 重製說明](docs/knowledge/golden-box-remake-for-chinese-readers.md)：
  把原本需要查紙本手冊的資訊整合進遊戲與文件。

## 開發與研究文件

- [`AGENTS.md`](AGENTS.md)：**工作規則的單一入口**（完成定義、證據標準、Git 紀律、
  驗證門檻、compact 恢復流程）。
- [剩餘工作盤點](docs/knowledge/coab-remake-todo.md)：逐項可執行的 TODO，
  含現況實測數字與建議順序。
- [完整度矩陣](docs/knowledge/coab-re-coverage-matrix.md)：全遊戲 RE／重建完整度的
  單一權威矩陣（`R1` 原始定位 → `R5` 玩家驗證）。
- [`WORKLIST.md`](WORKLIST.md)：反組譯盤點階段的執行順序。
- [`CONTEXT.md`](CONTEXT.md)：現況與已被推翻的斷言；歷史分冊在 [`docs/context/`](docs/context/)。
- [格式規格索引](docs/spec/README.md)：每份規格標示 `READY`／`DRAFT`、證據與
  `exact`／`strong inference`／`hypothesis`／`unknown` 邊界。
- [歷輪 README 報告](README-history.md)：長篇里程碑與研究紀錄，僅供歷史查閱，
  **不是目前狀態的權威來源**。

## 執行原則

1. 先讓正常玩家路徑可走，再補 fidelity。
2. 遊戲文字、選項、物品、事件與原版資料表由 CoAB JSON／locale 提供；
   Go／engine 只保留可重用機制，不複製劇情常數，也不 hardcode 原版資料表。
3. **先盤點後語意**（2026-08-13 起）：兩平台全模組先建庫盤點，每個函式都要進
   覆蓋台帳，再依玩家可見性排序做語意閉合。不再「遇到問題才反組譯那一段」。
   口徑見 [`AGENTS.md`](AGENTS.md) §2.5。
4. 每份結論都要能回指到哪個模組、哪個位址、哪幾個 byte；IDA／decompiler 的輸出
   本身不是證明。規格要寫明「明確不宣稱」的邊界。
5. 驗收採分段：每段可用直入旗標進入，但每段都要走到該段的正常結束狀態，
   而且**段與段之間的狀態交接本身也要列成一段驗**。
6. 每個重大、可展示的里程碑才集中 commit／push；兩個 repository 分開提交。

## 目前明確未完成

- **開場到結局的主線串接**是最大的一塊：目前走到眼魔洞穴與散提爾堡世界選單，
  後續章節、最終戰與結局序列尚未串完。
- **戰鬥回合生命週期**：ECL 的 `24h COMBAT` 有 199 處只做了分派，回合本身的
  initiative／held／delayed／guard／quick 逐項對照仍開著。
  （法術那一半已經收斂：可宣告 73 支全部有 handler、視覺與音效。）
- **ECL 的 11 個部分完成 opcode**，共 1,057 條指令：`CLEARMONSTERS`、`PICTURE`、
  `COMBAT`、`CALL`、`PRINT RETURN`、`TREASURE`、`DAMAGE` 等。
- **手札**：45 則有中英文，只有 15 則接上 ECL 觸發來源；手札 59 的地圖尚未繪製。
- **存檔**：角色記錄還有 29 bytes 未解讀；跨遊戲角色轉移未做。
- **表現層與音訊**：畫面逐張對照、每個場景與戰鬥 phase 的音效綁定。
  已知兩處未量的錨點：SPRIT 畫布相對於戰場格的位置（現在只在沒有 CPIC 時才會
  走到那條路），以及第一人稱每一塊牆磚的 `WALLDEF` 美術對應——
  幾何已與 spec 1006 對齊，選圖還沒比。
- **戰鬥佈陣**：`SETUP MONSTER` 的距離與 occupancy 表還沒解，兩隊的編制格位置
  仍是本作自訂的 fallback（現在至少會避開站不上去的格）。
- 三平台發行包。因此目前不製作正式 release 或宣傳片。

逐項狀態與建議順序見[剩餘工作盤點](docs/knowledge/coab-remake-todo.md)。

---

原始遊戲、手冊、圖片與音樂的權利仍屬各自權利人；本專案保存研究證據與新撰寫的
繁中資料，不重新散布未授權的原始遊戲映像。
