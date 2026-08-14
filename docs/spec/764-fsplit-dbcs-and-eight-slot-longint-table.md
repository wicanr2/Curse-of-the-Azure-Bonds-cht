# 764 — PC-98 把 `FSplit` 改成能認 Shift-JIS，以及 `DS:6F70h` 八格的第三種用法

- 證據等級：`exact`（以補洞後的匯出逐條讀完，見 spec 761）

## `PC98-GAME.EXE:18E59h` — DBCS 版的 `FSplit`

`retf 10h`（8 個 word ＝ 4 個遠指標）。三個輸出的長度上限分別是 `43h`（67）、
`8`、`4`——正好是 Turbo Pascal `DOS` 單元的 `DirStr`／`NameStr`／`ExtStr`，
所以這支就是 **`FSplit(Path, Dir, Name, Ext)`**。

```
lds si, 路徑字串                       { 第一個宣告的參數 }
長度 := [si++]
bx := 0；最後分隔位置 := 0
while bx <> 長度 do begin
    al := [si + bx];  inc(bx)
    call sub_18E44h                    { 回 CF }
    if CF then inc(bx)                 { 前導位元組 → 跳過後續位元組 }
    else if (al = '\') or (al = ':') then 最後分隔位置 := bx
end
取出 [0, 最後分隔位置) → Dir   （上限 43h）
從該處往後找第一個 '.' → Name  （上限 8）
其餘 → Ext                      （上限 4）
```

關鍵在 `sub_18E44h`：**每個位元組都先問它是不是 Shift-JIS 的前導位元組，是的話
連下一個位元組一起跳過**。原版 Turbo Pascal 的 `FSplit` 沒有這一步，會把
`0x5C`（`\`）當成路徑分隔字元——而 `0x5C` 正是許多 Shift-JIS 雙位元組字的
**後續位元組**。PC-98 版把 RTL 補成 DBCS-aware，就是為了避開這個經典的
「日文檔名被切壞」問題。

**對中文化的意義**：Big5 的後續位元組同樣涵蓋 `0x5C`，所以這個防護在中文版
一樣需要。`sub_18E44h` 的判定範圍是針對 Shift-JIS 寫的，換成 Big5 要重寫。

## `overlay-21:0A03h`（DOS）／`0A1Dh`（PC-98）— 扣一筆並累計

兩平台 50 條指令的助憶碼完全相同，只差 DS 位址。`retf 8`：

```pascal
procedure 扣(idx: byte; 數量: word; p: 遠指標);
begin
    p^[0FBh + idx * 2] := p^[0FBh + idx * 2] − 數量;   { word 陣列 }
    本模組 003Bh(數量, p);
    if (DS:4FBAh = 6) or (DS:4FBAh = 1) then
        四位元組表[DS:6F70h + idx * 4] :=
            四位元組表[DS:6F70h + idx * 4] + 數量;      { 32-bit 加法 }
end;
```

**減法沒有下限檢查**：`p^[0FBh + idx*2]` 不足時會迴繞成很大的無號值。

`DS:6F70h`（PC-98 `DS:0A00Ah`）這塊八格 4-byte 區域到這裡已經出現三種用法：

| 出處 | 用法 |
|---|---|
| spec 756 `overlay-21:0F5Bh` | 當**遠指標**檢查前 7 格是否有非 NIL、第 8 格單獨檢查 |
| spec 757 `overlay-21:019Eh` | 清前 5 格，再把某值的 `div 5` 寫進第 4 格、`mod 5` 寫進第 3 格 |
| 本輪 `overlay-21:0A03h` | 依索引當 **32-bit 累加器**加上數量 |

三種用法互相衝突（`0F5Bh` 的「非 NIL」判斷會被累加出來的數值影響），但它們都
在程式碼裡。remake 不能挑一種當「正確的型別」——要照抄。

`DS:4FBAh`（PC-98 `DS:7F27h`）就是既有的狀態變數：戰鬥開場設 5（spec 750）、
PC-98 `overlay-17:0115h` 設 4（spec 758）。這裡只有 **1 或 6** 才累加。

## `overlay-16:09EBh`（DOS）— 兩筆記錄的合併

`retf 8`，兩個遠指標（`a` 是目的、`b` 是來源）：

```pascal
n := b^[11h] − 1;                       { word }
for i := 0 to n do begin
    a^[17h + i]     := a^[17h + i]     or  b^[17h + i];
    a^[13h]^[i]     := a^[13h]^[i]     and b^[13h]^[i];
end;
```

同一個迴圈裡**一個欄位取聯集、另一個取交集**：`+17h` 起是內嵌的 byte 陣列，
`+13h` 是指向另一塊陣列的遠指標。長度來自來源的 `+11h`。

`b^[11h] = 0` 時 `n` 會是 `0FFFFh`，但迴圈的進入條件是
`0 > n`（無號）為假才進——`n = 0FFFFh` 時會跑 65536 次。**這是一個真實的
邊界問題**，不過只有在來源的計數為 0 時才會發生。

## 明確不宣稱

- 沒有宣稱 `p^[0FBh]` 那個 word 陣列與 `DS:6F70h` 八格各自代表什麼。
- 沒有宣稱 `DS:4FBAh` 的 1 與 6 是哪兩個狀態。
- 沒有宣稱 `overlay-16` 合併的兩筆記錄是什麼。
- 沒有宣稱 `sub_18E44h` 的判定範圍就是標準 Shift-JIS 前導位元組區間——本輪
  只讀到「它回 CF、為真就跳一個位元組」。
- `dos overlay-04:0047h` 在補洞後仍有 5 bytes 的缺口，本輪**沒有**判讀。
