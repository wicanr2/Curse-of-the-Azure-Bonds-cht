# 《青色枷的詛咒》目前工作清單

更新日期：2026-08-20（第 605 輪：怪物的物品鏈接進 `Fighter`——`MON*ITM` 六章 **44 個區塊全部帶物品**（最多一隻 5 件），開戰時依**怪物 ID 的章節命名空間**解析、每一隻各自一份（共用 slice 會讓一隻換裝影響全部）；⚠ 剩規則側的派生值重算；第 604 輪：`SEG-33` 後半完成——PC-98 的 `BGMPLAY` 直接讀 `CURRENTECL`，所以「哪一段該播哪一首」查得出來（spec 355）；game pack 的宣告逐格對回那張表，23 個段界播的都是對的 selector；★ 再加一條「整條主線至少 4 首不同的曲子」——只驗「有一首在曲目表裡」擋不住「每一段都播同一首」；第 603 輪：`SEG-31` 後半完成——**裝備、記憶法術與效果**在 23 個段界逐段存下去讀回來、段界之間也沒有一處變動；★ 閘門**先擋非空再擋變動**，否則攜帶物整批讀不到時會全綠（空的等於空的）；第 602 輪：世界地圖改成**從 13 個地點各掃一次**，走到的分支 17 → **762**，敘述與選項都沒有一條落回原文（又抓到 `BY BOAT`）；★ 閘門同時釘住**分支數下限**——走訪壞掉會變成「語系全綠但什麼都沒驗到」；第 601 輪：`4C9D` 解出來了（spec 1143：出發地 × 4 ＋ 選的方向，一條有向邊的編號，四張表都用它當索引）；世界地圖分支逐條驗，抓到 **3 個沒中文化的選項**（`MEET THEM`／`PASS`／`HIDE`）與 **11 個城市的場所提示**，都補上了——★ **選項要單獨判語系**，敘述中文化了選項還是英文的情況真的發生過；第 600 輪：分派器的配對條件改成「`ON GOTO` 分派的那一格＝`AND` 寫進去的那一格」——原本用「緊接兩條之內」**同時漏掉 5 個真的分派器、又把 3 組不相干的指令湊成一對**；有每格事件的 block 14 → **16**，逐格實測 194 → **250 個索引、235 格中文、落回原文 0 格**；第二種分派形狀（`GETTABLE` 查表）解讀完成，兩個世界地圖 hub 是「路段編號 `4C9D` → 14 個旅途場景」；第 599 輪：`cmd/ecl-cell-refs` 追旗標的寫入端——**事件目錄查 `4C06` 只有 18 處、走訪有 75 處**，拿目錄下「沒有人寫這個旗標」的結論會反過來；spec 1142 換 bank 的時機排除掉 `NEWECL` 與 `LOAD FILES` 兩個候選（各有一個反例），並釐清 `4C07`／`4C0E` 是抵達標記不是被污染的那一類；第 598 輪：`cmd/cell-sweep` 把逐格盤點變成全遊戲一次跑完（13 個 block、**194 個分派索引**逐格站上去，185 格中文、**落回原文 0 格**、9 格沒演出來且每格都印得出守衛），抓到 5 句沒接的原文；朝向刻度 `C04D ＝ 朝向 ÷ 2`——掃 0..3 只蓋一半，南西兩面整批假零；第 597 輪：`cmd/ecl-cell-events` 產出全遊戲「哪一格演哪一場」對照表（14 個 block、234 個場景對到格子），古熔岩洞逐格 11 格；第 596 輪：第六章三區逐格盤點合計 41 格（外城 11 格＋垂死者線索到藏匿處的前置劇情）；第 595 輪：墓園 15 格逐格內容盤點；語系閘門補上 `Prompt` 路徑，又抓到 6 句英文；第 594 輪：內城一樓九間＋二樓五間房逐間走過（九間出得來、五間被全域 `4C00` bank 壓掉，spec 1142）；第 593 輪：階段 3 的跨段不變量——語系（233 句落回原文 0）、隊伍連續性、音樂綁定；第 592 輪：**開場到結局的完整主線跑通了**——同一條 session 走到擊敗提朗瑟克斯的結局選單，23 個段測試；第 591 輪：主線接到第六章——連續 session 從眼魔洞穴經立石群走到密斯卓諾墓園入口，補上立石群灰袍男子的變數插入頁；第 590 輪：`SEG-10`／`SEG-11` 後半／`SEG-13`——整條主線跑拆成 12 個段測試、段界快照往返，並實機推翻「`beholder-cave` 的 `script_block` 指到別段」；第 589 輪：主線 25 段的接線盤點，入口文字沒有一段落回原文，計畫階段 2 的三條印象分類被量出來的數字推翻；第 588 輪：`SEG-12` 47 條 `NEWECL` 邊逐條交接測試；第 587 輪：`SEG-11` 前半——25 段的快照往返閘與 `-segment-snapshot` 交接；第 586 輪：`SEG-04` 統一直入旗標 `-segment`，25 段逐段進得去，順手修掉大圖與世界地圖段的兩個素材鍵錯誤，原作開場序幕第一次跑出來；第 585 輪：`SEG-03` 段落 id 與標籤，25 段逐條有原作敘述為證；第 584 輪：`SEG-02` block ↔ 地圖對照與段落清單；第 583 輪：`SEG-01` ECL block 轉移機制（spec 1141），事件目錄的轉移圖是假零；第 582 輪：乾淨 clone 的 engine bootstrap；第 581 輪：力量調整表與原版存檔 oracle；第 580 輪：七個戰鬥契約搬進共用 engine；第 579 輪：機會攻擊四道閘、隊員命中與 AC 表；第 578 輪：AC 與命中的刻度統一（spec 1139）；第 577 輪：第二個 AC `+19Bh` 的算式與資料驗證；第 573–576 輪：回合開始生命週期、突襲遮罩死碼、面向與背後攻擊、90° 扇形與機會攻擊面向閘；第 572 輪：第一人稱畫面與原版逐格比對 19／20；第 569–571 輪：README 五張圖的圖層對齊、第一人稱牆面第 0 段符號、戰鬥地圖同步；第 568 輪：第 2 批完成、雲格驗證、龍巫妖等級吸取）

## 目前執行順序（使用者 2026-08-16 指定，優先於下方所有舊清單）

### 第 0 批：把上一輪留下的邊界收乾淨

這三項不是新功能，是**讓前一輪的結論站得住**。**2026-08-16 全部完成**（spec 1110）。

| # | 項目 | 結果 |
|---:|---|---|
| 0-1 | **變長指令的長度守衛** | ✅ 新增 `ecl.RecordEnd`（唯一正確的「下一條在哪」）與 `VariableLengthCommands`，三條測試釘住：陷阱仍在（四個 opcode arity 仍是 0）、corpus 裡 **363 筆**變長記錄逐筆算得出結尾且一定大於 `Next`、個數不可知或越界時一律回錯誤 |
| 0-2 | **走訪截斷** | ✅ 分母與比對拆成兩趟。`walkPages` 不帶文字也不帶呼叫堆疊（子程式做成摘要），整份 corpus 都走得完 ⇒ 分母完整；`walkRuns` 碰到上限只會少判。`TestPageWalkCoversEveryPageTheRunWalkFinds` 印出差額，**實測 0 頁** |
| 0-3 | **16 頁 `variable-insert`** | ✅ 逐格判定值的來源：`7B01h` 目的地 14 個、`7B89h` 招牌 7 塊、`7B88h` 方向 4 個都是字串常數 ⇒ 逐項列舉（32 條規則）；`7F79h` 傳聞編號、`7F7Bh` 賭金、`7F82h` 距離、`7C00h` 隊員名無法列舉，寫成 `TestVariableInsertPagesAreWiredAtRuntime` 的 unhandled 清單並附理由。⚠ 列舉不會讓 `variable-insert` 的數字變小——靜態文字裡沒有那個值，這一類只能在執行期驗 |

### 第 1 批：使用者指定的主線（依序）

