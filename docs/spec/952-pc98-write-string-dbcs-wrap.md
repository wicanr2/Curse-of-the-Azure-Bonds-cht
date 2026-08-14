# 952 — PC-98 的字串輸出：全形字不跨行；以及 free list 的釋放

- 證據等級：`exact`（兩支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `19616h`、`1AB2Dh`
- 作法見 spec 783

## ★ `19616h`（104 行）：輸出一段字串

`es:di` ＝ 來源、`cx` ＝ 長度、`dl` ＝ 目前欄、`word_280C8h` ＝ 左界、
`word_280CAh` 的低位 ＝ 右界（spec 944 設的）。

```pascal
<sub_195FCh>();
byte_280ECh := byte_280E3h;                     { 前導旗標的初值 }
起 := di;   起欄 := dl;

repeat
    c := es:[di];

    if byte_280ECh <> 0 then begin              { 這個是全形的後半 }
        byte_280ECh := 0;
        inc(di);   dl := dl ＋ 2;               { ★ 全形佔兩欄 }
        if dl > 右界 then begin                  { 換行 }
            沖出();  換行();  dl := 左界;
        end;
    end
    else if (byte_280D0h <> 0) and <sub_1977Eh>(c) then begin
        byte_280ECh := c;                       { 記下前導 }
        inc(di);
        if dl >= 右界 then begin                 { ★ 放不下整個全形字 }
            沖出();  換行();  dl := 左界;        { → 整個字移到下一行 }
        end;
    end
    else case c of
        07h: begin 沖出();  <sub_19A07h>() end;             { BEL }
        08h: begin 沖出();  if dl <> 左界 then begin dec(dl); <sub_19B00h>() end end;
        0Ah: begin 沖出();  換行() end;                      { LF }
        0Dh: begin 沖出();  dl := 左界 end;                  { CR }
    end;
    inc(di);
    起 := di;   起欄 := dl;
    dec(cx);
until cx = 0;
沖出();  <sub_19604h>();
```

（「沖出」＝ `<sub_196C9h>`，把 `起`..`di` 之間累積的一段實際寫進 VRAM；
「換行」＝ `<sub_195B6h>`。）

### 兩件對中文化直接有用的事

1. **全形字佔兩欄**（`dl` 一次加 2）。
2. ★ **遇到前導位元組時先檢查 `dl >= 右界`**——放不下就**整個字移到下一行**，
   不會把一個全形字拆在兩行。

spec 945（退格）、spec 948（字元寫入）、本支（字串輸出）三支合起來，
PC-98 版的全形處理是完整的：**輸入、編輯、輸出三個環節都認得雙位元組**。
DOS 版三個環節都沒有。

`byte_280ECh` 是**跨字元保留的前導旗標**（初值來自 `byte_280E3h`），
所以一段字串可以從「上一次結尾是前導」的狀態接續下去。

## `1AB2Dh`（110 行）：把一段還回 free list

與 DOS 的 `1AA22h`（spec 931）**結構逐條對應**：

- `ax = 0` 直接返回。
- 用 `<sub_1AC62h>` 取得目前配置指標，做 `seg:para` 的進位正規化
  （`cmp ax, 10h` / `sub ax, 10h` / `inc dx`）。
- 確認要還的區間落在 `[word_23AE8h:word_23AEAh, word_23AECh:word_23AEEh]`
  之內，否則 `stc` 失敗返回。
- 掃 `dword_23AF0h` 的 free list（每筆 8 bytes：`+0`／`+2` 起、`+4`／`+6` 迄），
  與相鄰項合併，被吃掉的用 `<sub_1AC04h>` 移除。
- 還的正好是配置指標下緣就把指標往回退；否則用 `<sub_1ABE0h>` 取一筆新位置寫入。

**全域位址不同、邏輯相同**：

| 角色 | DOS（spec 931） | PC-98 |
|---|---|---|
| free list 遠指標 | `dword_2097Eh` | `dword_23AF0h` |
| 配置指標（para／段） | `word_2097Ah`／`word_2097Ch` | `word_23AECh`／`word_23AEEh` |
| 下界 | `word_20976h`／`word_20978h` | `word_23AE8h`／`word_23AEAh` |

## 明確不宣稱

- 沒有宣稱 `<sub_195FCh>`／`<sub_19604h>`／`<sub_196C9h>`／`<sub_195B6h>`／
  `<sub_19A07h>`／`<sub_19B00h>` 的內部。
- 沒有宣稱 `byte_280E3h` 由誰設定（spec 945 的退格也讀它）。
- 沒有宣稱 `<sub_1AC62h>`／`<sub_1AC04h>`／`<sub_1ABE0h>` 的內部。
