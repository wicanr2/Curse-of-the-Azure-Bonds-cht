# 1147 — `33h PRINT RETURN` 是「欄回 1、列 ＋1」，兩個分支做的事一樣

- 證據等級：`exact`（DOS `overlay-02:2CEAh`..`2D14h` 整支 14 條讀完）
- handler 位址取自 `docs/audit/ecl-opcode-handlers-dos.md`

## 全函式

```asm
2CED  inc   word ptr ds:4FB4h      ; PC 推進（每個 handler 都做，spec 1104）
2CF1  cmp   byte ptr ds:8B61h, 0
2CF6  jz    short loc_2D08
2CF8  mov   byte ptr ds:65A0h, 1   ; 欄 := 1
2CFD  inc   byte ptr ds:65A1h      ; 列 ＋1
2D01  mov   byte ptr ds:8B61h, 0   ; 清旗標
2D06  jmp   short loc_2D11
2D08  mov   byte ptr ds:65A0h, 1   ; ★ 跟上面一模一樣
2D0D  inc   byte ptr ds:65A1h
```

★ **兩個分支對游標做的事完全一樣。** `8B61h` 只決定要不要順手把它清成 0
（本來就是 0 的話不用清）。所以整條指令的可觀察效果是：

```pascal
DS:65A0h := 1;   { 欄回到 1 }
inc DS:65A1h;    { 列 ＋1 }
```

⇒ 它是**硬換行**，不是「把目前這一行結束掉」：連續兩條 `33h` 會多推一列，
也就是空一行。

## `65A0h` 是欄、`65A1h` 是列

spec 807 當時留了「沒有宣稱 `DS:65A0h` / `DS:65A1h` 是欄或列」。本輪關掉，
兩個獨立證人：

| 出處 | 用法 | 推得 |
|---|---|---|
| spec 826 | `DS:65A0h := DS:6506h^[0] ＋ 2`——拿**名字長度**算出來 | `65A0h` 是**欄** |
| spec 840 | 中段把行號重設成 `DS:65A1h ＋ 1`，逐行 `inc(行)` | `65A1h` 是**列** |

本規格的 `65A0h := 1` ／ `inc 65A1h` 與兩者一致：換行就是欄歸位、列前進。

## remake

`33h` 目前只把指令邊界記成 `PrintReturnCount`，**沒有游標模型**——版面由 UI
那一側決定。缺口因此不是「解讀」而是「表現層」。

### 缺口有多大：10 段會空行（第 751 輪）

`cmd/ecl-print-return-audit`（報表 `docs/audit/ecl-print-return.md`）把走得到的
`33h` 逐條數出來：**120 條、110 個換行段落，其中 10 段是連著兩條 ⇒ 會空行**。

⚠ 「連續」只認**位元組相鄰**，走訪跟不到的碼也不在分母裡 ⇒ 兩個方向都是下界。

那 10 段裡有 **7 段**兩側的文字都落在同一個顯示頁裡，也就是玩家真的看得到的
留白（其餘 3 段落在頁緣或是被跳進來的，兩側對不上同一頁）：

| ECL/區塊 | 位移 | 該頁的 `text_rule` | 空行落在 |
|---|---:|---|---|
| ECL3/0x10 | `0B01h` | `yulash.commander-business` | 「來尤拉什有何目的」／「要怎麼回答」之間 |
| ECL3/0x11 | `03CBh` | `pit.alias-leaves` | 「我們得走了。」之後 |
| ECL3/0x12 | `0FCDh` | `pit.its-your-dead-body` | 「好吧，是你們自己的命。」之後 |
| ECL4/0x20 | `040Bh` | `zhentil.sign.gorge-and-grog` | 招牌名與副標之間 |
| ECL4/0x20 | `0F2Fh` | `zhentil.fritz-accusation` | 「你們殺了弗里茲！」之後 |
| ECL4/0x21 | `0DBAh` | `zhentil.temple.door-below-altar` | 敘述與「要怎麼做？」之間 |
| ECL4/0x22 | `13A4h` | `dexam.departure.dimswart` | 對白與「奧莉芙與丁斯瓦特離開」之間 |

★ **這改變了缺口的性質**：七處全部落在 game pack 的 `text_rule` 服務的頁上，
而顯示出來的是 pack 的譯文、不是把原文段落 join 起來的結果
（`localizeECLText` 只在沒有規則命中時才 join）。UI 的 `wrapTextLines` 本來就
把 `\n` 當段落切、空段落算一列 ⇒ **要補的是那七則譯文的版面，不是 ECL VM，
也不是行模型**。

⚠ 動譯文之前要先看畫面：訊息框有 `maxLines` 上限，多推一列可能把後面的字
截掉。這一步要配實機截圖驗收，不能盲改。

## 明確不宣稱

- 沒有宣稱 `DS:8B61h` 是什麼旗標。本支只證明「它非 0 時會被清成 0」，
  而且**它不影響游標怎麼動**——兩個分支一樣。
- 沒有宣稱 `65A0h`／`65A1h` 的單位（字元格或像素），只宣稱哪個是欄、哪個是列。
