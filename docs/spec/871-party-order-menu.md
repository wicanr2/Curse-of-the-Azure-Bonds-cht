# 871 — 隊伍順序選單：Select／Place 兩頁，PC-98 熱鍵再一次獨立成 ASCII

- 證據等級：`exact`（兩平台各自逐條讀完）
- 作法見 spec 783

## `overlay-15:01834h`（DOS 103 條）↔ `overlay-15:017F6h`（PC-98 至 `retf` 為止）

`retf` ＝ 無參數。

| | DOS | PC-98 |
|---|---|---|
| 標題 | `'Party Order: '` | `'並び順：'` |
| 選定 | `'has been selected'` | `'を選んだ。'` |
| 選單（頁 0） | `'Select Exit'` | `'   選択                       抜ける    '` |
| 選單（頁 1） | `'Place Exit'` | `'   配置                       抜ける    '` |

```pascal
頁 := 0;  鍵 := 20h;
while <0A54h:08D4h>(離開鍵表, 鍵) do begin          { 以 ZF 回報 }
    暫 := 'Party Order: ';
    鍵 := <far 169Dh+2>(var 選中, 1, 1, 0Fh, 0Ah, 0Dh,
                        @DS:(4A4h ＋ 頁 * 29h));    { 選單字串 }
    if 選中 <> 0 then begin
        if 頁 = 0 then <far 1022h>(鍵)              { Select：記住是誰 }
        else if 鍵 = 47h then <本模組 158Ah>()      { Home }
        else if 鍵 = 4Fh then <本模組 16CBh>();     { End }
        <far 14F8h+2>(遠指標(DS:6506h));
    end
    else if <0A54h:08D4h>(切換鍵表, 鍵) then begin
        頁 := 1 − 頁;
        if 頁 <> 0 then begin
            暫 := 名字 ＋ 'has been selected';
            <far 1554h>(0, 0Ah);
        end else
            <far 1558h+1>();
    end;
end;
```

選單字串的 stride 是 **`29h` ＝ `string[40]`**（PC-98 `51h` ＝ `string[80]`），
與 spec 859 讀到的字串上限放寬一致。

## PC-98：熱鍵獨立成另一份 ASCII 字串（第五個出現點）

DOS 的 `'Select Exit'` 自帶可按的字母 `S`／`E`；PC-98 的
`'   選択                       抜ける    '` 是**固定欄位排版**（用半形空白對齊），
標籤裡沒有任何可按的字元。所以 PC-98 在迴圈開頭多做一段：

```pascal
if 頁 = 0 then 指派(DS:09FFh+1 → DS:0A335h, 0Ah)
          else 指派(DS:0A09h+1 → DS:0A335h, 0Ah);   { 0A65h:0262h }
```

**兩頁各一份 10 bytes 的熱鍵字串**，寫進同一個緩衝 `DS:0A335h`。
這與 spec 817／844／599／865 是同一個模式，這是第五個獨立出現點——
**可以當成 PC-98 在地化的通則，不是個案**。

中文版跟 PC-98 走：選單標籤用固定欄位排版，熱鍵字元另存一份 ASCII。

## ⚠ 三處兩平台真的不同

| | DOS | PC-98 |
|---|---|---|
| 移到最前的鍵 | `47h`（Home 掃描碼） | `38h` |
| 移到最後的鍵 | `4Fh`（End 掃描碼） | `32h` |
| `169Dh+2` 的第四個參數 | `1` | `0` |

鍵碼不同是 PC-98 鍵盤本來就不同；**remake 的鍵位對應要分平台**，
不能只做一套。第三項本規格不判斷影響。

## ⚠ PC-98 側的反組譯越界

PC-98 的函式在 `retf` 之後還被 IDA 當成程式碼繼續解，
解出來的是 Shift-JIS 字串 bytes（`int 8Ah`、`retf 0EA82h` 之類）。
**判讀以 `retf` 為止**；那 27 條不是程式。

## 明確不宣稱

- 沒有宣稱 `far 169Dh+2` 的七個參數各是什麼（形狀上是「畫選單並收一個鍵」）。
- 沒有宣稱 `far 1022h(鍵)`、`本模組 158Ah`／`16CBh`、`far 14F8h+2` 各做什麼。
- 沒有宣稱兩張鍵表（`loc_17D4`／`unk_1802`）的內容。
