# 復古遊戲知識路由

本檔只負責把任務導向一份入口與必要 reference；不要因看到路由就預讀整座
`docs/knowledge/` 或外部專案。

⚠ 右邊兩欄一律指**這個 repo 裡的檔案**。個人 skill 會改名、降級成 knowledge-base
或整個消失，所以這裡只留一個仍然叫得動的 skill 名（`re-retro-cht-rulebook`），
其餘一律用 repo 內路徑——路徑對不對，`ls` 一次就知道。

## 復古遊戲

| 任務關鍵字 | 第一入口 | 必要時再讀 |
|---|---|---|
| CoAB、Gold Box、ECL、DAX、GEO、SAVE、戰鬥、IDA | `re-retro-cht-rulebook` skill（它再路由到逆向／中文化的按需規則）| 先讀 `docs/knowledge/coab-re-coverage-matrix.md`，再依缺口讀對應 `docs/spec/` READY 規格；`gold-box-state.md` 只作主題／歷史補充 |
| ECL VM 內部、operand 編碼、work address、bank、helper、opcode handler | `docs/knowledge/gold-box-ecl-interpreter.md` | 個別函式狀態查 `docs/audit/coab-function-index.md`；要改反組譯流程先讀 `docs/spec/559` |
| 反組譯 overlay、IDA 建庫、TPOV、entry stub、函式覆蓋台帳 | `docs/spec/559-full-module-re-sweep.md` | 工具在 `tools/re-sweep.sh`、`tools/ida/`；不要重寫一次性 IDC |
| PC-98、640×400、16×15、中文字級、版面 | `docs/spec/348-original-dos-frame-pc98-type-density.md` | `docs/spec/406-dos-gui-draw-contract.md` |
| 配樂、音效、YM2203、S98、推廣片音樂 | `docs/knowledge/gold-box-audio.md` | `docs/knowledge/pc98-gold-box-music-reconstruction.md` 與對應音訊 spec |
| GUI、石框、人物頭身、3D viewport、戰鬥 HUD | `docs/spec/406-dos-gui-draw-contract.md` | `docs/spec/391-dos-head-body-character-stage.md`、`docs/spec/348-original-dos-frame-pc98-type-density.md` |
| 中文翻譯、Journal、攻略、長文分頁 | `docs/knowledge/golden-box-remake-for-chinese-readers.md` | `re-retro-cht-rulebook` skill 的中文化路由 |
| Demon’s Winter、冬之魔、Wasteland、跨 SSI RPG | `docs/knowledge/ssi-rpg-cross-project-lessons.md` | `re-retro-cht-rulebook` skill，再唯讀抽樣對應專案 |

## 跨專案界線

- `/home/anr2/cht/daemon_winter` 是比較樣本，不是 Gold Box VM 的既定消費者；
  已知 `DEMON.INT` 是原生 MZ 8086 executable，格式與 runtime 必須另證。
- Wasteland 尚未完成第二 adapter 驗證。可先沿用證據分級、中文資料 gate、字型、
  renderer contract、可重現測試與知識庫方法，不得預先宣稱共用 ECL／DAX／SAVE。
- 只有第二款合法持有的遊戲以最小 adapter 通過格式、座標、save 與 renderer
  contract，某機制才可從「候選共用」升級為 Golden Box engine 證實能力。
