# 1068 — 角色檔清單：用檔案長度過濾，名字直接從檔案裡 Seek 出來

- 證據等級：`exact`（DOS 側 382 條逐條讀完）
- 作法見 spec 783

## `dos overlay-16:0008Dh`（`retf`）

原本待解讀。這是 spec 1066（加入角色）與 spec 1057（存檔槽）清單的來源。

```pascal
function 掃檔(樣式:     字串;   { bp+10h，上限 80 }
              名字位移: word;   { bp+0Eh }
              旗標位移: word;   { bp+0Ch }
              期望長度: word;   { bp+0Ah }
              …兩個輸出鏈…)
```

## 流程

```pascal
<far 097F:0112h>(樣式, 0, @搜尋記錄);              { FindFirst }
if DS:8C98h <> 0 then exit;                        { 找不到就走 }
repeat
    完整 := 路徑 ＋ 檔名;
    Assign(f, 完整);  Reset(f, 1);
    if FileSize(f) <> 期望長度 then continue;      { ★★ 用長度過濾 }
    節點A := GetMem(2Eh);  節點B := GetMem(2Eh);   { ★ 各 46 bytes }
    FillChar(節點A^, 2Eh, 0);  FillChar(節點B^, 2Eh, 0);
    Seek(f, 名字位移);
    if DS:7584h = 2 then …Hillsfar 另一條路（先讀 10h bytes 再轉）…
    else            BlockRead(f, 節點A^, 10h);     { ★ 16 bytes 的名字 }
    Seek(f, 旗標位移);
    if DS:7584h = 2 then 旗標 := 0
    else            BlockRead(f, 旗標, 1);
    …
    if Copy(檔名, 9, 4) = '.SAV' then              { ★★ 副檔名 }
        後綴 := 'from saved game ' ＋ 磁碟機字母
    else
        後綴 := '';
    顯示名 := 名字 ＋ 空白補到 15 欄 ＋ 後綴;
    …串進兩條鏈…
until 沒有下一個;
```

> ★★ **清單只收「長度剛好等於 `期望長度`」的檔**
> ——這正是 spec 1043 那張長度表（`423`／`285`／`188`）的用途：
> 掃 `*.guy` 時只收 423 bytes 的、掃 `*.cha` 只收 285、`*.sav` 只收 188。
> ⇒ **檔案長度就是格式驗證**，遊戲不看檔頭。
>
> ★★ 另外兩個參數（名字位移、旗標位移）形狀上對應 spec 1043 的
> 另外兩張平行表（`DS:0BFCh` 的 `00 00 04 F7`、`DS:0BFFh` 的 `247／132／19／423`）
> ——⚠ **本規格不宣稱哪一張對哪一個參數**（那是 PC-98 側的另一支函式）。

## ★★ `'.SAV'` 會多一句 `'from saved game '`

★ 副檔名用 `Copy(檔名, 9, 4)` 取——**檔名固定 8.3，第 9 個字元起是 `'.'`**。
★ `'.SAV'` ＝ Hillsfar 的存檔（spec 1066），所以清單上會標
`'from saved game X'`（X ＝ 磁碟機字母）讓玩家分辨。

★ 名字欄補到 **15 欄**（`Copy(空白, 1, 0Fh − 長度)`）
——⚠ 中文全形只有 **7 個字**的空間。

## ★★ `DS:7584h = 2`（Hillsfar）走不同的讀法

Hillsfar 的檔案要先讀進一個暫存區再經 `sub_20h` 轉換，
而且第二個位移**不讀檔、直接給 0**。
⇒ **Hillsfar 的角色檔沒有那個旗標欄位。**

## 中文化

| DOS | 長度 | 備註 |
|---|---|---|
| `'.SAV'` | 4 | ⚠ **副檔名，不可翻** |
| `'from saved game '` | 16 | 「（來自存檔）」 |

⚠ 名字欄固定補到 **15 欄**，`'from saved game X'` 接在後面
——中文化要重算欄寬，否則兩欄會黏在一起。

## 明確不宣稱

- 沒有宣稱 `far 097F:0112h` 的第二個參數 `0`／`10h` 差在哪
  （spec 1067 用的是 `10h`）。
- 沒有宣稱 `sub_20h`（Hillsfar 名字的轉換）做什麼。
- 沒有宣稱兩條輸出鏈（各用 46 bytes 節點）的分工。
- 沒有宣稱名字位移／旗標位移各自對到 spec 1043 的哪一張表。
- 沒有宣稱那個 1 byte 的旗標是什麼。
