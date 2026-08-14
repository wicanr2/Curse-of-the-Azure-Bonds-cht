# 1014 — 物品顯示名稱怎麼組出來，以及 PC-98 為在地化加的那張覆寫表

- 證據等級：`exact`（DOS 側 315 條逐條讀完；PC-98 對側 420 條，
  多出來的 105 條是一整段 DOS 沒有的覆寫查表，已逐條讀完）
- 作法見 spec 783

## `dos overlay-24:00467h` ↔ `pc98 overlay-24:0046Ch`（`retf 10h`，＝ `entry#1`）

兩側原本都是待解讀。spec 1013 看到商店清單用完就叫這一支「把名稱還原」，
現在確定它做的事更完整：**從零組出物品的顯示名稱，就地寫進物品節點的 `+0`**。

### 簽章

`retf 10h` ＝ 8 word。第一個宣稱的參數在最高位址：

```pascal
procedure 組名稱(未用:     遠指標;   { bp+12h，本體從不讀 }
                 物品:     遠指標;   { bp+0Eh }
                 欄, 列:   byte;     { bp+0Ch, bp+0Ah }
                 顯示已裝備: byte;   { bp+08h }
                 要印出來: byte);    { bp+06h }
```

spec 1013 的呼叫 `(NIL, 物品, 0, 0, 0, 0)` ⇒ 只重建、不印 ✓。

## 流程

```pascal
物品^ := '';                                       { ★ 名稱就地重建 }

if 顯示已裝備 <> 0 then
    if 物品^.ready <> 0 then 物品^ := ' Yes  '
                        else 物品^ := ' No   ';

{ ── 隊伍裡有人帶著效果 5 時，魔法／詛咒物品前面加一顆星 ── }
偵測 := false;
p := 遠指標(DS:650Ah);
while (p <> NIL) and not 偵測 do begin
    if <sub_2396>(@x, 5, p) then 偵測 := true;
    p := p^[189h];                                  { 隊伍鏈的 next }
end;
if 偵測 and ((物品^.plus > 0) or (物品^.plussave > 0) or (物品^.cursed <> 0)) then
    物品^ := 物品^ ＋ '*' ＋ ' ';

if 物品^.數量(+39h) > 0 then
    物品^ := 物品^ ＋ Str(數量) ＋ ' ';
```

欄位名（`ready`／`plus`／`plussave`／`cursed`）全部來自 spec 1011 的除錯畫面。

## ★★★ `identified` 是逐段的隱藏遮罩

```pascal
遮罩 := 0;
for i := 1 to 3 do
    if (物品^.namenum(i) <> 0)
       and ((物品^.identified shr (3 − i)) and 1 = 0) then
        遮罩 += 1 shl (i − 1);
```

> **`identified`（`+35h`）不是布林值，是三個位元：
> bit 2 蓋住 `namenum(1)`、bit 1 蓋住 `namenum(2)`、bit 0 蓋住 `namenum(3)`。**

這就是「未鑑定的物品只看得到基本名」的實作——鑑定一部分就露出一部分。

## ★★★ 名稱表：`DS:1040h`，筆距 21

```pascal
for i := 3 downto 1 do                              { ★ 由後往前接 }
    if (遮罩 shr (i − 1)) and 1 <> 0 then begin
        物品^ := 物品^ ＋ 字串(DS:1040h ＋ namenum(i) × 21);
        if (數量 < 2) or 已加過複數 then
            物品^ := 物品^ ＋ ' '
        else
            …選 `' '` 或 `'s '`…
    end;

if 要印出來 <> 0 then 印字(欄, 列, 0Ah, 0, 物品^);
```

表的內容就是 AD&D 的物品詞彙：

| 編號 | 內容 |
|---|---|
| 1 | `Battle Axe` |
| 2 | `Hand Axe` |
| 3 | `Bardiche` |
| 4 | `Bec De Corbin` |
| 5 | `Bill-Guisarme` |
| 6 | `Bo Stick` |
| 7 | `Club` |
| 8 | `Dagger` |
| 9 | `Dart` |
| 10 | `Fauchard` |
| 11 | `Fauchard-Fork` |
| 12 | `Flail` |
| 13 | `Military Fork` |
| … | … |

**筆距 21 ⇒ 長度 byte ＋ 最多 20 字元。**
PC-98 把它移到 `DS:20E7h`、**筆距加大到 `1Fh` ＝ 31**（30 bytes ＝ **15 個全形字**）。

## ★ 英文複數：`'s '`

分隔符有兩個：`' '`（`CS:0462h`）與 **`'s '`（`CS:0464h`）**。
只有在**數量 ≥ 2 且這一輪還沒加過複數標記**時才可能選 `'s '`，
選了就把「已加過」旗標立起來，後面的段一律用 `' '`。

選哪一個由一串條件決定（依程式的判斷順序）：

