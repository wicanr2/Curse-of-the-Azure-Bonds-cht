# 第二百七十三輪：real ECL corpus smoke gate

狀態：`READY`

## 證據

原始 `curseoftheazurebonds.zip` 含 ECL1–ECL6 共 25 個 decoded DAX blocks；每個 block 有五個 `vm_init_ecl` entry，共 125 個入口。2026-07-27 以目前 bounded VM、500-step limit 與無預設 menu selection 重跑後，所有入口皆以正常 EXIT、menu pause、COMBAT、PROGRAM boundary 或 NEWECL transition 返回，沒有 unsupported-opcode error。

ECL5 block `0x30` entry `+0x0014` 另以含 raw item type `0x5E` 的 party context 重跑，第一個 `FIND ITEM` resolved found，控制流進入含 `SUNLIGHT` 的暗精靈裝備腐朽事件。

## Contract

- `TestRealAllInitializationEntriesReachSupportedBoundary` 動態讀取六個 ECL members，要求 25 blocks／125 entries 且每個 `EntrySmokeReport.Err == nil`。
- 原始 image 缺席時 test skip；image 存在但 member、block count、entry framing 或 VM semantics regression 時必須失敗並指出 member／block／entry／PC。
- corpus smoke 證明所有 initialization entries 的目前可達 prefix 已受支援，不等於每條 menu choice、random branch、外部 PROGRAM side effect 或完整玩家劇情都完成。
- 第 277 輪證明「無 unsupported opcode」仍不足以驗證 framing：ADD NPC 少吃 morale
  operand 時曾把 `0x00` 假判為 EXIT，corpus gate依然全綠。現在另以 block 0x52 鎖定
  三筆 NPC request、53 steps、11 段文字、CALL／DELAY counts 與 COMBAT boundary。
- ECL5 real-party test 保護 `PartyContext.ItemTypes → FIND ITEM compare → sunlight text`，避免 synthetic tests 通過但真實 operand／branch framing 失配。
