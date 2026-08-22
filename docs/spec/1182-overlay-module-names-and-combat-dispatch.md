# 1182 每個 overlay 的原始單元名，以及 `24h COMBAT` 的第四支

狀態：`READY`

## ★★★ 先講被推翻的那一條：`overlay-04` 是**神殿**不是營地

spec 1030 把 `dos overlay-04:00DD6h` 標成「營地（Camp）主選單」，spec 1095／1149
沿用，並因此說 remake 把 `7EE2h` 標成 `TempleRequested` 是「位址對、語意不符」。

**反了。** 兩份一手資料都說那是神殿：

1. **PC-98 的 Borland 除錯符號**：`overlay-04` 裡只有兩個符號，
   `LOADTEMPLE` 與 `GOTEMPLE`。營地在 `overlay-15`（`LOADCAMP`／`DOCAMP`）。
2. **DOS `overlay-04` 自己的字串**：`how can we help you?`、
   `l View Pool Appraise Exit`（＝ `Heal View Pool Appraise Exit`）、
   `e Dead`（＝ `Raise Dead`），以及
   `ou leave a priest says, "Excuse me but you have left some money here"`。

Heal／Raise Dead／Appraise／**a priest** ——神殿。spec 1030 讀到的那張選單
（`H` Heal、`V` View、`T` Take、`P` Pool、`S` Share、`A` Appraise）也是神殿的，
不是營地的 Save／Rest／Magic／Alter。

⇒ **remake 原本的 `TempleRequested` 一直是對的**，被兩份規格「改正」成錯的。
常數 `addrCampRequest` 已更名回 `addrTempleRequest`。

## 全部 36 個 overlay 的單元名

`cmd/borland-symbols -overlay-names` 產 `docs/audit/overlay-module-names.md`。
認法：Turbo Pascal 每個 overlay 單元的第一個 entry 一律是 `LOADxxx`
（overlay manager 的載入 stub），`xxx` 就是單元名。

⚠ **DOS 版沒有除錯符號**，所以 DOS 側的 `overlay-NN` 一直是靠內容推的——
**推錯了不會有任何徵兆**，`overlay-04` 就是這樣錯了很久。兩版的 overlay 編號
一致，這一點有兩處獨立印證：DOS 側規格早就各自認出 `overlay-05` ＝ POSTCOM
（spec 1156）與 `overlay-30` ＝ THREED／`LOADWALLSET`（spec 1153），
與 PC-98 符號表完全吻合。

## `24h COMBAT` 是**四選一**

`overlay-02:179Ah` 的前三支見 spec 1149。第四支（兩個旗標都不成立）先前
只知道叫 `sub_184B`，內容沒讀。讀完之後：

```pascal
{ sub_184B }
if bank1^[5C4h] = 1 then begin        { ＝ ECL 格 7EE2h }
    bank1^[5C4h] := 0;                { 讀到就清零 }
    <overlay-04 TEMPLE>();            { 0059:002Ah／0059:0025h }
end else
    <overlay-05 POSTCOM>();           { 0060:002Ah／0060:0025h ＝ DOPOSTCOMBAT }

{ 兩路收尾都重繪，依畫面模式二選一 }
if bank0^[1CCh] = 0 then <overlay-27 OVERLAND entry#6>()
else                    <overlay-30 THREED   entry#11>();
```

所以完整的分派是

| 順序 | 條件 | 去哪 |
|---|---|---|
| 1 | `DS:8B69h` 或 `DS:8B56h` 非 0 | 打（`overlay-08` GOCOMBAT）|
| 2 | `bank1^[6D8h]`（`7F6Ch`）＝ 1 | 商店（`overlay-06` SHOP）|
| 3 | `bank1^[5C4h]`（`7EE2h`）＝ 1 | **神殿**（`overlay-04` TEMPLE）|
| 4 | 以上都不成立 | **戰後處理**（`overlay-05` POSTCOM）|

★ 第 4 支解釋了那 46 處「沒擺過怪的 `24h`」：腳本用 `27h TREASURE` 把戰利品堆
好，再用 `24h` 開**戰後處理／分配畫面**。`overlay-05` 的字串正是那一組——
`The party has found Treasure!`、`Each character receives `、`Money Items Exit`、
`View Take Pool Share`。

★★ 收尾的重繪閘門 `bank0^[1CCh]` 就是 spec 1181 從 `21h LOAD FILES` 認出來的
那一格（ECL 格 `4BE6h`，「現在是不是第一人稱畫面」）。兩支 handler 各自獨立
用它二選一，互相印證。

## remake 這一側

- 常數與註解改正（`addrTempleRequest`）。行為沒有變——那一支本來就發
  `TempleRequested`。
