# 《青色枷的詛咒》目前工作清單

更新日期：2026-08-25（第 714 輪：**Go 漢字 literal 債歸零**——1,938 次全部移進 `internal/tooltext`（stable ID＋內嵌 catalog），報表輸出逐位元組不變，閘門維持 fail-closed。第 713 輪：**戰鬥佈陣接進 `StartCombat`**＋**ECL `partial` 歸零**（`2Dh CALL` 轉 `done`，可達 14,177 條全部 `done`）——ECL 遭遇的正常開戰改走 spec 1200 的 COMPREP 佈署演算法（地圖朝向＋遭遇距離＋地面檢查＋地城牆面），產原生戰鬥地圖座標並切 reference 模式；帶預置座標的路徑維持原樣。下水道巨魔房實測距離 0、佈出來的形狀與原版擷取定性一致；全套測試綠。第 712 輪：**入口落點偏差修掉**——`ViewMirror.Block` 擋板放寬成「本次執行跑過的 block」，下水道／火刀據點落點改為原版判定的 (0,0)／(8,0)，猶拉什入口收尾朝向（朝東）一併接上；全套測試綠。第 711 輪：**oracle 批次四項全部收掉**——(1) 第一人稱 fidelity 覆蓋收滿：585 種牆面配置全部有原版畫面可比（補拍 24 張全部逐格相同、差 0 格）；(2) 下水道／火刀據點入口落點判掉：原版落腳本寫的第 0 列，remake 的 (0,1)／(8,1) 是偏差（spec 1184，修正待做）；(3) `+1A2h` 大型加值原版實測**不漂**（spec 1175）；(4) SPRIT 錨點判定無原版對應物、corpus 內走不到。第 710 輪：**按鍵重放的世界地圖死環拆掉**（spec 1201）。「卡住換下一項」夾在最後一項會被紮營「修改」×「改名」兩個選單互踢鎖死（12,000 幀有 11,000 幀在環裡）；改成環繞後同路線同幀數 137 → **314 格**、5 → **8 段**（重返巫師塔），無路線預設 94 → **309 格**，落回原文 0。下一個瓶頸換位置：`ECL2/0x03` 爬牆事件是盜賊技能的閘，六個預設戰士永遠失敗。第 709 輪：戰鬥佈陣演算法解讀完並轉錄（spec 1200；第 713 輪接線完成）。第 708 輪：段內支線接進主線，段 subtest 23 → 25。）

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
| 2 | `RE-04` | 劇情與全地圖事件的逐格盤點。**副作用的分母已建立**（`docs/audit/ecl-effect-coverage.md`）：可達 14,177 條指令中 `done` 13,153／`partial` 1,024／`consumed` 0。依出現次數排，✅ **`1Ch CLEARMONSTERS`(206) 逐條讀完了**（spec 1145：它連同還沒領走的戰利品一起丟掉，remake 已跟上；`partial` 只剩 `7603h := 8` 的語意）。其餘缺口是 `0Eh PICTURE`(199)、`24h COMBAT`(199→RE-06)、`2Dh CALL`(168→RE-03)、`33h PRINT RETURN`(120)、`27h TREASURE`(63)|
| 3 | `RE-06` → `ENG-07` | 戰鬥回合生命週期。**回合開始那一段的逐項對照已完成**（spec 1135）。**逃跑已接**（spec 1112／1113）：走出邊界即脫離，比「我的移動率」與「敵方最快移動率」，平手擲 1d2；新增 `StatusPartyFled`。**突襲量到「沒有人設遮罩」**：36 個 overlay 裡對 `+596h` 只有先攻讀它與一處清 0，常駐段那一側還沒掃（正對照失敗，見 spec 1113）。**突襲遮罩已判定為死碼**（spec 1136），**機會攻擊的設定端就是 GUARD**、三段都已接。✅ 面向與背後攻擊已接（spec 1138）。✅ **90° 扇形判斷讀完並接上**（spec 1002，PC-98 符號名 `INARC`）：離開接觸的機會攻擊現在先過「朝向 −2..＋2 五個方向」的 180° 檢查，先攻未歸零與動作計數 0 兩個旁路照抄；剩 `+19Bh` 的算法 |
| 4 | `RE-07` → `ENG-08` | 怪物 AI。**RE-07 其實大半已解讀**（COMPTACT 16/38，大的全在內）。**移動已接**（spec 1114）：模式 1..6、五個候選方向、半格成本、走到射程才攻擊、20 次上限。**決策已接**（spec 1116）：門檻 7 起／1d7 輪的掃描、施法每輪抽 3 個、用道具全掃、效果碼重映射、`+0F7h` 士氣編碼；分數取自 `ENG-09` 匯出的法術表 `+0Dh`。**目標挑法已接**：射程內候選均勻隨機（838 §五），沒有人在射程內才移動。**障礙已接**（spec 1119）：`far 193Eh+1` ＝ `overlay-32 entry#19`，兩種障礙就是地形碼 `1Eh`（豁免得過就走）與 `1Ch`（`有號(+0E5h) ≥ 7` 就走），兩張效果清單不同；豁免改走原作那一支（天然 1／20）。**士氣崩潰四段表已接**（spec 831）：1d100 分 10%/50%/20%/20%，逃走那一段掛效果 `23h`、`+10h := 1`、士氣補 `0B3h`、清目標。**士氣整條已接**（spec 1122）：檢定在 AI 回合開頭，門檻是 `100 − HP%`，過不了再比移動率決定跑不跑得掉，跑得掉就設 `+14h` 並印「驚慌逃竄」；四段表的觸發點是**效果碼 `23h`**（走 `CALLEFFECT`，所以沒有任何靜態 far call）。**自動換裝已接**（spec 1120）：類別表 `DS:5CF6h` 就是遊戲 image 的 `ITEMS` 成員（執行檔裡是 BSS，開機載入），欄位逐格對得上；評分、遠近取捨（門檻是 `分A > 分B ÷ 2`）與盾牌槽都實作了，順帶把遠程／投擲判斷從「類別 41..47」改成原作的旗標位元。**換裝已接進 AI 回合**（`internal/game/auto_equip.go`）：職業遮罩由職業組出、彈藥槽 ＝ 裝備槽第 11／12 格，換完重投影衍生值但不動生命與位置。✅ **怪物的物品鏈已經進 `Fighter`**（`combat.MonsterItem`，開戰時依怪物 ID 的章節命名空間解析、每一隻各自一份）。⚠ **剩下的是規則那一側**：`autoEquipBeforeAITurn` 目前只走 `partyRoster`，怪物換裝之後的派生值重算（命中／傷害）還沒接 |
| 5 | `ENG-09` | 全法術表。**資料半邊完成**（spec 1111）：100 筆匯出，16 個位元組逐欄有出處。**實作改成資料驅動**（spec 1117）：效果碼／持續時間／豁免類別直接讀表，一支法術不需要一段程式碼——一次接上 11 支（定身、加速、緩慢、隱形…），宣告數 12 → **23**／79。**效果碼語意其實已經解讀**：spec 1005 的分派表 147 個碼裡 141 個標「已解讀」並附規格號，法術表用到的 50 個碼**一個都沒落在未解讀那六個裡**——所以缺的是把語意接進戰鬥規則，不是再反組譯。**覆蓋台帳加上效果碼維度**：未宣告的 54 支拆成「只差宣告 0／碼看不懂 33／傷害類 21」，判讀範圍由 `combat.InterpretedAffectKinds` 定義並由測試逐條對回 `battle.go`。**保護法術參數化**：原本寫死法術編號 6／7 與職業牧師，改成讀 pack 宣告，法師版 16／17 接上（宣告數 23 → **25**／79）。**效果修正表已接**（spec 1123）：`CHECKFX(timing)` 給清單、handler 給數字，兩張表湊起來就是規則。141 個 handler 機械分類（decoded 6／partial 20／inert 12／unread 103），條件分支裡的指令一律不收——防護邪惡的 ＋2 是比過陣營才生效的，照字面收會得到錯規則。已接豁免、士氣、移動三處；判讀範圍 11 → **20**／50 個碼，宣告數 25 → **34**／79。**傷害骰表已接**（spec 1124）：法術分派表 `DS:72A0h` → handler → 兩支擲骰入口。數字取自**收尾的呼叫點**（`sub_F06` 傷害／`entry#22` 治療），不是取自擲骰——燃燒之手根本沒擲骰，治療中傷是 `2d8 ＋ 1`、重傷 `3d8 ＋ 3`，只看擲骰那三支每一支都偏低。骰數也**不一定是立即數**：`entry#9` 對骰數 0 是直接回 0，所以「骰數 0 ＝ 用施法者等級」是錯的讀法；真相是火球 `等級d6`、魔法飛彈 `(等級＋1) div 2` 顆、寒冰錐 `等級d4 ＋ 等級`、電擊觸手 `1d8 ＋ 等級`，這幾支的算式逐支讀完寫在 `internal/combat/spell_formula.go`。`heal_dice`／`damage_dice` 兩個 behavior 讀表施法，宣告數 34 → **38**／79。**旗標守衛**：抗寒／抗火的折半是**有條件**的（看傷害屬性旗標 `6F95h` 的位元），條件也收進表；傷害路徑接上 `CHECKFX(06h)` 並帶旗標。bit 0 火／bit 1 冷／bit 2 電各有兩個獨立證人，補上 spec 573 空著的那一格。**十呎半徑防護**（52／53／69）與**笨拙術**（`1Bh` 與定身家族同一支 handler，先前漏了）也接上。宣告數 **47**／79。**分母修準**：戰鬥可施放 79 支裡有 6 支**職業模型不支援**（德魯伊 3、表裡沒有職業 3），可宣告的分母是 **73**。**「已判讀」用最嚴的定義**：碼要在某個 timing 裡、那個 timing remake 真的有查、它動的那一格 remake 真的有讀（`wiredTimings` 一張表）。**範圍傷害**（一次骰套全場、帶傷害屬性旗標，抗寒真的會減半）與**範圍士氣崩潰**（混亂術走 spec 831 的四段表）都接上。宣告數 12 → **48**／73。**最後 23 支逐支讀完**（spec 1125）：解除魔法的對抗是「相同 50%、高出去每級 ＋5、低下去每級 −2」（兩個方向係數不一樣）；移除詛咒**先**拿掉效果 `24h`，拿掉了就結束，沒拿掉才解開第一件詛咒裝備的**已裝備**旗標（詛咒本身還在）；次元門先解開身邊的勾住標記再傳送；火焰護盾是熱／冷兩種形態各掛兩碼，電腦操作時擲 `1d10` 大於 5 選熱；寒冰錐的**半徑不在屬性表裡**（表裡是 0），是 `(等級＋1) div 2` 最小 1；疾病的 `2Bh`／`2Ch` 是**效果 `22h` 的 handler** 掛的，不是法術掛的，所以規則寫在效果那一側；法師版降咒術原作就只印訊息。**復原術補上兩個先前未定的角色欄位**：`+0E7h` 被吸的級數、`+0E8h` 連帶少的 HP，還一級 ＝ 還 `+0E8h ÷ +0E7h` 點並把最省經驗的職業加一級。**宣告數 73／73，可宣告的分母歸零**；`TestEveryDeclaredCombatSpellIsCastable` 逐支走一次 `BeginCombatCast` ＋ `CombatCast`，擋住「宣告好看但施法噴錯」。覆蓋稽核改掃整個 `internal/game` 並跟著分派表找真正的施法函式——先前只讀 `combat_state.go` 又靠函式名猜 behavior，把 48 支的 sound 讀成沒有（**量測的洞，不是覆蓋的洞**）；修正後 sound 73／73 observed。⚠ **visual 仍是 partial／missing**：runtime 有排視覺事件，但 pack 只宣告了 13 筆視覺資產，其餘法術的原作視覺資料還沒反組譯——那是 RE 的缺口，不是接線的 |
| 6 | `RE-05` → `ENG-10` | 存檔（spec 1115／1118）。**三份記錄逐位元組有台帳**：角色記錄 422 bytes（decoded 299／documented 123／unknown 0，`decoded` 由突變量測驗證；欄位名由 PC-98 除錯符號展開，spec 1164）、`.SWG` 63 bytes 與 `.FX` 9 bytes 都蓋滿且 `unknown` 為 0。**兩道存檔完整性的閘**：`Fighter` 每個欄位（反射）＋ 整局存檔的存→State→存往返，前者當場抓到 `StatusPartyFled` 讀不回來，後者守住 `SavePartyFile` 那 20 個位置參數。**原版匯入的編碼分流已接**（spec 1121）：`CHRDAT?.sav`／`.swg` 的版面同時被原版與 remake 自己的槽使用，光看檔案分不出來源，所以分成兩條路（`LoadOriginalSAVGAMSlot`／`ParseOriginalDOSPlayerFiles`／`ParseOriginalItems`），CLI 用 `-savgam-import` 指定且匯入後不寫回原槽。⚠ **英文樣本測不出這件事**（ASCII 兩條路一樣），所以測試用中文樣本 ＋ 一條 ASCII 正對照。⚠ PC-98 `CHARREC`（`1A7h`）不做——決策六：PC-98 只解讀 remake 需要的部分，而 remake 匯入的是 DOS 存檔。**`ITEM*.DAX` 也走同一條線**：`ParseTreasureItemBlocks` 與兩個 `cmd/` 的 ITEM 區塊改走 `ParseOriginalItems`，並加一條掃原始碼的閘（走錯路不會 panic、只會把中文名讀成亂碼）|

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

