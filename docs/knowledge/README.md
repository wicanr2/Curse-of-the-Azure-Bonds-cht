# 《青色枷的詛咒》知識庫入口

本目錄保存可重用的 Gold Box／CoAB 研究結論。它不是完成度清單；目前是否已足以
重建某個系統，一律先查
[`coab-re-coverage-matrix.md`](coab-re-coverage-matrix.md)。逐輪規格與原始證據
仍在 [`../spec/`](../spec/README.md)，不能用知識文章的敘述取代位址、bytes、xref
或 runtime trace。

## 文件權威順序

1. 原始 binary／DAX／GEO／save bytes 與可重現原版 runtime。
2. 標示輸入雜湊、工具版本、位址空間及推論等級的 `docs/spec/` 規格。
3. [`coab-re-coverage-matrix.md`](coab-re-coverage-matrix.md) 的目前閉合狀態。
4. 各主題知識文章。
5. `golden-box-reverse-engineering-worklist.md`、`gold-box-state.md` 等逐輪累積紀錄；
   這些只供追溯，遇到衝突時不得壓過較新的 READY 規格與完整度矩陣。

## 四份工作台帳對應

| 目的 | 本專案文件 |
|---|---|
| 研究問題、證據與勘誤 | `docs/spec/`、根目錄 `CONTEXT.md`、`docs/context/` |
| 目前交付 frontier | `remake-completeness-assessment.md` 與當期 `docs/audit/` |
| 歷史執行台帳 | 根目錄 `WORKLIST.md`（已封存，不可單獨重開待辦） |
| 全遊戲驗證／RE 完整度 | `coab-re-coverage-matrix.md` |
| 新研究的固定閉合格式 | `re-closure-record-template.md` |
| compact／交接入口 | 根目錄 `HANDOFF.md` → `AGENTS.md`；需要深層歷史才讀 `CONTEXT.md` |

## 主題索引

| 主題 | 入口 | 注意事項 |
|---|---|---|
| ECL 指令與 VM | [`gold-box-ecl-command-set.md`](gold-box-ecl-command-set.md) | parser coverage 不等於副作用完整；跨類型時序尚待有序事件紀錄閉合。 |
| **ECL 直譯器內部構造** | [`gold-box-ecl-interpreter.md`](gold-box-ecl-interpreter.md) | 原版反組譯讀出的分派、operand 編碼、記憶體 bank 與 helper API；結構可跨作品沿用，**常數必須各作品重解**。 |
| **全函式覆蓋台帳** | [`../audit/coab-function-index.md`](../audit/coab-function-index.md) | 模組／函式層完整度；狀態只來自人工台帳，預設 `待解讀`。 |
| ECL／城市／地圖狀態 | [`gold-box-state.md`](gold-box-state.md) | 長期累積文章；先看完整度矩陣再引用其中段落。 |
| 正常玩家路徑 | [`golden-box-normal-player-path.md`](golden-box-normal-player-path.md) | direct-entry 與固定 fixture 不算正常路徑。 |
| 戰鬥效果 | [`gold-box-combat-effects.md`](gold-box-combat-effects.md) | 數值、AI、動畫、聲音與 continuation 必須分別驗收。 |
| 圖像／字型／UI | [`gold-box-graphics.md`](gold-box-graphics.md) | 素材解碼與畫面 fidelity 是不同完成門檻。 |
| DOS 存檔 | [`gold-box-save-format.md`](gold-box-save-format.md) | 目前是 raw-preserving 局部格式，不是完整 schema。 |
| 重製存檔 | [`gold-box-remake-save-session.md`](gold-box-remake-save-session.md) | 不可把 remake continuation 相容冒稱 DOS save parity。 |
| 音訊 | [`gold-box-audio.md`](gold-box-audio.md) | 播放生命週期 4／4 與換曲點 13／13 已閉合；OGG 人耳循環驗收與散布權仍須分開判定。 |
| 中文長文與攻略 | [`golden-box-remake-for-chinese-readers.md`](golden-box-remake-for-chinese-readers.md) | Journal producer、顯示與翻譯 coverage 分開追蹤。 |
| 跨 SSI 專案經驗 | [`ssi-rpg-cross-project-lessons.md`](ssi-rpg-cross-project-lessons.md) | 第二款遊戲實證前，不升格為共用 engine 事實。 |
| engine／作品的實際切線 | [`engine-title-split.md`](engine-title-split.md) | 內容界線守住了、程式碼界線沒有；「按性質屬於共用層」是推論不是證據。 |
| **Gold Box 系列重用邊界** | [`gold-box-series-reuse.md`](gold-box-series-reuse.md) | 下一作品的短入口；第二 consumer 前不得把候選升格成共通事實。 |

## 新增或訂正文件的最低要求

新主題先複製
[`re-closure-record-template.md`](re-closure-record-template.md)，不要另造缺少 R1–R5
欄位的自由格式完成聲明。

- 標明問題、範圍與不在本輪宣稱的內容。
- 保存輸入檔名、SHA-256、工具版本與位址空間。
- 同列原始定位、附加語意、證據出處與
  `exact／strong inference／hypothesis／unknown`。
- 說明 writer／projection／consumer 是否已閉合。
- 指向 engine contract、CoAB JSON、測試與正常玩家路徑；缺少者明列為缺口。
- 推翻舊結論時，保留 superseded 索引與訂正原因，不讓兩個互斥結論並存。