- 第 4 支接上了：**沒擺過怪的 `24h` 改發 `PostCombatRequested`**，不再發
  `CombatRequested`。原作的 `24h` 沒有「零隻怪的戰鬥」這種東西——有怪才走
  `sub_1956`，所以那個組合本來就不該存在。
  上層把它導到戰利品分配（地城裡沿用「續跑存下的 ECL PC」那條路，
  原作那一支等於立刻結束並回到腳本）。
  ★ 改完之後 10 個既有測試從「零隻怪的戰鬥」變成「戰利品分配」，
  每一個都是真的撿東西現場（外圍廢墟倉庫、逃亡者的暗格、賭場、內城臥房、
  埋葬谷的斯瑞克林營地、火刀辦公室、洞穴裡死去精靈的皮袋）——這比任何合成
  測試都更能說明那 46 處是什麼。
## ★★ 商店與神殿旗標的 producer 是**腳本自己**

spec 1095 寫過「這兩格在 1,355 條 ECL 指令裡沒有任何一條寫過（已全掃確認）」
——**那是假零**。掃全 corpus 得到 `7EE2h`（神殿）**4 處**、`7F6Ch`（商店）**9 處**，
慣用法一模一樣：

```
CLEARMONSTERS          ← 保證第一支（有怪就打）不成立
SAVE 01 <旗標>
COMBAT
```

例如 `ECL1/0x50:0822h`（世界地圖的酒館那一段）與 `ECL2/0x01:11E3h`。

⇒ 兩支服務**都走得到**，而且 remake 本來就接得起來——VM 讀 `memory`，
腳本用 `09h SAVE` 寫進同一張 map。`TestRealTavernTempleRequestReachesTheTempleBranch`
與 `TestRealShopRequestReachesTheShopBranch` 直接跑原版資料驗過，
連「讀到就清零」也一起釘住。

先前釘「沒有 producer」的 `TestCampRequestFlagHasNoProducer` 已經移除——它釘的是
一個假零。取代它的 `TestServiceRequestFlagsHaveScriptProducers` 是**正對照**：
掃出 0 就直接說「那是假零，先確認掃描而不是下結論」，數字與 4／9 對不上也紅。

## ★★ 第一支的第二個條件：`DS:8B56h` ＝ 決鬥旗標，而決鬥在這一款走不到

`24h` 的第一支是 `8B69h <> 0` **或** `8B56h <> 0`。前者是 `1Ch CLEARMONSTERS`
清的「有怪要打」旗標（spec 1145），remake 用**怪物鏈非空**代替。後者一直標著
「不宣稱」——其實答案早就在自己的規格裡：

- **`DS:8B56h` ＝ 是不是決鬥**（spec 917 判讀戰後結算的措辭時就解出來了：
  `8B56h <> 0` 時印 `You have won the duel.` 與 `The duelist receives `，
  否則印 `The party has won.` 與 `Each character receives `）。
- 整個 DOS 執行檔裡**只有一處**把它設成 1：`overlay-07 entry#25` ＝ **`GODUEL`**
  （spec 626；DOS 那一版複製出來的分身叫 `ROLF`，PC-98 叫 `ロルフ`）。
  其餘三處都是寫 0（`overlay-11` INIT 兩處、`overlay-02` 一處）。
- `GODUEL` 只有 `2Dh CALL` 的兩個選擇子到得了：`8000h` ⇒ `GODUEL(1)`、
  `8001h` ⇒ `GODUEL(0)`（spec 1150）——而 **corpus 兩個都沒用到**。

⇒ **`8B56h` 那一支在 CoAB 走不到**，所以 remake 用怪物鏈非空覆蓋了整個可達條件。
`TestDuelCallSelectorsAreUnusedWhileTheOthersAreNot` 釘住這件事，而且**先做正對照**
（那四個真的有在用的選擇子要掃到 125／19／13／11）才讓那兩個零算數。

⇒ `24h` 轉 **`done`**：四支全部接上，各有實機路徑。

⚠ 決鬥本身仍是 `2Dh CALL` 的未接目標之一，那屬於 `2Dh` 的帳。

## 回歸

`TestOverlayFourIsTheTempleNotTheCamp` 直接對**遊戲檔**斷言：DOS `overlay-04`
的字串裡必須有神殿那組（`priest`／`Raise Dead`／`Appraise`），而且不能被當成
營地。名字錯回去的話這條會紅。

## 不宣稱

- 沒有宣稱 `DS:8B56h` 是什麼（只知道它與 `8B69h` 並列，非 0 就去打）。
- 沒有宣稱 `overlay-04`／`overlay-05` 兩個入口 stub（`0025h`／`002Ah`）各自
  對應 `LOADxxx` 與 `GOxxx` 的哪一個——只確認它們都落在該 overlay。
- 沒有宣稱 `overlay-00`／`overlay-01` 的單元名（那兩個沒有 `LOADxxx` 符號）。
- 沒有宣稱 spec 1030 讀到的那張選單在遊戲裡的完整流程有沒有其他錯處，
  只確認它是**神殿**的選單。
