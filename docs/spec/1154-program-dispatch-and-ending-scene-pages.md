# 1154 — `38h PROGRAM` 的四個分派，與結局過場的五頁分頁點

- 證據等級：`exact`（DOS `overlay-02:30DDh`..`321Eh` 104 條逐條讀完；
  結局過場的字串順序與等鍵位置由 `overlay-18:10FFh` 逐條取出）
- 結局台詞全文見 [spec 1082](1082-ending-scene.md)
- handler 的 pascal 骨架見 [spec 1087](1087-ecl-handlers-six.md)

## 邊界（這次 IDA 是對的）

`retf` 在 `321Eh`，下一支函式從 `321Fh` 開始，104 條。⚠ 同一份 audit 表裡
`37h` 那一列的邊界是錯的（見 [spec 1153](1153-load-pieces-wall-set-params.md)），
所以每一支都要各自驗，不能因為一支對了就推論整張表都對。

## 四個分派值，corpus 全用得到

```pascal
if DS:47E2h <> 0 then begin                 { 進場先還原目前角色指標 }
    DS:6506h := 遠指標(DS:47E7h);  DS:47E2h := 0;
end;
READVAR(1);  n := ADDRESSVALUE(1);
case n of
  0: 重新初始化 ＋ 主選單 ＋ 重繪;
  8: 結局序列（見下）;
  9: 存下 DS:4FB4h → 跑 CAMP → 還原 DS:4FB4h → 重繪;
  3: DS:4FC7h := 1;  重繪;
end;                                         { 其餘值什麼都不做 }
```

corpus 13 處（block 清單取自 `ecl-effect-coverage.json` 的 `0x38` 條目）：

| 運算元 | 處數 | 出處 |
|---|---:|---|
| `09` | 7 | `ECL1/0x50`×2、`ECL1/0x51`、`ECL2/0x02`、`ECL3/0x12`、`ECL4/0x20`、`ECL4/0x23` |
| `00` | 4 | `ECL1/0x50`、`ECL1/0x51`、`ECL2/0x01`、`ECL2/0x03` |
| `03` | 1 | `ECL1/0x52:0391h` |
| `08` | 1 | `ECL6/0x43:13E7h` |

⇒ 四個分派值全部用得到，而且**沒有一處落在什麼都不做的預設值**。

## ⚠ `DS:4FC7h` 是離開旗標，不是「訓練免費」旗標

`PROGRAM 3` 設的是 `DS:4FC7h`。那一格在
[spec 1045](1045-dungeon-main-loop-dos.md) 是地城主迴圈的離開旗標
（`until DS:4FC7h <> 0` 之後清 0），在
[spec 1095](1095-ecl-main-loop-and-combat-continuation.md) 是跨迴圈的全域停止，
而 [spec 1152](1152-ecl-damage-targets-saves-and-repeat-attacks.md) 讀出
`2Eh DAMAGE` 的收尾在全隊都倒下時把它設 1。

⇒ **`PROGRAM 3` ＝ 宣告隊伍全滅、停掉迴圈。**

訓練所的三個「這次不收費」旗標是 `DS:7587h`、**`DS:4FC8h`**、`DS:8B6Fh`
（[spec 1084](1084-training-hall-dos.md)）——**`4FC8h` 不是 `4FC7h`**，差一個位址。

## `PROGRAM 8` 的完整序列

```pascal
<overlay-18:10FFh>;                          { 結局過場，spec 1082 }
DS:8B6Fh := 1;                               { 通關 ⇒ 訓練所不收 1000 gp }
bank0^[3FAh] := 0FFh;                        { 已通關（主選單用它印字，spec 1085）}
bank1^[550h] := 0FFh;                        { 訓練所收所有職業（spec 1084）}
p := 隊伍鏈頭;
while p <> NIL do begin                      { 全隊復活補滿 }
    p^[1A4h] := p^[78h];  p^[195h] := 0;  p^[196h] := 1;
    p := p^[189h];
end;
<主選單>;
<印 CS:30BAh>；取鍵；if 鍵 = 'Y' then <存檔>;
<06EAh:0000>;                                { 結束程式 }
```

## ★★★ 結局過場的分頁點

[spec 1082](1082-ending-scene.md) 給了全部台詞，但只寫「…印一段…」沒有邊界。
把 `overlay-18:10FFh` 裡的字串位移與等鍵呼叫（`542h:0946h`）依序取出來之後，
分段是**五頁、四次等鍵**：

| 頁 | 行 | 內容起訖 |
|---|---:|---|
| 1 | 4 | `Tyranthraxus' spirit coalesces…` ～ `…escape through the pool.` |
| 2 | 4 | `As you reach for the Pool of Radiance…` ～ `…contracts and shatters it.` |
| 3 | 4 | `'I am trapped without escape…` ～ `…crumbles into nothingness.` |
| 4 | **8** | `You are certain he is destroyed…` ～ `…to a fine feast.'` |
| 5 | 4 | `You are teleported to Shadowdale…` ～ `'You have won!'` |

- ★ **第 4 頁是 8 行，其餘四頁各 4 行**——它把「枷印消失」與「騎士們抵達」
  合在同一頁，不是兩頁。
- 第 5 頁演完**沒有**等鍵：直接落回 `PROGRAM 8` 的主選單與存檔詢問。
- 換 bigpic 的 `<overlay-29 entry#9>('z')` 落在第 4 頁的等鍵與第 5 頁之間
  ——慶功宴那張圖在最後一頁才換上。

## 中文化

⚠ **spec 1082 的中文草稿用的譯名與專案 glossary 不一致**
（提蘭斯拉克蘇斯／密斯卓諾／影谷／艾米斯特／莫安德／萊桑德／光輝之池）。
以 [`docs/knowledge/coab-glossary.md`](../knowledge/coab-glossary.md) 為準：
提朗瑟克斯、烈焰之主、迷斯卓諾、暗影谷、伊爾明斯特、摩安德護手、
洛山達護符、光芒之池、加里、娜卡西亞。

⚠ 原作字串寫的是 `Amulet of Lythander`，而 glossary 收的是
`Amulet of Lathander`（洛山達護符）——原作拼錯了，中文用 glossary 的譯名。

## remake

- `PROGRAM 8` 先播結局五頁（`ending_page_1`..`5`，每頁一個「按任意鍵」），
  演完才進存檔詢問。先前直接跳到存檔詢問，**打通關的玩家一句結局都看不到**。
- 分頁照原作的等鍵位置切，不是照句子切。
- 主線通關測試會逐頁走過去，順便驗那五頁在真實玩家路徑上到得了、而且是中文。
- `PROGRAM 3` 的隊伍全滅語意 remake 本來就是對的（spec 1087 那句「訓練免費」
  是把 `4FC7h` 誤寫成 `4FC8h`）。

## 明確不宣稱

- 沒有宣稱 `DS:47E2h`／`47E7h`（進場還原目前角色指標）由誰設定。
- 沒有宣稱 `DS:4FB4h`（`PROGRAM 9` 存還的那一格）是什麼。
- 沒有宣稱印字那八個參數各自的意義（spec 1082 同樣留白）。
- 沒有宣稱 `bank0^[3FAh]` 在主選單以外還有沒有讀取端。
- remake 沒有接 `DS:8B6Fh`／`bank1^[550h]` 的通關後效果：前者是「訓練免費」，
  但 remake 的通關流程是終端的（存檔或結束），走不回訓練所；後者的
  「收所有職業」remake 本來就無條件成立（`trainerMask` 一律 `0FFh`）。
