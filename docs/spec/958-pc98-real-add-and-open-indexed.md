# 958 — PC-98 的 6-byte real 加法，與索引式資源檔開啟

- 證據等級：`exact`（兩支逐條讀完，來源是 `PC98-GAME.EXE.asm`）
- 位置：`PC98-GAME.EXE` 的 `1AFE4h`、`17960h`
- 作法見 spec 783

## `1AFE4h`（136 行）：6-byte real 的加減

輸入是兩個 `Real`：`dx:bx:ax`（指數在 `ch`）與 `di:si:cx`（指數在 `cl`）。

```
if cl = 0 then 離開（第二個是 0，直接回第一個）;
if al = 0 then 回傳第二個;
if al < cl then 交換兩個運算元;                  { 讓 al 是較大的指數 }

差 := cl − al;
if abs(差) >= 28h then 回傳較大的那個;           { ★ 差 40 個 2 冪就直接吃掉 }

記下符號的 xor（bp），兩邊補上隱含最高位（or 8000h）;
把較小的那個右移「差」位（先整 byte 搬移、再逐位 shr／rcr）;
if 同號 then 相加（add／adc／adc）
        else 相減（sub／sbb／sbb），借位就取二補數並翻轉符號;
結果為 0 → 直接返回;
正規化：最高位還沒立起來就左移（shl／rcl／rcl）並減指數;
指數溢位 → stc 返回;
```

★ **`28h` ＝ 40 ＝ `Real` 尾數的位元數**。兩個數的指數差超過 40 時，
小的那個加下去也影響不到任何一位，所以直接回傳大的——標準的浮點加法短路。

與 spec 948 的 `1B1ACh`（除法）、spec 947 的 `1A8E9h`（長整數除法）
同屬 Turbo Pascal 的軟體數學庫。

## `17960h`（123 行，far，`retf 14h`）：開啟索引式資源檔

**與 DOS 的 `16B24h`（spec 939）是同一支的 PC-98 版**。

```pascal
開資源(資料檔 arg_0, 索引檔 arg_4, @位移 arg_8, 副檔名 arg_C, 主檔名 arg_10);

主 := 限長(arg_10, 14h);
副 := 限長(arg_C,  14h);
名 := 主 ＋ 副;
if not <sub_1741Dh>(arg_4, 0, byte_20AEAh, 名) then begin 結果 := 0;  離開 end;

名 := 主 ＋ 副;                                  { ★ 一模一樣地再組一次 }
if not <sub_1741Dh>(arg_0, 0, byte_20AEAh, 名) then begin 結果 := 0;  離開 end;

BlockRead(arg_0, arg_8^, 2, 讀到幾個);           { <sub_1BFD3h>，spec 944 }
arg_8^ := arg_8^ ＋ 2;
Seek(arg_4, arg_8^);                             { <sub_1C03Bh> }
結果 := 1;
```

**與 DOS 版逐步對應**，兩個差別：

| | DOS（spec 939） | PC-98 |
|---|---|---|
| 路徑來源 | `DS:25AEh ＋ byte_21DAEh × 1Eh`（**一張表，可索引**） | **`byte_20AEAh` 單一固定路徑** |
| 額外參數 | `byte_21EB0h` | 常數 `0` |

★ **PC-98 版沒有那張路徑表**——只有一個固定路徑。
配合 spec 943 的 DOS 版有「兩個路徑輪流試 ＋ 軟碟換片提示」，
可以看出 **DOS 版要應付換磁片，PC-98 版假設檔案都在同一個地方**。

`<sub_1BFD3h>` 就是 spec 944 的 `BlockRead`，`<sub_1741Dh>` 是 PC-98 的
開檔常式（對應 DOS 的 `sub_16645h`／spec 943）。

## 明確不宣稱

- 沒有宣稱 `<sub_1741Dh>` 的內部（尚未讀，它是 231 條的待解讀函式）。
- 沒有宣稱 `byte_20AEAh` 的內容。
- 沒有宣稱 `<sub_1C03Bh>`（Seek）與 `<sub_1AC99h>`／`<sub_1AC7Fh>`／
  `<sub_1AD0Ch>`（字串限長／指派／串接）的內部。
- 沒有宣稱 `1AFE4h` 指數溢位時 `stc` 之後由誰處理。
