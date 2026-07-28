# 第三百三十一輪：希爾斯法世界轉場與紅羽衛

狀態：`READY`

## 原版資料證據

擊敗法師塔後，隊伍可由艾森布拉循 TRAIL 回立石群，再選
`JOURNEY ON → HILLSFAR → TRAIL`。世界 dispatcher 在這裡由 ECL block
`0x50` 透過 `NEWECL` 切到 `0x51`。TRAIL 觸發第二次 world event 12：

1. BIGPIC 121；
2. `YOU ARE AMBUSHED BY FIRE KNIVES DISGUISED AS FIGHTERS.`；
3. 六名 fighter `0x59`、icon `0x20`；
4. COMBAT；
5. 勝利後寫入 current-location `4C9B=11`，抵達 Hillsfar edge。

世界地圖以 A–N zero-based 編號，因此 Hillsfar（L）是 `11`，不是 State enum、
選單 index 或 ECL block ID。

進城使用 PICTURE 80，場所為
`INN / STORE / HALL / TEMPLE / BAR / LEAVE`。酒館位於碼頭邊，提供
`HAVE A DRINK / RELAX / EXIT`。隊伍仍帶有 Fzoul 的散塔林枷印時，選
`RELAX` 會有紅羽衛故意打翻酒水，要求隊伍清理；拒絕後必定與六名 fighter
`0x59` 戰鬥，勝利返回同一酒館。

隨附攻略與真實 ECL 一致：只要 Fzoul 枷印仍在，希爾斯法酒館的紅羽衛對峙
無論如何都會導致六人戰。本輪固定驗證拒絕分支，不推論另一選項的額外台詞。

## 引擎與中文契約

- `applyCitySelection` 必須接受 Area 1 的 `0..13` world values，並同時同步
  `Location`、`OriginalLocation` 與 `Area.CurrentCity`。
- block `0x50` 與 `0x51` 都是 Area 1 世界 dispatcher owner；跨 block 不得
  清除 shared memory、party、monster chapter resolver 或 selection state。
- 旅途伏擊與酒館紅羽衛戰都使用六名原版 fighter、原版 CPIC 小人與同一個
  combat continuation，不能用敘事佔位取代。
- 城外、城市場所、碼頭酒館與紅羽衛挑釁均以繁中顯示；原始 menu index、
  bond flag 與 encounter data 保持不變。

## 可沿用的 Gold Box 知識

世界地點 ordinal 與 remake enum 是兩個 namespace。應建立明確資料表，而不是
以 enum 數值、城市加入順序或 ECL chapter 猜測；投影時也要同步 engine area 的
current-city byte，讓商店價格、旅店待遇與酒館條件繼續使用原始 script 狀態。

大型世界 script 可能按地理區域拆成多個 ECL block，但仍共享同一 world memory。
`NEWECL` 是 VM continuation，不代表切換成地城或清空 State。其他 Gold Box
作品應從原始 dispatcher 與 world byte 實證 block ownership。

## 驗收

- 同一條 real-session regression 從火刀首領、法師塔、龍巫妖、艾森布拉酒館
  延伸至立石群與希爾斯法。
- 驗證 `0x50→0x51`、六名旅途伏擊、`4C9B=11`、`LocationHillsfar` 與
  PICTURE 80 城市入口。
- 驗證希爾斯法六場所、碼頭酒館、紅羽衛 YES／NO 選單、拒絕後六人戰及勝利
  回酒館。
- unit regression 覆蓋 Area 1 全部 `0..13` world-location 投影。
