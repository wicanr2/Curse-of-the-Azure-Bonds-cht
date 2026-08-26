# 1152 — `2Eh DAMAGE`：三種目標形式、豁免的兩套算法，與那個「連打 N 下」的分支

- 證據等級：`exact`（DOS `overlay-02:2942h`..`2C8Fh` 305 條逐條讀完；
  被呼叫的 `MAKESAVE`／`TRYTOHIT` 也各自讀完）
- handler 位址與條數見 [`docs/audit/ecl-opcode-handlers-dos.md`](../audit/ecl-opcode-handlers-dos.md)
- 豁免的既有規格見 [spec 582](582-makesave-and-losedude.md)，命中見
  [spec 577](577-attempttohit-and-effect-chain-walk.md)

## 被呼叫的四支叫什麼

`far <段>:<位移>` 先由 [spec 1112](1112-tpov-entry-stub-map.md) 的算式解成
`overlay-NN entry#K`，再查 PC-98 同一 overlay 的第 K 個 stub 拿 Borland 符號名
（方法與其風險見 [spec 1150](1150-ecl-call-external-routines.md)）：

| 呼叫 | 解析 | 名字 | 這一支做什麼 |
|---|---|---|---|
| `006B:002A` | `overlay-07 entry#2` | `READVAR(5)` | 取本指令的 5 個運算元 |
| `006B:0025` | `overlay-07 entry#1` | `ADDRESSVALUE(k)` | 讀第 k 個運算元的值 |
| `006B:00BB` | `overlay-07 entry#31` | `KILLTHEDUDE(傷害, 角色)` | 扣血 |
| `0141:0048` | `overlay-23 entry#8` | `MAKESAVE(角色, 種類, 調整)` | 豁免檢定 |
| `0141:003E` | `overlay-23 entry#6` | `TRYTOHIT(角色, 攻擊值)` | 命中檢定 |
| `0141:004D` | `overlay-23 entry#9` | `ROLLDICE(顆數, 面數)` | 擲骰 |

## 五個運算元

```pascal
旗標   := ADDRESSVALUE(1);   骰數 := ADDRESSVALUE(2);   骰面 := ADDRESSVALUE(3);
加值   := ADDRESSVALUE(4);   目標旗標 := ADDRESSVALUE(5);
傷害   := ROLLDICE(骰數, 骰面) ＋ 加值;
```

**旗標的第 7 位元決定它是不是旗標。**

## ★★★ 旗標 bit 7 清空 ⇒ 整個 byte 是「打幾下」

```pascal
for k := 1 to 旗標 do begin
    目標 := 隊伍鏈的第 ROLLDICE(1, 隊伍人數) 個;      { 每一下各自挑 }
    if TRYTOHIT(目標, 目標旗標) then KILLTHEDUDE(目標, 傷害);
    傷害 := ROLLDICE(骰數, 骰面) ＋ 加值;             { ★ 每一下之間重擲 }
end;
```

★ 這一支裡 `目標旗標` **整個 byte 是攻擊值**，不是豁免種類——形狀上是
「怪物連續攻擊隊伍 N 次」，不是環境傷害。corpus 用到兩處：`ECL6.DAX/0x40:0F1Eh`
（`02 01 06 06 35`，打 2 下）與 `ECL6.DAX/0x42:115Ah`（`0C 02 08 00 34`，打 12 下）。

★ **重擲的位置在扣血之後**，所以第一下用的是進 handler 時擲的那一份。

## ★★ 旗標 bit 7 設定 ⇒ 才是旗標

| 位元 | 意義 |
|---|---|
| bit 7 | 這是傷害封包（清空則同上，整個 byte 是次數）|
| bit 6 | 全隊 |
| bit 5 | 不擲豁免，直接吃 |
| **bit 4** | **豁免成功仍吃全額傷害** |
| bit 0..4 | 傳給 `MAKESAVE` 的豁免調整值（`and 1Fh`）|

⚠ **bit 4 同時屬於調整值欄位**——原作對同一個 byte 各做一次 `and 10h` 與
`and 1Fh`，兩者重疊。單看程式碼分不出「調整值只有 4 位元、bit 4 是旗標」與
「調整值有 5 位元、順便被當旗標用」哪一種是設計意圖。**corpus 24 處的低 5 位
一律是 0**，所以實務上它只是旗標，調整值從來沒被用過。

⚠ 豁免成功時吃的是**全額**，不是一半——`KILLTHEDUDE` 兩邊傳的是同一個變數。

## ★★★ 目標選擇有三條路，而豁免種類的算法在其中一條不一樣

