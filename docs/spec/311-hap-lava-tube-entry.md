# 第三百一十一輪：哈普地圖、熔岩洞入口與伏擊

狀態：`READY`

## 反組譯證據

- ECL5 block `0x31` 的 dungeon exit entry 在 `4C5E=1` 時，先詢問是否離開哈普，
  再提供 `CAVES / WILDERNESS`。選 CAVES 執行 `NEWECL 0x32`。
- ECL5 block `0x32` initial entry 載入 `LOAD FILES 0x32,2,0xFF`、
  `LOAD PIECES 8,0xFF,0xFF`，從來源 block `0x31` 進入時落在
  `(15,5,W)`，並描述覆滿火山灰的古老熔岩隧道。
- 若 `4C60 != 1`，入口由 MON5 `0x39×4 + 0x31×3` 建立四隻火蜥蜴與
  三名黑暗精靈戰士的強制戰鬥。

## 引擎契約

- `NEWECL` 後除了同步 target `C04B..C04D`，還必須清除來源 movement cycle 的
  `7ED5` exit-attempt 與 `7EC9` forced-move work bytes。否則新 block 戰勝後會誤走
  出口分支，直接把隊伍送回荒野。
- 同一個 selection transaction 內發生 block switch 時也要執行上述同步；不能只在
  dungeon per-turn lifecycle 處理。
- `-lava-tube` 使用真實 block initial entry 建立可重現畫面，不複製事件文字或假造
  renderer state。

## 中文與畫面

- 哈普出口、地圖路線、熔岩洞入口及伏擊敘事均提供繁中。
- `SALAMANDER` 顯示為「火蜥蜴」，保留 MON5 icon `0x39`。
- frontend 的 ECL menu 若同時有 narrative message，640×480 版面會先以 24px
  高解析中文字顯示敘事，再把選項移到下方；原始 sprite／圖片仍採整數倍率
  nearest-neighbour 放大。

## 驗收

- 正式長流程從提爾佛頓一路抵達哈普、招募阿卡巴、解放村莊、選 CAVES、
  打贏七名敵軍，再回到 block `0x32` 地城探索。
- `StartDungeonStoryPreview(0x32,0x31,5)` 真實執行同一 initial entry。
- 實機截圖為 `docs/screenshots/hap-lava-tube.png`，尺寸 640×480。