### 盤點（2026-08-24，第 704 輪）：還開著的全部在這裡

**每一列的數字都有產生它的工具**，看 [`docs/audit/remake-status.md`](docs/audit/remake-status.md)
（32 列，29 列有數字）。下面只列**不是 0** 的，依「擋不擋得住一次完整通關」排。

| 級 | 還開著的 | 量到的缺口 | 缺什麼才收得掉 |
|---|---|---|---|
| ✅ 收掉（第 713 輪）| ECL 的 `partial` opcode | **0**——可達 14,177 條全部 `done`（`2Dh CALL` 第 713 輪轉 `done`：四個走得到的選擇子全接上且有回歸測試，`ViewMirror.Block` 是文件化的刻意邊界不是未解項）| 四張宣告 spawn **第 707 輪判掉並刪除**（拆宣告整套測試不變綠 ⇒ 宣告值是腳本執行期落點的抄寫，spec 1184 增補）。`CALL` 剩 `ViewMirror.Block` 這一格：兩個入口第 711 輪用原版**判掉並修掉了**——原版落腳本寫的第 0 列（下水道 (0,0)、火刀據點 (8,0)），擋板放寬成「本次執行跑過的 block」（`SessionRanBlockIDs`），(0,1)／(8,1) 偏差消失、猶拉什收尾朝向一併接上；那一格**判定留著**，只剩「更早執行的殘留舊座標不可蓋掉三張 `area-default` 載檔補值」這一件事（spec 1184／1172，全套測試綠）|
| ✅ 收掉（第 710 輪）| 第一人稱 fidelity | 585 種牆面配置**已比過 585**（最後 6 種補拍 24 張，逐格相同 24／24、差 0 格）| — |
| ✅ 收掉（第 711 輪）| 怪物側的自動換裝 | `+1A2h` 用原版量掉了：20 輪連打大型目標，第 6、7 擊仍打出滿值 8 ⇒ **原作不漂**，remake 的只讀實作行為一致（spec 1175）| — |
| ✅ 判定收掉（第 711 輪）| UI 與原版 fidelity（非第一人稱）| SPRIT 錨點**沒有可量的原版對應物**：原作不把 SPRIT 畫上戰場格（round-75 分工：戰場圖示一律 CPIC），remake 的 SPRIT-on-battlefield 只是「沒有 CPIC」的 fallback，而素材普查零筆退回（20 個 checkpoint）⇒ corpus 內走不到 | 若未來出現沒有 CPIC 的怪物再回來（維持第 186 行的判定）|
| ✅ 收掉（第 714 輪）| Go 漢字 literal 債 | **0**（1,938 → 0：71 檔全部移進 `internal/tooltext` 內嵌 catalog，printf 家族改走 `tooltext.Format`／`Errorf`，報表輸出逐位元組不變；閘門維持 fail-closed）| — |
| **P2** | 三平台發行 | 沒有數字 | 專案明文決定：遊戲完整可玩之前不做 |
| 暫緩 | 按鍵重放 | 12,000 幀 137 格，第 1,437 幀之後停在世界地圖 | 見上一節；使用者 2026-08-24 指示先 pending |

