# 1178 物品名稱是三個編號組出來的：`DS:1040h` 那張 255 筆的詞表

狀態：`READY`

## 結論

物品的顯示名稱**不含物品類別**。名稱完全由紀錄裡的三個 byte（`+2Fh`／`+30h`／
`+31h`，本專案叫 `NameNumbers`）各自去查一張詞表，再串起來。

詞表在 DOS 版的資料段：

| 項目 | 值 |
|---|---|
| 第 N 筆的位址 | `DS:1040h + N × 15h` |
| 每筆的形式 | Pascal `string[20]`（1 個長度 byte ＋ 20 個字元 byte） |
| 編號範圍 | `01h`..`FFh`，共 255 筆（`00h` ＝ 沒有這個成分） |
| 第一筆 | `01h` ＝ `Battle Axe`（`DS:1055h`） |
| 最後一筆 | `FFh` ＝ `Cursed`（`DS:252Bh`） |
| 空白的格子 | `3Eh`／`3Fh`／`90h` 三筆長度是 0 |

`DS:2540h` 起是**另一張表**（`N`、`NW`… 方位名），不是名稱成分。編號是一個
byte，所以 `01h..FFh` 已經是全部。

## 證據

指令在 `overlay-24:063Bh`（DOS）：

```
0630  26 8A 45 2E     mov al, es:[di+2Eh]   ; di ＝ 物品 ＋ N ⇒ 讀 +2Eh+N
0636  BA 15 00        mov dx, 15h
0639  F7 E2           mul dx                ; ax ＝ 編號 × 15h
063B  8B F8           mov di, ax
063D  81 C7 40 10     add di, 1040h         ; ← 基底
0641  1E 57           push ds; push di
0643  9A C1 06 54 0A  call far 0A54:06C1    ; Pascal 字串相接
```

`1040h` 是 **0 那一格的位址**，所以編號直接當索引用、不用減一。

資料側的正對照：把六章 `ITEM*.DAX` 全部 253 件物品照這個規則組回去，**208 件與
紀錄裡存的名字逐字元相同**；基底改成 `1040h ± 15h` 或 `± 2×15h` 的話，相同的件數
是 **0**。剩下 45 件的差異全部有解釋（見下面「存的那個名字是什麼」）。

## 組名規則（`overlay-24:0467h` ＝ `PRINTITEMNAME`）

> `overlay-24` 的原始單元名是 **GENERIC**，而這一支的 Borland 符號叫
> `PRINTITEMNAME`（spec 1182 的 overlay 單元名表）。

```
if Count(+39h) > 0:            名字 := Str(Count) + ' '
for N := 3 downto 1:
    if NameNumbers[N-1] == 0:            跳過
    if HiddenNameFlags(+35h) bit (3−N):  跳過
    名字 := 名字 + 詞表[NameNumbers[N-1]]
    if Count >= 2 且還沒加過複數:  名字 := 名字 + 's '
    else:                          名字 := 名字 + ' '
```

四件容易做錯的事：

- **成分由後往前**：`+31h` → `+30h` → `+2Fh`。`Long Sword +1` 的三格是
  `{00, A2h, 24h}`，先出 `Long Sword` 再出 `+1`。
- **數量連 `1` 都印**。原作的 `1 Oil`、`2 Javelin` 不是排版失誤，是這一條。
- **`+35h` 不是布林**，是三個 bit：`4h`／`2h`／`1h` 分別藏住第一／第二／第三個
  成分。魔法物品出廠時是 `6h`，鑑定之後才露出加值與附魔名。
- **`+32h`（Plus）與 `+36h`（Cursed）不進名字。** 加值是走名稱編號
  `A2h..A6h`（`+1`..`+5`）那幾個**可以被藏起來**的成分。原版 72 件帶 Plus 的
  物品裡只有 18 件同時帶 `+N` 成分——剩下 54 件在原作從頭到尾不顯示加值。

複數那一段還有三個特例（`+2Eh == 56h` 的油瓶、`NameNumbers[2] == 87h`、
`+2Eh ∈ {49h, 1Ch}` 且 `NameNumbers[2] == B1h`）決定 `'s '` 加在哪一個成分上。
中文沒有複數，remake 這一段整段不適用。

## 存的那個名字是什麼：**輸出欄位，不是資料來源**

