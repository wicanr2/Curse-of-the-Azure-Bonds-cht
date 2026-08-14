# 885 — 第二個速度選單：遊戲自己印出「0＝最快 9＝最慢」

- 證據等級：`exact`（兩平台各自逐條讀完）
- 作法見 spec 783

## `overlay-15:01B4Dh`（DOS 146 條）↔ `overlay-15:01B94h`（PC-98 138 條）

`retf` ＝ 無參數。與 spec 879 的 `overlay-08:011DFh` 是**同一個功能的第二個入口**
（另一個選單畫面），動的是同一格 `DS:4FA9h`。

| | DOS | PC-98 |
|---|---|---|
| 標題 | `'Game Speed = '` ＋ 數字 ＋ `' (0=fastest 9=slowest)'` | `'速さ＝'` ＋ 數字 ＋ `'　（０＝速い　９＝遅い）'` |
| 選項標題 | `'Game Speed:'` | `'速度：'` |
| 變快 | `' Faster'` | `' 速く '` |
| 變慢 | `' Slower'` | `' 遅く '` |
| 離開 | `' Exit'` | `'抜ける'` |

```pascal
repeat
    標題 := 'Game Speed = ' ＋ <far 1522h>(DS:4FA9h) ＋ ' (0=fastest 9=slowest)';
    <far 0542h:0352h>(0, 0Ah, 12h, 1);

    選項 := '';
    if DS:4FA9h > 0 then 選項 := 選項 ＋ ' Faster';    { 無號 }
    if DS:4FA9h < 9 then 選項 := 選項 ＋ ' Slower';    { 無號 }
    選項 := 選項 ＋ ' Exit';
    暫 := 'Game Speed:';

    鍵 := <far 169Dh+2>(var 有選中, 1, 1, 0Fh, 0Ah, 0Dh, 選項);

    if 有選中 <> 0 then begin
        if      (鍵 = 50h) and (DS:4FA9h > 0) then dec(DS:4FA9h)   { ↓ }
        else if (鍵 = 48h) and (DS:4FA9h < 9) then inc(DS:4FA9h);  { ↑ }
    end
    else if 鍵 = 'F' then dec(DS:4FA9h)
    else if 鍵 = 'S' then inc(DS:4FA9h);
until not <0A54h:08D4h>(離開鍵表, 鍵);
<far 1558h+1>();
```

## `' (0=fastest 9=slowest)'` 把兩個舊推論一次坐實

spec 879 從「`Slower` 走 `inc`」推出「值越大越慢」，
spec 881 從「門檻 ＝ 每格延遲 × (值 ＋ 3)」給出機制證據。
**這一支是遊戲自己印給玩家看的字，第三份、也是最直接的一份。**

值域 0..9 也由兩道邊界再確認一次。

⚠ **兩個入口的邊界檢查不一致**：
本支的 `'Faster'` 條件是 `DS:4FA9h > 0`、`'Slower'` 是 `< 9`；
spec 879 那一支寫反了顯示條件（`'Slower'` 配 `< 9`、`'Faster'` 配 `> 0`）
——**兩支其實是同一組條件，只是選項排列順序相反**。
不過本支在方向鍵那條路徑上有做邊界檢查，字母鍵那條路徑**沒有**
（`'F'` 直接 `dec`、`'S'` 直接 `inc`），所以**用字母鍵可以把值推出 0..9 之外**。
spec 879 那一支同樣沒有檢查。兩平台一致，remake 要補。

## PC-98：標籤與熱鍵是兩塊平行陣列，不可用時兩邊都寫 0

這一支把 spec 879 看到的機制補完整了：

```asm
; 可用時
mov di, 0A369h            ; 標籤欄位 A ← ' 速く '
mov byte ptr ds:0A33Bh, 46h   ; 熱鍵 'F'
...
mov di, 0A370h            ; 標籤欄位 B ← ' 遅く '
mov byte ptr ds:0A33Ch, 53h   ; 熱鍵 'S'

; 不可用時
mov byte ptr ds:0A369h, 0
mov byte ptr ds:0A33Bh, 0
mov byte ptr ds:0A370h, 0
mov byte ptr ds:0A33Ch, 0
```

所以 PC-98 的選單是**兩塊平行的固定陣列**：
標籤欄位（`DS:0A369h`／`0A370h`／`0A37Eh`，每欄 6~7 bytes）與
熱鍵字元（`DS:0A33Bh`／`0A33Ch`／…）。
**選項不可用時把標籤與熱鍵都寫 0**，而不是像 DOS 那樣「不串接那一段」。

中文版照這個做法：欄位固定、不可用就清 0。

## 明確不宣稱

- 沒有宣稱本支與 spec 879 那一支各在什麼場合被叫。
- 沒有宣稱 `far 0542h:0352h(0, 0Ah, 12h, 1)` 與 `far 1522h`／`169Dh+2`
  的完整簽章。
