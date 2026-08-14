# 755 — 角色的 84 格陣列、聲音驅動安裝／卸載、Text I/O 派工

- 證據等級：`exact`（逐條讀完；成對者兩平台互相比對，可疑處回查原始 bytes）

## 角色記錄 `+1Eh` 起是 84 格的陣列

三支互相印證，範圍完全一致：

`overlay-15:0188h`（兩平台位元組相同，`retf 4`）：

```pascal
procedure 清掉超界值(p: 遠指標);
begin
    for i := 0 to 53h do
        if p^[1Eh + i] > 7Fh then p^[1Eh + i] := 0;   { 無號比較 }
    p^[72h] := 0;
end;
```

`1Eh + 53h = 71h`，緊接著就是 `+72h`——陣列 84 格、後面一個 byte，**位址上剛好
相接**。

`overlay-14:06B1h`（DOS）／`067Dh`（PC-98），`retf 6`：

```pascal
function 找格子(v: byte; p: 遠指標): byte;
begin
    i := 0;
    while (i <= 53h) and (p^[1Eh + i] <> v) do inc(i);
    if i <= 53h then 找格子 := i else 找格子 := 0FFh;
end;
```

同一個 `1Eh`／`53h`，找不到回 `0FFh`。

`overlay-15:023Eh`（兩平台，`retf`）：從隊伍鏈頭 `DS:650Ah`（PC-98 `DS:9598h`）
沿 `+189h`（PC-98 `+18Ah`）走整條鏈，對每個成員叫本模組的 `0188h` 與 `01CBh`。
所以這個清理是**整隊逐一**做的。

> 讀 `0188h` 的匯出時，DOS 版的 `mov byte ptr [di+72h], 0` 看起來少了 `es:`
> 前綴，像是兩平台的真實差異。回查原始 bytes：兩邊都是
> `26 C6 45 72 00`，`26`（ES 前綴）在 DOS 匯出裡被吃掉了。**匯出上的平台差異
> 一律要回原始 bytes 確認**。

## `overlay-08:1314h`（DOS）／`1375h`（PC-98）— 清掉同陣營的目標

`retf 4`：

```pascal
procedure f(p: 角色指標);
begin
    p^[198h] := 1;
    q := p^[18Dh];                       { 戰鬥狀態 }
    if q^[0Ah] <> NIL then                { 目前目標 }
        if q^[0Ah]^[197h] = p^[197h] then { 同陣營 }
            q^[0Ah] := NIL;
end;
```

`+197h` 是陣營欄位（既有結論：只會是 0 或 1），`+18Dh^[0Ah]` 是目前鎖定的目標
——與 spec 719 那支「選目標」寫進同一個欄位一致。PC-98 三個位移都 ＋1
（`199h` / `18Eh` / `198h`），符合既有的 PC-98 角色記錄位移。

## `overlay-26:0074h`（DOS）／`0073h`（PC-98）— 數鏈長

`retf 4`，走遠指標鏈數節點數，回傳 word。**next 欄位兩平台不同**：DOS 在
`+2Ah`，PC-98 在 `+52h`。DOS 的 `+2Ah` 正是物品節點的 far next，所以這支在數的
是物品鏈。PC-98 的物品節點顯然重新排過，引用位移前必須各自確認。

## `overlay-21:0FAFh`（DOS）— 1d20 的兩段分支

`retf`，不收參數：

```pascal
r := ROLLDICE(1, 20);          { overlay-23 entry#9 }
if (r >= 1) and (r <= 14) then f := 1
else if (r >= 15) and (r <= 20) then f := 2;
{ 兩個範圍都不中時 f 沒有被指派 }
```

1d20 的值域是 1..20，兩段合起來剛好覆蓋，所以未指派的路徑走不到。**但 local
確實沒有初始化**——這是本專案第四個同形狀的例子（「一串 `if 範圍 then 指派`、
沒有 `else`、local 從不初始化」）。remake 若把骰子換成別的值域，這裡會回傳堆疊
殘值。70 / 30 的分法（14 : 6）值得記下來。

## `overlay-28` 的第三張查表（DOS `0088h`）

`retf 4`，一樣只用第一個宣告的參數：`0Bh`→`8`、`9`→`6`、`0DBh`→`6`、`6`→`1`、
其他→`6`。與 spec 754 的 `0016h` / `004Fh` 同一族。

