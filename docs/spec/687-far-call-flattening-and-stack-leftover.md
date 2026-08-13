# 第六百八十七輪：far call 的假名字，與靠堆疊殘留補參數

狀態：`READY`。等級：`exact`（呼叫圖解析）／`strong inference`（殘留參數的語意）。
日期：2026-08-14
工具：`scripts/overlay_call_graph.py`、`docs/audit/overlay-call-graph.json`

## 一、`call sub_1953` 不是呼叫 `1953h`

DOS `overlay-02:197Ah` 的位元組是 `9A 93 00 8C 01`，也就是
`call far 018Ch:0093h`。IDA 印出來的卻是 `call sub_1953`——因為
`018Ch×16 + 93h = 1953h`，剛好落在同一個 overlay 的線性位址上，於是它套了那裡
的標籤。

這不是偶發：raw overlay 的反組譯**沒有段的概念**，每一個 `9A` far call 都會被
攤平一次。所以

> **反組譯輸出裡 far call 那一行的名字一律不可信。**

而且它壞得很安靜：`197Ah` 那行看起來像是呼叫自己內部的 `1953h`（`jmp 1A4Fh`），
語意上完全講得通——「跳到收尾段」。真正的目標是 `overlay-32` 的 entry#23。
照這種標籤畫出來的呼叫圖會是自我封閉的假圖。

## 二、解回模組與 entry 編號

VROOMM 的控制區版面（spec 562）給了機械解法：`+20h` 起是 stub 陣列，
一個 stub 5 bytes。所以

```text
entry 編號 = (off − 20h) / 5      條件：off >= 20h 且 (off − 20h) mod 5 = 0
模組       = 段值 × 16 + 基底 H，去 manifest 的 executable_offset 查
```

基底 H 由 manifest 反解（取讓最多 far call 命中的值）：**DOS 1968、PC-98 2464**。

驗證用的是三個**沒有參與擬合**的條件：offset 必須落在 stub 格線上、編號必須小於
`entry_count`、目標 `code_offset` 不得是 `0FFFFh`。結果是**段值命中模組的
far call，100% 也通過這三關**（DOS 2346/2346、PC-98 2417/2417）。基底若取錯，
光是 5 bytes 的格線就會刷掉約八成。

| 平台 | 解出的 overlay→overlay 邊 | 呼叫點 | resident／未知 |
|---|---|---|---|
| DOS | 1730 | 2346 | 2942 |
| PC-98 | 1811 | 2417 | 3737 |

剩下的一半是打到 resident（`0542h`、`0713h`、`0A54h` 等固定段），本工具不猜。

## 三、參數不是全部由呼叫端推入的

`0A54:064Eh`（字串複製）在 `overlay-02:09D3h` 被推入 **5 個 word**：

```text
push cs; push di(unk_970)     ← 來源
push ss; push di(@var_104)    ← 目的
push ax(0FFh)                 ← 上限
call far 0A54h:064Eh
```

但同一支在 `09B8h` 只被推入 **3 個 word**（目的 ＋ 上限），`overlay-03:0380h`
與 `03B1h` 也是 3 個。兩邊的共通點是**前一條指令是另一個 far call**：

```text
call far 0542h:0722h          ← 消耗自己的參數後，留 2 個 word 在堆疊上
lea  di, [bp+var_104]
push ss; push di
push ax(0FFh)
call far 0A54h:064Eh          ← 來源就是上一個 call 留下的那 2 個 word
```

這是 spec 628／682／684 之後的**第四種數參數陷阱**，而且方向和前三種都不同：
前面三種是「推入的比呼叫端自己用得到的多」，這種是「推入的比被呼叫端要的少，
差額由前一個呼叫留下」。**只看單一 call 前面連續幾個 `push`，會把參數數到短
兩個，而且短的正好是最前面那個（來源），剩下的參數還會整組錯位。**

函式結尾的 `mov sp, bp` 讓這種殘留不會累積成堆疊失衡，所以執行期也不會出事——
沒有任何症狀可以提示你數錯了。

`0542:0722h` 本身的參數個數還沒定（`overlay-02` 推 7 個、`overlay-03` 推 3 個，
中間還隔著另一個可能同樣留殘留的呼叫），本輪不宣稱。

