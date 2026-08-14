# 954 — PC-98 的圖形矩形清除，與帶 `HeapError` 重試的配置

- 證據等級：`exact`（兩支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `141A0h`、`1AA72h`
- 作法見 spec 783

## `141A0h`（99 行，far，`retf 0Ch`）：清一個圖形矩形

`(填值 arg_0, 平面 arg_2, 下 arg_4, 右 arg_6, 上 arg_8, 左 arg_A)`，六個 byte。

```pascal
存 := byte_241E0h;

起位移 := longint(上 × 50h) shl 3 ＋ 左;         { 32 位元運算 }
行數   := ((下 − 上 ＋ 1) shl 3) − 1;
<sub_14187h>(平面);
<sub_14180h>();

for i := 0 to 行數 do begin
    <sub_1C180h>(word_241DEh : 起位移, 右 − 左 ＋ 1, 填值);
    起位移 := 起位移 ＋ 50h;
end;

byte_241E0h := 存;
<sub_14187h>(byte_241E0h);                        { 還原平面 }
```

★ **`50h` ＝ 80 bytes／掃描線**——PC-98 的圖形是 **640 像素寬、
每平面 1 bit／像素**（640 ÷ 8 ＝ 80）。
`shl 3` 表示**一個「列」仍然是 8 條掃描線**，與 spec 937／940／942／949
的縱向單位完全一致。

`byte_241E0h` 是目前選中的平面，本支**先存後還原**，
所以呼叫端不必自己保護。`word_241DEh` 是圖形 VRAM 的段。

⚠ 起位移用 32 位元算（`cbw`／`cwd` 之後 `add ax,cx` ＋ `adc dx,bx`），
但送給 `<sub_1C180h>` 的只有低位字——640×400 單平面是 32000 bytes，
**在 16 bits 之內，所以不會溢位**。

## `1AA72h`（116 行）：從 free list 配置（帶 `HeapError` 重試）

與 DOS 的 `1A986h`（spec 931）同一套骨架，但多了一層重試：

```pascal
loop:
push(要求大小);
<sub_1AC62h>();
{ 掃 dword_23AF0h 的 free list，找 (+6:+4) − (+2:+0) >= 要求 的一筆 }
{ 找到 → 從起點切走，切完剛好用光就 <sub_1AC04h> 移除該筆，clc 返回 }

{ 沒找到 → 檢查配置指標 word_23AECh/word_23AEEh 上面還夠不夠 }
if 夠 then begin 指標往上推;  clc 返回 end;

{ ★ 還是不夠 → 呼叫 HeapError }
if dword_23AF6h = NIL then begin stc;  返回 end;
r := <dword_23AF6h>(要求大小);
if r = 0 then begin stc;  返回 end;               { 0 ＝ 放棄 }
if r = 1 then begin bx := 0;  返回 end;           { 1 ＝ 回 NIL 但不報錯 }
goto loop;                                        { 其他 ＝ 已騰出空間，重試 }
```

★ `dword_23AF6h` 就是 Turbo Pascal 的 **`HeapError`** 向量，
回傳值的三種語意（0 ＝ 執行期錯誤、1 ＝ `New` 回 `nil`、其他 ＝ 重試）
是 Turbo Pascal 的標準約定。

**DOS 版的 `1A986h`（spec 931）沒有這一段**——它找不到就直接 `stc` 失敗。
PC-98 版多掛了 `HeapError`，所以記憶體不足時有機會由上層釋放快取再重來。
這與 spec 951 的覆疊 LRU 是同一個方向：**PC-98 版對記憶體壓力做了更多處理**。

## 明確不宣稱

- 沒有宣稱 `<sub_14180h>`／`<sub_14187h>`（設平面）與 `<sub_1C180h>`（填值）的內部。
- 沒有宣稱 `arg_2` 的平面編號對應哪一個顏色平面。
- 沒有宣稱 `<sub_1AC19h>`／`<sub_1AC62h>`／`<sub_1AC04h>` 的內部。
- 沒有宣稱 `HeapError` 實際被設成什麼常式。
