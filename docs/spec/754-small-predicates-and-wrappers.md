# 754 — 一批小判斷式、包裝函式與 `DS:4FCAh` 命令區塊

- 證據等級：`exact`（逐條讀完；兩平台成對者已互相比對）
- 讀法沿用 [`700`](700-turbo-pascal-parameter-order.md)：**第一個宣告的參數在最高
  位址**，`retf N` 是參數位元組數的唯一依據。

## 範圍判斷（兩對）

`overlay-30:0556h`（DOS）／`04ADh`（PC-98），`retf 4`：

```pascal
function 在格內(a, b: integer): boolean;
begin 在格內 := (a >= 0) and (a <= 15) and (b >= 0) and (b <= 15) end;
```

比較是有號的，界限含端點，範圍是 **16 × 16**。

`overlay-10:11FBh`（DOS）／`11EFh`（PC-98），`retf 4`：

```pascal
function f(a, b: integer): boolean;
begin
    if (a >= 0) and (a < 11) then f := false
    else if (b >= 0) and (b < 6) then f := false
    else f := true;
end;
```

注意這一支**不是**「兩個都在範圍內」——是「兩個都不在範圍內才回 true」。兩個
範圍不同（`a` 是 `0..10`，`b` 是 `0..5`），也不是同一個維度。沒有宣稱 11 與 6
代表什麼。

## `overlay-25:13BCh`（DOS）／`14C4h`（PC-98）— 與門檻比大小

`retf 4`，參數是一個遠指標：

```pascal
function f(p: 遠指標): boolean;
begin f := 本模組的 sub_134Ah(p) > p^[0E6h] end;
```

比較是有號的（`jg`）。被呼叫的 `sub_134Ah` 以 `push cs` ＋ `call near` 呼叫，
結果在 `AL`。

## `overlay-29:00B3h`（兩平台位元組相同）— 加 5 的包裝

`retf 8`：

```pascal
procedure f(a, b: byte; p: 遠指標);
begin 本模組的 sub_Ch(a, b + 5, 0, p) end;
```

`a`、`b` 都是零延伸成 word 才推入；被包的程序收 5 個 word。

## `overlay-13:0F12h`（DOS）／`0F77h`（PC-98）— 依相位進位的折半

`retf 2`：

```pascal
function f(x: byte): byte;
begin f := (x + (DS:758Dh and 1)) div 2 end;
```

`shr al, 1` 只是拿 `DS:758Dh` 的 bit 0 到 CF，之後那個值就丟掉了。`DS:758Dh`
在戰鬥開場（spec 750）被設成 0，所以它是個會翻轉的相位旗標：相位為奇數時
折半改成進位。PC-98 對應位址是 `DS:0A81Fh`。

## `overlay-20:0385h`（兩平台）— 上限 99 的累加

`retf`，不收參數：

```pascal
本模組的 sub_2D8h(@DS:7566h);
if DS:7570h > 0 then begin
    inc(DS:756Eh, DS:3E32h * DS:7570h);
    DS:7570h := 0;
    if DS:756Eh > 99 then DS:756Eh := 99;
end;
```

全部是 word、無號比較。PC-98 對應 `0A652h` / `0A65Ch` / `680Ch` / `0A65Ah`。
99 這個上限值得注意：中文化若牽涉到顯示，這是個百分比類的量。

## `overlay-28:0016h` 與 `004Fh`（DOS）— 兩張三項查表

兩支都是 `retf 4`，只用 `[bp+8]`（第一個宣告的參數）當索引，**第二個參數整支
沒讀**：

| 輸入 | `0016h` 回傳 | `004Fh` 回傳 |
|---|---|---|
| `08h` | `00h` | `08h` |
| `41h` | `00h` | `07h` |
| `0E8h` | `0Fh` | `07h` |
| 其他 | `0Fh` | `07h` |

兩支的最後一個 `cmp` 都是多餘的——`0E8h` 的分支與 default 給同一個值。原始碼
多半是 `case` 有 `else`，而 `0E8h` 那一支被改成與 `else` 相同後沒有刪掉。

