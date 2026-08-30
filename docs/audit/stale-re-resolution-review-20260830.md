# 較早 RE 未回填稽核（2026-08-30）

## 結果

本輪確認一個會改變玩家流程的真遺漏：

- DOS `overlay-17:03EE3h` 在 spec 1093 仍被稱為未知 `sub_3EE3`；後出的 spec 1037
  已用 DOS／PC-98 exact 證據閉合成 `SETACTIVEICON`，卻沒有回填建角骨架。

已修正 spec 1093、建立 spec 1247 並接回 remake。另抽查 `docs/spec/` 中仍含
「沒有宣稱 sub_*」的項目；大部分是刻意保留的內部、硬體停止線或尚無後續 exact
證據，不能只因同名 `sub_83` 出現在別的 overlay 就自動宣告已解。spec 1043 已有
正確的「spec 1073 讀完」勘誤，納入正對照。

## 防遺漏契約

每次把函式或欄位從 unknown／hypothesis 升為 exact／strong inference 時：

1. 使用「平台＋模組／overlay＋原始位址」作不可變鍵；禁止只用 `sub_xxx` 名稱。
2. 在 [`re-resolution-backlinks.tsv`](re-resolution-backlinks.tsv) 登記新 evidence spec、
   較早 spec 與該文件必須出現的勘誤文字。
3. 新 spec 必須列 callers、玩家垂直鏈與被推翻的舊斷言；不能只更新函式台帳。
4. 執行 `tools/re-resolution-backlinks.sh`；缺位址、缺 evidence 或舊 spec 未回填即失敗。
5. 若結論改變玩家流程，另加正常按鍵測試；單元測試不能取代玩家路徑。
6. completion／README 只消費已回填的目前 spec，不直接從歷史 worklist 推斷。

這份台帳刻意要求人工登記「哪一份舊文件受影響」；自動以短函式名全域 join 會把
不同 overlay 的同名 `sub_83` 合併，製造比漏回填更危險的錯誤勘誤。