**已判定、數字不會再降的**（留著是為了讓報表看得出來，不是待辦）：

| 項目 | 數字 | 為什麼不動 |
|---|---|---|
| 每格事件「沒演出來」 | 12 | 四種都是盤點的限制（守衛比移動前快照、旗標被前導覆蓋、守衛跨過處理常式、處理常式本來就不講話）；**還沒歸類 0** |
| 走訪可達性未達成 | 1 | 幾何上斷開，而逐格實測站上去是 `played` |
| 段交接沒有段宣告成入口 | 6 | 一段可以有好幾個入口，`EnterFrom` 只挑最順著劇情的那一個 |
| 音效對照 | 1 ＋ 1 | `SOUNDHALT` 是 remake 架構不需要、`CRASHFX` 是平台差異 |
| 分段驗收的交接 | 39 | 補回交接狀態之後剩下的，**全部歸得了因（歸不了因 0）** |

**收斂到 0、不用再排工作的分母**：ECL 文字 `unmatched` 0／run `orphan` 0、
ECL 副作用 `consumed` 0、反組譯台帳待解讀 0、角色記錄 `unknown` 0、
譯名不一致 0、逐格實測落回原文 0、戰鬥法術 73／73、充能物品效果沒接 0、
怪物特殊能力真的缺 0、換曲點 13／13、前端沒有路可以到的 `State` 進入點 0、
主線快照落回原文 0、主線快照按鍵推得動 114／114（推到一半回錯的相異成因上限釘 0）。


### 按鍵重放（第 710 輪重新開工，使用者 2026-08-25 指定順序）

世界地圖那一層的選單死環已拆（spec 1201）：「卡住換下一項」夾在最後一項會被
紮營的「修改」與「改名」兩個選單互踢鎖死，改成環繞後每一項（含出口）都週期性
被輪到。同一份路線、同樣 12,000 幀：**137 → 314 格、5 → 8 段（重返巫師塔）**；
無路線的套件預設同時 94 → 309 格。落回原文維持 0。

- **跑法**：
  ```bash
  COAB_DECISION_LOG=/src/workplace/campaign-frames/route-clean-709.json \
    ./tools/go.sh test ./internal/game/ -count=1          # 錄路線（整包，24,367 步）
  COAB_ROUTE_JSON=workplace/campaign-frames/route-clean-709.json COAB_KEY_FRAMES=12000 \
    ./tools/go.sh test ./cmd/azure-bonds-game/ \
      -run '^TestKeysDriveARealSessionFromTheTitle$' -count=1 -v
  ```
  ⚠ 路線檔要**跟程式碼同一版**（下水道那條路一改，舊路線的症狀是「走得比較短」
  不是報錯）。
- **下一個瓶頸**：第 9,984 幀之後停在 `ECL2/0x03` 的爬牆事件——六個預設戰士
  沒有盜賊，「牆面太滑…還有人要試嗎？」永遠失敗。兩個候選修法（見 spec 1201）：
  重放的預設隊伍帶一名盜賊，或把「同一格同一事件連續 N 次沒進展」列為放手條件。

### 盤點（2026-08-17）：目前真正還開著的工作

第 2 批收完之後重新對了一次帳。**下面每個數字都有產生它的工具**，不是印象。

先講已經收斂到 0 的分母——這幾條不用再排工作：

| 分母 | 現況 | 工具 |
|---|---|---|
| ECL 文字 | 控制流可達 1,022 頁，`unmatched` **0**（另 16 頁 `variable-insert`、7 頁 `subroutine` 靜態驗不到）| `cmd/ecl-text-coverage` |
| ECL 副作用 | 可達 14,177 條指令，`done` 13,153／`partial` 1,024／`consumed` **0** | `cmd/ecl-effect-coverage` |
| 反組譯台帳 | `待解讀` **0**（2,874 支：已解讀 2,137／不阻塞 162／邊界碎片 575）| `cmd/re-ledger` |
| 戰鬥法術 | 可宣告 73 支全部 `handler`／`visual`／`sound` observed | `cmd/combat-spell-coverage-audit` |

還開著的，依「擋不擋得住一次完整通關」排：

