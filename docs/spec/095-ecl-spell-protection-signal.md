# 第九十五輪：ECL SPELL／PROTECTION signal

狀態：`READY`（限 bounded operand decode 與 runtime signal）

reference command table 定義 `SPELL <spellID> <address1> <address2>`：尋找持有法術的角色，並將 spell slot／character index 寫入兩個 address；`PROTECTION <address>` 進入 copy-protection routine。`RunResult.SpellSearches` 現在保存 spell ID 與兩個 address，`ProtectionRequests` 保存 address，讓後續 party spell-slot resolver／copy-protection adapter 接入。

bounded runner 此輪不虛構 party spell slots，也不寫入 `address1/address2`；它只消耗正確 operands 並回傳 signal，避免在尚未載入原始 party spell record 時把「找不到法術」誤判成真實結果。

驗證：`internal/ecl/runtime_test.go` 覆蓋 SPELL 三 operands、PROTECTION address、word order 與正常 EXIT continuation；`go test ./...` 應通過。

格式依據：[Gold Box ECL command reference](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365#12.3.3)。