紀錄的 `+00h..+29h` 是一個 Pascal `string[41]`。`PRINTITEMNAME` 一進來第一件事就是

```
0471  26 C6 05 00     mov byte ptr es:[di], 0     ; 長度 byte := 0 ⇒ 名字清空
```

**無條件清掉它**，然後把組出來的名字一段一段寫回同一個欄位。所以那個欄位是
組名的**輸出**，`ITEM*.DAX` 出廠帶的值只是作者工具留下的快取，遊戲一顯示就覆蓋掉。

這解釋了 45 件對不上的來源，也說明**沒有東西需要額外翻譯**：

| 差異 | 例 | 為什麼 |
|---|---|---|
| 數量前綴 | 存 `10 Arrow`，成分組出 `Arrow` | 前綴是組名時才加的 |
| 過期的快取 | 存 `Deck`／`Beaker`／`Small Raft` | 這些字**不在詞表裡**，資料改過而快取沒更新 |
| 縮寫 | 存 `Magic User Scroll`，詞表是 `MU Scroll` | 同上，快取來自另一個版本的字串 |

⚠ 那幾個過期的字**不是未鑑定時的替身名**：`Deck` 那一件的 `+35h` ＝ `6h`，
未鑑定時組出來的是 `Gloves`，與 `Deck` 對不上。

## remake 這一側

`monster.LocalizedItemName` 改成照原作的模型組名：類別名只在三個成分都看不見
或都沒翻譯時當保險絲——**原版 253 件物品沒有一件走到那條路**（每一件都至少有
一個名稱編號）。`item_plus`／`item_cursed_suffix` 兩個語系鍵不再參與組名。

`assets/locale/zh-TW.json` 的 `item_name_XX` 由 54 筆補成 **252 筆**（詞表 255 筆
扣掉三個空白格）。先前那 54 筆是第 260 輪照一份 reference 實作的 `itemNames`
陣列填的，**與原作的編號對不起來**：`24h`..`3Bh` 那一段剛好正確，`9Fh` 之後整段
偏了 3（語系檔的 `item_name_9F` 寫 `+1`，而原作 `+1` 是 `A2h`），`3Ch`..`3Eh`
則是拿物品類別的編號去填的。玩家看得到的後果是**加值顯示錯數字**。

### 一個翻譯層的決定：連接詞前後對調

原作把 `of`（`A7h`）、`Of Prot.`（`E0h`）等連接詞當成一般成分排在中間，英文因此
讀成 `Wand of Fireballs`。中文的修飾語在中心語前面，照原順序會排成「魔杖之火球」。
所以**三個成分而中間是連接詞時，把前後兩段對調** →「火球之魔杖」。

這是翻譯層的規則，不動任何原始欄位；連接詞以**名稱編號**認定
（`A7h`、`C8h`、`E0h`、`EBh`、`F9h`、`FAh`、`FBh`），不是認翻譯出來的字。

## 回歸

- `cmd/item-name-audit` 產 `docs/audit/item-names.md`：253 件、**115 種相異名稱**、
  用到 126 個名稱編號、**缺譯 0**。
- `TestWordTableBaseAndStrideAreAnchored` 釘住基底與間距（含表尾 `FEh` ＝ `Pass`，
  防止把筆數往上調而吃到方位名那張表）。
- `TestEveryUsedNameNumberIsTranslated` 是 fail-closed 閘門：缺一個成分的翻譯，
  玩家看到的是**少一截的名字**而且不會有任何錯誤訊息。
- `TestComposedNamesMatchTheGoldenSample` 走完整條鏈（DAX 欄位 → 編號 → 詞表 →
  語系檔 → 組名規則）六個樣本。
- `TestLocalizedItemNameIgnoresTheItemTypeWhenNameNumbersExist`、
  `TestLocalizedItemNameDoesNotPrintPlusOrCursed`、
  `TestLocalizedItemNameSwapsAroundConnectors`。

## 不宣稱

- `PRINTITEMNAME` 開頭那個走 `DS:650Ah` 鏈、決定要不要標 `*` 的判斷（`cs:2396h` 收 `5`
  這個參數）到底在問什麼。remake 目前沒有 `*` 標記。
- `ITEM*.DAX` 出廠那份過期快取（`Deck`／`Beaker`…）是哪一版的資料留下來的。
- PC-98 版的詞表位址（本輪只查了 DOS 版的資料段）。
