# 403：外圍遺跡羅剎妖房間、坍塌陷阱與下水道

狀態：`READY`

## 問題

ECL6 block `42h` 的 terrain `0Bh／0Ch／8Ah／8Dh` 不是四段孤立文字。
它們包含可拒絕且會重現的誘餌、全隊坍塌傷害、部分地圖座標傳送、兩種大型
羅剎妖戰鬥、戰後財寶、五種交涉態度、手札 57，以及通往 ECL／GEO block
`43h` 的正式下水道入口。若只翻譯第一個畫面，玩家既拿不到手札與財寶，也
無法沿正常流程進入下一張地圖。

## 原始資料

| 來源 | SHA-256 | bytes |
|---|---|---:|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` | 22,175 |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` | 3,697 |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` | 1,424 |
| Adventurer's Journal PDF | `33471b8200d4aab93d2f52065601bebe548cb535a29064a3a0a47c2e833d88c9` | 6,753,869 |

PDF 第 13 頁（印刷頁 23–24）直接保存手札 57。它是羅剎妖提出的交易：
提朗瑟克斯奪走其族人的追隨者並威脅氏族；當玩家攻擊神殿時，氏族會奪回
追隨者，或許能削弱提朗瑟克斯。繁中正文依該頁翻譯，不從事件的一句摘要
自行擴寫。

## GEO 座標

| terrain | 座標 |
|---|---|
| `0Bh` | `(9,2)`、`(11,2)` |
| `0Ch` | `(5,6)` |
| `8Ah` | `(1,3)`、`(1,4)` |
| `8Dh` | `(9,6)` |

第 402 輪結束點 `(11,11)` 到四個事件的正常路線已由
`GEO6/42h.CanMoveDungeonWrapped` 逐步驗證。玩家先到 `(9,6)`，再到
`(9,2)`；坍塌後原作把隊伍移至 `(10,2,N)`，再向西走到 `(1,3)`，最後
經 `(1,4)→(3,7)→(5,7)` 抵達 `(5,6)`。

## terrain 8Dh：羅剎妖交易

- 初見在 `+160Bh` 寫 `4CDAh=1`，再提供
  `COMBAT／WAIT／FLEE／PARLAY`；因此不論採取何種行動，事件都只出現一次。
- `PARLAY` 的五種態度中，只有 `PARLAY_HAUGHTY` 成功。`+1741h` 寫入
  手札旗標 `4CBDh=1`，並顯示手札 57 摘要。
- `SLY／MEEK／NICE／ABUSIVE`，以及直接 `COMBAT`，都建立：

  ```text
  MON6CHA 44h HELL HOUND ×5
  MON6CHA 45h MARGOYLE   ×5
  MON6CHA 43h RAKSHASA   ×6
  ```

- `WAIT／FLEE` 直接返回地城，沒有戰鬥或手札。

## terrain 0Bh：石像鬼門廊陷阱

- 不攻擊石像鬼會立即離開，而且不寫完成旗標；重踏會再次看見誘餌。
- 攻擊會在 `+13C9h` 寫 `4CD8h=1`，殺死兩隻石像鬼後觸發門廊崩塌。
- ECL 接著在 `+13CFh` 寫 `C04Bh=0Ah`，在 `+13D5h` 寫
  `C04Dh=0`，但不寫 `C04Ch`；`+13DBh CALL 2E10h` 的 exact 意義是
  將隊伍放到 `(10,原 Y,N)`。從 `(9,2)` 進入時即為 `(10,2,N)`。
- 全隊坍塌傷害是：

  ```text
  DAMAGE flags=C0h, dice=3d10, bonus=0, saveFlags=01h
  ```

- 羅剎妖靠近時：
  - 裝死會出其不意地迎戰 MON6CHA `43h` RAKSHASA ×1；
  - 不裝死則羅剎妖發現崩塌未殺死玩家，轉身消失。

這個事件修正了第 402 輪過窄的座標提交條件。已驗證的通用 transaction 是：

1. 整次 session 不得跨 block；
2. CALL 前必須新寫 `C04Dh` 作為玩家 facing／position commit；
3. 只投影同 block、CALL 前實際新寫的 `C04B／C04C／C04D`，未寫欄位維持
   State 原值。

Filani 對話會在 `CALL 2E10h` 前把 `C04B/C04C` 當工作值清零，但不寫
`C04D`；正式回歸證明它不得把隊伍移到 `(0,0)`。這個反例也是方向提交標記
不可省略的證據。

## terrain 8Ah：賭局房

- 初見在 `+123Ch` 寫 `4CD7h=1`，顯示男性羅剎妖擲骰、女性羅剎妖休息、
  石像鬼觀戰的房間。
- 玩家可逃走；拒絕逃跑則迎戰：

  ```text
  MON6CHA 45h MARGOYLE ×8
  MON6CHA 43h RAKSHASA ×6
  ```

- 勝利後 exact 財寶服務是：

  ```text
  TREASURE 0,0,0,1200,2000,15,9,81h
  COMBAT
  ```

  依已驗證幣值換算，共 11,200 gold、15 gems、9 jewelry，以及 ITEM6
  block `81h` 的一件 deterministic random item。最後的無怪物 `COMBAT`
  是財寶服務邊界，不是第二場戰鬥。

## terrain 0Ch：下水道入口

- 初見在 `+1505h` 寫 `4CD9h=1`。玩家可放走落單石像鬼，或當場殺死牠；
  兩者都會露出下水道柵口。
- 玩家可拒絕進入；答應後，腦海中的聲音再次詢問是否確定前進。
- 最終答應會在 block `42h` 寫 `C04B/C04C=(15,15)`，執行
  `NEWECL 43h`。block `43h` 隨即：

  ```text
  LOAD FILES 43h,2,FFh
  LOAD PIECES 13,16,3
  ```

  並顯示下水道通往昏暗廚房、坍塌迫使隊伍離開地道。game-pack 以
  `myth-drannor.inner-ruins` 宣告 script／geometry block `43h`，不配置
  固定 spawn，讓原 ECL 寫入的 registers 擁有出生點。State 的正常玩家
  路徑已驗證 ECL／GEO `43h`、`(15,15,N)` 與中文訊息同步完成。

## 實作與驗證

1. 十七段事件文字與完整手札 57 以 stable ID 放進 CoAB game-pack。
2. `TestRealOuterRuinsRakshasaRoomsAndSewer` 直接執行原始 ECL bytes，覆蓋：
   - `0Bh` 拒絕重現、部分座標提交、`3d10`、裝死／不裝死；
   - `0Ch` 放走／殺死、兩層確認及 block `43h` handoff；
   - `8Ah` 兩組十四敵人與 exact 財寶；
   - `8Dh` 五種交涉態度、十六敵人與手札 57。
3. `TestRealPlayerPathStandingStoneToBurialGlen` 從遊戲正常入口沿合法 GEO
   逐步走完以上成功路徑，沒有直接指定事件 entry、強制給物或跳過戰鬥。
4. CALL adapter 回歸分別鎖定完整位置、部分位置、跨 block 拒絕，以及
   Filani 的無方向 scratch-coordinate 反例。

## 完成邊界

本輪完成的是 block `42h` 指定四種 terrain 與通往 block `43h` 的玩家
handoff，不代表內部遺跡、神殿或結局完成。block `43h` 廚房之後的 terrain、
羅剎妖／石像鬼完整 AD&D 特殊能力、坍塌動畫與音效、財寶隨機物的 DOS
選取序列，以及 `4CD7–4CDA` 的原版跨存檔邊界仍待後續驗證。