| 順序 | ID | 內容 |
|---:|---|---|
| 1 | `ENG-01` | 事件內容——**文字層已接完**（spec 1110），剩副作用（見下一列）|
| 2 | `RE-04` | 劇情與全地圖事件的逐格盤點。**副作用的分母已建立**（`docs/audit/ecl-effect-coverage.md`）：可達 14,177 條指令中 `done` 13,097／`partial` 1,057／`consumed` 23。依出現次數排，缺口是 `1Ch CLEARMONSTERS`(206)、`0Eh PICTURE`(199)、`24h COMBAT`(199→RE-06)、`2Dh CALL`(168→RE-03)、`33h PRINT RETURN`(120)、`27h TREASURE`(63)|
| 3 | `RE-06` → `ENG-07` | 戰鬥回合生命週期。**回合開始那一段的逐項對照已完成**（spec 1135）。**逃跑已接**（spec 1112／1113）：走出邊界即脫離，比「我的移動率」與「敵方最快移動率」，平手擲 1d2；新增 `StatusPartyFled`。**突襲量到「沒有人設遮罩」**：36 個 overlay 裡對 `+596h` 只有先攻讀它與一處清 0，常駐段那一側還沒掃（正對照失敗，見 spec 1113）。**突襲遮罩已判定為死碼**（spec 1136），**機會攻擊的設定端就是 GUARD**、三段都已接。✅ 面向與背後攻擊已接（spec 1138）。✅ **90° 扇形判斷讀完並接上**（spec 1002，PC-98 符號名 `INARC`）：離開接觸的機會攻擊現在先過「朝向 −2..＋2 五個方向」的 180° 檢查，先攻未歸零與動作計數 0 兩個旁路照抄；剩 `+19Bh` 的算法 |
| 4 | `RE-07` → `ENG-08` | 怪物 AI。**RE-07 其實大半已解讀**（COMPTACT 16/38，大的全在內）。**移動已接**（spec 1114）：模式 1..6、五個候選方向、半格成本、走到射程才攻擊、20 次上限。**決策已接**（spec 1116）：門檻 7 起／1d7 輪的掃描、施法每輪抽 3 個、用道具全掃、效果碼重映射、`+0F7h` 士氣編碼；分數取自 `ENG-09` 匯出的法術表 `+0Dh`。**目標挑法已接**：射程內候選均勻隨機（838 §五），沒有人在射程內才移動。**障礙已接**（spec 1119）：`far 193Eh+1` ＝ `overlay-32 entry#19`，兩種障礙就是地形碼 `1Eh`（豁免得過就走）與 `1Ch`（`有號(+0E5h) ≥ 7` 就走），兩張效果清單不同；豁免改走原作那一支（天然 1／20）。**士氣崩潰四段表已接**（spec 831）：1d100 分 10%/50%/20%/20%，逃走那一段掛效果 `23h`、`+10h := 1`、士氣補 `0B3h`、清目標。**士氣整條已接**（spec 1122）：檢定在 AI 回合開頭，門檻是 `100 − HP%`，過不了再比移動率決定跑不跑得掉，跑得掉就設 `+14h` 並印「驚慌逃竄」；四段表的觸發點是**效果碼 `23h`**（走 `CALLEFFECT`，所以沒有任何靜態 far call）。**自動換裝已接**（spec 1120）：類別表 `DS:5CF6h` 就是遊戲 image 的 `ITEMS` 成員（執行檔裡是 BSS，開機載入），欄位逐格對得上；評分、遠近取捨（門檻是 `分A > 分B ÷ 2`）與盾牌槽都實作了，順帶把遠程／投擲判斷從「類別 41..47」改成原作的旗標位元。**換裝已接進 AI 回合**（`internal/game/auto_equip.go`）：職業遮罩由職業組出、彈藥槽 ＝ 裝備槽第 11／12 格，換完重投影衍生值但不動生命與位置。✅ **怪物的物品鏈已經進 `Fighter`**（`combat.MonsterItem`，開戰時依怪物 ID 的章節命名空間解析、每一隻各自一份）。⚠ **剩下的是規則那一側**：`autoEquipBeforeAITurn` 目前只走 `partyRoster`，怪物換裝之後的派生值重算（命中／傷害）還沒接 |
| 5 | `ENG-09` | 全法術表。**資料半邊完成**（spec 1111）：100 筆匯出，16 個位元組逐欄有出處。**實作改成資料驅動**（spec 1117）：效果碼／持續時間／豁免類別直接讀表，一支法術不需要一段程式碼——一次接上 11 支（定身、加速、緩慢、隱形…），宣告數 12 → **23**／79。**效果碼語意其實已經解讀**：spec 1005 的分派表 147 個碼裡 141 個標「已解讀」並附規格號，法術表用到的 50 個碼**一個都沒落在未解讀那六個裡**——所以缺的是把語意接進戰鬥規則，不是再反組譯。**覆蓋台帳加上效果碼維度**：未宣告的 54 支拆成「只差宣告 0／碼看不懂 33／傷害類 21」，判讀範圍由 `combat.InterpretedAffectKinds` 定義並由測試逐條對回 `battle.go`。**保護法術參數化**：原本寫死法術編號 6／7 與職業牧師，改成讀 pack 宣告，法師版 16／17 接上（宣告數 23 → **25**／79）。**效果修正表已接**（spec 1123）：`CHECKFX(timing)` 給清單、handler 給數字，兩張表湊起來就是規則。141 個 handler 機械分類（decoded 6／partial 20／inert 12／unread 103），條件分支裡的指令一律不收——防護邪惡的 ＋2 是比過陣營才生效的，照字面收會得到錯規則。已接豁免、士氣、移動三處；判讀範圍 11 → **20**／50 個碼，宣告數 25 → **34**／79。**傷害骰表已接**（spec 1124）：法術分派表 `DS:72A0h` → handler → 兩支擲骰入口。數字取自**收尾的呼叫點**（`sub_F06` 傷害／`entry#22` 治療），不是取自擲骰——燃燒之手根本沒擲骰，治療中傷是 `2d8 ＋ 1`、重傷 `3d8 ＋ 3`，只看擲骰那三支每一支都偏低。骰數也**不一定是立即數**：`entry#9` 對骰數 0 是直接回 0，所以「骰數 0 ＝ 用施法者等級」是錯的讀法；真相是火球 `等級d6`、魔法飛彈 `(等級＋1) div 2` 顆、寒冰錐 `等級d4 ＋ 等級`、電擊觸手 `1d8 ＋ 等級`，這幾支的算式逐支讀完寫在 `internal/combat/spell_formula.go`。`heal_dice`／`damage_dice` 兩個 behavior 讀表施法，宣告數 34 → **38**／79。**旗標守衛**：抗寒／抗火的折半是**有條件**的（看傷害屬性旗標 `6F95h` 的位元），條件也收進表；傷害路徑接上 `CHECKFX(06h)` 並帶旗標。bit 0 火／bit 1 冷／bit 2 電各有兩個獨立證人，補上 spec 573 空著的那一格。**十呎半徑防護**（52／53／69）與**笨拙術**（`1Bh` 與定身家族同一支 handler，先前漏了）也接上。宣告數 **47**／79。**分母修準**：戰鬥可施放 79 支裡有 6 支**職業模型不支援**（德魯伊 3、表裡沒有職業 3），可宣告的分母是 **73**。**「已判讀」用最嚴的定義**：碼要在某個 timing 裡、那個 timing remake 真的有查、它動的那一格 remake 真的有讀（`wiredTimings` 一張表）。**範圍傷害**（一次骰套全場、帶傷害屬性旗標，抗寒真的會減半）與**範圍士氣崩潰**（混亂術走 spec 831 的四段表）都接上。宣告數 12 → **48**／73。**最後 23 支逐支讀完**（spec 1125）：解除魔法的對抗是「相同 50%、高出去每級 ＋5、低下去每級 −2」（兩個方向係數不一樣）；移除詛咒**先**拿掉效果 `24h`，拿掉了就結束，沒拿掉才解開第一件詛咒裝備的**已裝備**旗標（詛咒本身還在）；次元門先解開身邊的勾住標記再傳送；火焰護盾是熱／冷兩種形態各掛兩碼，電腦操作時擲 `1d10` 大於 5 選熱；寒冰錐的**半徑不在屬性表裡**（表裡是 0），是 `(等級＋1) div 2` 最小 1；疾病的 `2Bh`／`2Ch` 是**效果 `22h` 的 handler** 掛的，不是法術掛的，所以規則寫在效果那一側；法師版降咒術原作就只印訊息。**復原術補上兩個先前未定的角色欄位**：`+0E7h` 被吸的級數、`+0E8h` 連帶少的 HP，還一級 ＝ 還 `+0E8h ÷ +0E7h` 點並把最省經驗的職業加一級。**宣告數 73／73，可宣告的分母歸零**；`TestEveryDeclaredCombatSpellIsCastable` 逐支走一次 `BeginCombatCast` ＋ `CombatCast`，擋住「宣告好看但施法噴錯」。覆蓋稽核改掃整個 `internal/game` 並跟著分派表找真正的施法函式——先前只讀 `combat_state.go` 又靠函式名猜 behavior，把 48 支的 sound 讀成沒有（**量測的洞，不是覆蓋的洞**）；修正後 sound 73／73 observed。⚠ **visual 仍是 partial／missing**：runtime 有排視覺事件，但 pack 只宣告了 13 筆視覺資產，其餘法術的原作視覺資料還沒反組譯——那是 RE 的缺口，不是接線的 |
| 6 | `RE-05` → `ENG-10` | 存檔（spec 1115／1118）。**三份記錄逐位元組有台帳**：角色記錄 422 bytes（decoded 296／documented 101／unknown 25，`decoded` 由突變量測驗證）、`.SWG` 63 bytes 與 `.FX` 9 bytes 都蓋滿且 `unknown` 為 0。**兩道存檔完整性的閘**：`Fighter` 每個欄位（反射）＋ 整局存檔的存→State→存往返，前者當場抓到 `StatusPartyFled` 讀不回來，後者守住 `SavePartyFile` 那 20 個位置參數。**原版匯入的編碼分流已接**（spec 1121）：`CHRDAT?.sav`／`.swg` 的版面同時被原版與 remake 自己的槽使用，光看檔案分不出來源，所以分成兩條路（`LoadOriginalSAVGAMSlot`／`ParseOriginalDOSPlayerFiles`／`ParseOriginalItems`），CLI 用 `-savgam-import` 指定且匯入後不寫回原槽。⚠ **英文樣本測不出這件事**（ASCII 兩條路一樣），所以測試用中文樣本 ＋ 一條 ASCII 正對照。⚠ PC-98 `CHARREC`（`1A7h`）不做——決策六：PC-98 只解讀 remake 需要的部分，而 remake 匯入的是 DOS 存檔。**`ITEM*.DAX` 也走同一條線**：`ParseTreasureItemBlocks` 與兩個 `cmd/` 的 ITEM 區塊改走 `ParseOriginalItems`，並加一條掃原始碼的閘（走錯路不會 panic、只會把中文名讀成亂碼）|

條目的完整敘述與依賴見
[`docs/knowledge/coab-remake-todo.md`](docs/knowledge/coab-remake-todo.md)，
本檔只排順序。

### 第 2 批：`ENG-09` 收尾後由使用者指定（2026-08-17）

`ENG-09` 的宣告分母已經歸零，但台帳上還有兩欄不是綠的，加上一個規則有實作、
沒有來源的缺口。三項依序做：

| # | 項目 | 結果 |
|---:|---|---|
| 2-1 | **等級吸取的來源** | ✅ **這個遊戲沒有等級吸取**（spec 1127）。`+0E7h`／`+0E8h` 與職業等級陣列用**位元組層面**（ModRM disp16）掃過整份執行檔：兩個消費者（訓練升級、復原術）、**零個生產者**，職業等級只有兩個 `inc`、一個 `dec` 都沒有。正對照：`cmd/monster-roster` 匯出的 68 種怪物裡沒有屍妖／幽魂／幽靈／吸血鬼／暗影。手冊只寫「Cause disease … saps his **Strength and HP**」——CoAB 真正有的是力量／HP 被吸（效果 `2Bh`／`2Ch`，下限 3／1），不是等級。⚠ 第一版掃描條件寫窄（只掃 ModRM `80h..87h`）會漏掉 `cmp`／`inc`／`dec`，**查無與不存在長得一模一樣** |
| 2-2 | **visual RE** | ✅ 完成（spec 1126）。原作演出只有一個入口 `<overlay-24 entry#24>(槽)`，對同一個槽連放四格；**`槽 = COMSPR 區塊 + 13`**（`overlay-11` 開場的載入迴圈）。這條換算把既有規格的「icon 17h」翻譯得出來（＝ COMSPR `0Ah`／`8Ah`）。整個遊戲只用三個槽：18（共用施法投射物）、19（閃電電弧）、23（魔法命中）。★ **共用投射物在 `CASTSPELL` 裡、分派到 handler 之前就播**，所以每一支法術都有演出——不是七十幾筆各自的資產，而是一筆共用的加少數例外。pack 多一筆 trigger `spell_cast_shared`，台帳 `visual` 欄 6 → **73** |
| 2-3 | **確認 sound 真的修上去** | ✅ `TestEveryDeclaredCombatSpellEmitsSound` 不靠稽核工具，直接施放每一支法術再讀 `ConsumeSoundEvents`。**這條當場抓到兩件事**：雲霧那一類的聲音是由視覺時間軸發的（測試沒開時間軸就會誤判成沒有聲音），以及火球有自己的爆炸聲而不是 `SoundCast`（要求特定音效會把對的判成錯的）|

