# 1108 — 手札 producer 普查：59 則裡 45 則玩家在遊戲內讀得到

- 狀態：`READY`
- 證據等級：`exact`（producer 由窮舉掃描全 corpus 取得，逐筆可重跑；「有些條目是假的」
  由手冊本文佐證）；未被引用那幾則的**分類**是 `strong inference`
- 對照原文：`workplace/docs-txt/…/Curse of the Azure Bonds Adventure Journal.txt`
  （SSI 1994 合輯的 etext，含 1–59 除 32、42 之外的全文）

## ★★★ 一、編號是字面值，不需要解字串記憶體

原作有兩種寫法把手札編號交給玩家：

| 形狀 | 例 | 數量 |
|---|---|---:|
| 整句連編號寫死在一條 `PRINT` 裡 | `YOU ENTER THIS AS JOURNAL ENTRY 21.` | 20 |
| 共用子程式印固定句，**呼叫端**下一條 `PRINT` 印編號 | `GOSUB 9BC5h` ＋ `PRINT «58.»` | 22 |

第二種原本被當成擋路的 RE：以為編號來自 `81h` 運算元指向的字串記憶體。
實際看呼叫端就知道，編號是**下一條 `PRINT` 的字面值**，直接讀得出來。
⇒ **先看呼叫端再開 RE**；這一項不需要任何反組譯。

### 一之一、唯一一處是表驅動

`ECL1.DAX/0x50 +1A6Eh`「早上發現同伴胸前別著紙條」的編號來自
`2Ah GETTABLE 9D10h, [4C9Fh] → [7F7Fh]`，再 `PRINT` 那個變數。
計數器 `4C9Fh` 每次加一，`COMPARE [4C9Fh], imm(2); IF >; RETURN` 把它限成三次；
表的頭三個 byte 是 `18 23 2A` ＝ **24、35、42**。

★ remake 不需要為此加任何機制：`runtime.go` 的 `PRINT` 對非文字運算元會
`fmt.Sprint(value)` 併進 `result.Text`，而 `0x8000+` 的 code window 已經把 block
位元組灌進 `memory`（spec 282）。⇒ 三次的實機字串只差那個數字，**寫三條規則用
數字分辨**就能依序解鎖，不必替引擎加「有序解鎖」欄位。

## 二、普查結果

| 項目 | 數量 |
|---|---:|
| Adventurer's Journal 條目 | 59 |
| **corpus 會引用的** | **44** |
| 沒有任何引用 | 15 |
| game pack 已收錄 | **45**，且每一則都有 producer |
| └ 由 corpus 的手札引用觸發 | 44 |
| └ 掛在甦醒那一幕 | 1（第 1 則，見 §二之二） |

「每一則都有 producer」由 `gamepack.TestEveryJournalEntryHasAProducer` fail-closed
擋住。**沒有 producer 的條目是死資料**——玩家永遠讀不到，數字上卻看起來像做完了。

### ★★★ 二之一、corpus 沒有引用的 15 則

`1`、`2`、`6`、`8`、`10`、`13`、`14`、`16`、`28`、`37`、`39`、`40`、`43`、`47`、`49`。
窮舉掃過全 corpus 的每一個 byte offset（不受控制流可達性限制），
`JOURNAL`、`ENTRY`、`RECORD` 三個關鍵字都掃過，沒有第四種措辭。

| 類別 | 條目 | 依據 |
|---|---|---|
| 開場敘事 | `1` | 是「我開始寫一本新手札」那段導言，本來就不由手札引用觸發；已接上，見 §二之二 |
| 只有插圖 | `8`、`39` | etext 裡 `8` 只有 `<IMG SRC=...>`、`39` 是空的 |
| 別部作品的素材 | `2`、`6`、`10`、`13`、`16`、`28`、`40`、`43`、`47` | 內容講菲蘭、特什維夫、塔斯河、匕首瀑布、卡多納、提亞馬特、穆爾馬斯特眼魔軍團——都不是本作的地點與人物 |
| 本作風味但無引用 | `14`、`37`、`49` | 提到無名者、提朗瑟克斯、黑暗精靈神殿，但 corpus 沒有任何地方叫玩家讀它們 |

★★★ **手冊自己寫了這件事**，不必靠推論。〈JOURNAL ENTRIES〉開頭那段：

> During the game these entries are referred to by number. … **Do not read ahead
> to other Journal Entries; some tales are false, and may lead your adventurers
> astray.**

⇒「有些條目是假的」是 `exact`。至於**哪幾則**是假的，手冊沒有指名，
所以上表第三、四類的歸屬仍是 `strong inference`；
**第四類只宣稱「corpus 沒有引用」，不宣稱它們是誘餌**。

★ 這句話也是 remake 的設計依據：手札在遊戲內必須**逐則解鎖**，不能一次全開。
一次全開等於強迫玩家讀到假情報，比翻紙本手冊更糟。

### ★ 二之二、第 1 則接在甦醒那一幕

第 1 則的內容就是「舊手札連同全隊裝備一起不見了，所以我開始寫新的」——
指的正是開場甦醒。所以它掛在 `opening.new-game-awakening` 上，分三頁
（`journal.1.1`..`.3`）：南下提爾佛頓找娜卡西亞公主的緣由、路上的伏擊、
醒來後發現手臂上五個青色符號。

⇒ remake 玩家一開場就讀得到，不必回頭翻紙本手冊。

## 三、本輪修正的三則譯文

拿到英文全文之後回頭對照上一輪憑中文校訂稿寫的五則，三則要改：

| 條目 | 原本寫成 | 實際 |
|---|---|---|
| `21` | 「手上仍有足以對付奧提尤格的東西」 | 奧提尤格是**罵對方的話**：`a slime-ridden otyugh like you` |
| `23` | 競技場的參戰規則（一對一或三人共鬥） | 是**賭盤賠率**：押囚犯三比一、押怪物一賠一、無罪組雙邊二比一 |
| `44` | 併入了斯瑞克林與相位蜘蛛的警告 | 那段是**第 45 則**的內容，中文校訂稿把兩則併在一起了 |

⇒ 中文校訂稿是摘要重述，**遇到規則性細節（賠率、人數、費用）要回原文核對**。

## 四、明確不宣稱

- 沒有宣稱剩下 14 則在原版遊戲中完全無法讀到。手冊是紙本，玩家隨時可以翻；
  這裡宣稱的是「ECL corpus 沒有任何一處叫玩家去讀它們」。
- 沒有宣稱 45 則的解鎖時機都正確。producer 找到了，但「第一次進到那個場景才解鎖」
  這件事只有進過該事件的路徑驗過。
- 沒有處理手札的**圖**（第 4、8、39、52、59 條含地圖或插圖）。

## 五、回歸

| 測試 | 釘住什麼 |
|---|---|
| `gamepack.TestEveryJournalEntryHasAProducer` | pack 裡沒有讀不到的死條目 |
| `gamepack.TestJournalProducersUnlockTheirEntries` | 第一批五則的實機字串序列命中並解鎖 |
| `gamepack.TestSecondBatchJournalProducers` | 第二批八則，含三張紙條靠數字分辨 |
| `gamepack.TestOpeningAwakeningUnlocksJournalEntryOne` | 甦醒那一幕帶出手札 1 的三頁 |