```pascal
if 旗標 and 40h <> 0 then                       { 一、全隊 }
    p := 隊伍鏈頭;
    while p <> NIL do begin
        if 旗標 and 20h <> 0 then KILLTHEDUDE(p, 傷害)
        else if not MAKESAVE(p, 目標旗標 and 7, 旗標 and 1Fh) then KILLTHEDUDE(p, 傷害)
        else if 旗標 and 10h <> 0 then KILLTHEDUDE(p, 傷害);
        p := p^[189h];
    end
else if 目標旗標 and 80h <> 0 then               { 二、目前角色 DS:6506h }
    if 目標旗標 and 7 = 0 then KILLTHEDUDE(目前角色, 傷害)      { ★ 0 ＝ 不擲豁免 }
    else if not MAKESAVE(目前角色, (目標旗標 and 7) − 1, 旗標 and 1Fh) then …
else                                             { 三、隨機一名 }
    p := 隊伍鏈的第 ROLLDICE(1, 隊伍人數) 個;
    if not MAKESAVE(p, 目標旗標 and 7, 旗標 and 1Fh) then …
```

★★ **第二條路的豁免種類要減一，而且 `0` 代表不擲；另外兩條不減一，要不要擲
由旗標 bit 5 決定。** 這兩套算法在 corpus 上結果相同（走第二條路的封包
`目標旗標` 只有 `80h` 與 `84h`，而 `80h` ⇒ 不擲豁免、`84h` ⇒ 種類 3），
所以只驗行為分不出來，要讀 handler 才看得到。

★ 第三條路在原作裡是活著的程式碼，但 **corpus 24 處沒有一處走得到**
——所有單體封包的 `目標旗標` bit 7 都是設定的。

## corpus 的 24 處

沿 `ecl.TraceGraph` 走完可達的 24 處，`旗標` 只有八種值：

| 旗標 | 條數 | 意義 |
|---|---:|---|
| `C0h` | 10 | 全隊、擲豁免 |
| `A0h` | 4 | 單體（目前角色）、不擲豁免 |
| `E0h` | 3 | 全隊、不擲豁免 |
| `90h` | 3 | 單體（目前角色）、擲豁免、成功仍吃全傷 |
| `80h` | 1 | 單體（目前角色）、擲豁免 |
| `D0h` | 1 | 全隊、擲豁免、成功仍吃全傷 |
| `0Ch` | 1 | 連打 12 下 |
| `02h` | 1 | 連打 2 下 |

「目前角色」那 8 處的產生端一律看得見：`ECL5.DAX/0x32:0223h` 是
`SAVE 0 → LOAD CHARACTER → …DAMAGE… → COMPARE 與 7F3Eh`（`7F3Eh` ＝ 隊伍人數，
[spec 1096](1096-ecl-bank1-address-formula.md)）的**逐一走過整隊迴圈**；
`ECL5.DAX/0x32:036Dh` 是 `WHO "WHO WILL GO?"`；`ECL5.DAX/0x33:0D25h`／`0D46h`
各是一條 `LOAD CHARACTER`。

## `MAKESAVE` 與 `TRYTOHIT`（DOS 側逐條）

```pascal
function MAKESAVE(角色; 種類, 調整: byte): boolean;    { DOS overlay-23:12EBh }
    d20 := ROLLDICE(1, 20);
    if d20 = 1  then exit(false);                      { 自然 1 必敗 }
    if d20 = 20 then exit(true);                       { 自然 20 必成 }
    d20 := d20 ＋ 角色^[186h] ＋ 調整;
    CHECKFX(角色, 0Ch);                                { 效果可以再改 d20 }
    MAKESAVE := (角色^[0DFh ＋ 種類] <= d20);

function TRYTOHIT(角色; 攻擊值: byte): boolean;        { DOS overlay-23:11D7h }
    d20 := ROLLDICE(1, 20);
    if d20 <= 1 then exit(false);                      { 自然 1 必失手 }
    if d20 = 20 then d20 := 100;                       { 自然 20 ⇒ 必中 }
    CHECKFX(角色, 10h);
    if d20 < 0 then exit(false);                       { CHECKFX 壓得成負數 }
    TRYTOHIT := (d20 ＋ 攻擊值 > 角色^[19Ah]);         { ★ 嚴格大於 }
```

- 本檔的角色欄位一律是 **DOS 側**。對到 PC-98 要加 [spec 641](641-dos-field-offset-shift.md)
  的位移：`+186h`↔`+187h`（豁免修正）、`+189h`↔`+18Ah`（next）、
  `+196h`↔`+197h`（還能動）、`+19Ah`↔`+19Bh`（命中目標值）。
  spec 582 與 [spec 571](571-trytohit-attack-resolution.md) 轉錄的是 PC-98 那一組，
  兩邊因此看起來差一——不是矛盾。
- PC-98 側同一支是 `overlay-02:02ACEh`（[spec 609](609-ecl-area-effect-and-wipeout.md)），
  全滅訊息在那邊是「パーティーは全滅した！」。