## PC-98 `overlay-22:41F0h` 與 `4687h`—— 同一段的兩份參數化實例

```pascal
if DS:0A520h <> 0 then begin
    r := ROLLDICE(8, 2);                        { 4687h 是 ROLLDICE(8, 3) }
    if <overlay-23 entry#22 @2419h>(0, r + 1)   { 4687h 是 r + 3 }
        then <overlay-24 entry#29 @2776h>(遠指標@DS:0A521h);
end;
```

`DS:0A521h`／`0A523h` 是一個 4-byte 遠指標，`0A520h` 是它的有效旗標。

## PC-98 的三段機器相關程式

### `overlay-11:07CDh` — 16 色調色盤

`retn 2`（近呼叫）：

```
out 6Ah, 1                      { 切 16 色模式 }
si := 076Dh  或  079Dh          { 參數非 0 用前者，為 0 用後者 }
重複 16 次：
    out 0A8h, 索引
    out 0AAh, [si]   out 0ACh, [si+1]   out 0AEh, [si+2]
int 18h / AH=42h / CH=80h
out 68h, 8
int 18h / AH=40h
int 18h / AH=12h
```

`0A8h`..`0AEh` 是 PC-98 的類比調色盤埠（索引、綠、紅、藍）。因此
**`CS:076Dh` 與 `CS:079Dh` 各是一份 16 色 × 3 byte ＝ 48 bytes 的調色盤**，
remake 要重現配色就從這兩處取。索引是遞減寫入（`bx` 從 `0Fh` 往下）。

IDA 對 `int 18h` 註「TRANSFER TO ROM BASIC」是 IBM PC 的語意，**PC-98 不適用**
（`AH=42h` 顯示控制、`AH=40h` 顯示開、`AH=12h` 游標關閉一類）。

### `overlay-26:081Eh` — 清掉最下面一列文字

```
0A65h:1B30h(A000h:0F00h, 0A0h, 0);
0A65h:1B30h(A200h:0F00h, 0A0h, 0);
```

參數形狀是 `(遠指標, word 個數, byte 值)`，與 `FillChar` 完全一致；`A000h` 是
PC-98 文字 VRAM、`A200h` 是屬性平面，`0F00h` ＋ `0A0h`（160 ＝ 80 欄 × 2 bytes）
正好是最後一列。據此把 PC-98 resident 的 `0A65h:1B30h` 認成 `FillChar`
（DOS 對應 `0A54h:1AE0h`）——**是由參數形狀 ＋ 效果推得，不是位址換算**。

順帶記一個反例：`0A54h:0634h` ↔ `0A65h:062Fh`、`0A54h:064Eh` ↔ `0A65h:0649h`
都差 `-5`，但 `1AE0h` ↔ `1B30h` 差 `+50h`。**兩平台 resident 之間沒有固定位移**，
不能用位址換算轉移識別。

### `overlay-26:0858h` — 十筆固定長度欄位初始化

`for i := 1 to 10`：把 `CS:0851h` 的字串（6 bytes）複製到 `DS:0A338h + i * 7`，
再把 `DS:0A334h + i` 設成空白。兩個平行陣列：7 bytes 一筆的記錄，外加一個
byte 的旗標陣列。

## DOS `START.EXE`：聲音驅動與 Text I/O

### `19320h` — 安裝

存下 `INT 08h` 舊向量到 `CS:dword_1931Ch`，換成 `seg045:225Eh`；接著
`out 43h, 0B6h`、把 `13B1h` 分兩次寫進 **port 40h**。

`0B6h` 這個控制字選的是 **counter 2**（喇叭），但資料寫到 **counter 0**（系統
計時器）。counter 0 沿用先前的 RW 設定（BIOS 設的 LSB/MSB）所以兩個 byte 還是
寫得進去，功能上會動——但控制字與資料埠不一致。`13B1h` ＝ 5041，
1193182 / 5041 ≈ **236.7 Hz** 的中斷率。

### `19355h` — 卸載

還原 `INT 08h`；counter 0 寫回 `0FFFFh`（回到 18.2 Hz）；`in 61h` / `and 0FCh` /
`out 61h` 關掉喇叭閘；最後對 **port 0C0h** 依序寫 `9Fh`、`0BFh`、`0DFh`、`0FFh`
——這是 **SN76489（Tandy 三音）的四個聲道各自設成最大衰減**，也就是全部靜音。
所以 DOS 版的音效同時支援 PC 喇叭與 Tandy 音源，卸載時兩邊都要收。

