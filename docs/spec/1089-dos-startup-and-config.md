# 1089 — DOS 開機流程：`curse.cfg` 不存在時才問顯示卡／音效／存檔路徑

- 證據等級：`exact`（DOS resident 側 609 條逐條讀完）
- 作法見 spec 783；PC-98 對側見 spec 1088

## `dos START.EXE:1236Ch`（`retf`）

原本待解讀。這是 DOS 版的**開機序列 ＋ 首次設定畫面**。

★ resident 的反組譯有 **Borland 的 demangled RTL 名稱**（`@OVRINIT$q6String`、
`@FSplit$…`、`@Set@MemberOf$q4Byte`、`@MemAvail$qv`…），
所以這一支的 RTL 呼叫幾乎不必推測。

## ★★★ 一、與 PC-98 完全相同的路徑推導

```pascal
<1636:0206h>();
OvrInit('GAME.OVR');
路徑 := FExpand('*.exe');
FSplit(路徑, 目錄, 名稱, 副檔名);
if 路徑[1] >= 'C' then  工作路徑 := 路徑
else if 路徑[1] = 'A' then 工作路徑 := 'B:\'
else                       工作路徑 := 'A:\';
```

> ★★★ **這段與 spec 1088 的 PC-98 版一字不差**（連 `'*.exe'`／`'B:\'`／`'A:\'`
> 三個常數都一樣）——「兩片軟碟互為資料碟」是**兩平台共用的設計**。

## ★★★ 二、`curse.cfg` 不存在才進設定畫面

```pascal
if not <1636:04FFh>('curse.cfg') then begin        { 檔案不存在 }
    Assign(檔案, 'curse.cfg');  Rewrite(檔案);
    if IOResult <> 0 then begin
        ClrScr;
        WriteLn("Sorry, disk A can't be write-protected.");
        WriteLn;
        WriteLn('If these are your masters, use DISKCOPY to make duplicates.');
        Halt;
    end;
    …三個問題…
    Close(檔案);
end;
```

> ★★★ **設定值是寫進 `curse.cfg` 的**，所以**遊戲磁片不能防寫**
> ——那句 `"Sorry, disk A can't be write-protected."` 就是這麼來的，
> 而且後面直接建議玩家用 `DISKCOPY` 複製一份再玩。
> ⇒ remake 若把設定放在唯讀目錄，等於重現這個限制；建議改放使用者目錄。

### 三個問題

```pascal
ClrScr;  Write('GRAPHICS ADAPTER TYPE:  [1] CGA  [2] EGA  [3] Tandy  ');
n := <near 12110h>();                              { 讀一個數字 }
case n of 1: 寫 'C';  2: 寫 'E';  3: 寫 'T' end;    { 寫進 curse.cfg }

ClrScr;  Write('SOUND TYPE:  [1] PC  [2] Tandy 1000  [3] Silent  ');
n := <near 12110h>();
case n of 1: 寫 'P';  2: 寫 'T';  3: 寫 'S' end;

預設 := 工作路徑 ＋ 'SAVE';
Write('PATH TO SAVE DATA (DEFAULT - ' ＋ 預設 ＋ '):  ');
ReadLn(輸入);  輸入 := UpCase(輸入);
if (Length(輸入) >= 4) and (輸入[2] = ':')
   and (輸入[1] in ['A'..'P']) and (輸入[1] <> 工作碟)
   and (輸入[1] in ['A', 'B'] 的判斷…) then
    存檔路徑 := 輸入
else begin
    存檔路徑 := 預設;
    WriteLn('USING DEFAULT PATH: ' ＋ 存檔路徑);  Delay;
end;
if 存檔路徑 的最後一個字元 <> '\' then 存檔路徑 := 存檔路徑 ＋ '\';
WriteLn(檔案, 存檔路徑);
```

★ 兩個集合常數：**`CS:12294h` ＝ `['A', 'B']`**（軟碟）、
**`CS:122B4h` ＝ `['A'..'P']`**（所有磁碟機）。
⚠ 兩者都要用 `@Set@MemberOf` 讀成 **32 bytes 位元集合**（spec 1075／1076 同一個坑）。
★ 三個字元寫進 `curse.cfg` 的**前兩行 ＋ 第三行是存檔路徑**，
最後還會再寫一個字元（第四行）。

## 四、Tandy 提示與記憶體檢查

