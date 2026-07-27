# 第三百一十三輪：熔岩池、十五隻火蜥蜴與防火桶

狀態：`READY`

## 反組譯證據

- ECL5 block `0x32` per-turn entry 在 `C04F & 0x89 == C04D & 0`
  的方向條件成立時觸發；GEO5 `(0,5)` terrain `0x89` 必須面北。
- 事件顯示 PICTURE `0x39`：房內有活躍間歇泉、熔岩池與池中火蜥蜴。
- 接著提供原始 ENCOUNTER 順序 `COMBAT / WAIT / FLEE / PARLAY`。
  已驗證 WAIT 路徑顯示強烈熱浪並建立 MON5 `0x39×15` 戰鬥。
- 勝利後設定 `4C48 & 0x01`，發現六只防火桶並詢問是否派人前往。
- YES 進入 WHO roster selection。一般英雄未通過原 ECL 耐熱條件時會退回，
  再詢問是否換另一人；NO 返回 block `0x32` dungeon。

## 引擎契約

- terrain event 可同時依 cell 與 facing 觸發，不能只以 terrain byte dispatch。
- PICTURE 關閉後仍可能先停一個 PRINT RETURN，再進 ENCOUNTER menu；
  每個 input boundary 都必須保存同一 runtime PC。
- COMBAT continuation 可直接串接 YES／NO 與 WHO，不應回到泛用荒野選單。
- WHO 顯示中文角色名，但回傳原 roster index／ID，不改 script selection。

## 中文與畫面

- 間歇泉／熔岩池、灼熱熱浪、六只防火桶及熱度退回均提供繁中。
- PICTURE 57 與火蜥蜴 icon 57 使用原像素 nearest-neighbour 整數放大；
  中文敘事維持 640×480、24px 高解析字形。

## 驗收

- 正式長流程先完成入口與守門戰，再於原 GEO `(0,5,N)` 觸發事件。
- 驗證 PICTURE 57、四個原始選項、15 隻且僅有火蜥蜴、`4C48` bit 0、
  防火桶 YES、兩人 WHO、英雄熱度退回、拒絕重試及返回 dungeon。
