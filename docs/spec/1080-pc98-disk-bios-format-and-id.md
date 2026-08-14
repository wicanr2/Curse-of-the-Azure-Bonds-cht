# 1080 — PC-98 磁碟 BIOS：手工格式化 2HD 存檔片，再用 71 bytes 的簽章認出五種磁片

- 證據等級：`exact`（PC-98 側 90 條 ＋ 93 條，兩支各自逐條讀完）
- 作法見 spec 783

## 這兩支為什麼一直是「無 dump」

`scripts/annotated_dump.py` 讀的是 `overlays/prologue/` 那份匯出，
而這兩支沒有標準的 `push bp / mov bp, sp` 開場（它們自己管堆疊），
所以不在 prologue 匯出裡。**改讀 `overlays/full/` 那份就有完整指令**。
⇒ 台帳上「無 dump」不等於「反組譯不出來」，只代表工具走錯了那一份匯出。

⚠ 兩支通篇的 `int 1Bh` 被 IDA 註成 `CTRL-BREAK KEY`——**那是 IBM PC 的意思**。
PC-98 的 `int 1Bh` 是**磁碟 BIOS**。這與 spec 975 的 `out 0A6h`、
spec 946 的 `int 18h` 是同一類陷阱：**IDA 的週邊註解在 PC-98 上一律不可信**。

## `pc98 overlay-16:05588h`（`retf 2`）——格式化存檔磁片

spec 1043 記的 `sub_5588h`（「格式化／複製」）就是這一支。
資料段固定切到 `0ABF0h`（`ds` 與 `es` 都是），**整支不用區域變數，
所有狀態放在 `ds:0`／`ds:1`／`ds:2` 三個 byte**。

```asm
mov  ax, 0ABF0h ; mov ds, ax ; mov es, ax
call sub_574Fh                    ; 取磁碟機編號 → CS:576Eh
mov  al, cs:576Eh ; mov ah, 7 ; int 1Bh      ; ★ 磁頭復歸
```

### 逐軌格式化

```pascal
柱面 := 0;  磁頭 := 0;
repeat
    { 造 8 筆磁區 ID，每筆 4 bytes：C, H, R, N }
    for cl := 8 downto 1 do 寫入(柱面, 磁頭, 9 − cl, 3);   { ← ds:3 起 }
    AH := 5Dh;  BX := 20h;  CL := 柱面;  CH := 3;
    DL := 0E5h;  DH := 磁頭;  BP := 3;  int 1Bh;           { ★ 格式化一軌 }
    if 失敗 then goto 錯誤;
    磁頭 := 磁頭 xor 1;
    if 磁頭 = 0 then inc 柱面;
until 柱面 > 4Ch;
```

> ★★★ **`4Ch` ＝ 76 ⇒ 77 個柱面 × 2 個磁頭 × 8 個磁區 × 1024 bytes
> ＝ 1,261,568 bytes ——標準的 PC-98 2HD 軟碟。**
> 磁區 ID 的 `N = 3` 就是「磁區 1024 bytes」，`0E5h` 是填充位元組。
> ⇒ remake 若要產生相容的存檔片映像，就照這個幾何。

### 寫出檔案系統

```pascal
FillChar(ds:23h, 400h, 0);
資料 := 0C29h:0BA5h;  長度 := 47h;                { ★ 第一塊是 71 bytes 的簽章 }
i := 0;
repeat
    word[ds:23h] := 0;
    if (i = 1) or (i = 3) then word[ds:23h] := 0FFFEh;   { ★ FAT 的前兩個項目 }
    AH := 55h;  DL := (i and 7) ＋ 1;  DH := i shr 3;
    CX := 300h;  DX 的高位 := 1;  int 1Bh;               { 寫一個磁區 }
    if 失敗 then goto 錯誤;
    資料 := 0ABF0h:0023h;  長度 := 400h;          { ★ 之後都寫 1024 bytes 的 0 }
    inc i;
until i > 0Ah;
DOS 磁碟重置（int 21h, AH = 0Dh）;
結果 := 0;                                        { 錯誤路徑回 1 }
```

> ★★★ **這是手工鋪一份 FAT**：第 0 個磁區寫 71 bytes 的識別簽章，
> 第 1 與第 3 個磁區的開頭寫 `0FFFEh`（FAT12 的媒體描述 ＋ 結束標記），
> 其餘七個磁區清成 0（根目錄）。合計 11 個磁區。
> ⇒ 原作**沒有呼叫 DOS 的 FORMAT**，自己用磁碟 BIOS 造出可用的磁片。
> ★ 回傳 `0` ＝ 成功、`1` ＝ 失敗，與 spec 1043 記的
> 「`sub_5588h() = 1` 就顯示建立失敗」一致。

## ★★★ `pc98 overlay-16:05674h`（`retf 2`）——認出插進去的是哪一種磁片

```pascal
call sub_574Fh;  DOS 磁碟重置;
AH := 76h;  BX := 400h;  CX := 300h;  DX := 1;  BP := 23h;  int 1Bh;   { 讀一個磁區 }
if 失敗 then begin DOS 磁碟重置; exit(0) end;

for k := 0 to 4 do
    if 相同(0ABF0h:0023h, 0C29h:(0A89h ＋ k × 47h), 47h) then exit(2 ＋ k);
exit(1);
```

五個簽章各 `47h` ＝ **71 bytes**，等距排在 `0C29h:0A89h` 起：

| 簽章位址 | 回傳 |
|---|---|
| `0C29h:0A89h` | 2 |
| `0C29h:0AD0h` | 3 |
| `0C29h:0B17h` | 4 |
| `0C29h:0B5Eh` | 5 |
| **`0C29h:0BA5h`** | **6** |

> ★★★ **`0BA5h` 正是上面那支格式化程式寫進第 0 磁區的那一塊**
> ⇒ **自己格出來的存檔片會被認成 `6`**，而 spec 1070 記的
> 「`DS:0A816h = 0`（`.guy` 類別）時期望磁片種類 6」就此對上。
> ⇒ 其餘 2..5 是四張遊戲資料片。
>
> ★★ 回傳值只到 6；spec 975 的 `overlay-16:007DCh` 拿到結果之後，
> **另外用 `FindFirst('*.hil')` 成功時把它改成 7**——
> 所以「7」不是這一支產生的。⇒ spec 1070 的兩個期望值（6 與 7）
> 分別來自兩個不同的地方。
>
> ★ 讀不到（換片中／空磁碟機）回 `0`、讀得到但簽章不合回 `1`。

## 中文化

兩支都沒有字串。⚠ 那五個 71 bytes 的簽章是**磁片識別資料，不可更動**——
改了會讓原版磁片認不出來。

## 明確不宣稱

- 沒有宣稱 `sub_574Fh` 做什麼（只知道它會讓 `CS:576Eh` 有磁碟機編號）。
- 沒有宣稱兩支的 `retf 2` 那個 word 參數是什麼
  （兩支都不用 `bp` 框架，參數由 `sub_574Fh` 取用）。
- 沒有宣稱 `int 1Bh` 各功能碼（`07h`／`55h`／`5Dh`／`76h`）的完整暫存器約定
  ——上面只照組語列出實際填了哪些暫存器。
- 沒有宣稱那 71 bytes 的簽章內容（沒有把 `0C29h` 段的靜態資料抓出來對照）。
- 沒有宣稱 11 個磁區之後的目錄／FAT 版面細節。
- 沒有宣稱 DOS 側有沒有對應的磁片格式化路徑。
