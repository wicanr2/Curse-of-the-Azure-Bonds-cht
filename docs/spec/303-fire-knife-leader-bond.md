# 第三百零三輪：火刀首領、解除枷印與夢境

狀態：`READY`

## 證據邊界

ECL2 block 4 `SearchLocation` 的零起算 selector `0x07` 對應 GEO2 block 4
`(3,13)`、terrain `0x87`。事件先顯示 `PICTURE 0x0C` 並解鎖手札 11，
Continue 後建立 20 名 MON2 record 1 與 1 名 record 3；兩組 `PartyMask=0`，
不可只因手札提到公主便推定 record 3 是友軍。

腳本在 `COMBAT` 前建立 treasure packet：2000 gold、3000 platinum、8 gems、
4 jewelry、`ItemBlock=0x82`。這是遭遇戰勝利獎勵，不是
`CLEARMONSTERS → TREASURE → COMBAT` 的 treasure-service；State 只有在實際
monster spawns 為空時才可提前開 loot UI。依既有換算，勝利後入池 17,000 GP
與兩件隨機物品。loot UI 結束後才 resume 同一 ECL combat boundary，確保後續
手札、夢境及章節切換不被財寶選單截斷。

## 戰後 continuation

`[7EC7] <= 0x80` 進入勝利線：

1. 解鎖手札 54、寫 `4CFF=1`、顯示 PICTURE 14。
2. 解鎖手札 53、寫 `4C2A=1`、顯示 PICTURE 13。
3. 顯示城外第一夜與四張面孔的四段夢境；其中 BIGPIC block 120 是主人預言。
4. 全部 Continue 後才寫 `7F12=1`，再 `NEWECL 0x50`。

手札 11、53、54 依使用者提供的 Adventurer's Journal 來源翻譯，只有原始事件
文字抵達時才加入遊戲內手札。事件沿用 640×480 邏輯畫布：原始 PICTURE／BIGPIC
只做整數 nearest-neighbour 放大，繁中敘事直接以 24px 高解析字型重繪；狹窄欄位
可使用 16×15。

## 驗證

- real ECL regression 鎖定 monster descriptors、treasure packet、兩本戰後手札、
  PICTURE 13、BIGPIC 120，以及 `4CFF`／`4C2A`。
- State regression 鎖定遭遇獎勵戰前不可取得、勝利後才得到 17,000 GP、8 gems、
  4 jewelry 與兩件隨機物品，並從 loot 一路走完 `NEWECL 0x50`、恢復
  Tilverton 的 `ENTER CITY／JOURNEY ON／CAMP` 選單。
