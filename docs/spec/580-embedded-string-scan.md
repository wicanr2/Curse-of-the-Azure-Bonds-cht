# 第五百八十輪：overlay 內嵌字串全掃與英日對照

狀態：`READY`。日期：2026-08-14

## 字串在哪裡

Turbo Pascal 的 string 常數是「長度位元組 ＋ 內容」，編譯後**直接躺在 code
段裡**，用 `mov di, offset X` ＋ `push cs` 傳位址。所以它們不在任何資料檔，
而在 `.OVR` 的 overlay code 內——這也是為什麼先前找不到大部分玩家可見文字。

## 掃描判準（兩條同時成立）

1. **形狀對**：位址 `i` 的位元組是長度 `n`（`n >= 4`），其後 `n` 個位元組全
   是合法字元。DOS 要可列印 ASCII；PC-98 另外接受 Shift-JIS 雙位元組序列與
   半形片假名。
2. **有人引用**：某條 `BF lo hi`（`mov di, imm16`）的立即值等於 `i`。

第 2 條是關鍵。只靠形狀會把大量普通資料誤讀成字串；要求「有指令指著它」把
誤判壓到可用的程度。代價是**用其他方式取址的字串掃不到**——
`scripts/scan_pascal_strings.py` 的輸出是下界，不是全集。

結果：**1,635 條**（DOS 32 個模組 792 條、PC-98 31 個模組 843 條）。
清單見 [`../audit/embedded-strings.md`](../audit/embedded-strings.md)。

## 英日對照怎麼對

**不能按模組內的出現順序對。** 英文有單複數分歧
（`takes 1 point of damage` 與 ` points of damage `），日文只有一句，兩平台
的條數不同，按序號對會整段錯位。

可靠的對法是走已配對的函式（[spec 579](579-character-status-fields.md) 之前
建立的助憶碼配對）：配對保證兩支函式的指令序列完全相同，於是
`mov di, imm16` 出現的**位置與次數**也相同，第 k 個引用必然是同一件事。

結果：**102 條**高可信度對照，
見 [`../audit/string-pairs.md`](../audit/string-pairs.md)。引用數不一致的
函式對有 0 個——與「助憶碼相同」的前提一致，等於一道通過的自檢。

## 日文版不是逐句直譯

`overlay-08` 裡，DOS `0ED4h` 的 `Not with that weapon`（武器不對）在 PC-98
是 `0F06h`「そこへは進めない」（過不去）。兩者是同一支函式（67 條指令完全
相同）裡的同一條 `mov di, offset`，配對沒有錯——**是日文版換了說法**。同一
模組另有 `can't go there` ↔「そこへは行けない」。

**中文化以 DOS 英文為準**：英文是原作，日文是再創作。對照表的日文欄用來理解
語境，不是翻譯來源。

## 對台帳的用處

字串是函式語意的錨點。`overlay-23` 的
`starts to cough`／`is Poisoned`／`from Fire`／`from Electricity`／
`lost a spell`／`, and is Dying` 直接指出周邊函式在處理什麼，逐條閱讀的成本
因此下降。後續模組的解讀應**先看該模組的字串再讀函式**。