⚠ 2-3 排在最後不是因為最不重要，是因為它是**檢查 2-1／2-2 之外那次改動的正確性**——
稽核工具與被稽核的程式同時改動時，工具自己報綠不構成證據。

**覆蓋台帳現況：可宣告 73 支全部 `handler`／`visual`／`sound` 三欄 observed。**
剩下的不是接線缺口：13 筆占位玩家取不到、8 支只能紮營、6 支職業模型不支援。

### 盤點（2026-08-17）：目前真正還開著的工作

第 2 批收完之後重新對了一次帳。**下面每個數字都有產生它的工具**，不是印象。

先講已經收斂到 0 的分母——這幾條不用再排工作：

| 分母 | 現況 | 工具 |
|---|---|---|
| ECL 文字 | 控制流可達 1,022 頁，`unmatched` **0**（另 16 頁 `variable-insert`、7 頁 `subroutine` 靜態驗不到）| `cmd/ecl-text-coverage` |
| ECL 副作用 | 可達 14,177 條指令，`consumed` **0** | `cmd/ecl-effect-coverage` |
| 反組譯台帳 | `待解讀` **0**（2,874 支：已解讀 2,137／不阻塞 162／邊界碎片 575）| `cmd/re-ledger` |
| 戰鬥法術 | 可宣告 73 支全部 `handler`／`visual`／`sound` observed | `cmd/combat-spell-coverage-audit` |

還開著的，依「擋不擋得住一次完整通關」排：

| 級 | 工作 | 量到的缺口 | 下一個可驗收成果 |
|---|---|---|---|
| **P0** | **開場→結局的正常主線** | ✅ **通關了**（2026-08-20）：同一條 session 從開場走到擊敗提朗瑟克斯的結局選單，23 個段 subtest；報告見 [`docs/plan/seg-21-ending-report.md`](docs/plan/seg-21-ending-report.md) | 補段內支線（墓園盜墓者、內城各房間、外城下水道口）與 save／reload gate |
| **P0** | **戰鬥回合生命週期**（`RE-06` → `ENG-07`）| ECL `24h COMBAT` 199 處標 `partial`——分派點已接、回合生命週期未接；initiative／held／delayed／guard／quick 的逐項對照仍開著 | 每一項各一條回歸測試 ＋ 一條正常戰鬥路徑 |
| **P1** | **ECL 的 11 個 `partial` opcode** | 1,057 條指令：`CLEARMONSTERS` 206、`PICTURE` 199、`COMBAT` 199、`CALL` 168、`PRINT RETURN` 120、`TREASURE` 63、`DAMAGE` 24、`LOAD PIECES` 23、`LOAD FILES` 22、`APPROACH` 20、`PROGRAM` 13 | 依出現次數往下做；每一個都要 producer→state→consumer 三段齊全 |
| **P1** | **手札 producer** | locale 有 45 則手札，只有 **15 則**接上 ECL producer ⇒ 30 則還沒有觸發來源；手札 59 的地圖 renderer 未做 | 逐則從 ECL producer 接入，不因手冊有內容就提早揭露 |
| **P1** | **存檔剩餘欄位** | 角色記錄 422 bytes 裡 `unknown` 還有 **29**（decoded 294／documented 99）；`MOVEPARTY` 跨遊戲轉移未做 | 突變量測逐段收斂 `unknown`；轉移另開 |
| **P1** | **怪物側的自動換裝** | ✅ 資料接線完成：`MON*ITM` 已進 `Fighter.MonsterItems`（44 個區塊全部帶物品，最多一隻 5 件）。⚠ 剩規則側：怪物換裝後的派生值重算 | 把 `ChooseAutoEquipWeapon` 的結果套到怪物身上，需要武器 → 命中／傷害的重投影 |
| **P1** | **音樂／音效的 cue 綁定** | 戰鬥開始／全滅的語意已接；每個場景與戰鬥 phase 的 cue、播放生命週期未完成 | 先資料綁定 ＋ 可重播，再對 runtime phase |
| **P1** | **UI 與原版 fidelity** | 五張 manifest 圖的**圖層對齊**已收（spec 1130／1131／1132）：人像填滿框、地形與戰鬥員同一條相機路徑、佈陣看地面、戰場圖示是 CPIC、第一人稱補齊第 0 段符號與洋紅透明鍵。**逐張與原版比對仍未做**——本機沒有提爾佛頓第一人稱的原版畫面 | 先取原版 oracle（spec 530 的 `START STING Wooden` 測試模式可繞過翻譯輪），再逐格比牆磚選圖與天空層 |
| **P2** | 三平台打包、README／截圖／推廣片 | 等 P0／P1 收斂 | — |

⚠ 一個順手記下的小瑕疵：`cmd/re-ledger` 的「引用」欄掃 `docs/` 找「同時提到某
overlay 與某個十六進位值」的檔案，而**它自己的輸出就在 `docs/audit/` 底下**
——所以第一次跑完會改變下一次的輸入，要跑兩次才到不動點。工具本身是決定性的
（同樣輸入兩次輸出相同），只是「跑一次就 commit」可能留下非不動點的版本。
影響僅止於提示欄的檔名清單，不動任何狀態判定。

⚠ 三個**已量到但還沒排工作**的小缺口（都不擋主線，理由與代價都寫清楚了）：

| 缺口 | 現況 | 為什麼還沒排 |
|---|---|---|
| **第一人稱牆磚的逐格美術對照** | ✅ 提爾佛頓五格 × 四個朝向與原版逐格比對，19／20 完全相同（spec 1134）。其餘 17 張 `first_person` 地圖沒比過 | 原版 oracle 已建：容器常駐、逐步送鍵、讀畫面文字決定下一鍵。每張地圖約 20 次擷取 ＋ 20 次 remake 截圖 |
| **SPRIT 畫布相對於戰場格的錨點** | 戰場圖示改回 CPIC 之後，SPRIT 只在「沒有 CPIC」時才會走到。二十個 checkpoint 的素材普查是**零筆退回** ⇒ 目前跑不到那條路 | 跑不到的路徑先不投資；等真的有怪物沒有 CPIC 再量 |
| **第一人稱的天空層** | game pack 宣告提爾佛頓 `outdoor_sky_color = 3`（淺青）、`indoor_sky_color = 0`（黑），由地形位元選；目前畫面上方是黑的 | 沒有原版 oracle 就無法判斷「黑」是對是錯。與上面第一項同一個前置 |

⚠ 兩個**不列為工作**的項目，理由已經查證過：

- **等級吸取**：spec 1127 證明這個遊戲沒有（兩個消費者、零個生產者，怪物名單裡也沒有那一族）。復原術一定回「沒有可恢復的等級」是**原作行為**，不是 remake 缺口。
- **逐法術視覺資產**：不存在「七十幾筆待補的視覺資料」，但**要先問一支法術走哪一種演出通道**（spec 1126／1128 共四種）：四格動畫（`entry#24`：共用施法投射物槽 18、閃電電弧 19、魔法命中 23）、持續區域格（兩種雲，寫進地圖格）、產生器（睡眠 twinkle）、八方向投射物（弓箭）。**火球與閃電確實不一樣**——閃電有專屬電弧圖沿線逐格播，火球沒有專屬圖，靠共用投射物飛過去、再對範圍內每個目標依序播魔法傷害命中。**毒雲也確實有演出**，只是走的是持續區域格那條通道，只數 `entry#24` 會把它算成沒有。

### 後續建議順序（2026-08-18）

上面的盤點是「還開著什麼」，這一節是「接下來做哪一個」。判準是
**擋不擋得住一次完整通關**，其次才是代價。

