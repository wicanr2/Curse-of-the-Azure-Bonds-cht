# 共用引擎與作品之間，現在實際切在哪

盤點日期：2026-08-24。**每個數字都是量的**（`wc -l` 非測試碼、`grep` 內容外洩），
不是印象。重跑方式寫在最後。

## 一句話

**內容的界線守住了，程式碼的界線沒有。** engine repo 裡查不到任何 CoAB 的劇情、
座標或人名；但金盒子**共通**的那幾層——ECL 直譯器、DAX／原作記錄解碼、戰鬥規則、
角色與存檔記錄——**大部分住在作品 repo 裡**，不在 engine。第二款作品現在要嘛
import 這個作品的 module，要嘛把那幾層再寫一次。

## 量到的數字

### 兩邊各有多少非測試 Go

| repo | 行數 |
|---|---:|
| `golden-box-remake-engine` | **6,907** |
| `Curse-of-the-Azure-Bonds-cht` 的 `internal/` | **44,991** |

engine 逐套件：

| 套件 | 行數 | 內容 |
|---|---:|---|
| `engine` | 2,285 | game pack schema／載入／驗證、文字規則 |
| `combat` | 1,998 | 22 個規則子套件（先攻、抗性、命中後、睡眠、掃描順序…）|
| `audio` | 1,193 | S98／YM2203／PC-98 Sound BIOS |
| `graphics` | 505 | EGA／調色盤／點陣 |
| `viewport` | 381 | 第一人稱牆面投影 |
| `geometry` | 266 | GEO 解碼與移動判定 |
| `areamap` | 133 | 區域地圖 |
| `randomstream` | 99 | 可存檔的亂數續跑 |
| `worldmap` | 47 | 世界地圖游標 |

作品 repo 裡**按性質屬於共用層**的套件：

| 套件 | 行數 | 為什麼算共用 |
|---|---:|---|
| `internal/combat` | 7,155 | AD&D／金盒子戰鬥規則本體（engine 那 1,998 只是規則表的殼）|
| `internal/ecl` | 4,222 | **ECL 位元組碼直譯器**——金盒子系列的腳本格式 |
| `internal/pc98music` | 2,828 | PC-98 音樂驅動重建 |
| `internal/party` | 2,826 | 角色記錄（422 bytes 的原作版面）|
| `internal/monster` | 1,392 | 怪物／物品記錄 |
| `internal/eclcatalog` | 981 | ECL 目錄 |
| `internal/save` | 617 | SAVGAM 記錄 |
| `internal/mapdata` | 619 | 地城／荒野底圖生成 |
| `internal/eclcells` | 583 | 每格事件分派器 |
| `internal/gfx`／`sound`／`dax`／`geo`／`borlanddebug`／`audiomap` | ~1,600 | 解碼與對照 |
| **小計** | **≈ 22,800** | |

作品**自己**的那一層：`internal/game` **18,552** 行（`state.go` 6,936、
`combat_state.go` 4,816、角色建立 1,920、商店／神殿／訓練所／法術特例…）。

### 內容有沒有漏進 engine

對 engine 的非測試碼 grep `tilverton|azure|fire knife|moander|tyranthraxus|
naccacia|giogi|shadowdale|yulash|zhentil|myth drannor|coab`：**命中 2 處，兩處都是
註解在解釋界線**（`combat/resistance/rules.go:21`、`engine/pack.go:191`）。
沒有劇情旗標、座標、人名或翻譯。**這條守住了。**

反過來查作品 repo 有沒有把座標寫死：`internal/game` 裡

- 直接指派地城座標：**1 處**（`s.MapX, s.MapY = 0, 0`，回標題的重設）
- 寫死的 block 編號常數：**1 個**（`worldMapBlockFloor = 0x50`）
- `dataPack` 參照 **52** 處、語系 `catalog.Text` **271** 處

也就是說作品那一層**不是**靠硬寫座標在跑，是靠 JSON ＋ 原作資料。

## JSON 蓋到哪、沒蓋到哪

`gamepack/pack/` 四個檔、17,903 行：

