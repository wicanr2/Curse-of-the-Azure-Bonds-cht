# 第五百零三輪：PC-98 Quick 目標候選鏈與 legacy object 順序邊界

狀態：`READY`（只完成候選順序 adapter；不宣稱完整 Quick AI 或原版 tie／亂數相同）

## 本輪結論

PC-98 overlay 09 的 Quick 目標函式從 `04CCh` 開始：它先從施法者 action
record 的 far pointer 取得候選鏈，對候選逐項執行 suitability，最後把選出的
候選交給施法入口。這證明 Quick 目標不是單純把 Go map 中的角色依名稱排序，
也證明候選探索與法術 metadata／幾何合法性是不同邊界。

目前仍無法從已取得的 bytes 單獨證明 pointer-chain 的完整重抽次數、同分候選
tie、隨機 helper 的用途，以及候選欄位的所有正式名稱。因此本輪只把已由
`CHARACTERLIST／OBJECTLIST` 身份投影保存的 one-based `LegacyObjectID` 做成
可重用 engine＋JSON 契約。這是 `strong inference` 的 bounded adapter，不是
`exact` 的原版候選結果。

## 非破壞性輸入與位址空間

| 輸入／產物 | SHA-256 | 位址／用途 | 推論等級 |
|---|---|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | PC-98 resident／Borland symbol 輸入 | `exact` |
| `PC98-GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay／TPOV 輸入 | `exact` |
| `overlay-09.bin` | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | overlay-local code bytes | `exact` |
| `/tmp/pc98-quick-magic-overlay09.txt` | `3f7f6160e9aac8140ecf320ae45f34572b6d77a39bc7b5f1eb6c8f417c30b36c` | IDA Pro 9.4 連續指令報告 | `exact` |

IDA Pro 9.4 在 Docker 內以 `overlay-09.bin` 的副本執行既有
`scripts/ida/pc98_quick_magic_audit.idc`；原始 executable、overlay、基準
`.i64` 與 repo 內研究輸入均唯讀，database／暫存輸出在 `/tmp/coab503-ida`。
以下數值都是 **overlay-local address**，不能與 resident effective address、
file offset 或 ECL work address 混用。

## 已閉合的指令資料流

`04CCh..0624h` 的連續 raw／IDA 觀察如下：

1. `04E2h..04E8h` 呼叫 `3E01:142Dh`，傳入 `1` 與 `7`，並保存返回的
   `AL`；這是 `1..7` 的 helper 邊界，但 helper 的正式語意仍是 `unknown`，
   不能直接命名為「選第幾個目標」。
2. `04F0h` 由施法者 action record 的 `+14Eh` 取得目前候選 far pointer；
   `04F8h` 先檢查 action 的 `+02h`。未通過時在 `0602h` 前返回。
3. `0541h..05E2h` 沿目前候選的 `+52h` next pointer 走鏈。候選會讀取
   `+65h`、`+56h` 經 table 投影的值，並呼叫 `1701:113Eh`；其完整欄位
   名稱／helper 語意仍為 `unknown`。
4. `0596h..05A8h` 會排除 `+66h >= 80h` 與 `+5Ch == 0` 的候選；
   `05AAh..05B8h` 對另一個候選值執行 `>0`／`>38h` 分支，`38h` 之後減
   `17h`。這些是 raw predicate，尚不可直接翻譯成 HP、距離或陣營欄位。
5. `05C1h..05D3h` 呼叫同一 overlay-local `03D3h` suitability；成功才在
   `05D7h..05E0h` 保存候選 far pointer 為 best pointer，然後繼續鏈結探索。
6. `0602h..061Eh` 若 best pointer 非空，呼叫 `00FA:0048h` 並回傳選取結果；
   這是施法 handoff 邊界，不等於已證明完整的 `GETSPELLTARGETS` 結果。

## 推論分級與 remake 契約

- `exact`：`04CCh` 的候選鏈入口、`+52h` 的下一候選讀取、候選 predicate、
  `03D3h` suitability 呼叫與 `00FA:0048h` handoff 的連續控制流。它們保留在
  上表的 overlay-local 位址與報告中。
- `strong inference`：CoAB `StartCombat` 已依原始 `CHARACTERLIST／IDLIST`
  投影建立 one-based `LegacyObjectID`；以該身份順序作為 remake 目前的
  Quick legal-candidate order，比 `Fighters()` 的 lexicographic stable ID
  順序更接近 recovered object traversal。這不能證明原始 far-pointer chain
  的每一個 tie 結果。
- `unknown`：`3E01:142Dh` 的 helper 意義、完整 retry／pass 次數、同分候選
  tie、`1701:113Eh`、`+65h／+56h／+66h／+5Ch` 的正式 record 語意，以及
  `00FA:0048h` 後續 target list 的完整順序。

## 實作內容

- Golden Box engine 新增 `combat/quicktarget`：以 `Rule` 宣告
  `candidate_order=legacy_object_id`，驗證 non-zero、唯一 stable ID 與唯一
  legacy object identity，再以副本排序，不改寫呼叫端 slice。
- engine game-pack schema 新增 `combat_ai_target_rules`；CoAB JSON 宣告
  `coab.pc98.quick-target-candidate-chain`。共用 engine 不保存 CoAB 法術名、
  座標、中文或作品旗標。
- CoAB `Battle.FightersInCombatOrder` 保留原始 combat order；Quick 的 area、
  line、Curse、Cause Light Wounds、Protection from Evil／Good 先執行既有
  spell-specific legality，再套用 JSON 宣告的 object order。
- Magic Missile 的 `SelectCombatTarget` 隨機目標與 Cure 的九格／生命值／倒地
  專用規則本輪不改，因為它們需要另外關閉不同的 consumer／tie 證據。

## 驗證

- engine `combat/quicktarget`、`engine` 測試通過：排序不改寫輸入，缺失／重複
  legacy identity 會拒絕。
- CoAB game-pack、`internal/combat`、`internal/game` focused tests 通過；新增
  regression 以故意反轉字典序的 `z-first`／`a-second` 驗證 Quick line target
  依 legacy object order 選取，不依 fighter ID。
- 這只證明資料契約與 bounded adapter 的正常路徑；不證明完整 PC-98 target
  pointer chain、完整 Quick spell AI、敵方施法 AI 或整作可通關。