| 級 | 工作 | 量到的缺口 | 下一個可驗收成果 |
|---|---|---|---|
| **P0** | **開場→結局的正常主線** | ✅ **通關了**（2026-08-20）：同一條 session 從開場走到擊敗提朗瑟克斯的結局選單，23 個段 subtest；報告見 [`docs/plan/seg-21-ending-report.md`](docs/plan/seg-21-ending-report.md) | ✅ **save／reload gate 完成**（2026-08-21）：23 份段界快照逐份讀回來**真的走一步**（`stepRestoredCampaignState`），不是只比欄位。抓到真 bug——`eventReturnMode` 沒有進存檔，**存在事件畫面上的檔讀回來按下一步就噴 `event has no continuation`**，而欄位全對；由載入時已存下來的 `Area.InDungeon` 推導回去，舊存檔一併修好。✅ **段內支線也收了**（2026-08-21）：23 份段界快照上逐格站上去，9 段、**123 格演得出來、落回原文 0 格**（報告見 [`docs/plan/seg-26-side-branch-report.md`](docs/plan/seg-26-side-branch-report.md)）。⚠ 這一支量的是**另一次取樣**（once-only 已被主線設過），「沒演出來」不是缺口訊號。同一輪抓到 8 句玩家會看到的英文旁白（`29h ENCOUNTER MENU`，`cmd/ecl-encounter-text`）；✅ **`29h` 的距離挑句也接上了**（spec 1144）：16／20 處原本演錯句子，改對之後再補 14 句遠距旁白的譯文 |
| **P0** | **戰鬥回合生命週期**（`RE-06` → `ENG-07`）| ✅ **逐項對照收完了**（spec 1135／1136／1137／1138）：先攻算式、選誰動、`20 → 19`、DELAY、GUARD、QUICK、定身、面向與背後攻擊、機會攻擊全部已接且與原作相符；唯一沒接的突襲 −6 **已判定為死碼**（spec 1136）。⚠ 剩下的是 ECL `24h COMBAT` 的 199 處 `partial`。★ `24h` 沒有運算元（arity 0），所以那不是「參數沒對」——它是**三選一的服務分派點**：商店／營地／戰鬥。✅ **199 處已逐處分類**（`cmd/ecl-combat-sites`，沿控制流往回走找 `LOAD`／`SETUP MONSTER`）：**153 處真的要打、46 處走服務分派**（後面接的是「GOOD QUALITY TAVERN.」「MAY YOU ALWAYS STRIKE TRUE.」這類商店／神殿台詞），而且那 46 處每一處往回走都碰得到 `CLEARMONSTERS`。查完的結論是「營地跑完 ECL 再回來」在 remake **有等價物**（走 lifecycle entry 2／3，不是 `24h` 旗標），而 `0x7EE2` 那一格在正式程式碼裡**沒有任何 producer**（`TestCampRequestFlagHasNoProducer` 釘住）。`+07h`（保留的機會攻擊）不是缺口：remake 用 `CombatAction.Guarding` 表達，三端都對得上 spec 1136 且存得下去讀得回來 | 那 46 處服務分派的場景在 `route-sweep` 走得到而且是中文；剩下的是「remake 用別的機制到達」要不要收斂成原作的旗標路徑 |
| **P1** | **ECL 的 1 個 `partial` opcode** | 168 條指令：`CALL`（`PICTURE` 199 條第 706 輪收掉）（`PRINT RETURN` 已收：`65A0h`／`65A1h` 是 DOS 資料段的游標，全 corpus 沒有任何 ECL 指令碰得到，腳本觀察不到 ⇒ VM 這一側沒有東西可做；玩家看得到的空行那一半 spec 1147 已接完並有回歸測試）| handler 的位址與條數在 [`docs/audit/ecl-opcode-handlers-dos.md`](docs/audit/ecl-opcode-handlers-dos.md)——**照成本排順序**（✅ `33h PRINT RETURN` 已讀完（spec 1147，缺口在 UI 行模型）；✅ `0Eh PICTURE` 已轉 `done`（spec 1148：`4FBAh`／`4FBBh` 旁路模型建進 VM、關閉訊號接進 game 側三個套用點，先前 close 欄位在 session 聚合層整個被丟掉也一併補上）；✅ `24h COMBAT` 已轉 `done`（spec 1149／1182：四支全部接上並各有實機路徑；第一支的另一個條件 `8B56h` 是決鬥旗標，唯一的 producer `GODUEL` 只有 `2Dh CALL` 的 `8000h`／`8001h` 到得了而 corpus 兩個都沒用到）；✅ `2Dh CALL` 七支全部讀完（spec 1150，缺口收斂成 `2E10h` 的髒旗標模型）；✅ `27h TREASURE` 已轉 `done`（spec 1151，隨機表區間 bug、清單順序、造物品常式都補齊，最後一項「多筆 TREASURE 錢幣覆寫、物品累積」也對上了）；✅ `2Eh DAMAGE` 已轉 `done`（spec 1152，三種目標形式與「連打 N 下」全部接進正式路徑，含 corpus 走不到的「單體但隨機挑一名」）；✅ `37h LOAD PIECES` 已轉 `done`（spec 1153，牆面參數接進存檔，三支分派也全部照抄——分支由 VM 決定而不是上層猜）；✅ `38h PROGRAM` 已轉 `done`（spec 1154，結局過場照原作的等鍵位置分五頁播完）；✅ `2Dh CALL` 的 `2E10h` 髒旗標模型也建出來了（spec 1172／1183，啟發式整支拿掉；第 660 輪補上五條漏掉的暫存器同步點並移除唯一一個硬寫死的座標補丁，第 661 輪把投影時機對上原作（收尾投影）並**移除 `Written` 遮罩**，第 662 輪普查完 `720Fh` 的全部寫入者 ⇒ **原作沒有「換 block 的引擎進場放置」**，`ViewMirror.Block` 重新分類成保護 game pack `spawn` 的擋板；第 663 輪逐張盤點 ⇒ 它保護的只有**四個與腳本座標對不上的宣告值**，要拿掉得先用原版把那四張判掉（spec 1184））；✅ `1Ch CLEARMONSTERS` 已轉 `done`（spec 1173，`7603h` 是下一批怪物的圖示槽、不是遊戲狀態））；第 751 輪把三個 `partial` 的**缺口大小**量出來了：**`33h PRINT RETURN`** ⇒ 走得到的 120 條收斂成 110 個換行段落，其中 **10 段連著兩條會空行**，而 **7 段**兩側的文字落在同一個顯示頁裡（`cmd/ecl-print-return-audit`）——★ 那七頁全部由 game pack 的 `text_rule` 服務，顯示的是譯文而不是把原文 join 起來的結果，而 UI 的 `wrapTextLines` 本來就把 `\n` 當段落切 ⇒ **要補的是那七則譯文的版面，不是 ECL VM 也不是行模型**（⚠ 訊息框有 `maxLines` 上限，多推一列可能截掉後文，要配截圖驗收）；**`0Eh PICTURE`** ⇒ `4FBAh`／`4FBBh` 那個未定項解掉了：兩格是**現在／上一次的畫面模式**（3 非第一人稱、4 第一人稱，因為兩者都由 ECL 格 `4BE6` 的新值 0／非 0 決定，而 `4BE6` 就是第一人稱模式旗標）⇒ 旁路讀作「前後都還在第一人稱就不重繪」（spec 1148）；⚠ **`docs/audit/ecl-opcode-handlers-dos.md` 的狀態欄以前是手記的**，六支已經 `done` 的還記著 `partial`、摘要把「還剩多少」多報了四倍 ⇒ 現在由 `TestHandlerReportStatusMatchesOpcodeEffects` 對著 `internal/ecl.OpcodeEffects` 擋住；每一個都要 producer→state→consumer 三段齊全 |
| **P1** | **手札 producer** | ✅ **接完了**：locale 宣告 **64 則**，全部由 **48 條**內容規則的 `journal_message_ids` 解鎖（分頁的手札一則對多個 id）。✅ 六張原版插圖與地圖（含手札 59 的眼魔洞穴圖）也接完了：清單在 `gamepack/rules/journal-images.json`，檢視器在 `cmd/azure-bonds-game/journal_image.go`（縮放／平移／圖說跟語系走），四條閘門守著（清單↔目錄互為子集、en／zh-TW 內文與圖說、**PNG 解得開**、查詢對得上）| — |
| **P1** | **存檔剩餘欄位** | ✅ **角色記錄的 `unknown` 歸零**（spec 1164）：422 bytes ＝ decoded 299／documented 123，PC-98 除錯符號展得開 `CHARREC` 的 79 欄。✅ **SAVGAM 的 ECL 記憶體已雙向接上**（spec 1163／1176）：四塊 ＝ 區 0／1／2／3，第四塊是**位元組定址**的程式碼視窗（自己一對 codec、0 照收）——先前新開的遊戲匯出的那一塊整塊是 0。✅ DOS 比 PC-98 少的那一格也定案了（spec 1166：`+14Ch` 的 `MONSTERTYPE`）。✅ **`+0E6h` 對完了**（spec 1196）：PC-98 符號叫 `HIGHESTPREVLEVEL`，兩平台各 15 處讀取、逐模組逐數量相同，全部是雙職門檻比較；**remake 那一處把它當顯示等級的誤用也改掉了**——原版顯示常式（`LIBRARY`，PC-98 `overlay-19:0408h`／`041Ah`）印的是 `PREVIOUSLEVEL[槽] + CURRENTLEVEL[槽]`，`+0E6h` 在整支裡一次都沒出現（⚠ 純多職角色的 `+0E6h` 是 0 ⇒ 舊寫法在多職角色上**看不出錯**，只有雙職角色會露餡，所以那條測試是必要的）。⚠ 剩 `MOVEPARTY` 跨遊戲轉移未做。`+11Ch` 已由 spec 1180 對完（`BASEATTBLOWS[0]`），`+192h` 兩種讀法本來就相容（ECL 投影到 Pascal 的保留欄）| 轉移另開：`MOVEPARTY` 是跨 Gold Box 的角色轉移工具（spec 534），需要 Pool of Radiance／Hillsfar 的資料，不在 CoAB 的玩家路徑上 |
| **P1** | **怪物側的自動換裝** | ✅ 資料接線完成：`MON*ITM` 已進 `Fighter.MonsterItems`（44 個區塊全部帶物品，最多一隻 5 件）。✅ **那個沒答案的問題已經有答案**（spec 1174）：派生值重算讀的是**槽 0 的裝備中武器**，有就無條件把類別表的小型傷害三連寫進現值 ⇒ 13 隻的記錄骰是**放下武器時**的天生攻擊，不是矛盾。✅ **接線也做完了**（spec 1175）：開戰掛完 `MON*ITM` 就投影槽 0 的裝備中武器（`monster.ProjectMonsterWeapon`），大／小體型的切換集中在 `Fighter.WeaponDamageAgainst(目標)` 一處，自動換裝的重投影也帶新欄位。✅ 遠程那個保留也收掉了：整個戰鬥 overlay **只有一個**大／小切換點，與近戰／遠程無關。✅ `+1A2h` 漂不漂第 711 輪用原版量掉了：20 輪連打大型目標（鱷魚，不再生、以 AIM 面板 HP 差量讀每擊傷害），第 6、7 擊仍打出滿值 ⇒ **原作不漂**，remake 的只讀實作行為一致（spec 1175 增補）| — |
| **P1** | **音樂／音效的 cue 綁定** | ✅ **播放生命週期接完了**（spec 1192）：三格狀態的 17 處寫入收斂成 **4 種動作，4 種都發得出來、4 種玩家都碰得到**。「停」不是劇情資料而是**按鍵**（原作派曲常式查不到就 `ret`，唯一會停的是玩家關音樂）⇒ 補的是 **Ctrl+S 音效／Ctrl+O 音樂**兩個全域開關，不是 pack 裡的 binding。✅ **換曲點對回事件 13／13**：7 個查 `CURRENTECL` 派曲表（`ecl_blocks` binding）＋ 6 個事件驅動（**開戰**、開戰的 **`47h` 分岔**、**開場**、**角色建立**、**全滅**、**結局**，各用 `context` binding／cue，因為原作那幾處**不看場景**）。分母釘在 `TestEveryOriginalMusicChangePointHasARemakeCounterpart` 的表裡，不是敘述。✅ **DOS 側結案：DOS 版沒有 BGM**（94 個成員逐一列舉，沒有 `MSCDRV.EXE` 也沒有音樂資料檔；配正對照） 結局那一首掛在 `beginEndingScene()`（原作 PC-98 `overlay-18:168Dh` 寫在結局文字**正上方**）；全滅那一首掛在 `finishCombat()` 的 `StatusEnemyWon` 分支（原作 POSTCOM `1955h`，判準是 `PARTYDEAD && !ADUEL`，而非全滅那一條在 `18F6h` 直接跳過換曲點） | ✅ **「戰鬥 phase 的音樂同步」查掉了：原作沒有這回事**——位元組直掃戰鬥模組（`overlay-08` COMBAT／`overlay-13` COMSTUFF／`overlay-32` TACMAP）：`MSCPLAY` **0 處**、`SOUNDFX` 12 處，`MSCSTOP` 也只有常駐那一處（玩家關音樂）⇒ 曲子在 `INITCOMBAT`（開打前）選定，整場戰鬥只有音效；★ 那個 0 有正對照（同一次掃描在同幾個模組找到 12 處音效）⇒ 不是掃描的假零。⚠ **接上換曲點不等於接上那個畫面**：原作全滅之後回主選單、結局之後問存檔，remake 的全滅目前只顯示「戰鬥失敗。」就回到地圖。⚠ remake 用 `combat.StatusEnemyWon` 當全滅那一刻，**判準與原作逐位查 `CHARSTATUS` 不完全相同** |
| **P1** | **UI 與原版 fidelity** | 五張 manifest 圖的**圖層對齊**已收（spec 1130／1131／1132）。**原版 oracle 早就有了**：`tools/dos-oracle-jump-capture.sh` 用自製存檔把原版直接放到指定格，`workplace/dos-oracle/out/` 有 **755 張**擷取（提爾佛頓 `geo2-b01` 就佔 220 張），參考集 545 張收斂成 **249／585 種**牆面配置已比過。⚠ 那四張圖（`ECL5/0x33`／`0x35`／`ECL6/0x40`／`0x45`）握著 **336 個未比中的約 170 個**。第 749 輪逐格量過八十張：**不是「一面牆都不畫」，也不是缺存檔歷史**——選圖 14／17 的 `WALLDEF` 區塊各 1,560 bytes ＝ 兩個槽，`LOADWALLSET(1, 14)` 自己就填滿槽 1 與槽 2，所以 `15`／`18` 那兩個運算元是死資料（spec 1185 §`LOADWALLSET`）。修了三處：**天空選色**（該段 ECL 進入碼寫的 `4BFDh`／`4BFEh`，remake 載檔不重跑 ⇒ 停在別章的值；`cmd/ecl-sky-colours` 產表）、**多槽登記**（一個牆面組佔兩個槽時只登記在起始槽 ⇒ 牆型 6..10 整批不畫）、**回退路徑照段選塊**（`symbolImageForGlobalID` 原本一律取第 0 塊）。最後照原作的**平表**重做（`sub_39F` 從一塊平的三槽緩衝取位元組、`PUT8X8SYMBOL` 只看編號落在哪一段），四張合計 **76,782 → 1,190**，逐格相同 **23／80 → 60／80**，`geo5-b35` 二十張全部逐格相同 全量 **766 張**（第 750 輪 +221 張擷取，第 751 輪重拍 20 格）現況：**737／766 逐格相同、差 3,552 格**，十七張圖有**十三張逐格全等**。簽章覆蓋 249／585 → **382／585**。★ 第 751 輪換掉的是**擷取**不是 remake：那 20 格的舊擷取是**殘影**（跳格之後畫面還沒換完就拍，而殘影是一張完全正常的遊戲畫面）——`geo5-b33` 一張圖原本就佔 14,075 格的 94%，重拍後 14,075 → **3,552**。補了兩件事：`settle()` 連拍到兩張視窗完全相同才收（只比視窗，比整張永遠等不到）、以及用原版的**區域地圖**核對隊伍位置（`tools/dos_screen_areamap.py`，兩個正對照 ok、兩個對調的負對照 mismatch），而且**核對要放在拍完之後**（進出區域地圖會讓視窗停在舊畫面：先拍再核對 0 格差、先核對再拍 5,953 格差）。⚠ 不等於「第一人稱畫對了」：還有 203 種簽章沒有原版畫面可比 | 剩下的差異**換了形狀**：單張最大差 3,334 → 414，型態從「整面牆有無」變成**相鄰色階**（`11→7`、`7→6`、`8→7`）⇒ 是渲染層的色階問題，不是牆面有沒有畫。集中在 `geo5-b33`（2,773）、`geo5-b32`（437）、`geo5-e31-b32`（341）三張。⚠ **不要拿 GEO 幾何反推「擷取站錯格」**：兩個方向都有偽陽性（側牆不蓋到中間、遠牆正好蓋在中間），位置只能用畫面座標或區域地圖核對。★ 線索已經收斂：**29 張裡 28 張的視錐會走出 16×16 的地圖**（唯一例外把側向放寬到 ±3 也出界），但**不是繞回的錯**——`geo5-b33` 的 `wrap` 改成 `false` 之後 `(0,4)` 朝南從差 222 變成差 **2,101**，而且視錐出界的 265 張裡有 237 張逐格相同 ⇒ 「出界」本身不預測差異。逐張看圖是**同一塊東西畫在不同的地方**（差異分類雙向：`7→11` 與 `11→7` 同時出現）⇒ 下一個查**視錐外緣那一列的牆面塊選擇** |
| **P2** | 三平台打包、README／截圖／推廣片 | 等 P0／P1 收斂 | — |