| 區段 | 筆數 | 內容 |
|---|---:|---|
| `maps` | 20 | 18 張第一人稱、1 張區域地圖、1 張大地圖；含 spawn、`search_edges`、`external_exits` |
| `text_rules` | 1,284 | 原作英文 → 訊息 ID |
| `option_rules` | 127 | 選單選項 → 訊息 ID |
| `combatant_name_rules` | 68 | 怪物顯示名 |
| `combat_player_spells` | 73 | 法術行為表 |
| `combat_ai_spells` | 12 | 敵方 AI 法術 |
| `music_tracks`／`bindings`／`cues` | 12／13／3 | 曲目與觸發 |
| 其餘戰鬥規則表 | 11 組 | 修正值、抗性、命中後、目標選擇… |
| `locales` | 2 | en／zh-TW |
| `events` | 4 | |

大地圖**是** JSON：`moonsea.overland` 帶 14 個地點、座標與 `destinations`
鄰接表。

**沒有進 JSON 的是「事件本身」**——而這是刻意的。remake **不重寫劇情**，它
**直譯原作的 ECL 位元組碼**：對話、旗標、遭遇、寶物、每格事件全部從
`curseoftheazurebonds.zip` 讀出來執行。JSON 只補三件原作放在**引擎層**（不在腳本裡）
的東西：

1. **資源綁定**——哪張圖配哪個 GEO／牆面／天空／音樂；
2. **邊界宣告**——進場錨點、可搜尋的邊、走得出圖的格子（`external_exit`）；
3. **規則表**——原作寫在組語裡的常數表（法術、抗性、命中後…）。

所以「除了客製化 source code 之外其餘都是 JSON」這個描述**不成立，也不該成立**：
內容不在 JSON，在原作資料檔裡；JSON 是**繫結層**不是內容層。

## 這代表什麼

| 問題 | 答案 |
|---|---|
| engine 有沒有被作品內容污染？ | **沒有**（量過，2 處註解）|
| engine 夠不夠「通用」？ | **就它涵蓋的範圍是通用的，但涵蓋得不夠**——6,907 行對 22,800 行共用層 |
| 第二款金盒子作品能直接用嗎？ | **不能只靠 engine**：ECL 直譯器、戰鬥規則本體、角色／怪物／存檔記錄都在作品 repo |
| 作品那一層有多少是真的作品專屬？ | `internal/game` 18,552 行，其中商店／神殿／訓練所／角色建立／戰鬥狀態機是金盒子共通形狀，**只有旗標語意與流程順序是 CoAB 的** |

⚠ **「按性質屬於共用層」不等於「搬過去就能用」。** 沒有第二款作品當對照之前，
「這一層在別的金盒子作品上也成立」是**推論不是證據**：ECL 的 opcode 語意、
角色記錄的欄位版面、SAVGAM 的版本都可能逐作不同。要證明只有一個辦法——
**拿第二款作品跑一次**。在那之前，把這幾層搬進 engine 只是換個目錄，
不會讓通用性變成事實。

⚠ 這份盤點**不建議現在就大搬家**。目前 P0／P1 還開著（見
[`WORKLIST.md`](../../WORKLIST.md)），而搬移會讓那些工作全部帶著 import 路徑
改動。合理的順序是：先把這一款收斂，再以第二款作品的實際需求驅動搬移——
那時候「哪些真的通用」才有證據，而不是猜。

## 重跑

```bash
# 兩邊的非測試行數
find golden-box-remake-engine -name '*.go' -not -name '*_test.go' | xargs wc -l | tail -1
find internal -name '*.go' -not -name '*_test.go' | xargs wc -l | tail -1

# engine 有沒有作品內容外洩
grep -rniE 'tilverton|azure|fire knife|moander|tyranthraxus|naccacia|giogi' \
  --include='*.go' golden-box-remake-engine | grep -v _test

# JSON 各區段筆數
python3 -c "import json;d=json.load(open('gamepack/pack/00-core.json'));\
print({k:(len(v) if isinstance(v,(list,dict)) else 1) for k,v in d.items()})"
```
