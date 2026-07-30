# 第 387 輪：Burial Glen 紅網雙戰鬥 State 玩家路徑

狀態：`READY`（限正常 GEO 進入、蜘蛛／羅剎妖兩場勝利續跑、完成旗標與
原版素材戰鬥 checkpoint；戰敗分支、完整戰鬥 fidelity 及 Burial Glen
後續仍未完成）

## 1. 原始 ECL 與怪物證據

DOS `ECL6.DAX` block `40h` 的 ENTER 分支可由 decoded payload
`+07E3h` 線性追蹤：

1. `+07FDh` SETUP MONSTER 使用 icon block `41h`。
2. `+0805h` LOAD MONSTER 載入四隻 monster `42h`。
3. `+080Ch` COMBAT；返回後 `+080Dh` 只把 combat result `E8h` 保存到
   工作位址 `7F72h`。
4. 精靈幽魂路徑已設定 `4CBEh` 時，`+081Eh` 顯示 PICTURE 72，
   `+0821h..+0849h` 揭露幽魂其實是 rakshasa。
5. `+085Eh` SETUP MONSTER 與 `+0866h` LOAD MONSTER 載入一隻
   monster `43h`，`+086Dh` 開始第二戰。
6. 第二戰返回後，`+0875h` 顯示隊伍終於脫困，`+088Eh` 把
   `4CBFh` 設成 1，然後 EXIT。

真實 `MON6CHA.DAX` block `42h`／`43h` 分別解碼為 `GIANT SPIDER` 與
`RAKSHASA`。remake 測試與畫面均從這兩筆記錄及 ECL spawn descriptor
建立敵人，不在 frontend 重複寫死名稱、數值或數量。

## 2. 「強大力量」是敘事陷阱，不是規則加成

Journal 25 宣稱說出 `Krrkik` 再走入紅網會獲得強大力量；但實際 script：

- SPEAK 的 `INPUT STRING 8,[7F79h]` 後沒有比較字串，也沒有修改角色；
- ENTER 的完整 `+07E3h..+0894h` 指令序列只寫入 combat 工作值
  `7F72h`、幽魂狀態 `4CBEh` 與完成旗標 `4CBFh`；
- 沒有 SAVE 到角色力量、能力值、effect 或 party record destination；
- 實際結果是隊伍陷入網中，接著遭遇蜘蛛及假扮幽魂的羅剎妖。

因此「獲得強大力量」應保留為遊戲內可誤導玩家的原始手札內容，不能由攻略
文字推導出 Strength buff。`7F72h` 目前只可稱為 combat work cell；未完成
其外部 ABI 反組譯前，不得把它命名為力量或效果欄位。

## 3. Remake 玩家路徑

正式 State regression 不使用 direct-entry encounter：

1. Standing Stone 揭露提朗瑟克斯；
2. Journey On→Myth Drannor→Wilderness→Enter City；
3. 進入 ECL6／GEO6 block `40h` 的 `(2,15,E)`；
4. 沿原始 GEO 路徑遇見精靈幽魂並選擇 GREET；
5. 沿 `(4,14)→(5,14)→(6,14)` 到 terrain `82h`；
6. 選擇 ENTER IT，以真實 MON6 records 開始四蜘蛛戰；
7. 由一般 `CombatAct` 完成勝利，讓 State 自動續跑同一 ECL session；
8. 顯示 PICTURE 72／羅剎妖揭露後，繼續第二戰；
9. 再次由一般戰鬥勝利，顯示脫困文字、驗證 `4CBFh=1` 並回到地城；
10. 重新執行相同 terrain lifecycle，事件不會重播。

測試英雄的高命中、高 HP 與固定傷害只用來讓 regression deterministic；
敵方資料、兩場 encounter、ECL continuation、旗標與文字仍來自受測原始映像
及 game-pack stable ID。

## 4. 畫面證據

`docs/screenshots/burial-glen-red-web-spiders.png` 是正式
`cmd/azure-bonds-game -burial-red-web-battle` 在
Docker／Xvfb／ALSA null device 的 deterministic 第三幀：

SHA-256：`6096b091d1e2b0c4e9a654672b19a28e9dd2193e72b618f05647f8e73cb034a2`。

- 640×480 logical canvas；
- 原版 cracked-stone combat frame；
- 原生 24×24 CPIC 蜘蛛素材以 nearest-neighbour 整數放大；
- 四隻敵人及原版地城 combat terrain；
- 正式戰鬥 HUD、生命值、防護等級與行動訊息。

這張圖證明實際 State／renderer 玩家路徑與素材，不證明整場蜘蛛動畫、
AI、毒素、音效或 DOS 畫面逐像素完全一致。

## 5. 尚未完成

- 蜘蛛毒素、羅剎妖 AD&D 能力、AI、施法與完整原版戰鬥規則；
- 兩場戰敗、撤退、全滅及 save/resume 路徑；
- DOS／PC-98 同一遭遇的逐幀時間、音效與像素對照；
- `7F72h` combat result 外部 ABI 的 IDA／runtime 完整命名；
- Burial Glen 其餘 terrain、Tyranthraxus 最終遭遇與結局。
