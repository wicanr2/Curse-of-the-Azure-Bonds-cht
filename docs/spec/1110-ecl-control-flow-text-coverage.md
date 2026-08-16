# 1110 — ECL 劇情串接：把分母算對，然後把它接完

- 狀態：`READY`
- 證據等級：`exact`（分母由控制流走訪逐頁列出，逐筆可重跑；三個變長指令的長度公式
  與 `internal/ecl/runtime.go` 的 handler 逐行對齊）
- 產物：[`docs/audit/ecl-text-coverage.md`](../audit/ecl-text-coverage.md)、
  同名 `.json`；工具在 `cmd/ecl-text-coverage`
- 相關：spec 1104（一頁 ＝ 一次 `PRINTCLEAR`）、1106（`IF` 跳過下一條指令）、
  1105（game pack 分檔與 id 命名）

## ★★★ 一、`unmatched=0` 曾經什麼都不代表

上一版的分母是 **197 頁**，而且是這樣算出來的：`TraceGraph` 從五個 lifecycle
entry 走，跟循序與 `GOTO`／`GOSUB`，然後把走到的指令依 **offset 排序**，用
`12h PRINTCLEAR` 切頁。把 `unmatched` 清成 0 之後，看起來像做完了。

實際上那 197 頁只佔全部的兩成。三個獨立的缺陷疊在一起：

| # | 缺陷 | 後果 |
|---:|---|---|
| 1 | `25h ON GOTO`／`26h ON GOSUB` 的目的地沒被走訪 | 選單分支、隨機事件、街頭傳聞整批不在分母裡 |
| 2 | 同上三個 opcode 的**指令長度**沒算 | `Instruction.Next` 落在目標表的位元組上，把資料當程式解 |
| 3 | 頁的歸屬用 offset 順序 | 「一段位址」不等於「一次執行」 |

第 2 點特別值得記：`25h`／`26h`／`15h`／`2Bh` 在命令表裡的 arity 都是 **0**，因為
它們的長度要讀完 count 運算元才知道。`decodeInstruction` 於是給出
`Next = offset + 1`——**指向自己的第一個運算元**。任何相信 `Next` 的走訪器都會從那裡
繼續解碼，得到一串看起來合理的垃圾指令。這不會報錯，只會讓覆蓋率安靜地少一塊。

⇒ 長度公式補在 `ecl.BranchTargets`（`25h`／`26h`）與 `ecl.MenuEnd`（`15h`／`2Bh`），
兩者都照著 `runtime.go` 對應 handler 的運算元佈局寫，並讓 `TraceGraph` 一起用。

## ★★★ 二、offset 順序與控制流，只能二選一

修好前兩點之後，第三點就無法迴避了。兩個需求在 offset 順序上**互相矛盾**：

- 一張 `ON GOTO` 表底下的七個隨機事件在位址上緊鄰。實機一次只跳到其中一條，
  所以它們**必須切開**——不切開的話七段文字併成一份，
  **只寫一條規則，另外六段也會變成「已接上」**。
- if/else 的合流（`IF; GOTO x` 跳過 else，之後繼續印同一頁）要求**不能**在
  `GOTO` 處切開——切開的話一頁被腰斬，原本接得上的規則接不上。

同一個 `GOTO`，一邊要求切、一邊要求不切。⇒ 改成跟著控制流走
（`cmd/ecl-text-coverage/runs.go` 的 `walkRuns`）：從五個 entry 出發，
`IF` 兩條路都走、`GOSUB`／`RETURN` 用堆疊、`ON GOTO` 每個目的地各一條路，
會 `return result` 的指令（選單、戰鬥、寶物、輸入、換 block）就是一次 run 的結束——
那正是 runtime 呼叫 `MatchText` 的時機。

### 二之一、三個讓走訪安靜漏頁的坑

★ **「同一條路走過同一個位址就停」會砍掉正常流程。** 原作每一頁結尾都
`GOSUB` 同一支翻頁提示子程式，一條長路徑會踏過它十幾次。第一版設「走兩次就停」，
於是第三次呼叫時整條路被砍掉，**該子程式返回之後的劇情全部消失**
（`ECL1.DAX/0x50 +11AEh` 的暗影谷客棧就是這樣不見的，而它前後的頁都還在，
從結果完全看不出漏了東西）。迴圈防護交給「同樣的位址＋堆疊＋已累積文字不再走
第二次」與步數上限，兩者都不會誤殺「同一支子程式被呼叫很多次」。

★ **去重去掉的必須是文字，不是頁。** 同一份文字可以由不同路徑產生，
而且**印在不同的位址上**（共用敘述在 corpus 裡有好幾份）。第一版遇到重複就整條丟掉，
於是其餘那些位址沒有任何 run 認領，報告把它們算成「沒接上」——明明規則早就命中了
那份文字。改成把頁併進既有那一條。

