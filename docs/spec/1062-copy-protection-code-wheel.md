# 1062 — 防拷轉輪：完整的公式與 6 × 37 的答案表

- 證據等級：`exact`（DOS 側 318 條逐條讀完）
- 作法見 spec 783

## `dos overlay-03:0011Ah`（`retf`）

原本待解讀。這是《Curse of the Azure Bonds》的
**translation wheel（轉輪）防拷檢查**。

## 畫面

```pascal
<overlay-33 entry#2>('tiles', 1, 0, 1Ah);           { 載入 espruar 字模 }
<overlay-33 entry#2>('tiles', 2, 1Ah, 16h);         { 載入 dethek 字模 }
印字(3, 2, 'Align the espruar and dethek runes');   { 34 }
印字(3, 3, 'shown below, on translation wheel');    { 33 }
印字(3, 4, 'like this:');                           { 10 }

espruar := Random(1Ah);                              { 0..25 }
dethek  := Random(16h);                              { 0..21 }
<overlay-33 entry#3>(11h, 3, espruar,        0);     { 畫左邊那個符文 }
<overlay-33 entry#3>(11h, 7, dethek ＋ 1Ah,  0);     { 畫右邊那個符文 }

路徑 := Random(3);
圖樣 := ['-..-..-..', '- - - - -', '.........'][路徑];   { 各 9 bytes }
方框 := Random(6);
印字(3, 0Ch, 'Type the character in box number ' ＋ Str(6 − 方框));  { 33 }
印字(3, 0Dh, 'under the ');                          { 10 }
印字(0Eh, 0Dh, 圖樣);
印字(19h, 0Dh, 'path.');                             { 5 }
```

★ **26 個 espruar 符文 ＋ 22 個 dethek 符文**，字模都在 `'tiles'` 這個資源裡
（`overlay-33 entry#2` 一次載 `1Ah`／`16h` 個，第二組的索引要加 `1Ah`）。
★ 三種「路徑」圖樣、六個方框（畫面上顯示 `6 − 方框`，所以是 **1..6**）。

## ★★★ 答案公式

```pascal
索引 := (espruar ＋ 22h − dethek) ＋ 路徑 × 0Ch ＋ (5 − 方框) × 2;
while 索引 < 0    do 索引 := 索引 ＋ 24h;            { 36 }
while 索引 > 23h  do 索引 := 索引 − 24h;
答案 := byte[DS:0000h ＋ 方框 × 25h ＋ 索引 ＋ 1];   { ★ 25h ＝ 37 }
```

★ `22h` ＝ 34、`0Ch` ＝ 12、`24h` ＝ 36。**索引最後落在 `0`..`35`。**
★ `＋ 1` 是因為每一列的第 0 格是 `'$'`（佔位，不會被取到）。

## ★★★ 答案表：`DS:0000h`，6 列 × 37 bytes

```
列 0  $CWLNRTESSCEDCSHSISERRRNSHSSTSSNNHSHN
列 1  $LAASRDAIILIDSUGADAEEOEGRLSELIITESOIO
列 2  $LRUNIMMORIIGRRIUPTIIUELIMLHMIXACGRIL
列 3  $Z0LIOHEUVNODSGEOGXYWISIOCRARLRARRHOI
列 4  $AMTELRLUIYNAEOOITOUELRREREUIMADPPFAB
列 5  $ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890
```

> ★★★★ **這就是轉輪的全部資料。**
> `方框` ＝ `5` 那一列（畫面上顯示「box number 1」）正好是
> `A`..`Z` ＋ `1`..`0` 的完整字元集。
> ⇒ **remake 只要有這 6 × 37 的表與上面的公式，就能完整重現
> （或直接跳過）這道檢查，不需要實體轉輪。**

## 答錯

```pascal
輸入 := <far 0542:0722h>('type character and press return: ', 0Dh, 0, 1);
if 輸入 <> 答案 then begin
    …"Sorry, that's incorrect."（24）…
    …'An unseen force hurls you into the abyss!'（41）…
end;
```

★ 輸入提示 `'type character and press return: '`（33，結尾有空白）。

## 中文化

| DOS | 長度 | 建議中文 |
|---|---|---|
| `'Align the espruar and dethek runes'` | 34 | 「請把下方的 espruar 與 dethek 符文」 |
| `'shown below, on translation wheel'` | 33 | 「照下圖對齊到轉輪上」 |
| `'like this:'` | 10 | 「像這樣：」 |
| `'Type the character in box number '` ＋ 數字 | 33 | 「請輸入第」＋數字＋「格」 |
| `'under the '` ＋ 圖樣 ＋ `'path.'` | 10 ＋ 5 | 「在」＋圖樣＋「這條路徑下的字元」 |
| `'type character and press return: '` | 33 | 「輸入字元後按 Enter：」 |
| `"Sorry, that's incorrect."` | 24 | 「答錯了。」 |
| `'An unseen force hurls you into the abyss!'` | 41 | 「一股看不見的力量把你們拋進了深淵！」 |

⚠ `'espruar'` 與 `'dethek'` 是**專有名詞**（精靈文與矮人文的書寫系統），
中文版應保留原文或加註。
⚠ 三個圖樣（`-..-..-..` 等）是**畫面上的圖形**，不可翻。
⚠⚠ **答案表是 ASCII 大寫字母與數字，中文化不能動**——玩家輸入的也是 ASCII。

## 明確不宣稱

- 沒有宣稱 `overlay-33 entry#2`／`entry#3` 的完整介面（只知道實參）。
- 沒有宣稱答錯之後除了兩句訊息還做什麼（`03CEh` 之後跳到 `03ECh` 再到 `0446h`，
  中間 30 ＋ 90 bytes 是匯出沒有反組譯的區段）。
- 沒有宣稱 DOS 側的重試次數（`var_1` 每輪 ＋1，上限判斷落在未反組譯的區段）。

> ★★★ **PC-98 對側見 spec 1063**：`pc98 overlay-03:00131h`，
> **公式逐條相同**，答案表在 `DS:00FCh` 且與 DOS 的 `DS:0000h` **逐 byte 相同**；
> PC-98 側看得到**重試上限是 3 次**。
