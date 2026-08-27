# Spec 1208：公用資金池生命週期與購買者選擇

狀態：**READY**（2026-08-26）

## 原版證據

- spec 798（`POOLMONEY`，exact）把角色七種金錢搬入 MONEY overlay 的全域 Pool，
  並清空角色欄位。
- spec 863（`SHAREPOOL`，exact）才把 Pool 分回角色；裝不下的餘額仍留在 Pool，
  並依剩餘內容更新旗標。
- spec 814（`CASHPOOL`，exact）從傳入的七個 Pool longint 計算金幣價值，不會
  改動來源。

因此「進入另一間商店／神殿」不是 Pool 的消費端，也沒有清空契約。

## 找到的玩家缺陷

remake 原先在每次 `enterECLShop` 與 `enterECLTemple` 都執行
`moneyPool = 0`。玩家在第一間商店選 POOL 後，角色的幣別已依法清空；再進第二間
商店時公用池也被清掉，整隊資產便無聲消失。一般強度按鍵 trace 實際因此買不起
400 GP 板甲，甚至只能留下單件盾牌。

商店 BUY 還有另一個未接玩家路徑：`SetShopCharacter` API 可以指定收件角色，
但前端選單永遠使用預設的第 0 人，其他隊員無法取得購買物品。

## Remake 修正

- 商店與神殿入口不再清空公用資金池；只有玩家執行 SHARE／TAKE／購買等既有
  消費操作會改變它。
- 多人隊伍選 BUY 時先選購買角色，再進商品清單；單人隊伍維持直接進商品清單。
- 商品仍以未整備狀態交付，後續必須走 spec 1207 的營地整備 UI。

## 驗證

- `TestMoneyPoolSurvivesEnteringAnotherShopOrTemple`：500 GP Pool 依序進神殿與另一間
  ECL 商店後仍為 500。
- `TestShopBuyLetsPlayerChooseTheReceivingCharacter`：選第二名角色購買後，物品只進
  第二人的背包。
- 一般強度按鍵路徑已能在劇情扣款後、確認盔甲店後才 POOL，替六人各買皮甲與
  盾牌，並逐人整備；皇家衛兵戰開場六人均為 AC 7，證明「Typed coins → Pool →
  BUY recipient → inventory → READY → combat projection」垂直鏈已接通。

## 尚未證明

六名 2 級戰士交由 QUICK AI 仍會敗給五名皇家衛兵；這不證明手動戰術也必敗，
也不能用強化角色測試宣稱正常路徑已通關。後續須以玩家可操作戰術或原版同狀態
對照，區分 QUICK AI、遭遇數值及戰鬥規則差異。
