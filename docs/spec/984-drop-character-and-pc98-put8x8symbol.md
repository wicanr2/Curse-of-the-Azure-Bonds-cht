# 984 — 「丟棄角色」的雙重確認與四句台詞；PC-98 的 `Put8x8Symbol` 同一個洞

- 證據等級：`exact`（兩支逐條讀完；七個字串由 overlay 原始 bytes 讀出）
- 作法見 spec 783

## `dos overlay-17:0260Ch`（112 條，`retf 0`）— DROP CHARACTER

DOS 單邊（PC-98 沒有對應函式）。七個字串全在 overlay 內：

| 位址 | 長度 | 內容 |
|---|---|---|
| `CS:25A7h` | 5 | `'Drop '` |
| `CS:25ADh` | 10 | `' forever? '` |
| `CS:25B8h` | 14 | `'Are you sure? '` |
| `CS:25C7h` | 9 | `'You dump '` |
| `CS:25D1h` | 10 | `' out back.'` |
| `CS:25DCh` | 19 | `' bids you farewell.'` |
| `CS:25F0h` | 27 | `' breathes a sigh of relief.'` |

```pascal
if 遠指標(DS:6506h) = NIL then goto 收尾;
p := 遠指標(DS:6506h);                                  { 目前選中的角色 }

if <overlay-26 entry#6>('Drop ' ＋ p^ ＋ ' forever? ', 0Fh, 0Ah, 0Eh) <> 'Y' then goto 反悔;
if <overlay-26 entry#6>('Are you sure? ',              0Fh, 0Ah, 0Eh) <> 'Y' then goto 反悔;

if p^[196h] = 0 then <overlay-24 entry#19>('You dump ' ＋ p^ ＋ ' out back.')
else                 <overlay-24 entry#19>(p^ ＋ ' bids you farewell.');

<overlay-16 entry#5>(p);                                 { 實際移出隊伍 }
<本模組 3A56h>(1, 0);
goto 收尾;

反悔:
<overlay-24 entry#19>(p^ ＋ ' breathes a sigh of relief.');

收尾:
<overlay-24 entry#2>(p);                                 { 重畫角色列 }
```

### 三件事

1. **要問兩次。** 第一次帶名字（`'Drop <名字> forever? '`），第二次只有
   `'Are you sure? '`。任何一次不是 `'Y'` 都取消。
2. **死人和活人的台詞不同。** `p^[196h] = 0` 走 `'You dump <名字> out back.'`
   （直譯是「把他丟到後面去」），否則走 `'<名字> bids you farewell.'`。
   `+196h` ＝ 0 在 spec 698 是「狀態不在 `{0,1}` 時被清成 0」——
   也就是**不能行動／已死**。這一支等於再確認一次那個欄位的語意。
3. **取消也有台詞**：`'<名字> breathes a sigh of relief.'`（鬆了一口氣）。
   這句只在玩家反悔時出現，**而且 `p` 為 NIL 時不會走到這裡**。

### 中文化

四句台詞都是**名字夾在中間或前面的串接**：

```
'Drop ' ＋ 名字 ＋ ' forever? '
'You dump ' ＋ 名字 ＋ ' out back.'
名字 ＋ ' bids you farewell.'
名字 ＋ ' breathes a sigh of relief.'
```

中文的語序是「要永久丟棄〈名字〉嗎？」「〈名字〉向你道別。」——
**前兩句的兩個片段中文要對調**，不能只換字串內容。
這與 spec 693 的 `'s` 所有格是同一類問題：**英文文法被拆成片段硬接**，
中文化要改的是接法，不只是字。

`'Are you sure? '` 是唯一一句不含名字、可以直接換掉的。

## `pc98 overlay-35:0032Ah`（127 條）＝ DOS `overlay-35:00173h`（spec 781）

`Put8x8Symbol` 的 PC-98 版。函式名同樣由自己的錯誤訊息定住：
`CS:0307h` ＝ `'Bad symbol number in Put8x8Symbol.'`。

五段分組與 spec 781 完全相同，而且**分段起始值這次直接讀到了**：
`DS:488Ah` 起五個 word ＝ **`1, 46, 116, 186, 256`**，
與四道區間界線（`1..45`／`46..115`／`116..185`／`186..255`／`256..295`）逐格吻合。
（DOS 的對應表在 `DS:2680h`。）

| | DOS | PC-98 |
|---|---|---|
| 五個資源遠指標 | `DS:65B6h` | `DS:0964Ch` |
| 五個分段起始值 | `DS:2680h` | `DS:488Ah` |
| 直接畫 | `0297h:08F8h` | `02A8h:0712h` |
| 暫存資源 | `DS:65CAh` | `DS:9660h` |
| 暫存後畫 | `0297h:1110h` | `02A8h:0A86h` |
| 區間比較 | **有號** `jl`／`jg` | **無號** `jb`／`ja` |

### ⚠ 「組沒有被指派」這個洞兩平台都有，只是入口不同

spec 781 指出 DOS 在**編號為負數**時 `組` 沒被指派，之後拿堆疊殘值當索引。

PC-98 改用無號比較，乍看沒有負數問題，但錯誤判斷寫成

```
if 編號 = 0            then 報錯
else if 編號 < 128h     then 略過
else if 編號 > 7FFFh    then 略過
else                        報錯
```

**`編號 > 7FFFh` 一樣略過報錯，而且一樣沒有指派 `組`**——就是 DOS 那個負數區間
換個寫法。**同一個洞，兩種比較方式。** remake 要補的是同一道界限檢查。

### 資源版面再確認一次

`Move(圖庫^[17h ＋ 編號 × 圖庫^[11h]], 暫存^[17h], 圖庫^[11h])`
——表頭 `17h`、`+11h` 是每個符號的位元組數、資料從 `+17h` 起連續排列，
與 spec 781、spec 983 的 `overlay-24:01975h` 完全一致。
**「8x8」這個名字也對得上全專案的不變量：一列 ＝ 8 條掃描線。**

### PC-98 的錯誤路徑多做了幾件事

DOS 是「顯示訊息 ＋ 等一個鍵」。PC-98 是

```
<02A8h:0018h>();
WriteLn(輸出檔 DS:0BF7Eh, CS:0307h);        { 0A65h:178Bh ＋ 0A65h:16C3h }
暫 := <07E3h:0206h>();
<07E3h:0077h>();
```

`0A65h:178Bh` ＋ `16C3h` 是 Turbo Pascal 的 `Write` ＋ `WriteLn` 配對，
目標是 `DS:0BF7Eh` 這個 text 變數。**這是開發期留下來的斷言，隨遊戲出貨。**
本規格不宣稱 `07E3h:0206h`／`0077h` 是不是 `ReadKey` 與 `Halt`。

## 明確不宣稱

- 沒有宣稱 `overlay-26 entry#6` 的後三個參數（`0Fh`／`0Ah`／`0Eh`）是什麼。
- 沒有宣稱 `overlay-16 entry#5`（實際移除角色）與 `本模組 3A56h` 做什麼。
- 沒有宣稱 `DS:6506h` 之外還有沒有別的「目前角色」來源。
- 沒有宣稱 PC-98 錯誤路徑那四個 resident 呼叫的確切語意。