| 建議 | 工作 | 為什麼是這個順序 | 粗估 |
|---:|---|---|---|
| 1 | **原版第一人稱 oracle**（✅ 完成，spec 1134）| 提爾佛頓五格 × 四個朝向逐格比對：19/20 完全相同、1 張差一格。剩下的是把其餘 17 張 `first_person` 地圖的天空色與牆磚也比過 | 每張地圖約 20 次擷取 ＋ 20 次 remake 截圖，各約一分鐘 |
| 2 | **戰鬥回合生命週期**（`RE-06` → `ENG-07`）| ✅ 回合開始那一段收完了（spec 1135／1136）：先攻算式、選誰動、`20 → 19`、DELAY、GUARD、QUICK、定身、機會攻擊全部已接且與原作相符；**突襲遮罩是死碼**（全 build 一讀一清 0，先攻的 −6 走不到），CoAB 傳 0 是唯一有證據的值。✅ **收完**（spec 1137／1138）：面向、動作計數、累計轉向三個欄位接上，開場面向表、攻擊時轉向、機會攻擊的轉向記帳、回合開始只清後兩格都與原作相同；背後攻擊的三道條件是字面移植。**一處刻意留白**：`+09h` 的幾何意義（兩個寫入端互相矛盾）——欄位照字面移植，不解釋 | ✅ **spec 1010 的 180° 扇形檢查已接**（spec 1002 八個方向逐條讀完，兩平台一致；斜向並沒有「多一道條件」，八段是同一個形狀）。✅ **第二個 AC `+19Bh` 解出來了**（spec 1000 §七：`＝ +19Ah − 敏捷防禦調整 − 盾牌槽 − 2`，81 筆 `MON*CHA` 有 78 筆殘差 0、三筆是 ＋1／＋2 盾牌），已接進 `ArmorClassFacing`。✅ **AC 與命中的刻度與原作對齊**（spec 1139：命中式 `d20 ＋ 命中加值 ＋ 目標 AC >= 20`，FIRE KNIFE 對回原作的 18）。✅ **機會攻擊四道閘全部解讀並接上**（spec 1010：`entry#6` 查的四個碼是 `33h 34h 35h 1Fh`、`sub_1144` 就是 `CHECKFX(01h)` 的可見性否決、`4Ah`／`4Bh` 是撤退）。✅ **隊員的命中與 AC 改吃原作的表**（spec 1140：`+73h` 八槽查 `DS:3E38h` 取最大、`+124h ＝ 32h`、敏捷防禦調整走 spec 694 的表）。✅ **力量的命中／傷害調整也接上了**（spec 694／697，三張表在 engine 的 `combat/ability`；`+125h` 是開關，`MON*CHA` 81 筆裡 61 筆是 0）。★ 整條派生鏈用原版存檔逐項驗過：`+73h`／`+199h`／`+1A2h`／`+19Ah`／`+19Bh` 五格全中 |
| 3 | **開場→結局的主線續接** | 唯一真正決定「能不能通關」的工作，但它會**反覆撞上第 2 項**：每接一章就要打一場沒有完整回合生命週期的戰鬥。**驗證方式改成分段**（一個 ECL block 一段、一條 `NEWECL` 邊一段交接），計畫見 [`docs/plan/mainline-segmented-verification.md`](docs/plan/mainline-segmented-verification.md)。✅ `SEG-01` 完成（spec 1141）：主迴圈讀 `LastECL`、`NEWECL` 改它；**25 個 block、47 條出邊**，只有開場與結局沒有出邊；報告見 [`docs/plan/seg-01-verification-report.md`](docs/plan/seg-01-verification-report.md)。✅ `SEG-02` 完成：`LOAD FILES` 的第一個運算元就是那一段載的地圖區塊，game pack 的 `area_id`／`script_block` 與 block 編號 join 得起來，段落清單已產生；報告見 [`docs/plan/seg-02-verification-report.md`](docs/plan/seg-02-verification-report.md)。✅ `SEG-03` 前半完成：段的 id 一律 `ECL{成員}/0x{block}`（不綁地圖名），25 段的標籤逐條有原作敘述為證；報告見 [`docs/plan/seg-03-verification-report.md`](docs/plan/seg-03-verification-report.md)。✅ `SEG-04` 完成：`-segment <id>` ＋ 註冊表 `internal/segment`，25 段逐段一條測試都進得去；報告見 [`docs/plan/seg-04-verification-report.md`](docs/plan/seg-04-verification-report.md)。✅ `SEG-11` 前半完成：25 段的邊界狀態存得下去讀得回來，`-segment-snapshot` ＋ `-party-load` 交接實際跑通；報告見 [`docs/plan/seg-11-verification-report.md`](docs/plan/seg-11-verification-report.md)。✅ `SEG-12` 完成：47 條 `NEWECL` 邊每條一個交接測試，來源段存快照 → 讀回 → 帶著來源 block 當 `LastECL` 進目的段；報告見 [`docs/plan/seg-12-verification-report.md`](docs/plan/seg-12-verification-report.md)。✅ 階段 2 的現況盤點完成：`cmd/segment-coverage` 量出 25 段的入口，**22 段有中文劇情、3 段的入口不出文字（被別段帶進來的）、0 段落回原文**；報告見 [`docs/plan/seg-20-coverage-report.md`](docs/plan/seg-20-coverage-report.md)。剩下的接線工作在**段內**（每回合／搜尋生命週期、離開邊條件、段內戰鬥），✅ `SEG-10`／`SEG-11`／`SEG-13` 完成：那條 790 行的整條跑拆成 **12 個段 subtest**（正規化逐行比對證明一個遊戲動作都沒增刪），每段結束存一份快照並整批驗往返；報告見 [`docs/plan/seg-10-verification-report.md`](docs/plan/seg-10-verification-report.md)。✅ **主線接到第六章**：同一條 session 從開場走到 `ECL6/0x40` 密斯卓諾墓園入口（散提爾堡 → 猶拉什 → 希爾斯法 → 立石群 → 密斯卓諾），段測試 18 段、段界快照 18 份；報告見 [`docs/plan/seg-20-myth-drannor-report.md`](docs/plan/seg-20-myth-drannor-report.md)。✅ **通關了**：同一條 session 從開場走到擊敗提朗瑟克斯的結局選單（38 名戰鬥員的最終戰），23 個段測試守著；報告見 [`docs/plan/seg-21-ending-report.md`](docs/plan/seg-21-ending-report.md)。階段 3 的跨段不變量也接上了：語系（`SEG-32`，整條主線 233 句**落回原文 0 句**，第一次跑抓到 12 句英文並補譯）、隊伍連續性（`SEG-31`）、音樂綁定（`SEG-33`）。內城一樓九間＋二樓五間房也逐間走過了（九間出得來；五間被 spec 1142 的全域 `4C00` bank 壓掉，已釘住成因）。全遊戲「哪一格演哪一場」的對照表已經產得出來（`cmd/ecl-cell-events`：25 個 block 裡 14 個用地形碼分派、**234 個場景對到格子**）。✅ **逐格實測也全遊戲一次跑完了**（`cmd/cell-sweep`，27 秒）：13 個 block、**194 個分派索引**逐個站上去跑生命週期，**185 格演出中文、落回原文 0 格**，9 格沒演出來且每一格都印得出守衛——其中 **7 格指向同一個成因 spec 1142 的 `4C00` map-local bank**；報告見 [`docs/plan/seg-23-cell-sweep-report.md`](docs/plan/seg-23-cell-sweep-report.md)。✅ **spec 1142 的成因收斂了一半**：換 bank 的時機**不是** `NEWECL`（`ECL4/0x20` 前導讀的是第二章寫的 `4C01`）也**不是** `LOAD FILES`（`ECL3/0x10` 在它之前寫 `4C05`／`4C0C`，之後整段都在用），兩個候選各有一個反例；`4C07`／`4C0E` 則根本不是被污染的那一類，是 `ECL6/0x43` 依 `C04B`（隊伍 X 座標）設的抵達標記。剩下兩種可能（真的是 map-local 但時機未知／這一區就是全域旗標而那五間房在原作也演不出來）**靜態讀碼分不出來，要拿原版當 oracle**。✅ **第二種分派形狀（`GETTABLE` 查表）也解讀完了**（報告見 [`docs/plan/seg-24-dispatcher-detection-report.md`](docs/plan/seg-24-dispatcher-detection-report.md)）：兩個世界地圖 hub 是「路段編號 `4C9D` → 14 個旅途場景」，散提爾堡酒館是 `7F7C`。同一輪把分派器的配對條件修對（用目的地運算元，不是相鄰），涵蓋從 14 個 block 變成 16 個，逐格實測 **250 個索引、235 格中文、落回原文 0 格**。✅ **`4C9D` 也解出來了**（spec 1143，報告見 [`docs/plan/seg-25-world-map-localization-report.md`](docs/plan/seg-25-world-map-localization-report.md)）：抵達時 `MULTIPLY 4C9C, 4 → 4C9D` 算基底並把四個鄰居載進 `4C02..4C05`，選好方向後 `ADD 7F79, 4C9D → 4C9D`；四張表分別給旅行天數、走法數、到達地點與旅途場景。同一輪把世界地圖走得到的分支逐條驗過語系。✅ **覆蓋再擴一輪**：`State.ArriveAtWorldLocation` 直接抵達宣告的 14 個地點，從每個地點各掃一次，走到的分支 17 → **762**，敘述與選項**都沒有一條落回原文**。⚠ 仍是下界：迷斯卓諾要第六章進度才走得到，另有 161 條分支推不動。✅ **`SEG-31` 後半也完成了**：裝備／記憶法術／效果在 23 個段界逐段驗過（報告見 [`docs/plan/seg-31-33-invariants-report.md`](docs/plan/seg-31-33-invariants-report.md)）。✅ **`SEG-33` 後半也完成了**：選曲逐格對回 spec 355 的 PC-98 selector 表（報告同上）。**階段 3 的四個跨段不變量全部完成**。剩下的是 spec 1142 那個 oracle 實驗與中後期快照起跑 | 逐段推進，每段一個可重播快照 |
| 4 | **怪物側的自動換裝** | 小而完整：規則已經實作，缺的只是把 `MON*ITM` 接進 `Fighter`。適合插在大批次之間 | 一個資料接線 ＋ 一條回歸測試 |
| 5 | **手札 producer**（30 則）／**ECL 11 個 `partial` opcode**／**存檔剩餘 29 bytes** | 三條都可平行、都不擋主線；依出現次數與玩家可見性往下做 | 各自逐項收斂 |

#### AC 與命中的刻度（2026-08-20 完成，spec 1139）

原作把命中能力與護甲值都存成 `60 − 顯示值`，命中判定是
`d20 ＋ 攻擊者^[199h] ＋ v >= 目標 AC`（`overlay-23:123Fh`，spec 577）——
**儲存值越大越難打**，與 `sub_14E8` 那三段 AC 由難到易的排法一致。

remake 統一用**顯示刻度**（AC 越小越難打、命中加值越大越好），
命中式換算後是 `d20 ＋ 命中加值 ＋ 目標 AC >= 20`——同一個函式，
兩邊各減一個常數。FIRE KNIFE 自己打自己原作要 18，remake 也是 18。

換算點四處：怪物匯入（`CombatArmorClass`／`CombatAttackBonus`）、
`CanHitTarget` 的呼叫端、ECL 隊伍戰力、`CHECKFX` 的護甲記錄寫入。
★ 第二個護甲欄位的換算**由第一格決定**——它系統性地小 2..8，會掉出視窗。

修好之前的症狀（都沒有錯誤訊息）：怪物 AC 1..7 配命中加值 41..53 ⇒ 永遠命中；
敏捷越高的隊員反而越好打；護甲改善與防護法術讓自己更好打；背後攻擊比正面難打。
三條回歸測試各釘一段。

命中式裡的 `v` 是 ECL 位址 `7F70h`（敵方側）／`7F71h`（隊伍側）的當場命中修正
（bank 1 的映射 `(位址 − 7C00h) × 2`），由劇情寫入、remake 接在
`SetSideAttackRollModifier`（spec 390）。

隊員的命中與 AC 也改吃原作的表了（spec 1140）：`+73h` 是八個職業槽查
`DS:3E38h` 取最大值，`+124h` 建角寫 `32h`（AC 10），敏捷的防禦調整與力量的
命中／傷害調整都走 spec 694／697 那三張表（`+125h` 為 0 時力量那兩張不套用）。

★ **整條派生鏈有原版存檔當 oracle**：戰士 5 級、力量 18/18、敏捷 15 的樣本，
`+73h ＝ 44`、`+199h ＝ 45`、`+1A2h ＝ 3`、`+19Ah ＝ 51`、`+19Bh ＝ 48`
五格重算全中。

#### 原版第一人稱 oracle（2026-08-18 完成，spec 1134）

提爾佛頓（`GEO2` 區塊 `01h`）**五格 × 四個朝向、共二十張**畫面與原版逐格比對：
**十九張完全相同，一張差一格**。原版擷取收在
[`docs/reference/original-dos/first-person/`](docs/reference/original-dos/first-person/)，
重跑比對是一行：

```
python3 tools/fp-oracle-compare.py docs/reference/original-dos/first-person/index.tsv
```

修好的是天花板：`gamepack` 的 `tilverton.first-person` 把
`indoor_sky_color` 宣告成 0（黑），原版的 runtime 值是 **9**（`skyPalette[1]` ＝ 白），
`outdoor_sky_color` 3 → **11**。兩個數字取自原版在那張地圖上 `ENCAMP → SAVE`
寫出來的 `Area1`。★ 從**選單**存的那一份兩個都是 0——**同一支存檔常式，
存的時機不同，欄位意義不同**。

