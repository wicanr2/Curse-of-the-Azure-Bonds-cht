# 第三百五十一輪：兜帽女子與眼魔洞穴轉場

Status: `READY`

> 第 547 輪勘誤：本檔原先將洞穴可操作錨點寫成 `(4,5,N)`，那是尚未閉合 DOS
> virtual-map register bridge 時留下的 prototype 座標。現以 Cave E1 `(5,7,W)`
> 為正常轉場結果；原始證據、推論等級與後續 `(13,1,W)` 同區塊位置交易，見
> [`547-normal-beholder-cave-presentation-state.md`](./547-normal-beholder-cave-presentation-state.md)。
> 本檔其餘兜帽女子至護符離場的事件順序仍有效。

## 目標

延續迪姆斯沃特 story escort：在幽暗神殿答應兜帽女子帶路，經歷德克薩姆與
弗佐爾的衝突、手札 30／7、弗佐爾死亡與第四道枷印消退，最後進入可操作的
眼魔洞穴，而不是在神殿出口或跨 ECL pause 中斷。

## 原版執行證據

- `ECL4/GEO4 0x21`，迪姆斯沃特同行旗標成立後：
  - `(4,12)` terrain `0x86` 觸發 PICTURE 39 的兜帽女子。
  - `DO YOU FOLLOW?` 選 `YES`；PICTURE 33 顯示弗佐爾闖入。
- continuation 由 ECL block `0x21` 切至 `0x22`：
  - PICTURE 40 顯示德克薩姆與披甲牛頭人。
  - 德克薩姆演說解鎖手札 30。
  - 迪姆斯沃特指出祭壇上的洛山達護符，顯示
    `COMBAT/WAIT/FLEE/ADVANCE/PARLAY`。
  - 玩家尚未行動，弗佐爾率眾闖入並解鎖手札 7。
  - 德克薩姆殺死弗佐爾；PICTURE 33 顯示弗佐爾枷印消退。
  - 德克薩姆帶走護符，兩派人馬混戰，最後回到 `ModeDungeon`。
- remake 正常轉場最終狀態為 ECL4 block `0x22`、`GEO4/0x25 (5,7,W)`；Cave E1
  的命名與幾何錨點為 `strong inference`，詳見 spec 547。作品 pack 必須宣告該
  map，否則跨 block 聚合會殘留來源 GEO `0x21`。

## 交叉參考

- 隨附 *Adventure Journal* TXT 是手札 30、7 的翻譯原文。
- [GameFAQs Shrine of Bane details](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)
  交叉確認 location 7、拒絕後巡邏戰、接受後唯一離場路徑，以及 Cave of the
  Beholder E1。
- [GameBanshee Zhentil Keep walkthrough](https://www.gamebanshee.com/curseoftheazurebonds/walkthrough/zhentilkeep.php)
  交叉確認迪姆斯沃特是出寺必要條件、兜帽女子會反覆出現。
- 攻略只作導航 oracle；座標、圖片、pause、選項與 block transition 仍以真實
  ECL/GEO runtime 為準。

## 資料與引擎邊界

- 全部 CoAB 人名、翻譯、手札與三張 Area 4 first-person map definition 留在
  JSON game pack。
- 通用 State 只依 pack 的 exact area/block map definition 修正 NEWECL
  destination，不 hardcode 弗佐爾、德克薩姆或洞穴。
- 本輪沿用原 ECL 的五項 encounter menu；弗佐爾會打斷選擇，不另造捷徑。

## 繁中與長文

- 手札 30、7 各拆成兩頁；保留德克薩姆冷酷的黑色幽默與弗佐爾自大的政治
  宣言，不縮成「反派吵架」摘要。
- 統一術語：兜帽女子、德克薩姆、弗佐爾・錢布瑞爾、穆爾馬斯特、
  總督察長、貝恩霸權、洛山達護符。
- 正文沿用倚天 16×15 粗體與 640×480 長文版面。

## 驗收

1. 未建立迪姆斯沃特同行狀態時，不把兜帽女子當無條件出口。
2. `(4,12)` 顯示 PICTURE 39、繁中提議與「是／否」。
3. 答應後通過弗佐爾闖入及離開兩個 pause。
4. 真實 `0x21→0x22` 後顯示德克薩姆、牛頭人與手札 30 兩頁。
5. 五項 encounter menu 與迪姆斯沃特護符提示為繁中。
6. 解鎖手札 7 兩頁，顯示弗佐爾死亡與枷印消退。
7. 完成護符離場／混戰後，回到 `GEO4/0x25 (5,7,W)`；再以一般 GEO 走格抵達
   同區塊的後續位置交易。
8. focused real-image test、game-pack test 與完整 Docker/Xvfb regression 通過。

## 本輪邊界

完成範圍止於眼魔洞穴恢復操作。洞穴探索、固定戰、德克薩姆／梅杜莎／牛頭人
決戰、取得洛山達護符與返回荒野屬後續規格。
