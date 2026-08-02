# 408：內部遺跡二樓、提朗瑟克斯終戰與 PROGRAM 8

狀態：`READY`

## 範圍

本輪完成 ECL／GEO block `43h` 一樓後段、上下樓 transaction、二樓東北角
terrain `9Ah` 的提朗瑟克斯終戰，以及戰勝後 `PROGRAM 8` 勝利存檔 handoff。

這證明從 Standing Stone 已累積的正常玩家長路徑可以完成迷斯卓諾章節；它
不等於全遊戲已從新建角色一路驗證到結局，也不代表 37 名敵人的法術、特殊
能力、AI 與 DOS 動態演出已完整還原。

## 原始資料與位址

- `ECL6.DAX` block `43h`：5,363 bytes；SHA-256
  `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb`。
- `GEO6.DAX` block `43h`：1,026 bytes；SHA-256
  `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c`。
- `MON6CHA.DAX`：SHA-256
  `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b`。
- `PROGRAM 8` 的 engine-level 勝利語意沿用 READY spec 275 的 executable
  consumer 證據；本輪另以真實終戰 ECL continuation 驗證實際 caller。

terrain `86h` 位於 `(0,10)`，raw ECL 只做共用初始化與 redraw，並不是終戰
入口。二樓入口是 terrain `97h` 的樓梯：YES 寫 `(2,5,N)`；terrain `98h`
下樓則寫 `(10,7,E)`。真正終戰在二樓東北角 `(6,1)` terrain `9Ah`。

## 一樓其餘房間

- terrain `90h`：打斷高階祭司召喚班恩神力的儀式；敵軍為
  HIGH PRIEST `48h`×1、PRIEST OF BANE `46h`×3，寫 `4C0Ah=1`。
- terrain `91h`：食品儲藏室，寫 `4C0Bh=1`。
- terrain `92h`：腐朽書庫；沒有 one-shot writer，原作可重複顯示。
- terrain `95h`：停屍架與棺木，寫 `4C0Fh=1`。
- terrain `96h`：防腐室及光影錯覺，寫 `4C10h=1`。

上述文字、樓梯問題與終戰三段台詞均以 stable ID 存在 game-pack；State 與
測試不複製繁中顯示文字。

## 終戰 raw ECL

terrain `9Ah` 依序顯示：

1. 提朗瑟克斯宣稱枷印力量已恢復，命令隊伍跪伏。
2. 隊伍以意志力掙脫，他威脅把眾人碾成粉末。
3. 他承認洛山達護符能傷到自己，但宣稱仍不足以取勝。

之後 exact LOAD MONSTER 組成為：

| MON6CHA | 敵人 | 數量 |
|---|---|---:|
| `45h` | MARGOYLE | 28 |
| `47h` | TYRANTHRAXUS | 1 |
| `48h` | HIGH PRIEST | 8 |

`4CBD／4CC7` 皆為 0 與皆為 1 的 raw ECL 輸出相同，只證明這兩個盟友旗標
不改變終戰三組 LOAD MONSTER。它們會讓途中共用 minion encounter 的四組
count 成為零；原 DOS scheduler 對零敵人 COMBAT 立即勝利並續跑。remake
現在只在真實 dungeon ECL session 中重現這項行為；無 session 的未知零怪物
COMBAT 仍保留 unsupported fallback，避免掩蓋錯誤 continuation。

戰勝後 ECL 立即執行 `PROGRAM 8`。State 依既有 executable 證據恢復全隊、
設定 `GameWon`，並顯示保存勝利進度／不保存的結束選單。

## 正常玩家路徑

第 407 輪犬舍 `(1,9)` 戰後，沿 15 步合法 GEO 路線抵達 `(10,7)`，由
terrain `97h` 上樓至 `(2,5,N)`；再沿十步合法路線抵達 `(6,1)`：

```text
(2,5) → (2,4) → (2,3) → (2,2) → (2,1) → (2,0)
      → (3,0) → (4,0) → (5,0) → (6,0) → (6,1)
```

正式 combat scheduler 打敗全部 37 名敵人，戰後同一 ECL session 直接抵達
`PROGRAM 8`；沒有 teleport、直接設定終戰 PC 或強制勝利。

## Runtime 修正

- `BlockSession` 的 fresh `RunEntry` 現清除 compare registers 與各類互動
  offset，但保存 shared ECL memory 與同一 PRNG 串流。
- `RuntimeState` 保存跨 UI／engine boundary 的 monster setup；LOAD MONSTER
  在 PRESS／PROGRAM 等 pause 前發生時，後續 COMBAT 不再遺失敵軍。
- 新 dungeon movement transaction 清空舊 UI message／choice，避免空白 PRESS
  顯示上一事件文字。

## 驗證

- `TestRealInnerRuinsUpperFloorAndTyranthraxusFinalBattle`：房間、樓梯、兩組
  盟友旗標、37 名敵人與 `PROGRAM 8` raw ECL。
- `TestBlockSessionCarriesEncounterDescriptorsAcrossEngineBoundary`：跨 engine
  boundary 保留 LOAD MONSTER setup。
- `TestRealPlayerPathStandingStoneToBurialGlen`：Standing Stone 起始長路徑、
  一樓、西翼、樓梯、二樓、正式終戰 scheduler 與勝利選單。
- 第 449 輪已由正常 block 初始化、terrain `97h` 樓梯與十步 GEO 路線建立
  可重現的 37 人終戰 checkpoint，並保存 640×480 首領觀察畫面；它取代早期
  直接 ECL 起點且只得到戰敗頁的嘗試。證據等級與 capture-only 鏡頭限制見
  [`spec 449`](449-tyranthraxus-final-battle-camera-checkpoint.md)。

## 尚未完成

- 由全新角色／匯入隊伍從遊戲開場一路完成每章的單一無 debug 通關回歸。
- TYRANTHRAXUS、HIGH PRIEST、MARGOYLE 的完整法術、特殊能力、AI、死亡與
  DOS 動態演出。
- 三神器在 combat formula、提朗瑟克斯魔法抗性與光芒之池消滅敘事中的完整
  executable consumer 鏈；目前不能只憑攻略擴大宣稱規則已還原。