### `17206h` — 依裝置選擇決定要不要安裝

```
dword_21154 := seg045:0124h;
byte_21DAB  := byte_21DAA;      { 裝置選擇 }
byte_21DAC  := 1;
if byte_21DAA <> 2 then begin
    19320h();                   { 上面那支安裝 }
    196E0h(0);  1974Ah(0);
end;
```

`byte_21DAA` 是音效裝置代碼，**值 2 會整段跳過 PIT ＋ Tandy 的安裝**。

### `19BF5h` — Ctrl-Break

`byte_24E6C` 非 0 才動作：清旗標、用 `INT 16h` 把鍵盤緩衝區抽乾、印出 `^C`、
換行，然後 `INT 23h`。屬 RTL 的中斷處理，`retf 2`。

### `1BB59h` — Text 檔案的 I/O 派工

用 Turbo Pascal `TextRec` 的欄位配置可以直接對上：

| 位移 | `TextRec` 欄位 | 這支怎麼用 |
|---|---|---|
| `+04h` | `BufSize` | 回傳到 `DX` |
| `+08h` | `BufPos` | 進來先寫成 `BX` |
| `+0Ah` | `BufEnd` | 回傳到 `AX` |
| `+0Ch` | `BufPtr` | 回傳到 `ES:DI` |
| `+14h` | `InOutFunc` | `call dword ptr es:[di+14h]` |

呼叫結果非 0 就存進 `word_20996`（`IOResult` 的位置）。這是把「緩衝區用完了，
去叫裝置驅動」那一步獨立出來的內部程序。

### `1C08Ah` — 掃 PSP 命令列

從 `ES:0080h`（PSP 的命令列長度 ＋ 內容）開始，跳過空白、量一段非空白的長度，
`DX` 每數完一段減一；`BX` 累計段數，`SI`／`DI` 留下該段的起訖。這是
`ParamCount` / `ParamStr` 共用的掃描核心。空白判準是 `<= 20h`。

### `1AB0Eh` — 指標正規化 ＋ 下界夾制

與 spec 753 的 `1AAD5h` 同一族：把 `dword_2097Eh` 減去 `word_20982h`，低 4 bits
留在 `SI`、其餘右移 4 位加到段值，再與 `word_2097Ch:word_2097Ah` 這個下界比。
低於下界就整個夾到下界。`offset = 0` 有一條特別路徑（段值 `+1000h`）。

## `overlay-00:0017h`（DOS）／`001Bh`（PC-98）— 有一個沒被指派的回傳值

`retf 2`：

```pascal
0A54h:0634h(0Eh, 0, @本地緩衝區, CS:0000h);   { PC-98：0A65h:062Fh }
0542h:0946h();                                  { PC-98：0418h:12EFh }
<overlay-16 entry#11>();
if DS:5CF1h <> 0 then 06EAh:0000h();            { PC-98：07E3h:0077h }
結果 := [bp-2];
```

`sub sp, 1Ah` 配置 26 bytes 的 local，緩衝區從 `bp-19h`(＝`bp-25`) 起，回傳值取
`bp-2`。**這支函式本身沒有任何指令寫過 `bp-2`**。它是不是垃圾值，取決於
`0A54h:0634h` 實際往緩衝區寫了多少——若真的寫滿 26 bytes 就會蓋到 `bp-2`，
若只寫 `0Eh+1` bytes 就不會。**這一點沒有結論，實作前要實機確認。**

## 明確不宣稱

- 沒有宣稱 `0542h:0946h`、`06EAh:0000h`、`0418h:12EFh`、`07E3h:0077h`、
  `overlay-23 entry#22`、`overlay-24 entry#29` 是什麼。
- 沒有宣稱角色 `+1Eh` 那 84 格裝的是什麼。三支函式一致指出範圍與哨兵值
  （`> 7Fh` 視為壞值、`0FFh` 代表找不到），內容意義未定。
- 沒有宣稱 `byte_21DAA = 2` 對應哪個音效裝置。
- PC-98 那三段機器相關程式沒有實機驗證，只依 PC-98 硬體規格判讀。
