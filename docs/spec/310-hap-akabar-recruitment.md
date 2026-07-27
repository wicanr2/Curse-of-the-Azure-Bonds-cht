# 第三百一十輪：哈普阿卡巴入隊與祕密商路

狀態：`READY`

## 反組譯證據

- ECL5 block `0x31` terrain `0x8A` 在哈普尚未解放時設定
  `4C5F=1`、`7F7A=1`，顯示 PICTURE `0x3B`，YES 分支執行
  `ADD NPC 0x3B, 0x64`。
- `MON5CHA[0x3B]` 是 38 歲、五級人類魔法師 `AKABAR BEL AKAS`；
  `MON5ITM` 提供兩件裝備，角色記錄保存 11 個 known spells 與 `4/2/1` 法術容量。
- 解放哈普後，長老的法師塔提示會呼叫 block `0x31 +0x0E0A`。此子程序以
  zero-based `LOAD CHARACTER 0..7` 逐人比較 DOS 名稱；找到阿卡巴才顯示
  他知道可繞過法師塔的祕密商路。

## 實作契約

- 顯示名可翻譯成「阿卡巴・貝爾・阿卡什」，但 `Character.ScriptName`
  必須保存 DOS 腳本名稱，供 ECL `0x7C00` 字串比較。兩者不可混用。
- `LOAD CHARACTER` 的低七位是 zero-based TeamList index；bit 7 另保留
  restore／redraw 狀態，不得把 index 0 誤判為無角色。
- 畫面使用 640×480 logical canvas。原始 PICTURE／sprite 採整數倍率
  nearest-neighbour 放大；繁中敘事直接以 24×24 級字形重繪，狹窄欄位才使用
  16×15，不把中文字先壓入 320×240。

## 驗收

- 真實 ECL 長流程驗證入隊旗標、PICTURE、NPC 能力／裝備／法術、旅店返回、
  伊弗利特戰，以及解放後只有阿卡巴在隊時才出現的祕密商路訊息。
- 共用 runtime regression 鎖定 zero-based `LOAD CHARACTER` 與未找到時保留
  上次 selected name 的原版行為。