★ **P2 的閘門是「P0／P1 收斂」，不是一個要人拍板的決定。** P0 兩列都綠了；
P1 現在還開著的是這五項，每一項都有具名的下一步：

| P1 開著的 | 缺什麼 | 下一步要的東西 |
|---|---|---|
| ECL 的 1 個 `partial` opcode | 可達指令 14,177 條裡 `partial` 還有 **168**（`CALL`）。`PICTURE` 第 706 輪收掉（旁路模型＋關閉接線）；四張宣告值第 707 輪判掉並刪除；`CALL` 剩 `ViewMirror.Block`（護 `area-default` 與下水道／火刀入口）| `PRINT RETURN` 那七頁全部由 `text_rule` 服務、UI 的 `wrapTextLines` 本來就把 `\n` 當段落切 ⇒ **改的是那七則譯文的版面**，要配截圖驗收（訊息框有 `maxLines` 上限）。`CALL` 的四張宣告值仍要原版畫面當 oracle |
| 存檔剩餘欄位 | `+0E6h` **已對完**（spec 1196：`HIGHESTPREVLEVEL`，兩平台各 15 處讀取、逐模組逐數量相同，全部是雙職門檻比較，沒有一處拿去顯示）；✅ **remake 那一處誤用也改掉了**：原版的顯示常式（`LIBRARY`，PC-98 `overlay-19:0408h`／`041Ah`）印的是 `PREVIOUSLEVEL[槽] + CURRENTLEVEL[槽]`，`+0E6h` 在整支裡一次都沒出現；`MOVEPARTY` 跨遊戲轉移未做 | `MOVEPARTY`（跨 Gold Box 的角色轉移工具，spec 534）另開，它需要 Pool of Radiance／Hillsfar 的資料，不在 CoAB 的玩家路徑上 |
| 怪物側的自動換裝 | ✅ `+1A2h` 第 711 輪原版實測**不漂**（spec 1175） | — |
| 音樂／音效的 cue 綁定 | **13 個換曲點全部有落點**（`cmd/music-change-points`，分母在 `internal/audiomap`）：7 個查 `CURRENTECL` 派曲表 ＋ 6 個事件驅動（開戰、開戰的 `47h` 分岔、開場、角色建立、全滅、結局）| 「戰鬥 phase 同步」已查掉——原作戰鬥中不換曲也不停曲（`MSCPLAY` 0 處，配 12 處 `SOUNDFX` 正對照）。剩下的是**時機**：「接上了」不等於「在原作會發的那一刻發」，那要實機比對 |
| UI 與原版 fidelity | 已擷取的 **766 張逐格比對 737 張相同**、差 **3,552** 格，十七張圖有十三張全等（第 751 輪）。剩下的差異**不是牆面有無**而是相鄰色階，已定位到版面呼叫層級（`drawWallFar` 版面 0／9 與 `drawWallMid` 版面 4），而且**兩個方向同時發生** ⇒ 不是單向走訪錯誤。另一半是**擷取覆蓋率**：585 種牆面配置蓋到 **382**，203 種沒有原版畫面可比 | 拿原版畫面反查那幾格該是哪一個 8X8 符號（`rulebook/64` 的截圖 oracle），不要再猜走訪；覆蓋率照 `cmd/fp-screen-plan -covered` 的貪婪順序補拍（`tools/fp-capture-plan-751.sh`，64 格可再蓋 84 種）|