★ **狀態鍵不要塞整份文字。** `IF` 每一個都分岔，狀態數與文字長度相乘。
改成持久化串列（分岔只複製指標）＋增量雜湊之後，整份 corpus 的走訪從
26 秒降到 1 秒級，上限也才能從 40 萬拉到 400 萬。

## 三、分類：只留下真的驗不到的那一類

| 狀態 | 意思 | 數量 |
|---|---|---:|
| `matched` | 有規則命中它所在的 run | 999 |
| **`unmatched`** | **還沒寫規則——待辦** | **0** |
| `variable-insert` | 頁裡印的是執行期的值，靜態文字裡沒有那幾個字 | 16 |
| `subroutine` | 共用子程式的片段，實機一定被併進呼叫端那一頁 | 7 |
| 合計（控制流可達的頁） | | **1022** |

★ **`gosub-insert` 與 `branch-insert` 從狀態降級成註記。** 它們原本擋在
`unmatched` 前面，意思是「工具比對不了，規則可能早就寫好了」——那正是
**誤判做完的形狀**：待辦躲進一個看起來已處理的桶子。這兩種插入現在都由走訪器
展開，展開完還是沒命中就照樣算 `unmatched`。

★ `subroutine` 的判準要**兩個條件同時成立**：落在被 `GOSUB` 呼叫的範圍內，
**且**從來沒有和別的頁同屬一份 run。只看前者會把 `THE DOOR IS LOCKED` 這種
共用的整頁吞掉；只看後者會把真正的單句事件吞掉。

## ★★★ 四、三種「寫了規則反而更糟」

本輪補了 545 條 text_rule（464 → 1009）。過程中撞到三種會讓實機變差的寫法：

| 形狀 | 例 | 為什麼糟 |
|---|---|---|
| **片段跨越 run 邊界** | `['THANK YOU. RETURN SOON.', 'THE FINEST CORMYR STEEL']` | 兩句在實機是**兩次**執行，永遠不會同時出現在一份文字裡 |
| **片段短到會攔截別人** | `['YOU MOVE AWAY.']` | 商店、賢者、神殿的道別頁全部被攔走 |
| **把執行期的值換成固定句** | `YOU SEE A SIGN OVERHEAD <招牌名>` | 招牌名、金額、距離被吃掉，玩家拿不到他需要的資訊 |

處置：第一種拆成逐 run 的規則；第二種只保留一條並**排在 `text_rules` 最後**
（前面所有規則都優先，它只當保底）；第三種**不寫規則**，留在 `variable-insert`
——要嘛像 `world.night-note.24/35/42` 那樣按值列舉，要嘛等能插值的機制。
**沒有規則只是顯示原文，寫錯規則是把資訊刪掉。**

## ★★★ 五、分母與比對走兩趟

第一版把「有哪些頁」與「那一頁接上沒有」綁在同一次走訪裡，於是
`ECL1.DAX/0x50`／`0x51` 一碰到狀態上限，那個 block 就**同時**失去頁與比對，
報告只能寫「可能有頁沒進分母」——一句無法查證的保留。

拆成兩趟就沒有這個問題：

| 走訪 | 狀態 | 保證 |
|---|---|---|
| `walkPages`（分母） | 位址 ＋ 子程式摘要 ＋「這一頁開著沒有」一個位元 | 整份 corpus 都走得完，**頁不會漏** |
| `walkRuns`（比對） | 位址 ＋ 呼叫堆疊 ＋ 已累積文字 | 碰到上限只會**少判**，頁不會消失 |

★ 分母那一趟不能帶呼叫堆疊。**堆疊本身就是乘數**：世界地圖的頁尾翻頁提示
子程式有上百個呼叫點，每個返回位址都是一個不同的堆疊，(位址, 堆疊, 一個位元)
在 200 萬狀態就爆掉。改成 `GOSUB` 進去走一次、記下它產生哪些頁、回來從 `Next`
繼續（子程式摘要），狀態就只剩位址數×2。

★ 子程式回來之後那一頁是開是關取決於它印了什麼，分不清時兩種都推。
這是**過近似**：會多列頁（變成看得見的待辦），不會少列頁。
**分母寧可多，比對寧可少**——兩個方向刻意相反。

⇒ `TestPageWalkCoversEveryPageTheRunWalkFinds` 釘住 `walkRuns ⊆ walkPages`，
並把差額印出來。**實測差額是 0**：截斷目前沒有讓任何一頁比對不到。

## ★★★ 六、`variable-insert` 的值是哪一格來的

這 16 頁是靜態工具**結構上**驗不到的一類。要判斷它們能不能接，先看值的來源
（`cmd/ecl-text-coverage` 現在把運算元位址記進 `variable_inserts`）：

