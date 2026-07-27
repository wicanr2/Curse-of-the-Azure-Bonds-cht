# 第二百零一輪：ECL variable monster operands

狀態：`READY`

## 原始 image 證據

ECL3 smoke／linear scan 觀察到 `LOAD MONSTER` 與 `SETUP MONSTER` 的 operand
使用 `code 0x01`，例如 monster ID／icon 會從 `0xC04F`、`0x7F79` 等 memory
address 讀取；同一類 entry 在此前先執行 `SAVE`、`AND` 或 `OR` 寫入這些欄位。
因此把所有非 literal descriptor 一律拒絕，會在讀到真正遭遇前錯誤停止。

## Contract

- bounded runtime 以既有 `operandValue` 解析三個 numeric descriptor operand；
- `code 0x00` 讀 byte literal，`code 0x01`／`0x03` 讀 runtime memory；
- 每個 descriptor value 必須 `<= 0xFF`，否則回傳明確 error；
- `DecodeMonsterSpawn`／`DecodeMonsterSetup` 的舊 API 仍保持 literal-only，避免
  沒有 runtime context 的 caller 默默讀出零值；
- runtime 使用新增的 `...FromMemory` API，保留每個 ECL block 的 bounded memory。

這只解決 descriptor operand resolution，不宣稱完成 monster table、external
CALL、battle rules 或 ECL party memory。