⚠ **原版 oracle 不是瓶頸**——`tools/dos-oracle-*.sh` 一整套都在，`workplace/dos-oracle/out/`
有 755 張擷取。真正卡住的是各自具體的東西：兩項要**原版實機量一個值**
（`+1A2h` 會不會漂、`+0E6h` 的讀法），一項要**共用 engine 收一個新 signal**，
一項要**表現層的模型**，而 UI fidelity 已經不卡在渲染——**卡在擷取覆蓋率**
（585 種牆面配置蓋到 249，其餘 336 種沒有原版畫面可比；spec 1185）。

⚠ 一個順手記下的小瑕疵：`cmd/re-ledger` 的「引用」欄掃 `docs/` 找「同時提到某
overlay 與某個十六進位值」的檔案，而**它自己的輸出就在 `docs/audit/` 底下**
——所以第一次跑完會改變下一次的輸入，要跑兩次才到不動點。工具本身是決定性的
（同樣輸入兩次輸出相同），只是「跑一次就 commit」可能留下非不動點的版本。
影響僅止於提示欄的檔名清單，不動任何狀態判定。

⚠ **`docs/audit/ecl-event-catalog.*` 這份 committed 產物比現況落後 3.4 倍，而且
重生不出來**（2026-08-21 第 619 輪量到，成因在本輪之前）：

| | committed 的那份 | 現跑 `cmd/ecl-event-catalog` |
|---|---:|---:|
| 靜態可達 instruction | 4,222 | **14,177** |
| 跨 effect-kind 候選 | 154 | **701** |
| corpus 出現的 opcode | 55 | **61** |
| 其中 `2Dh CALL` | 78 | **168** |

現跑的 14,177 與 `cmd/ecl-effect-coverage` 的分母**完全一致**，所以現跑那份才是對的；
committed 那份停在走訪修好之前。**任何拿事件目錄數出來的數字都要先確認是不是這個坑**
——它會系統性少報，而少報的方向剛好是「看起來已經數完了」。

重生被兩道 fail-closed 閘擋著，兩道都是設計上該擋：

- 帶預設的 `-reviews`：`candidate review ECL4.DAX/0x23/0x0130-0x019D does not match
  the current corpus`。候選 ID 由 `member/block/start-end` 組成，走訪一變 ID 就跟著變；
  照 `AGENTS.md` §12 必須失敗即關閉，要重新審那個候選，不能模糊比對。
- 關掉 reviews（`-reviews=`）：`ordered-effect phase ledger is missing corpus opcodes
  0x1E, 0x26, 0x2C, 0x30, 0x34, 0x3B`。`internal/eclcatalog/phases.go` 的表少這六支，
  而它們在**現在的** corpus 裡是可達的。