| 格 | 內容 | 可否列舉 | 處置 |
|---|---|---|---|
| `7B01h` | 目的地名（14 個字串常數，ECL1 `0x50`／`0x51`）| ✅ | 逐地列舉 `*.edge`／`world-route.*` |
| `7B89h` | 提爾佛頓店家招牌（7 個字串常數，ECL2 `0x01`）| ✅ | 逐塊列舉 `tilverton.sign.*` |
| `7B88h` | 方向（`NORTH.`／`EAST.`／`SOUTH.`／`WEST.`）| ✅ | 逐向列舉，兩種檢查哨各一組 |
| `7F7Fh` | 手札編號（24／35／42）| ✅ | 已列舉（spec 1108 §一之一）|
| `7F79h` | 酒館傳聞編號 | ❌ 尚未 | 那一格在 corpus 裡被許多不相干流程寫過，要先追出這一頁的寫入來源 |
| `7F7Bh` | 競技場賭金 | ❌ | 算出來的數，無法列舉 |
| `7F82h` | 光球距離 | ❌ | 逐回合遞減，同上 |
| `7C00h` | 隊員名字 | ❌ | 玩家自己取的 |

★ **列舉不會讓 `variable-insert` 的數字變小。** 靜態文字裡沒有那個值，工具永遠
比對不了——這一類的驗收只能在**執行期**做。所以改用
`gamepack.TestVariableInsertPagesAreWiredAtRuntime`：依 `runtime.go` 的
`result.Text` 形狀重建實機字串序列再問 `MatchText`。已接的必須命中，
**還沒接的必須不命中**——「已知沒接」與「忘了接」在報告裡長得一樣，
只有把它寫死才分得開。

## 七、明確不宣稱

- **沒有宣稱後四類（傳聞編號、賭金、距離、隊員名）不需要。** 它們是玩家看得到
  的資訊，現在維持原文；要接需要規則支援佔位符，那是引擎 schema 的改動。
- **沒有宣稱譯文正確。** `matched` 只表示有一條規則的 `all_contains` 全部命中。
- 沒有宣稱事件副作用（解鎖、旗標、戰鬥編成）已還原，那是另一條線。
- 78 條規則在報告裡從未出現在 `rule_id` 欄。多數是 `variable-insert` 家族
  （靜態驗不到），少數是被更早的規則先命中同一份 run。**沒有逐條核過**。

## 八、回歸

| 測試 | 釘住什麼 |
|---|---|
| `ecl.TestRealECLCorpusHasNoUnknownReachableCommands` | 加進 `ON GOTO`／選單長度之後，全 corpus 仍然沒有解不出來的指令 |
| `ecl.TestVariableLengthCommandsAreNotWalkableByNext` | 四個變長指令的 arity 仍是 0（陷阱還在），`RecordEnd` 是唯一正確的走法 |
| `ecl.TestRecordEndCoversEveryVariableLengthRecordInTheCorpus` | corpus 裡 363 筆變長記錄逐筆算得出結尾，且一定大於 `Next` |
| `ecl.TestRecordEndRejectsUnknowableAndOutOfRangeLengths` | 個數是記憶體參照、表被截斷、選項數 0 時一律回錯誤，不回一個看起來合理的數字 |
| `TestPageWalkCoversEveryPageTheRunWalkFinds` | 分母 ⊇ 比對，差額是印出來的數字不是一句保留 |
| `gamepack.TestVariableInsertPagesAreWiredAtRuntime` | 16 頁裡已接的在實機命中、還沒接的明確不命中 |
| `gamepack.TestNoTextRuleIsShadowedByAnEarlierOne` | 沒有規則被更早的規則整個遮蔽 |
| `gamepack.TestDefaultPackMergesAllCommittedParts` | 合併後的條數等於各分檔加總，且 id 不重複 |
| `game.TestRealNewGame*`（三條實機路徑） | 提爾佛頓那一段的每一頁都拿 game pack 的文字比對 |

★ `TestDefaultPackMergesAllCommittedParts` 原本釘死 `text_rules=464`、
`locales=724/724`。那種快照**擋不住任何真正的錯誤**，卻會在每一批內容之後變紅——
註解寫的是「合併不能遺漏或重複」，那才是該釘的東西，本輪改成直接檢查它。

★ 三條實機路徑測試原本斷言的是 `localizeECLLine` 的逐行譯文
（`ecl_tilverton_inn_welcome` 這一類）。整頁規則接手之後那些斷言全部要改成
對著 game pack。**逐行表沒有退休**：它仍然接著還沒寫規則的行，兩者不是二選一
（`internal/game/state.go` 的註解已寫明）。
