# 第 389 輪：Burial Glen 黛米爾公主幽魂

狀態：`READY`（限正常 GEO 抵達、一次性事件、四選單、好感／`4CBBh`
原始寫入、資料化繁中與畫面；`4CBBh` 的戰鬥 consumer 尚未證明，因此
命中加減值仍未接入 runtime）

## 1. 原始資料與正常地圖路徑

輸入：

- DOS `START.EXE` SHA-256：
  `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5`
- DOS `GAME.OVR` SHA-256：
  `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215`
- `ECL6.DAX`／`GEO6.DAX` block `40h`

`GEO6.DAX` 的 Princess Daemir 事件位於 `(13,14)`、terrain `03h`。
新增的作品中立 CLI 稽核：

```text
azure-bonds -member GEO6.DAX -block-id 64 -geo-grid \
  -geo-path 6,12:13,14
```

會使用 `CanMoveDungeonWrapped` 的牆／門規則，重建下列最短可行路徑：

```text
(6,12) E→(7,12)→(8,12)→(9,12)→(10,12)→(11,12)→(12,12)→(13,12)
       S→(13,13)→(13,14)
```

正常玩家 regression 不是直接設定 ECL PC：它由 Standing Stone 前往
Myth Drannor、完成 Burial Glen 入口、紅網雙戰與墳墓事件，再逐格走完上述
九步；途中若出現一般隨機遭遇，選擇 FLEE 後仍由同一 `BlockSession`
續跑。

## 2. ECL 控制流

terrain selector `03h` 對應 decoded payload `+0991h`：

- `+0991h`：`4CC0h==1` 時直接離開，因此黛米爾只出現一次；
- `+0999h`：把 `1` 寫入 `4CC0h`；
- `+099Fh`：`PICTURE 48h`，即人物圖 block 72；
- `+09DAh`：以偏移中立值 `80h` 比較精靈認可度 `4CBAh`；
- `<80h` 顯示「褻瀆葬林／接受寬恕」，否則詢問是否接受祝福；
- `+0A0Fh` 的動態四選單是
  `ACCEPT／REJECT／KILL／FLEE`，selection destination 為 `7F7Ah`。

選單字串結束於 `+0A2Dh`，後續 `ON GOTO` 分支為：

- `ACCEPT` → `+0A3Fh`；
- `REJECT` → `+0AE7h`；
- `KILL` → `+0AE7h`，與拒絕共用同一分支，沒有額外戰鬥；
- `FLEE` → 共用離開路徑。

real-image `BlockSession` 測試鎖定五種有效結果：

| 初始 `4CBAh` | 選項 | 結果 |
|---:|---|---|
| `80h` | ACCEPT | 顯示祝福，`4CBAh=85h`、`4CBBh=02h` |
| `7Fh` | ACCEPT | 顯示寬恕，`4CBAh=80h`，不寫 `4CBBh` |
| `80h` | REJECT | 顯示武器詛咒，`4CBAh=76h`、`4CBBh=FEh` |
| `80h` | KILL | 與 REJECT 完全相同 |
| `80h` | FLEE | 離開，`4CBAh` 不變且不寫 `4CBBh` |

## 3. `4CBBh` 的證據界線

事件文字說武器會扭轉，攻略則把 `02h／FEh` 解釋成 Myth Drannor 期間的
`+2／-2 to hit`。但 ECL 目前只證明 writer，攻略只屬次級交叉參考，尚未
證明 DOS 戰鬥公式的 consumer。

次級來源：
<https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365>

依專案規則在 Docker 中以 IDA Pro 9.4 fresh-load 兩個原版 binary，執行
`scripts/ida/dos_daemir_work_cells_audit.idc`：

- `START.EXE`：`4CBAh` 有 3 個 instruction operands，`4CC0h` 有 18 個；
  合計 21 個 instruction hits／21 個 raw little-endian hits；
- `START.EXE` 沒有 `4CBBh` literal operand；
- `GAME.OVR` 以 8086／16-bit raw binary 載入時沒有可自動辨識的
  instruction hit，只在 file `25569h`（IDA EA `256C9h`，load base
  `160h`）出現一個 `C0 4C` raw pattern；
- `GAME.OVR` 沒有 `4CBBh` literal pattern。

這個陰性結果不代表 `4CBBh` 未使用：ECL work memory 很可能由 interpreter
基底指標／索引間接存取。反之，`4CBAh／4CC0h` 的 literal hits 也只是
START resident 資料位置，不能直接命名成 Daemir 規則 consumer。因此
本輪保留 `4CBBh` 原始值與測試，不把它投影成 remake 的 `AttackBonus`；
必須先找到間接讀取端或 DOS runtime 的相鄰值戰鬥實驗。

## 4. 資料化繁中與測試

黛米爾的兩種提問、祝福、寬恕、詛咒，以及四個選項都由
`gamepack/events/pit-of-moander.json` 的 stable message ID、
`text_rules` 與 `option_rules` 驅動。State／frontend 沒有新增作品專屬
中文 switch。

產品層測試不複製繁中字串，而由目前 game pack 的 stable ID 取得期望顯示；
原始英文只保留在 real-image ECL oracle 測試，用來驗證原文、分支與位址
副作用。

## 5. 畫面證據

`docs/screenshots/burial-glen-princess-daemir.png` 是正式
`cmd/azure-bonds-game -burial-daemir` 在 Docker／Xvfb 中產生的
640×480 確定性 checkpoint。

SHA-256：
`920a3dfd9ccc6d996fccc1af8fb1b3ba3f2099b0d920a227457ca5c713289804`。

畫面證明：

- 原版 DOS cracked-stone adventure frame；
- `PICTURE 48h`／block 72 原始黛米爾人物圖以 nearest-neighbour 放大並
  cover 左上內格；
- 16×15 倚天粗體繁中提問；
- party roster 與下方文字區沿用正式 640×480 adventure layout。

它是 `material-exact/layout-reconstructed`，不是整張 DOS 畫面的逐像素
exact 複製，也不證明後續選單或戰鬥公式。

## 6. 尚未完成

- 追出 `4CBBh` 的間接 consumer，證明有效範圍、符號解讀、累加／覆寫及
  save/load 行為；
- 以 DOSBox 修改相鄰值，對照 Myth Drannor 內外的命中骰與角色狀態頁；
- 黛米爾選單按鍵、延遲、音效與淡出逐幀 fidelity；
- Burial Glen 紅羽戰士、其他 terrain、最終區域、Tyranthraxus 與結局。