⇒ 連帶影響：phase 台帳（55 列、25 支已分類、30 支 `unknown`）也重生不出來，對 `0Dh`
（spec 1146 已讀完，表上仍是 `unknown`）與 `2Dh`（spec 1150）是**過期的**。改
`phases.go` 但重生不了產物只是把不一致換個地方放，所以這一輪沒有動它。要收是一件事：
補那六支的分類 ＋ 重審那個 candidate ＋ 重生四份產物，順便把所有引用 4,222／154 的
文件一起改。

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
| 3 | **開場→結局的主線續接** | 唯一真正決定「能不能通關」的工作，但它會**反覆撞上第 2 項**：每接一章就要打一場沒有完整回合生命週期的戰鬥。**驗證方式改成分段**（一個 ECL block 一段、一條 `NEWECL` 邊一段交接），計畫見 [`docs/plan/mainline-segmented-verification.md`](docs/plan/mainline-segmented-verification.md)。✅ `SEG-01` 完成（spec 1141）：主迴圈讀 `LastECL`、`NEWECL` 改它；**25 個 block、47 條出邊**，只有開場與結局沒有出邊；報告見 [`docs/plan/seg-01-verification-report.md`](docs/plan/seg-01-verification-report.md)。✅ `SEG-02` 完成：`LOAD FILES` 的第一個運算元就是那一段載的地圖區塊，game pack 的 `area_id`／`script_block` 與 block 編號 join 得起來，段落清單已產生；報告見 [`docs/plan/seg-02-verification-report.md`](docs/plan/seg-02-verification-report.md)。✅ `SEG-03` 前半完成：段的 id 一律 `ECL{成員}/0x{block}`（不綁地圖名），25 段的標籤逐條有原作敘述為證；報告見 [`docs/plan/seg-03-verification-report.md`](docs/plan/seg-03-verification-report.md)。✅ `SEG-04` 完成：`-segment <id>` ＋ 註冊表 `internal/segment`，25 段逐段一條測試都進得去；報告見 [`docs/plan/seg-04-verification-report.md`](docs/plan/seg-04-verification-report.md)。✅ `SEG-11` 前半完成：25 段的邊界狀態存得下去讀得回來，`-segment-snapshot` ＋ `-party-load` 交接實際跑通；報告見 [`docs/plan/seg-11-verification-report.md`](docs/plan/seg-11-verification-report.md)。✅ `SEG-12` 完成：47 條 `NEWECL` 邊每條一個交接測試，來源段存快照 → 讀回 → 帶著來源 block 當 `LastECL` 進目的段；報告見 [`docs/plan/seg-12-verification-report.md`](docs/plan/seg-12-verification-report.md)。✅ 階段 2 的現況盤點完成：`cmd/segment-coverage` 量出 25 段的入口，**22 段有中文劇情、3 段的入口不出文字（被別段帶進來的）、0 段落回原文**；報告見 [`docs/plan/seg-20-coverage-report.md`](docs/plan/seg-20-coverage-report.md)。剩下的接線工作在**段內**（每回合／搜尋生命週期、離開邊條件、段內戰鬥），✅ `SEG-10`／`SEG-11`／`SEG-13` 完成：那條 790 行的整條跑拆成 **12 個段 subtest**（正規化逐行比對證明一個遊戲動作都沒增刪），每段結束存一份快照並整批驗往返；報告見 [`docs/plan/seg-10-verification-report.md`](docs/plan/seg-10-verification-report.md)。✅ **主線接到第六章**：同一條 session 從開場走到 `ECL6/0x40` 密斯卓諾墓園入口（散提爾堡 → 猶拉什 → 希爾斯法 → 立石群 → 密斯卓諾），段測試 18 段、段界快照 18 份；報告見 [`docs/plan/seg-20-myth-drannor-report.md`](docs/plan/seg-20-myth-drannor-report.md)。✅ **通關了**：同一條 session 從開場走到擊敗提朗瑟克斯的結局選單（38 名戰鬥員的最終戰），23 個段測試守著；報告見 [`docs/plan/seg-21-ending-report.md`](docs/plan/seg-21-ending-report.md)。階段 3 的跨段不變量也接上了：語系（`SEG-32`，整條主線 233 句**落回原文 0 句**，第一次跑抓到 12 句英文並補譯）、隊伍連續性（`SEG-31`）、音樂綁定（`SEG-33`）。內城一樓九間＋二樓五間房也逐間走過了（九間出得來；五間被 spec 1142 的全域 `4C00` bank 壓掉，已釘住成因）。全遊戲「哪一格演哪一場」的對照表已經產得出來（`cmd/ecl-cell-events`：25 個 block 裡 14 個用地形碼分派、**234 個場景對到格子**）。✅ **逐格實測也全遊戲一次跑完了**（`cmd/cell-sweep`，27 秒）：13 個 block、**194 個分派索引**逐個站上去跑生命週期，**185 格演出中文、落回原文 0 格**，9 格沒演出來且每一格都印得出守衛（✅ 第 647 輪起「沒演出來」自動分得出成因，spec 1177）——其中 **7 格指向同一個成因 spec 1142 的 `4C00` map-local bank**；報告見 [`docs/plan/seg-23-cell-sweep-report.md`](docs/plan/seg-23-cell-sweep-report.md)。✅ **spec 1142 的成因收斂了一半**：換 bank 的時機**不是** `NEWECL`（`ECL4/0x20` 前導讀的是第二章寫的 `4C01`）也**不是** `LOAD FILES`（`ECL3/0x10` 在它之前寫 `4C05`／`4C0C`，之後整段都在用），兩個候選各有一個反例；`4C07`／`4C0E` 則根本不是被污染的那一類，是 `ECL6/0x43` 依 `C04B`（隊伍 X 座標）設的抵達標記。剩下兩種可能（真的是 map-local 但時機未知／這一區就是全域旗標而那五間房在原作也演不出來）**靜態讀碼分不出來，要拿原版當 oracle**。✅ **第二種分派形狀（`GETTABLE` 查表）也解讀完了**（報告見 [`docs/plan/seg-24-dispatcher-detection-report.md`](docs/plan/seg-24-dispatcher-detection-report.md)）：兩個世界地圖 hub 是「路段編號 `4C9D` → 14 個旅途場景」，散提爾堡酒館是 `7F7C`。同一輪把分派器的配對條件修對（用目的地運算元，不是相鄰），涵蓋從 14 個 block 變成 16 個，逐格實測 **250 個索引、240 格中文、落回原文 0 格、沒演出來 10 格（逐格看過都是盤點的限制，spec 1177）**。✅ **`4C9D` 也解出來了**（spec 1143，報告見 [`docs/plan/seg-25-world-map-localization-report.md`](docs/plan/seg-25-world-map-localization-report.md)）：抵達時 `MULTIPLY 4C9C, 4 → 4C9D` 算基底並把四個鄰居載進 `4C02..4C05`，選好方向後 `ADD 7F79, 4C9D → 4C9D`；四張表分別給旅行天數、走法數、到達地點與旅途場景。同一輪把世界地圖走得到的分支逐條驗過語系。✅ **覆蓋再擴一輪**：`State.ArriveAtWorldLocation` 直接抵達宣告的 14 個地點，從每個地點各掃一次，走到的分支 17 → **762**，敘述與選項**都沒有一條落回原文**。⚠ 仍是下界：迷斯卓諾要第六章進度才走得到，另有 161 條分支推不動。✅ **`SEG-31` 後半也完成了**：裝備／記憶法術／效果在 23 個段界逐段驗過（報告見 [`docs/plan/seg-31-33-invariants-report.md`](docs/plan/seg-31-33-invariants-report.md)）。✅ **`SEG-33` 後半也完成了**：選曲逐格對回 spec 355 的 PC-98 selector 表（報告同上）。**階段 3 的四個跨段不變量全部完成**。剩下的是 spec 1142 那個 oracle 實驗與中後期快照起跑 | 逐段推進，每段一個可重播快照 |
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
2. `P0-RE-2`：靜態層已完成 6 DAX／25 block／125 entry／14,177 instruction 的可重生
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
解析；原版手札地圖 bitmap 已放進 renderer。舊 `(15,1)` 搜尋邊只保留為
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
| 手札地圖的資料誠實性 | 遊戲內同時提供繁中圖例摘要與**原版掃描裁出來的地圖本身**（`assets/journal/entry-59.png`）。⚠ 圖已到位不等於出口 route 已解：圖面拓撲是 `layout-only`，不能由它反推座標或牆值。 |
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
| 開場到結局的正常玩家主線 | ✅ **通關**：同一條 session 從開場走到擊敗提朗瑟克斯的結局選單（立石群 → 密斯卓諾墓園的紅網與黛米爾公主 → 外城遺跡 → 內城遺跡的儀式與爪牙戰 → 二樓東北角的最終戰，38 名戰鬥員）。拆成 23 個段 subtest，每段結束存快照並驗往返。 | save／reload gate、段內支線、`29h` 的距離旁白（spec 1144）與 `1Ch` 的戰利品清除（spec 1145）都已完成。剩下的是 spec 1142 的 oracle 實驗，以及 `29h` 的重畫迴圈（原作按 `ADVANCE` 會距離減一、重新出選單並換旁白，remake 還沒接）。 |

