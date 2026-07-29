# 第三百五十二輪：眼魔洞穴德克薩姆雙戰與洛山達護符

Status: `READY`

## 目標

延續 ECL4 script block `0x22`／GEO4 geometry block `0x25` 的可操作洞穴，
完成德克薩姆固定決戰、戰利品服務、取得洛山達護符及散提爾堡部隊第二戰，
最後回到同一洞穴，而不是把單人研究 probe 的敗北誤判成 continuation 錯誤。

## 原始檔案與 runtime 證據

- `GEO4.DAX` 只有 blocks `0x20`、`0x21`、`0x25`；洞穴不是 GEO `0x22`。
- ECL4 `0x22` 在 GEO4 `0x25`：
  - `(15,1)` terrain `0x90` 與 `(15,2)` terrain `0x8F` 都可進入固定遭遇；
    focused regression 使用 `(15,1,N)`。
  - PICTURE 49：
    `YOU HAVE MET UP WITH DEXAM AND HIS MINIONS! THE HOODED WOMAN
    REVEALS HERSELF TO BE A MEDUSA!`
  - 兩個 PRINT RETURN pause 後顯示 `THEY ATTACK!`。
  - 第一戰由 `MON4CHA.DAX` 建立：
    - `0x27` HOODED MEDUSA ×1；
    - `0x28` BEHOLDER ×1；
    - `0x29` MINOTAUR ×10。
- 第一戰勝利後先進入 TREASURE service，真實 item block 解析為四件普通
  戰利品；離開 loot menu 後 script 顯示從德克薩姆遺骸取回洛山達護符。
- 下一次 ECL resume 不是安靜退場，而是明確顯示
  `YOU ARE ATTACKED BY FORCES OF ZHENTIL KEEP!`，並建立第二戰：
  - `0x20` ZHENTIL FIGHTER ×11；
  - `0x21` ZHENTIL MAGE ×4；
  - `0x22` ZHENTIL CLERIC ×3；
  - `0x48` HIGH PRIEST ×1。
- 第二戰勝利後按 Continue，回到 ECL4 `0x22`／GEO4 `0x25` 的同一洞穴位置。

## 已排除的錯誤假設

早期 probe 只建立一名 100 HP 英雄。第一戰初始化時敵方先攻自動回合把他降至
15 HP；第二戰初始化時其餘敵方先攻再將他擊倒，因此 Battle 正確回報
`StatusEnemyWon`。Raw post-amulet result 沒有 DAMAGE request，故：

- 不是上一個 DAMAGE packet 重複套用；
- 不是 HP sync 或 ECL continuation 破損；
- 第二批散提爾堡部隊是原作真實戰鬥，不可刪除或改成背景敘事。

正式 regression 使用六名高耐久 deterministic test fighters，讓測試能跨過
兩次合法的敵方先攻，再由測試邊界強制結算敵方 HP；這只縮短測試，不改寫
production encounter、initiative 或 AI。

## 資料與程式邊界

- 四段劇情文字、繁中翻譯與 ECL 原文 matcher 留在 CoAB JSON game pack。
- 梅杜莎、眼魔、牛頭人、散提爾堡職業與大祭司名稱走 locale stable IDs。
- 戰後 continuation 必須呼叫 State 的 data-pack-aware localization；不得退回
  只含舊 Go switch 的 fallback，否則戰前可繁中、戰後卻洩漏英文。
- 共用 engine 只擁有 script/geometry block 分離、VM、monster spawn 與
  battle continuation 機制，不得含德克薩姆或 Area 4 常數。

## 驗收

1. 真實 ZIP、ECL4、MON4CHA、GEO4/25 可重現 `(15,1,N)` PICTURE 49。
2. 揭露、攻擊、護符取得與第二戰訊息全部顯示繁中。
3. 第一戰敵方精確為 1／1／10，第二戰精確為 11／4／3／1。
4. 第一戰勝利後顯示四件 ordinary loot 與離開選項，再取得劇情護符。
5. 第二戰不能被略過；勝利後回到 ECL `0x22`／GEO `0x25`。
6. focused real-image tests、game-pack validation 與完整 regression 通過。

## 本輪邊界

本規格證明遭遇組成與控制流程，並不證明眼魔射線、梅杜莎凝視、所有法術、
弓箭投射動畫或整張戰鬥畫面已達原版 fidelity。這些需要 DOS runtime／原版
影片逐幀 oracle 與各自 READY 規格。