- ★ **這一支答掉了 spec 582 留著的未定項之一**：`種類` 的取值範圍由呼叫端
  `目標旗標 and 7` 夾成 **0..7**。

## ★★ 收尾：全隊倒下的判定與那句英文

```pascal
全滅 := true;
p := 隊伍鏈頭;
while p <> NIL do begin
    if p^[196h] <> 0 then 全滅 := false;               { +196h ≠ 0 ＝ 還能動 }
    p := p^[189h];
end;
if 全滅 then begin
    <某常駐常式>;  游標欄 := 2;  游標列 := 2;
    <印> 'The entire party is killed!';
    <延遲 3000>;
end;
DS:6506h := <進 handler 時存下來的目前角色>;            { ★ 還原 }
<印> 'press <enter>/<return> to continue';
```

- `+196h ≠ 0 ＝ 還能動` 與 [spec 828](828-drop-from-party.md) 的兩句道別同一個判準。
  ⚠ 這是 **DOS** 的 `+196h`（站著且能行動旗標）；PC-98 的 `+196h` 是
  `CHARSTATUS`——兩平台差一格 offset（spec 1166），對照表與這面旗標的
  生產者（KILLDUDE／STANDUP）見 [spec 1204](1204-party-wipe-criteria.md)。
- 游標欄 `65A0h`／列 `65A1h` 的角色由 [spec 1147](1147-ecl-print-return-hard-newline.md) 定案。
- ⚠ **handler 會把 `DS:6506h` 還原**，所以 `2Eh` 不改變「目前角色」。

## 中文化

兩句都在 DOS `overlay-02` 的 `CS:2903h`／`CS:291Fh`：

| 位址 | 原文 |
|---|---|
| `2903h` | `The entire party is killed!` |
| `291Fh` | `press <enter>/<return> to continue` |

⚠ 後面那句 **PC-98 沒有**（台帳已記；PC-98 `overlay-02:02ACEh` 少掉那 11 條）。

## remake

- VM 早就把五個運算元原樣保存，欄位名也對（`Flags`／`DiceCount`／`DiceSize`／
  `Bonus`／`SaveFlags`）。`internal/party` 的規則也已經照原作寫：次數分支、
  重擲順序、`and 1Fh` 調整值、`and 7` 種類、第二條路的減一與 `0 ⇒ 不擲`、
  自然 1／20、`>= 表值`、bit 4 的成功仍吃傷——逐條對得上。
- 這一輪補的是**接線**與一個順序錯誤：
  1. 正式路徑先前只結算 `Flags and 0C0h = 0C0h`（全隊）那 14 處，另外 10 處
     （8 處目前角色、2 處連打）永遠留在 pending。現在三種形式都結算，只有
     三種形式都結算。
  4. ★ **第三條路「單體但隨機挑一名」也接上了**：擲一顆 `1..隊伍人數` 挑人，
     豁免種類**不減一**、要不要擲由 `旗標` bit 5 決定——與全隊那一路共用同一個
     結算器。corpus 24 處走不到這一路，所以它沒有實機路徑背書，是照 handler 寫的；
     兩條回歸（挑到的人是擲出來的那一個、種類不減一）都突變驗過。
  2. ★ 目標改成由**封包自己**帶（`DamageRequest.SelectedPlayerIndex`）。腳本會
     把 `LOAD CHARACTER` ＋ `DAMAGE` 包在走過整隊的迴圈裡，一次執行累積好幾組，
     事後只看最後一次選的人會把整批傷害算到同一位身上。
  3. ⚠ 結算的呼叫點**維持**在 `applyECLLoadCharacterSignals` 之前。看起來該
     移到後面（腳本是 `LOAD CHARACTER` 緊接 `DAMAGE`），但封包已經帶著發出
     當下的選定角色，本次執行的 `LOAD CHARACTER` 對它們沒有影響；而**沒帶**的
     封包代表 VM 到那一刻為止沒人被選過，該退回的是這一次執行**之前**的選定
     角色。移到後面反而會讓那種封包改用本次最後選到的人。

## 明確不宣稱

- 沒有宣稱 `旗標` bit 0..4 那個調整值在設計上是 4 位元還是 5 位元
  （corpus 全 0，兩種讀法都自洽）。
- 沒有宣稱豁免種類 0..7 各自對應哪一類豁免（spec 582 的另一半未定項）。
- 沒有宣稱 `KILLTHEDUDE` 內部怎麼分配傷害與死亡狀態（另一支，208 條）。
- 沒有宣稱 `CHECKFX(0Ch)`／`CHECKFX(10h)` 會怎麼改寫骰值。
- 沒有宣稱收尾那個 `01A0:0000` 常駐常式做什麼。
- 沒有宣稱角色 `+19Ah` 的完整語意（[spec 577](577-attempttohit-and-effect-chain-walk.md) 稱它命中修正）。
