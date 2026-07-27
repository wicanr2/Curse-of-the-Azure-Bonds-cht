# 第三百零五輪：阿沙本福德與立石群

狀態：`READY`

## Ashabenford city service

ECL1 block `0x50` 在 world current-location `4C9B=2` 時，`ENTER CITY` 顯示
PICTURE `0x50`，再進入抽象場所選單；沒有 GEO、LOAD FILES 或 NEWECL。
目前真實路徑可見：

- INN／STORE。
- HALL：`PROGRAM 0` 是訓練服務；不能按全域 ID 誤判為返回主選單。
- TEMPLE。
- BAR：河畔酒館的 `HAVE A DRINK／RELAX／EXIT`。
- LEAVE：PICTURE `0x79`，回到 Ashabenford edge menu。

固定 seed 的 RELAX regression 抵達 Tavern Tale 28；遊戲內直接顯示使用者
Adventurer's Journal 的內容：「兩艘前往暗影谷的船失蹤，河道已變得危險」，
不再只要求玩家自行翻閱手冊。

## Standing Stone route

Ashabenford 的 `JOURNEY ON` 提供 Tilverton、Shadowdale、The Standing Stone。
選 Standing Stone 的 TRAIL 後，真實 State 首先顯示一隊偽裝成巡邏兵的火刀伏擊，
再建立六名 MON1 monster `0x59`：

- raw name `FIGHTER`，繁中顯示「戰士」。
- count `6`。
- icon block `0x20`。

勝利後抵達立石群。灰袍男子指出隊伍仍受四位主人控制，要求消滅另外三位後再回來；
選擇 `THANK HIM` 後，他補充「往南方尋找紅色之人」。world dispatcher 將
`4C9B=4`，State 投影為 `LocationStandingStone`。這是通往 Essembra／Hap、
紅袍法師 Dracandros 主線的導航。

## UI contract

Ashabenford PICTURE 80、離城 BIGPIC 121 與戰鬥 icon 均使用原始像素素材的
nearest-neighbour 整數放大；城市／酒館／立石群文字則直接在 640×480 畫布以
24px 繁中排版，緊湊欄位仍可使用 16×15。
