# 第二百零三輪：ECL event text localization bridge

狀態：`READY`

## 證據

真實 ECL3 block 16 entry 4 在 `CALL 0x2E10` 後輸出三段 packed text：

```text
SMOKE RISES FROM BEHIND THE RUINED WALLS
OF YULASH. THE SOUND
OF BATTLE RINGS OUT FROM INSIDE
```

原本 State 只把最後一段放進 `OriginalEvent`，不會在繁中事件訊息顯示；這輪
把已驗證句子加入 zh-TW catalog，並將未知 ECL text 保持原文，避免未反組譯
內容被錯誤翻譯。

## Contract

- `localizeECLText` 按原始 segment 順序以空格合併；
- 已知 segment 使用 catalog key；
- 未知 segment 原樣保留；
- 不改變 ECL VM 的 raw `RunResult.Text`，因此反組譯／回歸仍可比對原文。

這是事件文字顯示橋接，不代表全部 ECL 對話已中文化，也不取代後續完整劇情
translation pass。
