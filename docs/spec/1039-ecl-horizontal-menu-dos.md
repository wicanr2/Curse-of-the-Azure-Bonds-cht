# 1039 — ECL `2Bh`（HORIZONTAL MENU）DOS 側：波浪號熱鍵標記是這裡組出來的

- 證據等級：`exact`（DOS 側 161 條逐條讀完）
- 作法見 spec 783；PC-98 對側見 spec 605

## `dos overlay-02:01082h`（`retf`）

原本待解讀，PC-98 對側（`overlay-02:010C2h`）已由 spec 605 讀完。

## 變長指令的骨架（與 PC-98 一致）

```text
<overlay-07 entry#2>(2)                    ← READVAR(2)
目的 := <overlay-07 entry#9>(DS:76C6h, DS:7706h)   ← ADDFNC(high[1], low[1])
n    := <overlay-07 entry#1>(2)            ← ADDRESSVALUE(2)，選項數
dec DS:4FB4h                               ← ECL_PC − 1
<overlay-07 entry#2>(n)                    ← 再讀 n 個 operand
```

⇒ spec 564 從解碼器反推的變長模式，在 **DOS 側也是一模一樣**。

## ★★★ 選項行是執行期串出來的，而波浪號就是熱鍵標記

```pascal
選項行 := '';
for i := 1 to n − 1 do begin
    選項行 := 選項行 ＋ '~' ＋ 文字表[i] ＋ ' ';      { 上限 32h ＝ 50 }
end;
選項行 := 選項行 ＋ '~' ＋ 文字表[n];                 { ★ 最後一個不接空白 }
鍵 := <overlay-07 entry#20>(…, 選項行, 0Dh, 樣式…);
<overlay-07 entry#15>(鍵, …);
```

★ `'~'` 在 `CS:107Eh`、`' '` 在 `CS:1080h`，各 1 字元。

> ★★★ **spec 1030 在營地選單看到的 `'~Yes ~No'` 波浪號標記，
> 產生端就在這裡**：ECL 的水平選單**逐個選項在前面補一個 `~`**。
> 也就是說波浪號不是某幾個畫面的特例寫法，而是
> **DOS 版水平選單的通用熱鍵標記**——`~` 後面的第一個字元就是熱鍵。
>
> ★★★ 對照 spec 605：**PC-98 把這套整個換掉**，
> 改成把熱鍵寫成 `DS:0A334h ＋ i := i ＋ 40h`（`'A'`、`'B'`、`'C'`…），
> 選項文字則搬進 `DS:0A339h ＋ i × 7`。
> ⇒ ⚠ **中文化必須跟 PC-98**：`~` 後面接 Big5 首位元組會把熱鍵判成亂碼。

## ★ 文字表：兩平台都是 stride 256

```asm
1157  mov  di, ax          ; i
1159  mov  cl, 8
115B  shl  di, cl          ; ★ i × 100h
115D  add  di, 7648h
```

| | 表位址 | 筆距 | 索引 |
|---|---|---|---|
| DOS | `DS:7648h` | `100h` ＝ 256 | 從 1 起算 |
| PC-98 | `DS:0A8DAh` | `100h` ＝ 256 | 從 1 起算 |

## ★★ `n = 1` 的特例

```pascal
if n = 1 then begin
    樣式 := (1, 0Fh, 0Fh);
    if 字串(DS:7748h) <> 'PRESS BUTTON OR RETURN TO CONTINUE.' then
        DS:7748h := 'PRESS <ENTER>/<RETURN> TO CONTINUE';
end else
    樣式 := (0, 0Fh, 0Ah);
```

★ **DOS 有兩句「按鍵繼續」**：
`'PRESS BUTTON OR RETURN TO CONTINUE.'`（35，含句點）與
`'PRESS <ENTER>/<RETURN> TO CONTINUE'`（34，不含句點）。
**已經是前者就不動**（形狀上是「有搖桿時的說法留著」），否則換成後者。
⇒ PC-98 只有一句「リターン・キーを押してください」（spec 605）。

★ 樣式參數 `(1, 0Fh, 0Fh)` vs `(0, 0Fh, 0Ah)` 與 spec 605 記的
「`0` 而不是 `0Fh`／`0Ah`」對得上。

## ★★★ 關掉 spec 605 的一個未定項

spec 605 記「`DS:BDF4h`／`BDF5h` 如何決定樣式參數 `var_3C`」沒有宣稱。
DOS 側寫得很直白：

```pascal
var_3C := ord((DS:8B62h <> 0) and (DS:8B63h <> 0));
```

⇒ **兩個旗標都非零才是 1**（PC-98 對應 `DS:0BDF4h`／`0BDF5h`）。

## 收尾

`0542:0B4Ah(18h, 27h, 18h, 0)`，與 spec 605 記的 PC-98
`0418:14C7h(18h, 27h, 18h, 0)` **四個參數完全相同**。

## 中文化

| DOS | 長度 | PC-98 |
|---|---|---|
| `'PRESS BUTTON OR RETURN TO CONTINUE.'` | 35 | — |
| `'PRESS <ENTER>/<RETURN> TO CONTINUE'` | 34 | 「リターン・キーを押してください」 |
| `'~'`／`' '` | 1／1 | ★ 不存在（PC-98 走熱鍵表） |

⚠ 選項行的上限是 **`32h` ＝ 50 bytes**，`n` 個選項加上 `~` 與空白全部要塞進去
——中文全形字**整行最多 25 個字**，且要扣掉分隔符。
⚠ 每個選項的原始文字上限是 256 bytes（表的筆距），但**組出來的行只有 50**。

## 明確不宣稱

- 沒有宣稱 `overlay-07 entry#15`（把選擇結果交出去的那支）做什麼。
- 沒有宣稱 `DS:7648h` 那張 256-bytes-per-entry 的表由誰填
  （PC-98 側是 packed text 解出來的，見 spec 605）。
- 沒有宣稱 `DS:8B62h`／`DS:8B63h` 兩個旗標各自是什麼。
- 沒有宣稱 `DS:7748h`（那句提示字串）還被誰用。
- 沒有宣稱選項超過 50 bytes 時會發生什麼。