驅動方式從「猜一整串定時按鍵」換成**容器常駐、逐步送鍵、每一步讀回畫面文字**
（`tools/dos-oracle-session.sh` ＋ `tools/dos_screen.py`），一步約一秒。
`CURSE.CFG` 是四行純文字（`E`／`S`／`C:\SAVE\`／`F`），留著它開機三題整個跳過。

要站到任意格子：`cmd/dos-save-export -base <原版存檔>` 只覆寫座標與朝向；
`cmd/geo-probe` 先印出哪些格子走得動——開場那格 `(7,13)` 只有一個可走方向
（劇情要求先找出口），在那裡按方向鍵不會動，**那不是按鍵沒送進去**。

還開著的：

- 其餘 `GEO` 地圖沒有逐格比過（本輪只驗提爾佛頓五格）。
- `(1,7)` 面西、原生 `(88,70)` 那一格：原版 EGA 8、remake 0。
  **背景那一層是對的**（`SKY.DAX` `0FCh` 的 `(64,6)` 讀出來就是 8，
  `cmd/picture-probe`），蓋掉它的是牆面 stamp——對應牆格 (列 5, 欄 8) 的
  局部像素 `(0,6)`。還沒查那一格的符號編號與它走哪一段。
- `outdoor_sky_color`／`indoor_sky_color` 只改了提爾佛頓那一張地圖的宣告，
  其餘 17 張 `first_person` 地圖的值還是舊的（多半是 0 ＝ 黑天花板）。
  **正確的收法不是逐張填常數，是找出原作在哪裡寫 `Area1` 的 `1FAh`／`1FCh`**——
  查到之後 remake 自己就會算出來，pack 的宣告可以整批刪掉。已經確定的：
  - **那兩個欄位不是純存檔狀態**：把存檔裡的它們清成 0，載入、進地圖、
    紮營存回來，值會變回 11／9。原版每次進地圖都會重算。
  - **改存檔的欄位到不了別的地圖**：`Current3DMapBlockID` patch 成 3／4 之後
    存回來還是 1（同樣被重算），改 `GameArea` 則會讓載入整個失敗。
    ⇒ 要拿別張地圖的值就得**真的把劇情推到那裡**，沒有捷徑。
  - `GEO` 區塊被 `Parse` 丟掉的兩個標頭位元組是 `00 04`（＝ payload 長度），
    十六個區塊全部一樣，**不是**天空色（`cmd/geo-probe -prefix`）。
  - ECL 的記憶體寫入指令寫不到那兩個欄位：`overlay-07` `0DCB`–`0DD9` 是
    `Area1 指標 + 索引×2 + 6A00h`，選擇子 0 走 Area1、1 走 Area2（`0DEE`）。
    九個「指標算術型」的 Area1 存取站點全部是 `+6A00h` 那一族。
  - ⚠ **「全 overlay 找不到寫入」只能當弱證據**：disp16 的靜態掃描結構上
    抓不到指標間接寫入。下一步要用 **runtime watchpoint**（DOSBox debugger）
    看誰寫那兩個位元組，不是再掃一次靜態。
- 取值的自動化已經做好，缺的只是「怎麼走到那張地圖」：
  `tools/harvest-sky.sh <底檔> <area> <block>` 會載入→進遊戲→紮營存檔→
  把 `1FAh`／`1FCh` 讀回來；提爾佛頓的正對照回 11／9。
  ⚠ 它用 `tools/dos-oracle-session.sh wait`／`keyuntil` 等畫面文字，
  **不要改回固定延遲**——開機時間會漂，漂掉的那一鍵會讓後面每一步錯位。
- 示範模式走的是**示範專用地圖**（地名 `NOWHERE IN THE REALMS`），
  而且那一幕是 **PIC（夜營畫面）不是第一人稱視野**。
  `docs/reference/original-dos/tilverton-first-person-demo.png` 的檔名雙重誤導。

⚠ 這一項先做**不是**因為它最重要，而是因為它便宜且會**改變後面的判斷**：
有了原版畫面，「UI fidelity」才有辦法從「尚未逐張比對」變成一個數字。

⚠ 第 2 與第 3 項的順序可以對調，代價不同：先做 3 會讓主線先走通、但每一章的
戰鬥都要在回合規則補上之後回頭重驗；先做 2 會晚一點看到新章節，但走過的章節
不用重驗。**這是使用者的取捨，不由本檔決定。**

#### 兩份戰鬥地圖與編制帶列起點（spec 1132）

戰鬥地圖是由**地城座標**生出來的。checkpoint 的劇情流程會在開戰之前把隊伍移到
別的格，而那時地圖還是啟動時用舊座標生的——於是**佈陣的地面檢查讀舊地圖、
畫面讀新地圖**。這是 spec 1130 那一類「兩層各自算」的第五個來源，只是換成
兩份地圖。三支戰鬥地形投影改成查詢前先同步（座標沒變只做兩次整數比較）。

編制帶的列從第 0 列改成第 2 列起算：原本兩隊都貼在七列視野的上緣，單人隊伍
會被推到右上角。

順帶做了一次素材缺漏普查：二十個 checkpoint × 六個記錄點（戰鬥員圖示、戰鬥
地形圖層、牆面 piece set 與版面、共用符號段、人物舞台、PIC）**全部零筆**。

#### 第一人稱牆面：第 0 段符號與洋紅透明鍵（spec 1131）

spec 1130 收尾時留了一句「牆磚選圖還沒比」。比下去發現不是選錯圖，是
**該畫的格子有一大半根本沒畫出來**：深度 0 的正面牆 56 格只畫了 7 格。

`Put8x8Symbol` 把符號編號切成五段（`1..45`／`46..115`／`116..185`／`186..255`／
`256..295`，spec 781）。中間三段由 `LOAD PIECES` 供應，**第 0 段先前完全沒載**
——而 `WALLDEF` 用第 0 段的編號鋪滿每個牆位的第一列（天空格）與側牆的斜邊
（透視收邊）。六個 `8X8D*.DAX` 裡只有一個 45 項的區塊（`8X8D1.DAX` 的 `0CBh`），
範圍與內容都對得上。

另一半是 **EGA index 13（洋紅）是透明鍵不是顏色**：`0CBh` 的第 0 項是一整格
洋紅（天空格），斜角磚的洋紅切出三角形輪廓。不遮罩就會在天空與牆角畫出桃紅色塊。

修好之後四個朝向都是完整走廊，西向還看得到拱門；README 的第一人稱截圖因此
改回**正常流程的朝向**，不再需要 `-dungeon-facing` 覆寫。

⚠ 影子 `PieceSet` 的作法讓十個牆位的版面表**只有引擎那一份**——遊戲這一側
不抄第二份表。

#### README 四張圖的四個對齊錯誤（spec 1130，使用者 2026-08-18 指出）

使用者逐張指出問題，量下去是同一類：**兩層各自算座標，然後對不上。**

| 圖 | 症狀 | 真正的原因 |
|---|---|---|
| 旅店 | 黃框裡人像沒佔滿 | sub-image 目的地座標被扣了兩次（Ebiten 的 sub-image 只裁切）|
| 遭遇戰 | 人站在綠色灌木上 | 地形層沒過相機與鏡像 |
| Burial Glen | 人站在牆裡、四隻蜘蛛全不見 | 同上 ＋ 編制格 `7..9` 落在七欄視野外 ＋ 佈陣不看地面 |
| 第一人稱 | 像沒有外框、3D 組成可疑 | 兩者都沒錯（框是點狀線、幾何與 spec 1006 逐項相符）|

★ 「人站在牆裡」其實是**牆被畫到他腳下**：戰鬥員走 `鏡像(相機(地圖格))`，
地形層卻直接拿畫面欄列當地圖座標。

順帶挖出三件事：

- 三支戰鬥地形投影原本裝在 `RunGame` 前一行，而確定性 checkpoint 在那之前就
  開戰——**那幾場戰鬥完全沒有地形**（佈陣不看地面、AI 不算成本、閃電不撞牆反彈）。
- 戰場格上的圖示應該是 CPIC 不是 SPRIT（round-75 早就寫明分工）。CPIC 恰好是
  格子尺寸，SPRIT 的畫布高 73..80 且圖案貼在底部，當格子圖示會落在下方一格半。
- `-encounter` 的荒野座標停在 `(0,0)`，戰場大半沒有地形資料。

新增 `-dungeon-facing`（與 `-dungeon-x`／`-dungeon-y` 同一族），README 的第一人稱
改拍朝北的走廊，有縱深比一片平牆有資訊量。

⚠ `tilverton-first-person-demo.png` **不能**當這個位置的 oracle：位置列印是
`7,13 N`，畫面卻是洋紅天空的戶外林地。**檔名相符不等於場景相符。**

#### 雲格已驗（spec 1128）

用 `cmd/combat-tile-export` 把三個戰鬥地形圖集逐格匯出來看：`RANDCOM` 第 2 格是
**藍白色的雲**（致命毒雲）、第 4 格是**綠色的雲**（惡臭之雲），顏色與法術語意
對得上，game pack 早先宣告的兩格是對的。

⚠ 這裡差點下了錯結論：雲格在 `RANDCOM`，不在 `COMSPR`。拿 COMSPR 的同號區塊去看
會看到一支箭與一顆投射物，然後判定「宣告錯了」。**`source_file` 要跟著一起看，
只比對區塊編號會比錯圖集。**

#### 順帶關掉的一個缺口：障礙格終於有生產者（spec 1128）

追毒雲的演出時發現兩支雲霧 handler 除了兩個常數外一模一樣，而那兩個常數是
**寫進地圖格的地形碼**：惡臭之雲 `1Eh`、致命毒雲 `1Ch`——正是 spec 1119 那兩種
障礙格。spec 1119 當時讀完了「怎麼走過去」卻不知道**是誰鋪的**，現在補上了。

★ 更要緊的是 remake 這一側：`ObstacleTerrainBlocks` 規則完整，但地形碼查詢
**從來沒有人掛上**，所以那段規則一次都沒跑過。兩種雲就是缺掉的生產者，
已接上並加測試（`obstacleBlocks` 沒有外部來源時退回戰鬥自己鋪的雲）。

#### 決策：復原術保留，等級吸取給龍巫妖（使用者 2026-08-17，已完成）

- **復原術照原作實作保留**。同系列後續作品有等級吸取時共用引擎直接可用。
- **等級吸取加給龍巫妖**（spec 1129）。★★ 這是**刻意偏離原作**——CoAB 沒有這回事
  （spec 1127）。三個防止日後混淆的作法：
  1. 宣告放 `gamepack/rules/house-rules.json`，**不進共用 engine 的 schema**；
     house rule 待在它屬於的那一款遊戲裡。
  2. 每一條都寫「為什麼偏離」與「誰決定的」，程式與資料註解一律以 ★★ 標出。
  3. 載體選天生效果碼 `7Eh`——**龍巫妖獨有**（`7Dh` 與 ALIAS／DRAGONBAIT 共用，會誤傷）。
- **資料形狀沿用原作**，所以 `restoreDrainedLevel`（復原術）與訓練升級那條路
  **一行都沒改**就能還回去；測試釘住一來一回歸零。

### 反組譯盤點（第 559 輪起）的殘項

轉向 remake 主線之後仍未關的幾條，不擋第 1 批，可平行：

1. ~~ECL opcode → handler 全表~~ ✅ 第 560 輪完成（`ecl-opcode-dispatch.md`）。
2. **`INTERPET` 內部函式盤點**：29 筆 arity 與 `KnownCommands` 不一致的 handler
   待逐一讀（`ecl-handler-operand-audit.md`）。⚠ **第 0-1 項與這裡同源**——
   spec 1110 找到的四個變長指令就是「arity 表說 0、實際不是 0」的那一類。
3. ~~external `CALL` registry selector 層~~ ✅ 第 561 輪完成；7 個分支主體仍 `待解讀`。
4. **`code = 80h`（packed text）長度規則與 bank 1 計算 routine**——補齊 VM 核心的最後兩條。
5. **未定義區段判定**：DOS 16,044／PC-98 20,319 bytes 逐段判定（`RE-11`）。

每完成一批就更新 `docs/audit/re-function-ledger.json` 並重跑 `cmd/re-ledger`；
台帳裡的 `待解讀` 數字是這個階段唯一的進度指標。


本檔是 compact、交接與每輪開工時的執行順序入口。全遊戲 RE／重建完整度以
[`docs/knowledge/coab-re-coverage-matrix.md`](docs/knowledge/coab-re-coverage-matrix.md)
為單一權威矩陣；詳細反組譯歷史仍在
[`docs/knowledge/golden-box-reverse-engineering-worklist.md`](docs/knowledge/golden-box-reverse-engineering-worklist.md)；
可驗證的歷史與每輪規格見 [`docs/project-status.md`](docs/project-status.md)。
本檔只保留目前有效的工作，不把歷史輪次的舊 blocker 當成現況。

## 目前階段：先封閉知識庫，再擴張 remake

2026-08-12 稽核確認：現有研究足以支援多個真實資料垂直切片，但不足以保證
整作重建。接下來先依完整度矩陣補 R1–R3（原始定位、原版語意、READY 規格），
再進 R4–R5（engine＋JSON、正常玩家驗證）；停止用固定 fixture 或局部 parser
coverage 代替全系統閉合。

第一批工作依序是：

1. `P0-RE-1`：第 558 輪已確認三組 `TREASURE → COMBAT` 由既有第 255／257／258
   輪 transaction 覆蓋，並補 PC-98 IDA、真實 DAX pause／resume 與候選審查台帳；
   下一組改為 ECL2 block `0x02 +04BC..+053A` 的 `COMBAT → text`，再逐步建立全域
   ordered effects／exactly-once 規格。不得重做已標 `covered/exact` 的三組候選。
2. `P0-RE-2`：靜態層已完成 6 DAX／25 block／125 entry／1,355 instruction 的可重生
   清冊；下一步回填動態 edge、條件旗標、座標／terrain、external routine、resume
   與每項 R1–R5，不把 33 個靜態候選冒稱 runtime order。
3. `P0-RE-3`：統一 spec 狀態、IDA 腳本引用與可重生報告；舊逐輪文章只作歷史。
4. 建立 external-call registry 與逐區 `area-event-coverage`，再依矩陣補戰鬥、存檔、
   音訊、畫面與中文內容。

只有不影響玩家可見結果的 compiler/runtime helper 可列為 `不阻塞`；這項收斂不
允許略過 D&D 規則、事件分支、戰鬥、畫面、聲音、存檔或正常路徑所需的 consumer。

第 557 輪權威規格：
[`ECL 全事件靜態清冊與有序副作用稽核`](docs/spec/557-ecl-event-catalog-and-ordered-effects-audit.md)；
第 558 輪：
[`PC-98 ECL TREASURE／COMBAT 邊界`](docs/spec/558-pc98-ecl-treasure-combat-boundary.md)；
第 564 輪：
[`ECL opcode 有序副作用相位`](docs/spec/1104-ecl-opcode-ordered-effect-phases.md)；
生成物與審查台帳在 [`docs/audit/ecl-event-catalog.md`](docs/audit/ecl-event-catalog.md)、
[`docs/audit/ecl-ordered-effect-reviews.json`](docs/audit/ecl-ordered-effect-reviews.json)。

## 一句話結論

重製尚未完成整作通關。現在已經有多條可重播的正常玩家垂直切片，並完成
`SEARCH`／`LOOK`、wall=09 候選橋接、E2、火刀 E1、戰後世界地圖與 save/load
的 engine＋JSON 接線；本輪再完成 25 個 ECL block／125 個 entry 的 parser／控制流
稽核、16 個原始 GEO block 的 game-pack 宣告、ECL 戰鬥開始／隊伍全滅音效意圖，
以及 14 個世界點位的 ECL1 到達／JSON 有向路網基線；第 542 輪把同一新遊戲
session 從火刀首領後接到阿沙本福德城內、立石群與艾森布拉城外；第 543 輪再把
同一 session 接到 Hap 村落、熔岩洞、巫師塔、回洞穴與熔岩池第二次戰鬥。仍缺
完整 ECL side effects／外部 routine、全城市／全房間 coverage、完整結局同
session、完整戰鬥與原機音訊、全量繁中校對、完整存檔相容與三平台發行。

第 548 輪保留 DOS IDA Pro 已證實的 `C04B..C04F` 虛擬地圖 bridge，並更正 A2
事件時序：同一新遊戲 session 由 E1 `(5,7,W)` 走到 `(5,9)` 後，先經原始砲擊
敘事的三次 `PRESS` boundary，才在 ECL4 block `0x22 +061B` 寫入
`C04B/C04C/C04D=13/1/3` 與 `4C06=1`。中立 engine 的 `continue_result` 使 JSON
`set_map_position` 投影 `(13,1,W)`、`wall=08`、`terrain=C0` 時保留同一份死精靈
選單 continuation；它不攜帶 CoAB 劇情。這不是完整洞穴：Dexam、其他傳送／隨機
事件、出口與重訪仍待正常路徑驗證；`0x4C00` 仍是與目前玩家結果無關的 `unknown`，
不列為 remake blocker。

第 549 輪已將 README 五張代表圖逐張重拍與校正：冒險文字／命令列、角色建立的
錯誤分欄、倚天 ASCII／全形標點 fallback、戰鬥 footer 與第一人稱 stage 都有
明確 contract；角色建立現在使用原版單一大面板。這是目前 UI 的
`material-exact/layout-reconstructed` 基線，仍不等於所有畫面、戰鬥演出或整作
fidelity 完成。

第 550 輪把同一新遊戲 session 從洞穴 A2 的死精靈提示延伸至
`EXAMINE REMAINS → PICK UP POUCH`：原始 ECL4 block `0x22` 依序顯示皮袋、
氣體陷阱、標有 Dexam 祭壇與外出路徑的地圖，解鎖遊戲內手札 59，再進入無生怪
`COMBAT`／戰利品服務。選擇離開戰利品後，session 以同一 `(13,1,W)` 地城狀態
續跑，沒有座標注入或推測性搜尋邊。手札文字、陷阱與選項均由 CoAB JSON stable ID
解析；原始手札地圖 bitmap 尚未放入 renderer。舊 `(15,1)` 搜尋邊只保留為
`strong inference`，不再當作 Dexam 或洞穴出口的正常路徑證據。

第 551–552 輪建立低成本模型可安全承接的三個機器稽核。locale audit 證實
目前正式 UI literal key、game-pack `en／zh-TW` 對稱、`message_id` 與 stable
option binding 沒有硬性違約；靜態 orphan 只列資訊，不能冒稱未使用。玩家戰鬥
法術 audit 將 12 個正式 stable spell ID 對照目前 remake handler／visual／sound
callsite，只有 2 筆達到三者皆可觀察，另外 10 筆如實列為 incomplete；這不是
原版規則或逐幀 fidelity 證據。截圖 manifest 鎖定 README 五張 640×480 圖的
SHA-256、生成模式與證據等級，並把正常 `VIEW`、`AREA`、overland 與法術關鍵幀
保留為 planned 缺口。三項 focused Docker gate 均通過；P0 洞穴正常路徑沒有因
audit 而假接，仍須先證明手札 59 後到 Dexam 的原版可走 route。

## 狀態與證據規則

- `已完成（remake contract）`：重製程式、JSON、測試與玩家路徑已閉合；不代表
  原版每個 byte 已逐一證明。
- `exact`：原始 bytes 與 consumer／runtime trace 已閉合。
- `strong inference`：多項證據一致，但仍少一段原版資料流或 runtime oracle。
- `待實作`：目前玩家路徑或產品功能仍未完成。
- `待研究`：只有在要支援該功能或原版 fidelity 時才逆向；不可先把假說寫入
  正式規則。

## 第 540 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| ECL corpus parser／控制流稽核 | 原始 `ECL1..6.DAX` 的 25 個 block／125 個 lifecycle entry 都可由 `EntryPoints` 取得並交給 `TraceGraph`；靜態可達 opcode 都有 command metadata，`0x00..0x40` table 也有 coverage。這不是完整 opcode side effect、外部 routine 或整作通關。 |
| 全原始 GEO block 的 game-pack identity | 16 個原始 `GEO2..6.DAX` block 都有 first-person declaration，且 `script_block`／`geometry_block` 分離；ECL3 `0x12` 共用 GEO3 `0x11` 的幾何也有明確映射。這不是所有地形事件、出口、世界旅行或持久重訪。 |
| 戰鬥開始／隊伍全滅音效 intent | ECL encounter 進入戰鬥排入 `SoundCombat`，`PROGRAM 3` 排入 `SoundCrash`；PC-98 selector 對應留在 adapter，DOS 缺少 14／15 WAV 時安全略過。這不是完整原機音效、混音、時序或全場景音樂。 |

權威規格：[`第五百四十輪 ECL／GEO／戰鬥音效邊界`](docs/spec/540-ecl-map-combat-audio-corpus-closure.md)。

## 第 541 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| ECL 外部 routine 分層 | 已將 `CALL`／`NEWECL`／`PROGRAM`／資源請求／typed side-effect request 與 CoAB address／caller context 分開；共用 engine 保留有序 raw／typed boundary，`0x2E10`、`0xC01E`、`0xB200`、`PROGRAM` 語意不下沉成跨作品事實。完整決策與推論等級見 [`spec 541`](docs/spec/541-ecl-external-routine-engine-boundary.md)。 |
| 全世界點位到達 | `TestRealOverlandArrivalAndRouteGraphCoverage` 由原始 ECL1 arrival entry 執行 `moonsea.overland` 全部 14 個 native location values，並驗證 Area／Location／world state 投影。 |
| 世界旅行路網 | JSON adjacency 的所有 destination 都有宣告，且從 Tilverton 的 directed graph 可達全部 14 點；`arriveAtWorldLocation` 在 ECL1 entry 前後提交 `4C9B`，修正部分抵達後沿用舊城市路由列的 bug。 |

這不是「全地圖事件完成」：所有城市設施、隨機遭遇、區域／地城房間、出口、重訪
旗標與完整主線仍要沿正常輸入逐區驗證；既有城市事件測試與後續 vertical slice
繼續累積，不能用路網可達性替代劇情事件證據。

## 第 542 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| 火刀首領後正常世界出口 | 同一新遊戲 ECL session 完成戰後夢境、`PATROL FOREST` 戰鬥續跑、`JOURNEY ON` 與阿沙本福德抵達；`4C03=0x80` 的前置共享旗標被保留，沒有用 frontend 特判清除。 |
| 阿沙本福德正常城市 handoff | 正常抵達後進城、進河畔酒館、選 `RELAX` 觸發 Tavern Tale 28、按鍵續跑、離開酒館與離城均由 game-pack stable option ID 驅動。 |
| 立石群／艾森布拉正常主線骨架 | 阿沙本福德離城後沿 `THE STANDING STONE`，完成提爾隘口戰鬥、灰袍男子／尋紅線索，再沿 `ESSEMBRA` 到達城外 edge；同一 ECL session 未重播 block 起點。 |
| 固定事件與正常路徑的證據分層 | 長固定整合測試仍涵蓋哈普、熔岩洞、法師塔、希爾斯法、尤拉什、摩安德之坑、散提爾堡等大量事件；第 542 輪規格明確標出它們不能取代一條從新遊戲到結局的正常 session。 |

權威規格：[`第 542 輪正常主線與城市／地城 handoff`](docs/spec/542-normal-campaign-spine-and-city-dungeon-handoff.md)。

## 第 543 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| Hap／熔岩洞／法師塔正常 session | `TestRealNewGameRunsToTheEnding` 從第 542 輪的艾森布拉城外沿正常世界路由進入 Hap，以 `MoveDungeon` 逐格完成民宅、阿卡巴、旅店、伊弗利特與村落出口；經 JSON external exit 進入熔岩洞，完成伏擊／守門戰；到 `(6,15,W)` 進巫師塔，完成庭院、德拉坎德羅斯、黑龍敘事、攻擊巫師與屋頂 CAVES 回程；同一 session 再完成熔岩池 `WAIT→PARLAY_NICE`、重訪 `COMBAT`、15 隻火蜥蜴、防火桶 WHO 與耐熱失敗，並沿正常故事 handoff 經 Cave E1 `(5,7,W)` 到死精靈格 `(13,1,W)`。未直接寫入劇情旗標、未 direct-entry 戰鬥；Dexam／洞穴其他事件尚未納入此 gate。 |
| ECL 故事重繪座標邊界 | `original.geo5.block-33` 由 JSON 宣告 `spawn=(7,15,W)`；engine 以 map anchor 保留地城 live cursor，避免同區塊 ECL redraw 的暫存 `C04B/C04C/C04D` 改寫玩家位置。這是可重用契約，不是 CoAB 座標特判。 |
| 外部出口 presentation selector | 獨立 engine 新增 `ExternalExitDefinition.RoofType` 與 schema；CoAB Hap `(15,5,E)` 宣告 `roof_type=2`，正常邊界使用 `wall_type`／`roof_type` 而不把原始 GEO terrain 假稱 exact。engine 本地提交為 `9cf5fa5`；GitHub push 需待外部目的地審核通過。 |
| 正常路徑與固定夾具分層 | 新增 spec 543，明確把同一 session coverage 與既有固定 Hap／Myth Drannor／`PROGRAM 8` fixture 分開；目前仍不能宣稱全城市、全地城或整作結局。 |

權威規格：[`第五百四十三輪正常主線 Hap／熔岩洞／法師塔 coverage`](docs/spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md)。

## 第 544 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| opaque raw memory engine contract | Golden Box engine 新增 `set_memory`／`Runtime.MemoryWrites`／schema validation；CoAB JSON 宣告 `zhentil-keep.inner-city.route-memory-reset`。`0x4C00` 原版語意仍是 `unknown`，不列為 D&D 規則，也不再為它建立反組譯 blocker。 |
| 散提爾堡至眼魔洞穴死精靈格 | 同一正常新遊戲 session 通過 Olive、暗神殿、Dimswart、hooded woman 與 block `0x22`／GEO4 `0x25` handoff，抵達 Cave E1 `(5,7,W)`，再以 ECL 的 `C04B/C04C/C04D=13/1/3` 位置交易到 `(13,1,W)`；測試沒有 `0x4C00`／`4BF2` 直接注入。 |
| 洞穴驗收邊界 | 依公開攻略只保留 Dexam／傳送／隨機事件作 nearby 導航線索；未把攻略座標直接寫成 GEO edge，洞穴內部房間、戰鬥續跑與出口不宣稱完成。 |
| 角色 saving throw 資料分層 | 五欄 saving throw threshold 已由 character template JSON、engine schema 與 State projection 驗證；不與 `0x4C00` 混為同一語意問題。 |

權威規格：[`第五百四十四輪 raw memory route boundary／Dexam 洞穴入口`](docs/spec/544-opaque-memory-route-boundary-and-cave-entry.md)。

## 第 548 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| DOS `C04B..C04F` virtual-map bridge | IDA overlay-07 原始 bytes 證明 setter 將 `C04B/C04C/C04D` 對應到 `DS:720F/7210/7211`，getter 將 `C04E/C04F` 讀回 `DS:7212/7213`；既有 vector 4／6 與 GEO four-plane evidence 將後兩者接到 wall／terrain cache。這消除玩家路徑所需的 map-register blocker，但不宣稱每個 renderer dirty flag 或 DOS redraw 都已 pixel-exact。 |
| E1／A2 至死精靈選單正常 handoff | 正常新遊戲 session 由 `(5,7,W)` 一般走到 source cell `(5,9)`，先完成 A2 的三次 `PRESS`，再由原始 ECL `+061B` 寫入位置。game-pack 只投影該交易，最終為 `(13,1,W)`、`wall=08`、`terrain=C0`，並以 `continue_result` 保住死精靈選單與 `C04B..C04F`／event exactly-once。這不含完整洞穴、出口、重訪或完整通關。 |
| `0x4C00` 範圍 | 本輪沒有新增 `0x4C00` 逆向或命名；依第 546 輪結果維持 `unknown`，只要不影響玩家可見 D&D 規則／路由／存檔，就不列為 remake blocker。 |

權威規格：[`第五百四十八輪 A2 續跑與地圖 handoff`](docs/spec/548-ecl4-cave-a2-continuation.md)。

## 第 550 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| 死精靈皮袋至手札 59 的正常續跑 | 同一新遊戲 session 在 `(13,1,W)` 選擇 `EXAMINE REMAINS → PICK UP POUCH`，依原始 ECL 先看到氣體陷阱，再取得手札 59。`dexam.dead-elf.gas-trap`、`dexam.dead-elf.map` 與 `journal.59` 均由 game-pack stable ID／locale resolver 驅動，沒有把中文文字塞入 State。 |
| 無生怪 `COMBAT`／戰利品服務 | 原始 ECL 的 `COMBAT` request 在這段沒有 monster spawn；remake 以既有 service boundary 提供兩件 pending item 與 `TREASURE_EXIT`。離開後地城 lifecycle 清空暫存狀態並留在 `(13,1,W)`，正常 session 沒有遺失 ECL continuation。 |
| 手札地圖的資料誠實性 | 遊戲內先提供繁中圖例摘要並指向原始 Adventurer's Journal；原始 PDF bitmap 尚未嵌入 Journal renderer，故不能宣稱遊戲內地圖圖像或出口 route 已完成。 |
| 搜尋邊勘誤 | `(15,1)` 的搜尋邊沒有被當作 normal-path evidence；它維持 `strong inference`，Dexam 固定夾具也繼續明確區分為局部 regression。 |

權威規格：[`第五百五十輪死精靈、手札 59 與戰利品續跑`](docs/spec/550-ecl4-dead-elf-journal59-treasure-continuation.md)。

## 第 539 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| 火刀據點入口→首領戰前正常路徑 | `TestRealNewGameBeginsAtGlobalBlockOne` 從真實開場、下水道、E2 block 4 `(6,1,S)` 出發，以同一個 `MoveDungeon`／ECL session 走 29 步到 `(3,13)`、terrain `0x87` 的首領事件；路徑實際經過 `0x99` 刀刃區、`0x9A` 冰凍房與 `0x94／0x95` 相位蜘蛛區，完成必要的選單／戰鬥／續跑後抵達首領戰前。測試以 stable message ID 與敵方物件資料驗證 20 名火刀＋首領，共 21 名敵人；沒有直接注入座標或直接進入首領。原版對照等級是 `layout／route reconstructed`，不是整張地圖 pixel-exact。 |
| 中文 GUI 與截圖校正 | renderer 依倚天字形實際 glyph advance 做 rune-safe 換行與單行裁切；第 549 輪再修正 frame 合成順序、角色建立原版單一面板、ASCII／全形標點 companion raster、戰鬥 footer 與 README 五張圖。這是 `material-exact/layout-reconstructed`，不是所有狀態逐像素 exact。 |
| 讀檔位置投影 | `LoadPartyFile` 依保存的 `Area.CurrentCity` 重建 `LocationName`／`OriginalLocation`，不再只恢復數字 enum；火刀首領固定 fixture 的 save/load 另驗證原始地點保留。正常完整 session 的戰後世界選單尚未因此宣稱閉合。 |

權威規格：[`第 538 輪火刀入口至首領路徑`](docs/spec/538-fire-knife-normal-leader-route.md)、
[`第 539 輪中文 GUI 寬度與溢框`](docs/spec/539-cjk-gui-width-clipping.md)、
[`第 549 輪角色建立與截圖校正`](docs/spec/549-dos-character-creation-and-screenshot-polish.md)。

## 第 537 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| engine＋JSON 的 `SEARCH`／`LOOK` 分離 | `S` 持續切換 `DungeonSearchEnabled`，`L` 是一次性 `LookDungeonLocation`；地圖 edge、external exit、locale 與 save v12 均有資料契約。這是重製契約已完成；原版 Search 成功率與 wall writer 仍未 exact。 |
| 下水道至 E2 | 從 `(13,10)` 正常逐格移動，經 wall=09 候選橋接抵達 `(8,15,S)`，再由 E2 進 ECL2 block 4；未用直接設定座標完成這條重製路徑。 |
| 火刀 E1 回返 | block 4 北側 E1 候選可由正常移動越界，回到下水道 `(10,15,N)`；三個 E1 座標仍屬 `strong inference`。 |
| 戰後 handoff 與存檔 | 首領勝利後的 ECL 夢境、Tilverton 世界地圖選單與 Search／edge 狀態 save/load 已有固定 fixture 回歸；第 539 輪另外完成入口到首領戰前的正常路徑，但完整 session 在首領戰後仍會出現原始 `PATROL FOREST／JOURNEY ON／CAMP` 分支，尚未與預期的世界地圖出口契約閉合。 |
| 既有資料分層 | 開場選項、玩家法術入口與多個事件已使用 stable ID、locale JSON 與 engine resolver；不能因此宣稱全量中文化。 |
| UI 基線 | 640×480、原版裂紋石框、倚天粗體 16×15、人物 HEAD／BODY 分層與多張對照截圖已建立；目前多為 `layout-reconstructed`，不是所有狀態逐像素 exact。 |

權威規格：[`docs/spec/537-search-look-e2-fire-knife-normal-route.md`](docs/spec/537-search-look-e2-fire-knife-normal-route.md)。

## 剩餘工作總表

### P0：先讓主線繼續走，不再用座標輔助掩蓋缺口

| 工作 | 現況 | 下一個可驗收成果 |
|---|---|---|
| 火刀據點完整正常路徑 | 入口→首領→戰後世界→阿沙本福德→立石群→艾森布拉城外的正常 session 已接通；仍未覆蓋所有可選房間、全部寶物、失敗／重訪分支。 | 以原始 GEO 路徑補齊火刀可選房間與重訪，再把可驗收結果寫入 coverage matrix。 |
| 火刀據點出口、返回世界地圖與重訪 | 正常 session 的 `PATROL FOREST`、`JOURNEY ON`、阿沙本福德抵達與後續城市 handoff 已閉合；Tilverton 固定 fixture 的 save/load 回歸仍保留。 | 將同一正常 session 的存檔／重載與重訪延伸到世界路由，並分離固定夾具與正常主線證據。 |
| 開場到結局的正常玩家主線 | ✅ **通關**：同一條 session 從開場走到擊敗提朗瑟克斯的結局選單（立石群 → 密斯卓諾墓園的紅網與黛米爾公主 → 外城遺跡 → 內城遺跡的儀式與爪牙戰 → 二樓東北角的最終戰，38 名戰鬥員）。拆成 23 個段 subtest，每段結束存快照並驗往返。 | 補段內支線（墓園盜墓者、內城的臥房／辦公室／廚房／犬舍／雕像室／禮拜堂、外城下水道口）與 save／reload gate。 |

### P1：補齊可玩規則、資料與原版行為

| 工作 | 目前缺口 | 驗收方向 |
|---|---|---|
| 全 ECL 與外部 routine | 25 個 block／125 個 entry 的 parser／控制流 corpus gate 已完成；`C04B..C04F` virtual-map adapter 已閉合，但 `CALL`、`NEWECL`、剩餘地圖服務、劇情旗標、NPC 離隊、輸入與 continuation 的完整 consumer 仍未閉合。與玩家結果無關的 raw work address（目前如 `0x4C00`）不列為 blocker。 | 只逆向會改變玩家結果的 producer→state→consumer；每個完成事件都要有 raw bytes／runtime trace、JSON contract、stable ID 測試與正常輸入路徑。 |
| 全地圖與世界旅行 | 16 個原始 GEO block 已在 game-pack 宣告；14 個世界點位的 ECL1 到達、Tilverton→全點 directed adjacency，以及新遊戲→阿沙本福德→立石群→艾森布拉的正常主線已通過 Docker gate。仍缺所有城市／地城房間 coverage、TRAIL／WILDERNESS／EXIT 全分支、隨機遭遇、所有入口出口、持久 map state 與原版 fidelity。 | 建立每座城市／每個 GEO block 的正常事件 coverage matrix，保存 flag／座標／資源 handoff，並補全世界旅行與重訪回歸；不把攻略座標直接寫成規則。 |
| 戰鬥規則、AI、法術效果與動畫 | 法術那一半**已收斂**：可宣告的 73 支全部宣告，`handler`／`visual`／`sound` 三欄 observed（spec 1111／1117／1123／1124／1125／1126）。視覺不是七十幾筆待補資產——原作只有一段共用施法投射物（COMSPR 區塊 5）加閃電電弧（6）與魔法命中（10）。仍缺的是**回合生命週期**（initiative／held／delayed／guard／quick 逐項對照，`RE-06`）、怪物側物品鏈進 `Fighter`，以及弓箭原版時序矩陣。 | 回合生命週期每項一條回歸測試 ＋ 正常戰鬥路徑；影片只能證明演出，數值要回 bytes／DOSBox。 |
| 存檔、角色規則與跨遊戲轉移 | remake save v12 已保存 Search／edge；DOS／PC-98 `SAVGAM`、角色 sidecar、完整 record、年齡／職業／特殊能力、刪除／改名與 `MOVEPARTY` 跨遊戲 transfer 尚未完整 round-trip。 | 先完成版本化 parser／serializer 與 save mutation diff，再以角色檔跨 Gold Box 來源做 stable transfer contract；不能把 `MOVEPARTY` 靜態 helper 直接當秘密門。 |
| 全量繁體中文化與遊戲內手札 | 第 556 輪修正七行裁切：長手札依真實字寬自動分頁，來源 stable ID 與 save 不變；摩安德之坑真實 producer 已接通手札 46。目前 59 則中 31 則有 en／zh-TW stable ID 與事件解鎖，另 28 則尚缺 producer 接線；手札 1 來源仍為 `unknown`。全 ECL／物品／法術／怪物／地名／UI 校對與手札 59 地圖 renderer 仍未完成。 | 以 stable `message_id` 做 coverage／orphan／source-drift audit；逐條從 ECL producer 接入剩餘 28 則，不因手冊存在就提早揭露；手札 59 原圖依來源與版面規格加入 renderer。 |
| 音樂與音效 | YM2203、S98、PC98 sound BIOS、cycle PCM 等 engine 知識與部分合成測試已有；戰鬥開始／隊伍全滅 semantic intent 已接通，但完整 DOS／PC-98 producer、播放生命週期、音效與戰鬥 phase 同步仍未完成。 | 先完成每個場景／戰鬥 cue 的資料綁定與可重播播放，再用 DOS／PC-98 runtime 對照 phase、音量、音效次序；合成器測試不能冒稱硬體 exact。 |
| UI、素材與原版 fidelity | **第一人稱已有逐格數字**：提爾佛頓五格 × 四朝向 19／20 與原版完全相同（spec 1134）。冒險／戰鬥／地圖／對話／頭像的其餘狀態仍未逐張比對；palette cycle、sprite timing、完整戰鬥地形與 PC-98 密度仍需抽樣校準。 | 每張對照標示平台、狀態、save／seed、theme 與 `exact`／`nearby`／`layout-only`；原版 theme 與美化 theme 分開，先完成原版忠實驗收。逐格比對走 `tools/fp-oracle-compare.py`。 |

第 551–552 輪新增的工作分派基線：

- 便宜模型可承接：stable-ID coverage、locale reference/drift、截圖檔案完整性、
  bounded schema／validator、已有 contract 的測試補強與素材索引。
- mentor／高階模型保留：原版 bytes／runtime 的語意升級、洞穴 route、AI 選敵與
  RNG、逐幀時序、音效來源、架構／玩家體驗取捨，以及所有完成聲明與合併。
- 子代理不得自行 commit／push；mentor 覆核 diff、修正整合衝突、跑正式 gate，
  再按重大 milestone 集中提交。

### P2：完成後才做的發行工作

| 工作 | 門檻 |
|---|---|
| Windows／Linux／macOS 打包 | P0 主線、P1 規則／資料／音訊與存檔通過後，才做三平台可重現 build、資產授權檢查、存檔位置與首次啟動 smoke。 |
| README／截圖／40–60 秒推廣片 | 截圖只使用目前版本可重播狀態；推廣片只在可玩整合完成後製作。8 小時錄影不是本專案目標。 |
| 日後美化 theme 與 donate | 原版忠實 theme 永久保留；美化與 donate 只作後續／local 設定，donate 資訊不得上傳 GitHub。 |

## 仍需逆向，但不應阻塞目前 remake 路徑的項目

這些是原版 parity 或跨作品知識庫工作，不是重新打開第 537 輪已接通的路徑：

1. `wall=09` 第三平面 before／after writer、Search 成功率、原版 E1 精確座標與
   同版重訪 trace。現有 CoAB edge 是 `strong inference`，不要擴大成所有
   `wall=09` 都可走。
2. DOS `CALL 2E10h` 的剩餘 redraw／dirty-flag 時序與 runtime pixel trace。第 547
   輪已關閉 `C04B..C04F` 對 `DS:720F..7213` 的 adapter；這項只屬原版 fidelity
   稽核，除非某個玩家結果被阻塞，不再逐行追無關 overlay。
3. PC-98 `MOVEPARTY` 的角色轉移 selector／record／save round-trip。中文說明書
   已證明產品功能邊界，但尚未證明每個 raw helper 與 transfer record 的一對一
   runtime 對應。

## 明確不做的事情

- 不為了湊「完整反組譯」而逐行解讀與玩家結果無關的 function。
- 不把 `BDF1`、`SEARCHREC`、`MOVEPARTY`、相同十六進位數字或單一 xref 重新命名成
  秘密門、detail、年齡、旗標或地圖 owner。
- 不以 direct-entry、固定座標、注入戰鬥、測試模式或窄測試宣稱完整通關。
- 不在 JSON 尚未成為真相來源前，把劇情文字、裝備、法術或測試期望值硬編碼回 Go。
- 不在遊戲完整可玩前花時間做三平台 release 或長篇推廣影片。

## 完成聲明的共同驗收門檻

至少要通過：

1. 新隊伍／正式角色能由開場以正常輸入走到結局，包含移動、互動、裝備／使用、
   戰鬥、治療／休息、存檔、退出、重載與一個後期任務重訪。
2. 所有已宣稱支援的內容由 CoAB JSON／locale 與 engine contract 驅動；未支援
   行為明確失敗即關閉，不以 fallback 假裝完成。
3. 原版／remake 的畫面、動畫、音樂、音效、規則與存檔比較都有證據等級；近似畫面
   不標成 pixel-perfect。
4. Docker 內完成受影響套件、代表性正常玩家路徑、save round-trip、截圖／包裝 smoke；
   再集中 commit＋push 兩個 repository。

下一個最小可重現工作：沿第 543 輪同一個 ECL session 從防火桶返回的熔岩洞續接
尤拉什、摩安德之坑、散提爾堡與 Myth Drannor；同時以手札 59 的房間鄰接圖、
GEO4/25 cell graph 與 ECL4/22 producer／consumer 對位，閉合 `(13,1,W)` 到
德克薩姆及 `(6,3)` 出口的唯一正常 route；只有靜態證據仍有多解時才用修改存檔
做受控 DOSBox 抽樣。之後建立
城市／GEO 事件 coverage matrix，逐項標示 normal、fixed fixture 或
coordinate-assisted。不要把本輪
Hap／巫師塔局部路徑或既有 `PROGRAM 8` fixture 擴大解讀成完整結局。
不要把 static corpus／路網 gate 擴大解讀成完整 ECL，也不要先深挖與玩家結果無關的
反組譯。
