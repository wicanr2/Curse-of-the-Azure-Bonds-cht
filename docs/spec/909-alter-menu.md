# 909 — `Alter:` 主選單，並找到 spec 871／885 的呼叫端

- 證據等級：`exact`（兩平台各自逐條讀完；**相似度 0.079，UI 結構真的不同**）
- 作法見 spec 783

## `overlay-15:01D0Ah`（DOS 200 條）↔ `overlay-15:01D4Ch`（PC-98 204 條）

`retf` ＝ 無參數。

| 熱鍵 | DOS | PC-98 標籤 | 呼叫 |
|---|---|---|---|
| `O`（`4Fh`） | — | `' 順序 '` | `本模組 1834h` ＝ **spec 871 隊伍順序選單** |
| `D`（`44h`） | — | `' 離脱 '` | `本模組 199Ch` |
| `S`（`53h`） | — | `' 速度 '` | `本模組 1B4Dh` ＝ **spec 885 遊戲速度選單** |
| `I`（`49h`） | — | `' ｱｲｺﾝ '` | `far 0F14h` ＋ `本模組 15A9h` |
| `P`／`A` | 子選單 | `' ｱﾆﾒ無'`／`' ｱﾆﾒ有'` | 切換兩個開關 |
| `E` | `'Exit'` | — | 離開 |

標題是 `'Alter: '`。DOS 的選單字串放在 `DS:0479h ＋ 1`（資料段），
PC-98 逐格寫進固定欄位陣列。

```pascal
鍵 := 20h;
while <0A54h:08D4h>(離開鍵表, 鍵) do begin
    暫 := 'Alter: ';
    鍵 := <far 169Dh+2>(var 有選中, 1, 1, 0Fh, 0Ah, 0Dh, @DS:0479h＋1);

    if 有選中 then begin
        <far 1022h>(鍵);  <far 14F8h+2>(遠指標(DS:6506h));
    end
    else if 鍵 = 4Fh then <本模組 1834h>()          { Order }
    else if 鍵 = 44h then <本模組 199Ch>()          { Drop? }
    else if 鍵 = 53h then <本模組 1B4Dh>()          { Speed }
    else if 鍵 = 49h then begin <far 0F14h>();  <本模組 15A9h>() end
    else if 鍵 = 50h then                            { Pics → 子選單 }
        repeat
            if DS:4FBEh <> 0 then begin
                選項 := 'Pics on  ';
                if DS:4FBFh <> 0 then 選項 := 選項 ＋ 'Animation on  '
                else                  選項 := 選項 ＋ 'Animation off  ';
            end else
                選項 := 'Pics off  ';
            選項 := 選項 ＋ 'Exit';
            鍵2 := <far 169Dh+2>(…);
            if      鍵2 = 50h then DS:4FBEh := 1 − DS:4FBEh
            else if 鍵2 = 41h then DS:4FBFh := 1 − DS:4FBFh;
        until not <0A54h:08D4h>(離開鍵表, 鍵2);
end;
```

## 兩個規格的呼叫端

- `本模組 1834h` ＝ `overlay-15:01834h`，**spec 871 的隊伍順序選單**（`Order`）。
- `本模組 1B4Dh` ＝ `overlay-15:01B4Dh`，**spec 885 的遊戲速度選單**（`Speed`）。

兩支先前都只知道「是一個選單」，現在知道是從 `Alter:` 進去的。

## ⚠ PC-98 把子選單攤平進主選單

DOS 的 `P` 進去之後才有 `Pics on/off` 與 `Animation on/off` 兩個開關；
PC-98 直接在主選單放兩項：

- `' ｱｲｺﾝ '`（`I`）
- `' ｱﾆﾒ無'` ／ `' ｱﾆﾒ有'`（`A`）——**標籤本身就顯示目前狀態**

`DS:4FBEh`（PC-98 `DS:7F2Ch`）是那個開關；PC-98 切換時另外
`遠指標(DS:7F05h)^[1FEh] := 遠指標(DS:7F05h)^[1FEh] xor 1`，
並依開關把 `DS:0A2C7h` 抄到 `DS:0A2C6h` 或直接寫 1。

PC-98 的 `I` 分支還多一整段：先看 `遠指標(DS:9594h)^[0F7h]`，
為 0 才畫（`far 02A8h:1174h(遠指標(DS:7F56h), @DS:4830h, 0, 0A8h, 2Ah, 0, 0)`），
否則印一句訊息。DOS 沒有這一段。

**這是 spec 783 第 d 類裡幅度最大的一支**（相似度 0.079）。
remake 要選一邊的 UI；本規格不判斷哪一邊比較好。

## 中文化

PC-98 的六個標籤都是 6 bytes：`' 順序 '`、`' 離脱 '`、`' 速度 '`、
`' ｱｲｺﾝ '`、`' ｱﾆﾒ無'`、`' ｱﾆﾒ有'`。後三個**已經被逼到用半形片假名**
（`ｱｲｺﾝ` 4 個半形 ＋ 2 個空白；`ｱﾆﾒ無` 3 個半形 ＋ 1 個全形）。
中文三個字剛好，例如「順序／離隊／速度／頭像／動畫關／動畫開」。

## 明確不宣稱

- 沒有宣稱 `本模組 199Ch`（`D`）與 `far 0F14h`／`15A9h`（`I`）各做什麼。
- 沒有宣稱 `DS:4FBEh`／`DS:4FBFh` 控制的是哪一種顯示。
- 沒有宣稱 PC-98 那段 `02A8h:1174h` 畫的是什麼。
