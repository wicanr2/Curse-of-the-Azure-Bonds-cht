# 1247：主選單讀檔與建角戰鬥圖示編輯器

狀態：`CONFORMED`
日期：2026-08-30

## 勘誤與範圍

spec 1093 把姓名後的 `sub_3EE3` 留成未知；spec 1037 後來已以 DOS／PC-98
兩平台證明它就是 `SETACTIVEICON`。因此原版建角骨架應勘誤為：

```text
種族 → 性別 → 職業 → 陣營 → 能力值／HP → 姓名
→ 戰鬥圖示編輯（READY／ACTION）→ 儲存角色 → 回隊伍功能選單
```

同時，spec 1085 已證明角色管理主選單的 `L` 是載入存檔；spec 1246 為了先閉合
建立／加入垂直鏈而排除它，不代表原版沒有或 remake 可以永久省略。

## 一手證據

| 行為 | 原始定位 | 等級 | 來源 |
|---|---|---|---|
| 建角姓名後呼叫圖示編輯器 | DOS `overlay-17:00782h` 呼叫 `03EE3h` | `exact` | spec 1093＋1037 位址閉合 |
| 圖示編輯器與 READY／ACTION 雙預覽 | DOS `overlay-17:03EE3h`；PC-98 `overlay-17:045D0h`／`SETACTIVEICON` | `exact` | spec 1037，兩平台逐條讀完 |
| 頭 0..0Dh、武器 0..1Fh、大小 1／2 | 角色記錄 `+141h`、`+142h`、`+144h` | `exact` | spec 1037／1076 |
| ACTION 使用 READY block `+80h` | CHEAD／CBODY consumer | `exact` | spec 178／179 |
| 主選單 `L` 呼叫載入存檔 | PC-98 `overlay-17:002CBh` → overlay-16 entry 10；DOS 對側 `001DDh` | `exact` | spec 1085／1076 |
| 原版只在新遊戲空隊伍顯示 `L` | `bank0^[1E4h] == 0` 且隊伍空 | `exact` | spec 1085 |

輸入檔雜湊、IDA 版本與位址空間沿用
[`coab-function-index.md`](../audit/coab-function-index.md) 的固定清冊；本輪不新增
推測性名稱，也不改寫原始函式定位。

## Typed 行為

### 主選單讀檔

- 建立隊伍外層新增 `L 載入遊戲`，與建立、加入、開始冒險同頁。
- 本按鍵載入 remake 目前的版本化遊戲存檔；成功後直接進入存檔記錄的模式、地圖、
  ECL、戰鬥與音訊狀態，沿用既有 `LoadPartyFile` 交易。
- 檔案不存在、格式錯誤或內容不合法時採失敗即關閉：留在外層選單並顯示錯誤，
  不清空當前角色庫或建立中的隊伍。
- 原版 SAVGAM A..J 匯入仍由既有命令列入口處理；本輪不把兩種格式靜默混在一起。
- F9 全域快捷鍵保持不變；主選單 `L` 是原版可發現入口，不是另一套 loader。

### 建角圖示編輯

- 姓名確認後先進 `CreationStepIcon`，確認／取消圖示後才進儲存詢問。
- 以三列編輯頭、武器與體型：上下選欄，左右循環；頭為 14 種、武器為 32 種，
  體型為 Small／Large。
- 畫面同時用實際 CHEAD＋CBODY 顯示 READY 與 ACTION；ACTION 使用相同選擇的
  `+80h` block。F2 必須原子切換 original／modern-a6，且不得改變選擇。
- Enter／K 保留選擇；Esc／E 還原進入編輯器時的頭、武器與體型，再進儲存詢問。
- 圖示欄位寫進獨立角色庫與正式遊戲存檔；加入隊伍及戰鬥投影不得再退回預設圖示。

## 已知邊界

- spec 1037 已證明六個部位各有兩個 4-bit 顏色，但 remake 角色 schema 與 renderer
  尚未具備已驗證的原版 recolor palette transaction。本輪不猜 palette 映射；
  顏色編輯另立窄 spec 後接入。
- 原版完整角色管理有十二個命令；本輪只補玩家指出且已有現成交易的讀檔，不把
  Delete／Modify／Train／Change Class／View／Remove／Save／Exit 一次混入。
- remake 使用 640×480 CJK 安全矩形，屬 `material-exact/layout-reconstructed`，
  不宣稱 320×200 逐像素相同。

## 驗收

1. State：姓名後必到 icon；頭 14、武器 32、大小 2 均環繞。
2. State：Keep 保存選擇；Exit 還原備份；兩者都進 Save 而不跳過角色保存詢問。
3. Save：圖示選擇經角色庫保存、重新載入、加入隊伍後仍一致。
4. UI：同頁可見 READY／ACTION，且兩者使用 normal／`+80h` block。
5. 按鍵：正常標題→建角→圖示→保存→加入→出發路徑通過；圖示頁切 F2 不漂移。
6. 讀檔：外層 `L` 成功載入既有 remake 存檔；不存在／損壞時留在外層並顯示錯誤。

## 實作與驗證結果

- 四語系外層新增 `L`；與 F9 共用 `loadCurrentGame`，成功與不存在檔案兩條按鍵
  路徑均通過。
- 姓名後進入圖示頁；頭 14、武器 32、大小 2 的環繞、Keep／Exit、角色庫
  round-trip 與加入隊伍後欄位保存均有測試。
- 正常按鍵 session 在圖示頁實際切換 F2、改選頭部、保存、加入並進入第一張地圖；
  400 幀走過 47 格，英文回退 0、全滅 0。
- Docker/Xvfb 正式畫面抽拍在
  `workplace/game-tester-20260830/creation-icon-ready-action.png`：READY／ACTION 的
  CHEAD＋CBODY sprite 同頁可見，文字與下框不重疊。
- 第一版抽拍因把未完成 draft 丟進完整 `Fighter()` 驗證而靜默不畫 sprite；改成
  直接由已驗證圖示欄位建立 preview fighter 後修正。這證明狀態測試不能取代視覺抽樣。
