# 1190 — 文字規則的覆蓋要以**執行路徑**計，不是以頁計

狀態：`READY`（工具與規則已實作，數字由 `cmd/ecl-text-coverage` 產生）

## 問題

`text_rules` 的 `all_contains` 是「這幾個片段全都出現才算命中」。原作常把一頁的
敘述拆成好幾條 ECL 字串，於是規則很自然地寫成「A 片段 ＋ B 片段」。

問題出在**同一頁不只一條執行路徑**。A 與 B 是不是一起印，取決於執行期的條件：

| 例 | 兩句一起印 | 只印其中一句 |
|---|---|---|
| `standing-stone.grey-man` | 到立石群 ＋ 灰袍男子在 | 灰袍男子不在，只印「你們來到立石群」 |
| `pit.lemon-tang` | 檸檬味 ＋「這附近一定有出口」 | 只印檸檬味 |
| `pit.alias-dragonbait-tendrils` | 藤蔓 ＋「愛麗雅絲**與**龍餌」 | 愛麗雅絲不在隊上，只纏住龍餌 |
| `hillsfar.red-plumes-spill-drinks` | 打翻酒 ＋ 命令清乾淨 | 只打翻酒 |

規則要求兩個片段，走到「只印一句」那條路徑時就命中不了，**玩家看到英文**。

## 為什麼逐頁的覆蓋率看不出來

`cmd/ecl-text-coverage` 原本把處置判在**頁**上：一頁只要曾經被某一條命中的 run
經過就算 `matched`。上表那些頁在「兩句一起印」的路徑上已經被標成 `matched`，
「只印一句」那條路徑就不再有人問。

**玩家走的是一條 run，不是所有 run 的聯集。** 分母錯了，`unmatched = 0` 這個
零就不承載任何保證——把任何一條規則從 pack 拿掉，它仍然是 0。

## 判準

覆蓋率改以 **run**（一次執行累積、**一次**交給 `MatchText` 的那份文字）為單位。
沒命中的 run 分四類，前三類沿用逐頁那一套豁免，只是提升到 run：

| 處置 | 意思 |
|---|---|
| `variable-insert` | run 裡有頁印執行期的值（城名、隊員名、編號），靜態文字裡沒有那幾個字 |
| `subroutine` | 整條 run 就是一頁共用子程式片段，實機一定被併進呼叫端那一頁 |
| `gosub-insert` | run 裡有插入點，手上這份文字**不完整**，不能拿來判缺陷 |
| `truncated` | 停在句子中間，或是另一條 run 的嚴格前綴 |
| **`orphan`** | **走得到、文字完整、沒有任何規則命中——這才是待辦** |

`truncated` 的依據：原作把累積文字交出去的時機是**選單、戰鬥、寶物、輸入、
換 block、離開**，那些點一定落在一句話講完之後。所以停在句子中間的 run 不是
玩家看得到的一頁，是走訪器在 `IF` 分岔上切出來的半截。收尾的引號要看**前一個
字元**：`HE CONTINUES, '` 收在引號上，但引號前是逗號，話還沒講完。

## 這個零撐得住

拿掉任何一條規則重跑，`runs_orphan` 就從 0 變成拿掉的條數，而且逐條指名是哪一份
文字；逐頁的 `unmatched` 在每一種情況下都還是 0。

| 拿掉的規則 | `unmatched`（逐頁） | `runs_orphan`（逐 run） |
|---|---:|---:|
| 一條都不拿 | 0 | 0 |
| `wizard-tower.efreet-band` ＋ `tilverton.sewers.spoils-and-master-alone` ＋ `standing-stone.arrival` ＋ `zhentil.arena.manticores-attack` | 0 | 4 |

## 寫規則時

新規則一律**加在檔案最後**：`MatchText` 是照檔案順序先命中先贏，既有的多片段
規則（兩句一起印那條）因此永遠先命中，單獨印的情況才落到新的那條。新規則彼此
之間，**片段較多／文字較長的排前面**，否則短的會攔截長的
（`YOU WAKE UP IN A DARK, DREARY CELL` 是 `YOU ARE DRAGGED THROUGH THE TEMPLE.
YOU WAKE UP IN A DARK, DREARY CELL` 的一部分）。

## 相關

- `cmd/ecl-text-coverage`：`classifyRun`、`truncatedRun`
- `TestOrphanRunAccountingCatchesTheAloneSiblingBug`、`TestClassifyRunExemptions`、
  `TestTruncatedRunTellsHalfSentencesFromRealPages`
- `docs/audit/ecl-text-coverage.md`：逐條 orphan run 清單
- spec 1106（控制流走訪）、spec 1121（原版 DAX 解碼）
