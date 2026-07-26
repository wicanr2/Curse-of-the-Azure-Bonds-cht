# 《青色枷的詛咒》中文金盒子歷史筆記

## 原作定位

《Curse of the Azure Bonds》是 Strategic Simulations（SSI）於 1989 年推出的 Forgotten Realms／Advanced Dungeons & Dragons Gold Box 電子角色扮演遊戲，承接《Pool of Radiance》的冒險。故事圍繞五枚蔚藍枷印展開：隊伍被迫追查枷印的來源，並在達倫地區尋找解除方法。

原作將桌上 AD&D 的角色、法術、裝備、城鎮事件、荒野移動與戰術戰鬥放進 DOS 時代的資料驅動引擎。ECL 腳本、DAX 容器、圖片／精靈資料與 overlay 程式共同組成遊戲內容；這也是本專案採用「反組譯 → 規格 → Go／Ebiten 重製」流程的原因。

## 本專案的保存方向

1. 保存原始資料格式的可驗證觀察，不把猜測寫成事實。
2. 以繁體中文呈現開場、地點、冒險手札、物品與戰鬥介面。
3. 將原本需要翻閱紙本 Adventurer's Journal 才能理解的背景，整理成遊戲內手札與本文件。
4. 保留原始 DOS 映像與掃描手冊作為本地研究素材，不在 repository 重新散布未授權副本。

## 目前里程碑

- 已建立 DAX 解析、ECL bounded VM、繁中 locale、開場／城市／暗影谷 state。
- 已解碼實際 `MON*CHA`、`MON*ITM`、`MON*SPC` records，並從 ECL encounter 建立敵人。
- 已建立基本可操作戰鬥與繁中畫面；`PROGRAM 9` 的 CAMP 目前已被辨識為外部 routine boundary。
- 完整 party creation／import、地圖 tile、CAMP 休息規則、法術、物品效果、存檔與全劇情流程仍待完成。