```pascal
if 遮罩 = 1 shl (i − 1)                              then 用 's '
else if (i = 1) and (遮罩 > 4) and (itemptr <> 56h)  then 用 's '
else if (i = 2) and odd(遮罩)                        then 用 's '
else if (i = 3) and (itemptr = 56h)                  then 用 's '
else if namenum(3) = 87h                             then 用 ' '
else if (itemptr in [49h, 1Ch]) and (namenum(3) = 0B1h) then 用 ' '
else                                                      用 's ';
```

⚠ **中文沒有複數形**——這一整段對中文版是純粹的雜訊，
而且 `'s '` 會直接接在漢字後面。中文化必須把它換成空字串或直接跳過。

## ★★★ PC-98 為在地化加了一張覆寫表

PC-98 在函式最前面多了 105 條，DOS 完全沒有：

```pascal
n1 := 物品^.namenum(1);  n2 := namenum(2);  n3 := namenum(3);
命中 := 0;
for i := 1 to 0A1h do
    if (byte[40BDh ＋ i × 3] = n1)
   and (byte[40BEh ＋ i × 3] = n2)
   and (byte[40BFh ＋ i × 3] = n3) then begin
        命中 := i;
        n1 := word[42A6h ＋ 命中 × 8];              { ★ 換掉三個編號 }
        n2 := word[42A8h ＋ 命中 × 8];
        n3 := word[42AAh ＋ 命中 × 8];
        if 物品^.identified <> 0 then
            物品^.identified := byte[42ACh ＋ 命中 × 8];   { ★ 連遮罩也換 }
        跳出迴圈;
    end;
```

之後組名稱時用的是 **`n1`／`n2`／`n3` 這三個區域變數**，不是物品節點裡的原值。

| 表 | 位址 | 筆距 | 內容 |
|---|---|---|---|
| key | `DS:40BDh` | 3 bytes | 要比對的三個 `namenum` |
| value | `DS:42A6h` | 8 bytes | 換上去的三個編號（word）＋ 新的 `identified`（`+6`） |

最多 **161 筆**（迴圈上限 `0A2h`）。前幾筆長這樣：

```
key (0,   0, 1) → value (0,   0, 1, 遮罩 0)      { 原樣 }
key (0, 162, 1) → value (0, 162, 1, 遮罩 6)      { ★ 只改遮罩：藏掉前兩段 }
key (0,   0, 2) → value (0,   0, 2, 遮罩 0)
```

> ★★★ **英文的物品名可以用「三段拼接」組出來，日文不行——
> 語序不同、也不是每個組合都能逐段翻。
> PC-98 的做法不是改組裝邏輯，而是加一層「特定組合 → 改寫成另一組編號＋遮罩」的表。**

**中文化直接沿用這個機制**：中文同樣是「長劍 ＋1」而不是 `Long Sword +1` 的語序，
`Ring of Protection` 要變成「防護戒指」——正好是覆寫表要解的問題。
表在資料段、筆距固定，可以整批換。

## 兩平台位址與欄位

| | DOS | PC-98 |
|---|---|---|
| 名稱表 | `DS:1040h`，筆距 **21** | `DS:20E7h`，筆距 **31** |
| 覆寫表 key／value | — | `DS:40BDh`／`DS:42A6h` |
| 隊伍鏈頭／next | `DS:650Ah`／`+189h` | （`+18Ah`） |
| `namenum(1..3)` | `+2Fh`／`+30h`／`+31h` | `+57h`／`+58h`／`+59h` |
| `plus`／`plussave` | `+32h`／`+33h` | `+5Ah`／`+5Bh` |
| `ready`／`identified`／`cursed` | `+34h`／`+35h`／`+36h` | `+5Ch`／`+5Dh`／`+5Eh` |

物品節點一律 `+28h`（spec 1011／1013）。

## 中文化

三件事按重要性排：

1. **名稱表 `DS:1040h`**（DOS 20 字元／PC-98 30 bytes ＝ 15 全形字）——
   物品詞彙全在這裡，整批可換。**要跟 PC-98 的 31 筆距。**
2. **覆寫表 `DS:40BDh`／`DS:42A6h`**——語序與特例靠它，PC-98 已經備好機制。
3. **`'s '` 要拿掉**，中文沒有複數形。

`' Yes  '`／`' No   '` 各 6 字元（＝ 3 個全形字），是欄位對齊用的固定寬度，
中文只能用 3 個字（例如「已裝備」「未裝備」正好）。

## 明確不宣稱

- 沒有宣稱 `sub_2396(@x, 5, 角色)` 是什麼檢查，只知道效果碼 5 會讓魔法物品前面出現 `*`。
- 沒有宣稱名稱表有幾筆。
- 沒有宣稱 `itemptr = 56h`／`49h`／`1Ch` 與 `namenum(3) = 87h`／`0B1h`
  這幾個特例各對應哪些物品。
- 沒有宣稱覆寫表實際用到幾筆（迴圈上限 `0A2h` 是編譯期常數，不代表資料筆數）。
- 沒有宣稱第一個參數（`bp+12h` 的遠指標）為什麼存在。