## 四、`overlay-02:179Ah`：場景切換的分派

`GUARD(f)` 是全篇重複的形狀：

```text
if DS:5BECh <> 0 then <far 0713:0020h>(word ptr DS:25AAh)
f()
if DS:5BECh =  0 then <far 0713:00D6h>()
```

**前後兩半是用同一個旗標的相反值當閘門**，所以除非 `f` 自己改動 `DS:5BECh`，
兩者只會執行其中一個。這不是配對的存檔／還原，抄成對稱結構就錯了。

主體：

```text
DS:4FB4h := DS:4FB4h + 1
if DS:8B69h <> 0 或 DS:8B56h <> 0 then goto 重載路徑
if  [4F9Dh]^[6D8h] = 1 then [4F9Dh]^[6D8h] := 0；GUARD(overlay-06 entry#3)；overlay-06 entry#1
elif [4F9Dh]^[5C4h] = 1 then [4F9Dh]^[5C4h] := 0；GUARD(overlay-04 entry#2)；overlay-04 entry#1
else                                              GUARD(overlay-05 entry#2)；overlay-05 entry#1
if [4F99h]^[1CCh] = 0 then GUARD(overlay-27 entry#6)
else                       GUARD(overlay-30 entry#11)
→ 收尾
```

三條分支是同一個形狀套在三個不同模組上（06／04／05），只有 overlay-06 那條用
entry#3 而不是 entry#2。

重載路徑（`1956h`）依序呼叫 overlay-08 e2、overlay-10 e2、overlay-09 e2、
overlay-13 e34、overlay-32 e23、overlay-31 e7，然後：

```text
n := MAXRANGE(DS:7211h, DS:7210h, DS:720Fh)         ← overlay-07 entry#6 @04BDh
if n < bank1^[582h] then bank1^[582h] := n          ← 只會把值**調小**
```

比較是 `cmp ax, [582h]` ＋ `jnb`，所以只在 `n` 比現值小的時候才寫入——是往下
夾，不是設定。`bank1^[582h]` 是**剩餘可走步數**：`MAXRANGE` 在非地城時直接填
2，`0801h`（spec 685）每走一步遞減一次。

收尾（`1A4Fh`）：

```text
DS:4FBAh := 4 或 3         ← 依 [4F99h]^[1CCh] 是否為零
[4F9Dh]^[594h] := [4F9Dh]^[594h] and 1              ← 只留最低位
<far 0A54h:1AE0h>(DS:8B48h, 2, 0)
DS:8B62h := 0
<overlay-24 entry#37 @2A6Dh>()
```

## 五、`overlay-02:0972h`：讀一行字串寫回運算元 2

```text
READVAR(2)                                     ← overlay-07 entry#2
buf := 0
dst := ADDFNC(DS:7707h, DS:76C7h)              ← 運算元 2 的位址（spec 686）
<far 0542h:0722h>(@line, @buf, 0Ah, 0, 28h)    ← 讀入，28h = 40 字元上限
複製（前一呼叫留下的來源）→ text，上限 0FFh
if text 的長度位元組 = 0 then 複製 CS:0970h → text    ← 空輸入時填預設字串
STORESTRING(dst, @text)                        ← overlay-07 entry#17
```

⚠ 推入順序是**位址在前、字串在後**。PC-98 側那筆的記法寫成 `STORESTRING(s, addr)`，
順序相反；本輪讀到的 DOS 位元組是 `push [bp+var_4]`（位址）再 `push @text`。
兩者必有一邊的記法寫顛倒，引用前要回去對 PC-98 的位元組。

空字串的判斷是 `cmp [bp+var_104], 0`——**讀的是 Pascal 字串的長度位元組**。

## 明確不宣稱

- `0542h`／`0713h`／`0A54h` 這些 resident 段各自對應什麼模組。
- `0542:0722h` 的參數個數。
- `DS:5BECh`、`DS:8B69h`、`DS:8B56h`、`DS:4FBAh`、`[4F9Dh]^[594h]` 的語意。
- `overlay-04`／`05`／`06` 三者的差別。