```pascal
ClrScr;  WriteLn('TANDY 1000 USERS - Make sure NUMLOCK is on.');  Delay;

Assign(檔案, 'curse.cfg');  Reset(檔案);
Read(檔案, 顯示卡字元);  ReadLn(檔案);
INTR(…, 暫存器);                                   { 偵測實際顯示卡 }
if 顯示卡字元 = 'E' then <1297:00D7h>()            { EGA }
else if 顯示卡字元 = 'T' then begin Move(…);  <1297:00D7h>() end   { Tandy }
else <1297:00D7h>();                               { CGA }

if MemAvail − OvrGetBuf 不夠 then begin
    OvrSetBuf(…);
    Write('Not enough memory to run game.  You need ');
    Write(差額);  Write(' more bytes.');  WriteLn;  Halt;
end;
if OvrResult <> 0 then begin
    WriteLn('Overlay error, program abort!');  Halt;
end;

Read(檔案, 音效字元);  ReadLn(檔案);               { 'T'／'P'／其他 }
Read(檔案, 存檔路徑);  ReadLn(檔案);
Read(檔案, 第四個字元);  ReadLn(檔案);             { 與 'F' 比較 }
Close(檔案);
```

★ **顯示卡的選擇會被 `INTR` 的實際偵測結果覆蓋／校正**，不是完全信設定檔。
★ `curse.cfg` 的第四行是一個字元，與 `'F'` 比較——本規格不宣稱它是什麼。

## ★★ 兩平台的設定項目完全不同

| | DOS | PC-98（spec 1088） |
|---|---|---|
| 設定時機 | **`curse.cfg` 不存在時**，只問一次 | **每次開機**都顯示 |
| 顯示卡 | **CGA／EGA／Tandy** 三選一 | 無（PC-98 只有一種） |
| 音效 | **PC／Tandy 1000／Silent** 三選一 | BGM ／ 効果音 兩個開關 |
| 存檔路徑 | **可輸入** | 無 |
| 其他 | Tandy 的 NUMLOCK 提示 | 魔法範圍顯示、液晶螢幕 |
| 切換方式 | 只能刪掉 `curse.cfg` 重設 | **遊戲中用 Ctrl 熱鍵隨時切** |

## 中文化

| DOS | 長度 | 建議 |
|---|---|---|
| `"Sorry, disk A can't be write-protected."` | 39 | 「抱歉，A 磁碟不能設成防寫。」 |
| `'If these are your masters, use DISKCOPY to make duplicates.'` | 59 | 「若這是原版磁片，請用 DISKCOPY 複製一份再玩。」 |
| `'GRAPHICS ADAPTER TYPE:  [1] CGA  [2] EGA  [3] Tandy  '` | 53 | 「顯示卡類型：　[1] CGA　[2] EGA　[3] Tandy　」 |
| `'SOUND TYPE:  [1] PC  [2] Tandy 1000  [3] Silent  '` | 49 | 「音效類型：　[1] PC　[2] Tandy 1000　[3] 靜音　」 |
| `'PATH TO SAVE DATA (DEFAULT - '` ＋ 路徑 ＋ `'):  '` | 29 ＋ 4 | 「存檔資料路徑（預設 － 」…「）：　」 |
| `'USING DEFAULT PATH: '` | 20 | 「使用預設路徑：」 |
| `'TANDY 1000 USERS - Make sure NUMLOCK is on.'` | 43 | 「Tandy 1000 使用者－請確認 NUMLOCK 已開啟。」 |
| `'Not enough memory to run game.  You need '` ＋ 數字 ＋ `' more bytes.'` | 41 ＋ 12 | 「記憶體不足，還需要 」…「 位元組。」 |
| `'Overlay error, program abort!'` | 29 | 「Overlay 錯誤，程式中止！」 |
| `'GAME.OVR'`／`'curse.cfg'`／`'*.exe'`／`'A:\'`／`'B:\'`／`'SAVE'`／`'\'` | — | ⚠ **檔名與路徑，一律不可翻** |

⚠ 這一段全部走 **DOS 主控台（`ClrScr`／`Write`／`WriteLn`）**，不經過遊戲的字型
——**中文要顯示得出來，得先掛好 DOS 中文系統**，否則只能留英文。
⚠ 三個問題的答案是**按數字鍵 `1`／`2`／`3`**，數字本身不能翻。

## 明確不宣稱

- 沒有宣稱 `near 12110h`（讀一個數字）的細節。
- 沒有宣稱 `1636:0206h`（進門第一支）、`1297:00D7h`／`1297:0107h`
  （顯示卡設定）、`1542:0073h`／`1542:00A4h` 做什麼。
- 沒有宣稱 `INTR` 的中斷號與它讀回來的內容。
- 沒有宣稱 `curse.cfg` 第四行那個與 `'F'` 比較的字元是什麼。
- 沒有宣稱 `unk_24E6Eh`（記憶體檢查之後那個 word 判斷）是什麼。
- 沒有宣稱 `dos START.EXE:12C1Dh`（另一支待解讀的 resident 函式）與本支的關係。
