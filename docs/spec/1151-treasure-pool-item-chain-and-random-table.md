# 1151 — `27h TREASURE`：錢幣池是覆寫、物品鏈是前插、隨機表兩端都夾

- 證據等級：`exact`（DOS `overlay-02:1B53h`..`1F45h` 整支 398 條讀完）
- 戰利品池由 `1Ch CLEARMONSTERS` 清見 spec 1145；物品節點版面見 spec 1036
- handler 位址取自 `docs/audit/ecl-opcode-handlers-dos.md`

## 骨架

```pascal
READVAR(8);
for i := 0 to 6 do                          { 1B6Bh }
    LongInt(DS:[6F70h + 4i]) := ADDRESSVALUE(i + 1);   { ★ 覆寫，不是累加 }

n := ADDRESSVALUE(8);                        { 1B95h }
if n < 80h then          載入 ITEM<片>.DAX 的區塊 n，把裡面**每一筆**掛上鏈
else if n = 0FFh then    什麼都不做
else                     隨機產生 n − 80h 件，最後逐件印出名字
```

七個錢幣運算元寫成 **32-bit**，總共 28 bytes——正好是 `1Ch` 歸零的那一段。

## `n < 80h`：把整個 ITEM 區塊掛上鏈

檔名由 `DS:5BEEh`（資料片編號）組成，載不到就印訊息並結束（`1C1Ch`、`1C21h`）。
接著逐筆 63 bytes 搬進鏈：

```asm
1C2B  Move(區塊 + 位移, @暫存, 3Fh)          ; 每筆 63 bytes
1C4F  if 鏈頭 = NIL then GetMem(@DS:6F8Ch, 3Fh); Move(暫存 → 鏈頭, 3Fh)
1C7C  else 舊頭 := 鏈頭; GetMem(@DS:6F8Ch, 3Fh); Move(暫存 → 新頭, 3Fh);
1CB3       新頭^[2Ah] := 舊頭                ; ★ next 在 +2Ah，而且是前插
1CBB  直到走完整個區塊；收尾 FreeMem(區塊)
```

⇒ **鏈的順序是區塊順序的反序**。

## `n > 80h`：隨機表

`n − 80h` 件，每件擲一次表。三個被叫到的常式名字由 PC-98 Borland 符號表依
entry 槽號對回來：`overlay-23 entry#9` ＝ `ROLLDICE`、`overlay-21 entry#18` ＝
`CREATERNDTREASURE`（spec 1036）、`overlay-24 entry#1` ＝ `PRINTITEMNAME`。

```text
a := ROLLDICE(1, 100)
  a ∈ [1,60]   → b := ROLLDICE(1, 100)
                   b ∈ [1,47] ∪ [50,59]  → 物品 b（b = 45 改成 59）
                   b ∈ [60,90]           → c := ROLLDICE(1, 10)
                                            c ∈ [1,4] → 36   c ∈ [5,7] → 35
                                            c = 8     → 34   c = 9     → 37
                                            c = 10    → 38
                   b ∈ [91,94]           → 73
                   b ∈ [95,97]           → 93
                   b ∈ [98,100]          → 77
                   其餘（b ∈ {48,49}）    → 59
  a ∈ [61,85]  → 61
  a ∈ [86,92]  → 62
  a ∈ [93,98]  → d := ROLLDICE(1, 15)；d ∈ [1,9] → 71，d = 10 → 84，d ∈ [11,15] → 79
  a ∈ {99,100} → 59
```

★ **每一段都是兩端都夾的區間，不是 `else if` 一路往下。** 原作在 `1D61h` 明寫
`cmp ax, 3Ch / jl`，所以第二段是 `[60,90]`；`b ∈ {48,49}` 四段都不收，一路掉到
`1DEFh` 的預設值 **59**。

⚠ 第三段 `a ∈ [91,98]` 被前一段 `[86,92]` 遮住，實際只吃得到 93..98。原作就是
這樣寫的，不是筆誤——照抄。

產生的每一件都走 `CREATERNDTREASURE(物品編號, @暫存)` 建成 63 bytes 的節點，
再以同樣的前插掛上鏈（`1E79h`..`1EF3h`，第一件會明寫 `+2Ah := NIL`）。

**只有隨機這一路**在最後沿鏈逐節點叫 `PRINTITEMNAME`（`1EFEh`..`1F40h`）；
`n < 80h` 那一路直接跳到收尾。

## 顯示順序：從鏈頭走

戰後撿東西的清單是 `overlay-05 POSTCOM` 的 `0CF5h`..`0D35h`：從 `DS:6F8Ch` 開始、
每個節點叫一次 `PRINTITEMNAME`、沿 `+2Ah` 往下。**head-first**。

⇒ 玩家看到的順序 ＝ 鏈的順序 ＝ 載入／產生順序的**反序**。

⚠ `SHOWLOOT`（`overlay-19 entry#4`，DOS `0597h`）只畫**錢幣那一格**（七個名稱在
`DS:0F93h`，每筆 `0Bh` bytes），完全沒有碰物品鏈——名字聽起來像，但不是這條路。

## remake

- 隨機表的區間照抄了。★ 先前 `b ∈ {48,49}` 會掉進劍那一段（`itemRoll <= 90` 寫成
  `else if`），現在回 59。`TestRandomTreasureItemTypeFollowsTheOriginalRanges`
  逐點釘住 30 個邊界。
- 戰利品清單改成**前插**，對上原作的鏈序與 head-first 顯示。
- `TreasureRequest` 的八個運算元由 `operandValue` 解析，所以 corpus 裡把數量或
  `ItemBlock` 放在 ECL 格的用法（`7F79`／`7F80` 那幾處）本來就走得通。

- 隨機那一路接上了 `CREATERNDTREASURE`（`monster.BuildRandomTreasureItem`，
  spec 1036）：擲完類別之後照原作補加值、名稱三段索引、重量、價值、卷軸法術
  與特殊物品範本，撿東西畫面上顯示得出 `＋N` 與名稱修飾。
  兩邊共用同一條 seeded 亂數流，擲骰順序也照原作（`plus` 先擲、`15h` 的 1d5 後擲、
  `47h` 的 1d8 在範本選擇裡擲），順序被 `scriptedRolls` 逐顆釘住。

★ **多筆 TREASURE 的模型也對上了**（`done`）：**錢幣池覆寫、物品鏈累積**。
那在原作是兩套資料結構——七個池是 `LongInt(...) := 運算元`，而物品是前插進
`DS:6F8Ch` 的鏈——remake 先前把錢幣也當成累加。corpus 沒有觀察到「兩筆之間沒有
`1Ch`」的情形，所以今天看不出差別；照抄是因為那就是原作的行為。
`TestECLTreasureCoinsOverwriteWhileItemsAccumulate` 兩側各釘一半，突變驗過。

## 明確不宣稱

- 沒有宣稱 63 bytes 節點裡除了 `+2Ah`（next）以外的欄位——那是 spec 1036 的範圍。
- 沒有宣稱錢幣池那 28 bytes 怎麼換算成金幣；本 handler 只把運算元原樣寫進去。
- 沒有宣稱 `n < 80h` 那一路載入失敗時 `06EAh:0` 做什麼（只知道印完訊息就走它）。
- 沒有宣稱隨機表產生的物品編號各自是什麼東西。