## `START.EXE:1586Eh`（DOS）— 跳過開頭空白

`retf 2`，參數是一個 **SS 相對**的位址：

```pascal
while (ss:[p-103h] < ss:[p-101h]) and (ss:[p + ss:[p-103h] - 100h] = ' ') do
    inc(ss:[p-103h]);
```

形狀是一個放在堆疊上的記錄：`-103h` 是游標、`-101h` 是長度、`-100h` 起是內容。
比較是無號的。

## `DS:4FCAh` 是一個 16-byte 命令區塊

`overlay-18:0000h`（`retf 6`）與 `0032h`（`retf 4`）是同一個機制的兩個包裝：

```pascal
procedure 送出命令_0C(x, y: word; c: byte);
begin
    DS:4FCBh := 0Ch;      { 命令碼 }
    DS:4FCAh := c;
    DS:4FCDh := 0;
    DS:4FCEh := x;
    DS:4FD0h := y;
    097Fh:000Bh(@DS:4FCAh, 10h);
end;

function 送出命令_0D(x, y: word): byte;
begin
    DS:4FCBh := 0Dh;
    DS:4FCDh := 0;
    DS:4FCEh := x;  DS:4FD0h := y;
    097Fh:000Bh(@DS:4FCAh, 10h);
    送出命令_0D := DS:4FCAh;       { 同一個 byte，這次當回傳值讀 }
end;
```

`10h` 是區塊長度（16 bytes，`4FCAh`..`4FD9h`）。`+0`（`4FCAh`）在 `0Ch` 是輸入、
在 `0Dh` 是輸出，`+1` 是命令碼。掃過全 DOS overlay，**只有這兩支寫
`DS:4FCBh`，也只有這兩支呼叫 `097Fh:000Bh`**——`overlay-18` 是這個介面的唯一
包裝層。

## PC-98 `overlay-18` 的三段機器相關程式

`0000h`（`retf`，不收參數）：`02A8h:1392h(9, 0Fh)`；`08EEh:0379h(1)`；
`02A8h:1392h(9, 9)`。

`1951h`（**近呼叫** `retn 4`）：把一個 Pascal 短字串就地改成 NUL 結尾——
`cx := 長度; si := 位址+1; rep movsb` 把內容往前搬一格蓋掉長度 byte，再補
`[di] := 0`；之後 `bx := 段`、`cx := 位移`、`dl := 另一個參數`，呼叫
`sub_17F3h`。這是「Pascal 字串轉 C 字串」的典型寫法，**中文化時要注意內容會被
就地覆寫，原字串在呼叫後已經不完整**。

同一段匯出裡還接著兩支獨立的小程序（各自 `retn`）：
`1979h` 是 `INT 18h`／`AH=42h`／`CX=0C000h`，`1985h` 是 `INT 18h`／`AH=42h`／
`CX=8000h` 再 `out 68h, al`（`al = 8`）。IDA 的註解寫「TRANSFER TO ROM BASIC」
是 IBM PC 的 `INT 18h` 語意，**在 PC-98 上不適用**——PC-98 的 `INT 18h AH=42h`
是畫面顯示控制，`68h` 埠是 CRT 模式暫存器。這兩支是切換繪圖畫面模式。

## 明確不宣稱

- 沒有宣稱 `097Fh:000Bh`、`02A8h:1392h`、`08EEh:0379h` 是什麼。
- 沒有宣稱命令碼 `0Ch` / `0Dh` 的意義；只知道兩者共用同一個區塊，一個寫入一個
  讀出。
- 沒有宣稱 `overlay-28` 那兩張表的輸入 `08h` / `41h` / `0E8h` 是什麼編碼。
- `DS:3E32h`（PC-98`680Ch`）只知道是被乘的 word，沒有追它從哪裡來。
- PC-98 `1979h` / `1985h` 兩支的實際畫面效果沒有實機驗證，只依 `INT 18h` 與
  `68h` 埠的機器規格判讀。
