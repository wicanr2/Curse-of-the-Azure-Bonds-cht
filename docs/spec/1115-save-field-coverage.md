# 1115 — 存檔：角色記錄的逐位元組台帳，與「新增欄位就會紅」的存檔閘

- 證據等級：`exact`（台帳每一段都有出處；`decoded` 由位元組突變逐段量測驗證）
- 產出：`docs/audit/dos-save-field-coverage.{md,json}`（`cmd/save-field-coverage`）
- 對應工作項：`RE-05`（讀懂原版存檔）與 `ENG-10`（remake 自己的存檔完整）

## 一、`RE-05`：422 個位元組，每一個都有狀態

DOS 的 `CHARREC` 是 `1A6h` ＝ 422 bytes。先前的說法是「仍缺全欄位與未知 byte 的
讀取語意」——那句話回答不了「**還剩多少不知道**」。現在有數字：

| 狀態 | 位元組 | 意思 |
|---|---:|---|
| `decoded` | 294（69%）| remake 的解析器真的讀它，值會影響匯入結果 |
| `documented` | 99（23%）| 有規格說得出它是什麼，但解析器沒讀 |
| `unknown` | 29（7%）| 沒有任何出處。**不是「沒有用」**，只是還沒查到 |

台帳在 `internal/party.DOSPlayerRecordFields`，`ValidateDOSPlayerRecordFields`
擋住洞與重疊——**少一段就會讓 `unknown` 的數字偏低**，而那個數字正是這一項的
進度指標。

### `decoded` 是量出來的，不是宣告的

`cmd/save-field-coverage` 用**位元組突變**驗證：改一個位元組、重新解析、比對
結果，有差別就是有讀。

★ 為什麼不掃程式碼：`data[0x12D+class*5]` 這種算出來的索引掃不到，得到的
「未讀」清單裡會混進一堆假缺口，而假缺口看起來就像真缺口。突變法不看程式碼
長什麼樣。

⚠ 量測是**下界**：只在某些職業／種族才被讀的位元組，基準記錄沒涵蓋就會被判成
沒讀。所以基準有五份（戰士／法師／牧師／盜賊／雙職），而且對帳是雙向的：

- 量到有讀、台帳卻標 `documented`／`unknown` ⇒ 台帳低估，硬錯。
- 台帳標 `decoded`、卻沒有任何基準量到 ⇒ **補一份能觸發它的基準記錄**，
  不是把台帳改成 `documented`。

對帳跑在 `go test ./...` 裡（`cmd/save-field-coverage` 的 `main_test.go`）——
只有人去跑報告才會發現的漂移，等於沒有閘。

### 理解程度 ≠ 保真度

**匯入不會因為 `unknown` 而遺失資料。** `State.LoadSAVGAMSlot` 保留整份原始
記錄，`PatchDOSPlayerRecord` 只寫已知位移（spec 185），既有測試用
`+190h = 0AAh` 這個未定位元組釘住「寫回之後它還在」。

所以這份報告衡量的是**看得懂多少**，不是**存得住多少**。兩者要分開講，
否則 7% 的 `unknown` 會被誤讀成 7% 的資料流失。

## 二、`ENG-10`：存檔完整性的閘

`Fighter` 的每一個匯出欄位都要能存進快照、再讀回來
（`TestEveryFighterFieldSurvivesASaveRoundTrip`）。用反射掃欄位而不是逐欄斷言：

> 逐欄斷言只擋得住「已經想到的欄位」。真正會出事的是**下一個新增的欄位**——
> 加了欄位、忘了存檔，症狀是讀檔之後某個狀態悄悄回到零值，而所有既有測試照樣綠。

### 這條閘當場抓到一個真的 bug

本輪新增 `StatusPartyFled`（隊伍逃離，spec 1112）之後，`RestoreBattle` 的
`if snapshot.Status > StatusDraw` 沒有跟著改——**逃離結局的存檔存得下去、讀不
回來**，而且錯誤發生在讀檔那一刻，離成因很遠。上限已改成跟著 `Status` 的最後
一個值走，並補了 `TestFledBattleSurvivesASaveRoundTrip`。

### 五個欄位故意不進存檔

`DamageRules`、`ConditionalModifierRules`、`MagicResistanceRules`、
`PostHitRules`、`MonsterSpellRules` 是 game pack 的規則表，讀檔之後由遊戲層重新
注入。存進去只會讓存檔跟著資料檔一起過期。

豁免不是白名單就算了——`TestPackDerivedFieldsAreReinjectedOnLoad` 去
`internal/game/creation.go` 的讀檔路徑找對應的 `SetXxxRules` 呼叫，注入被拿掉時
豁免就不再成立（那正是「讀檔之後怪物突然不會施法」的成因）。反向也擋：
`Fighter` 上以 `Rules` 結尾的欄位若沒登記在豁免表，測試會要求先確認它是存檔
狀態還是 pack 衍生。

## 二之二、兩份 sidecar 也逐位元組蓋滿

| 記錄 | 大小 | `decoded` | `documented` | `unknown` |
|---|---:|---:|---:|---:|
| `.SWG` 物品 | 63 | 59 | 4 | **0** |
| `.FX` 效果 | 9 | 5 | 4 | **0** |

兩者的 `documented` 都是**鏈結指標**（物品 `+2Ah`、效果 `+05h`）：執行期的遠指標，
匯入時原樣保留、不解讀。

★ 物品名稱只到 `+29h`。 原作走物品鏈用的就是 `q := q^[2Ah]`（spec 1000／832），
所以名字是 42 bytes 不是 46。多讀四個位元組在英文原版只會看起來像「名字後面有
幾個怪字元」，中文資料則會直接壞掉——`TestItemNameStopsBeforeTheChainPointer`
用非零指標值把它釘住。

這兩份沒有突變探針：解析器逐欄直讀、沒有算出來的索引，台帳蓋滿就足以說明覆蓋。

## 三、決策四的落點

使用者的決策四是「remake 讀舊版 DAT、存成自己的格式，不必互通」。因此：

- **不做**原版 round-trip gate，也不驗「寫回原版可讀」。
- 匯入路徑（`LoadSAVGAMSlot`，spec 182／185）與寫回同一個 slot 的能力保留——
  後者只寫已知位移、保留未知位元組，是既有行為，不是本輪新增的承諾。

## 明確不宣稱

- 沒有宣稱 29 個 `unknown` 位元組是空的或沒有用途。
- 沒有宣稱 PC-98 的 `CHARREC`（`1A7h`，比 DOS 多一個 byte，spec 626）可以套用
  同一張台帳。
- 沒有宣稱遊戲層快照（`save.CurrentGameVersion`）的每個欄位都有同等的反射閘——
  本輪只對 `combat.Fighter` 做。