### P1：補齊可玩規則、資料與原版行為

| 工作 | 目前缺口 | 驗收方向 |
|---|---|---|
| 全 ECL 與外部 routine | 25 個 block／125 個 entry 的 parser／控制流 corpus gate 已完成；`C04B..C04F` virtual-map adapter 已閉合，但 `CALL`、`NEWECL`、剩餘地圖服務、劇情旗標、NPC 離隊、輸入與 continuation 的完整 consumer 仍未閉合。與玩家結果無關的 raw work address（目前如 `0x4C00`）不列為 blocker。 | 只逆向會改變玩家結果的 producer→state→consumer；每個完成事件都要有 raw bytes／runtime trace、JSON contract、stable ID 測試與正常輸入路徑。 |
| 全地圖與世界旅行 | 16 個原始 GEO block 已在 game-pack 宣告；14 個世界點位的 ECL1 到達、Tilverton→全點 directed adjacency，以及新遊戲→阿沙本福德→立石群→艾森布拉的正常主線已通過 Docker gate。仍缺所有城市／地城房間 coverage、TRAIL／WILDERNESS／EXIT 全分支、隨機遭遇、所有入口出口、持久 map state 與原版 fidelity。 | 建立每座城市／每個 GEO block 的正常事件 coverage matrix，保存 flag／座標／資源 handoff，並補全世界旅行與重訪回歸；不把攻略座標直接寫成規則。 |
| 戰鬥規則、AI、法術效果與動畫 | 法術那一半**已收斂**：可宣告的 73 支全部宣告，`handler`／`visual`／`sound` 三欄 observed（spec 1111／1117／1123／1124／1125／1126）。視覺不是七十幾筆待補資產——原作只有一段共用施法投射物（COMSPR 區塊 5）加閃電電弧（6）與魔法命中（10）。回合生命週期的逐項對照也收完了（spec 1135～1138），怪物側的物品鏈已進 `Fighter`。戰鬥的 `使用` 整條走得通：充能 18/18（spec 1169／1170）＋ 卷軸 12 件（spec 1171）。怪物換武器的傷害已由 spec 1174／1175 收斂（記錄的骰是放下武器時的天生攻擊），怪物側的武器投影也接上了。**每回合攻擊次數已改成原作模型**（spec 1180）：半次單位 ＋ `ADJUSTBLOWS` 的回合奇偶換算，先前讀錯欄位讓 68 種怪物全部每回合只打一下。隊伍側的 `BASEATTBLOWS`（戰士／聖武士 7 級、遊俠 8 級 → 一次半）與遠程的彈藥數上限也接上了。仍缺的是 `PREVIOUSLEVEL`（雙職角色靠前職業拿到的一次半）與 spec 771 的 `overlay-13:1CE2h` 加成。 | 回合生命週期每項一條回歸測試 ＋ 正常戰鬥路徑；影片只能證明演出，數值要回 bytes／DOSBox。 |
| 存檔、角色規則與跨遊戲轉移 | remake save v12 已保存 Search／edge；DOS／PC-98 `SAVGAM`、角色 sidecar、完整 record、年齡／職業／特殊能力、刪除／改名與 `MOVEPARTY` 跨遊戲 transfer 尚未完整 round-trip。 | 先完成版本化 parser／serializer 與 save mutation diff，再以角色檔跨 Gold Box 來源做 stable transfer contract；不能把 `MOVEPARTY` 靜態 helper 直接當秘密門。 |
| 全量繁體中文化與遊戲內手札 | 第 556 輪修正七行裁切：長手札依真實字寬自動分頁，來源 stable ID 與 save 不變；摩安德之坑真實 producer 已接通手札 46。目前 locale 宣告的 64 則**全部**有 en／zh-TW stable ID 與 ECL producer 解鎖。**物品名稱已收斂**：名稱是三個名稱編號查 `DS:1040h` 那張 255 筆詞表組出來的（spec 1178），語系檔的 `item_name_XX` 由 54 筆補成 252 筆並改正第 260 輪填錯的編號，原版 253 件物品用到的 126 個編號**缺譯 0**、115 種相異名稱逐件列在 `docs/audit/item-names.md`。全 ECL 頁文字亦已閉合（1022 頁、unmatched 0）。**戰鬥員名字亦已收斂**：原版六章 `MON*CHA` 共 68 種，pack 先前只宣告 23 種、其餘 45 種在戰鬥畫面是英文；現在 68 種全部有 zh-TW／en，逐種列在 `docs/audit/combatant-names.md`，並由三條閘門守著（spec 1179）。**手札的原版插圖與地圖也接完了**：六張（含手札 59 的眼魔洞穴圖）由 `gamepack/rules/journal-images.json` 對應，檢視器支援縮放與平移，四條閘門守著（含「PNG 解得開」——renderer 那一側刻意吞錯，壞掉的圖不會讓任何東西紅）。法術／地名／UI 校對仍未完成。 | 以 stable `message_id` 做 coverage／orphan／source-drift audit。 |
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

下一個最小可重現工作：**重排眼魔洞穴那一段的內容順序，然後把 spec 1160 那
六件事一起推上去**。前置（GEO 目錄）已經做完，六件事與九處斷言、巫師塔的走法
都推導好了，唯一剩下的是死精靈那個一次性選單（`ECL4/0x22:0651h` 的 `4C06`）
會在更早的地方被消費掉——要先查清楚主線在哪一步先踩到 `(13,1)`，再決定那一段
的斷言該擺在哪裡。

六件事：① 拿掉 `projectFreshDungeonCoordinatesBeforeCall` 的 `Spawn` early
return；② `case 0x2E10` 的守衛改成逐條看 `CallRequests[i].BlockID`；
③ `2Dh CALL C01Eh` 改讀腳本當下的 `C04D` 並把走完的座標寫回暫存器；
④ 重畫要消費座標寫入（記「上一次重畫吃到哪個執行序」）；⑤ 宣告的 spawn 改成
後備；⑥ 拿掉 `ECL2/0x02` 的座標位移。走訪器那三條規則（落點不符就交回控制權、
開鎖門子路徑也記被拒的邊、隨機格排除清單逐圖）也要一起加。

驗收：主線通關測試全綠，而且每一段路線的起點都由原作行為推出來，不是把失敗值
抄進期望值。⚠ 查「誰跳到這個位址」一律用 `cmd/ecl-window -into`。

⚠ 不要把「主線跑得完」擴大解讀成全城市、全地城或全事件完成；也不要先深挖與玩家
結果無關的反組譯。上面那張表的「量到的缺口」欄才是分母，不是印象。
