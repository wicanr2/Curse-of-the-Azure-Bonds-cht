# 953 — ★ PC-98 全形字在文字 VRAM 的編碼，與 `OvrInit`

- 證據等級：`exact`（兩支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `196C9h`、`19EC4h`
- 作法見 spec 783

## ★★ `196C9h`（113 行）：把累積的一段文字寫進 VRAM

spec 952 的「沖出」就是這一支。**全形字實際怎麼放進 PC-98 的文字畫面，
答案在這裡。**

```pascal
if si = di then 離開;                            { 沒有累積的東西 }

word_280E8h := bx;   word_280EAh := dx;          { 記下起點 }
cx := di − si;                                    { 這一段的長度 }
di := (起列 × byte_280DAh ＋ 起欄) × 2;
if byte_280DAh <> 50h then di := di shl 1;        { 40 欄模式 }
bx := word_280D6h;                                { ＝ 2000h，屬性平面偏移 }
dh := byte_280C6h;                                { 目前屬性 }
<sub_18FA3h>(dh);
ds := 來源段;   es := word_280D4h;                { ＝ 0A000h }
ah := byte_280E3h;                                { 跨呼叫保留的前導旗標 }

repeat
    al := ds:[si];  inc(si);

    if ah <> 0 then begin                         { ★ 全形：ah 是前導、al 是後續 }
        屬性(es:[bx＋di]) := dh;
        屬性(es:[bx＋di＋2]) := dh;
        <sub_19791h>();                           { 兩個 byte → 內碼 }
        xchg(ah, al);
        al := al − 20h;
        stosw;                                    { 左半格 ＝ 內碼 }
        ax := ax or 8000h;
        stosw;                                    { ★ 右半格 ＝ 內碼 or 8000h }
        if byte_280DAh <> 50h then begin          { 40 欄：每格再重複一次 }
            ax := ax and 7FFFh;  stosw;
            ax := ax or  8000h;  stosw;
        end;
        ah := 0;
    end
    else if (byte_280D0h <> 0) and <sub_1977Eh>(al) then
        ah := al                                  { 這是前導，留到下一圈 }
    else begin                                    { 單位元組 }
        屬性(es:[bx＋di]) := dh;
        stosw;                                    { 字碼 }
        if byte_280DAh <> 50h then stosw;
    end;
until cx 用完;

byte_280E3h := ah;                                { ★ 把未配對的前導留給下一次 }
```

### PC-98 全形字的 VRAM 編碼規則

| | 內容 |
|---|---|
| **左半格** | `<sub_19791h>` 換算出的內碼，再 `xchg ah,al` 並 `sub al, 20h` |
| **右半格** | **同一個值 `or 8000h`** |
| 屬性平面 | 兩格都寫同一個屬性（位址 ＋ `word_280D6h` ＝ `2000h`） |

`sub al, 20h` 是把 JIS 的 `2020h` 偏置扣掉的標準換算；
**bit 15 是「這是全形字的右半」的標記**——PC-98 的 GDC 靠它把兩格連成一個字。

40 欄模式時每一格都要寫兩次（`stosw` 兩次），因為一個邏輯格佔兩個實體格
——與 spec 950 的 `欄 × 2` 一致。

**這是繁中化最關鍵的一層**：Big5 要嘛換算成 PC-98 的內碼，
要嘛整個改走圖形輸出。`byte_280E3h` 讓一段文字可以跨呼叫接續，
所以分段輸出不會把全形字拆壞。

## `19EC4h`（102 行，far，`retf 4`）：`OvrInit`

```pascal
if word_23AD4h = 0 then begin 結果 := −1;  離開 end;      { ovrError }

依序試 <sub_19F7Eh>／<sub_19F8Ah>／<sub_19FC5h> 三種方式開檔；
全部失敗 → 結果 := 0FFFEh（−2 ＝ ovrNotFound）;

讀標頭（<sub_1A038h>）;
if 標頭[0] = 5A4Dh then begin                             { ★ 'MZ' }
    { 覆疊資料附在 EXE 影像後面：算出影像大小並 LSEEK 過去 }
    位置 := 標頭[4] × 200h − ((−標頭[6]) and 1FFh);
    int 21h AX=4200h;
    重讀標頭;
end;

if (標頭[0] <> 5054h) or (標頭[2] <> 564Fh) then 關檔並回 −1;   { ★ 'TP' 'OV' }

word_23AE4h := 檔案 handle;
dword_23AD0h := 覆疊資料的起始位置;
dword_28104h := cs:loc_1A119h;                            { 覆疊讀取常式 }
if dword_23AD8h = NIL then dword_23AD8h := cs:locret_1A196h;
int 21h AX=253Fh, DS:DX = cs:02E7h;                       { ★ 掛 INT 3Fh handler }
結果 := 0;
word_23AC6h := 結果;
```

### 兩個直接的閉合

1. **`5054h` ＋ `564Fh` ＝ `'TPOV'`**——Turbo Pascal Overlay 的簽章，
   與 `~/.claude/knowledge-base/retro/borland-tpov-overlay-re.md` 一致。
   本專案的 `GAME.OVR` 就是這個格式。
2. **`int 3Fh` 的 handler 在這裡掛上**（`AX=253Fh`，指向 `cs:02E7h`）。
   spec 928 記的那 871 個 entry stub 全部以 `CD 3F` 開頭，
   **它們最終跳進來的就是這一個 handler**。

覆疊資料**可以附在 `.EXE` 後面**（`'MZ'` 分支），也可以是獨立檔
（直接就是 `'TPOV'`）。本作是獨立的 `GAME.OVR`，所以走後者。

回傳值與 spec 944 的 `OvrInitEMS` 共用同一組 `OvrResult` 常數，
一起寫進 `word_23AC6h`。

## 明確不宣稱

- 沒有宣稱 `<sub_19791h>` 的內碼換算細節（只知道之後要 `xchg` 與 `− 20h`）。
- 沒有宣稱 `<sub_1977Eh>` 判前導的範圍。
- 沒有宣稱 `<sub_19F7Eh>`／`<sub_19F8Ah>`／`<sub_19FC5h>` 三種開檔方式的差別。
- 沒有宣稱 `cs:02E7h`（INT 3Fh handler）與 `loc_1A119h`（讀取常式）的內容。
- 沒有宣稱 `dword_23AD8h` 那個預設成空常式的向量是什麼。
