# 1118 — 整局存檔的往返閘：20 個位置參數是這條測試存在的理由

- 證據等級：`exact`（機制層，不涉原版語意）
- 對應工作項：`ENG-10`

## 問題出在哪

`State.SavePartyFile` 用 **20 個位置參數**呼叫 `EncodeGameWithAdventureState`：

```go
partySave.EncodeGameWithAdventureState(s.partyRoster, areaState, uint8(s.Mode),
    uint8(s.Location), s.MapX, s.MapY, s.DungeonX, s.DungeonY, s.DungeonDirection,
    s.DungeonWallType, s.DungeonWallRoof, s.gameClock, s.gameAgeCycles, …)
```

新增欄位忘了傳、或同型別的兩個參數對調，**編譯器都不會吭聲**。症狀是「讀檔之後
某個狀態悄悄不見了」或「地城座標的 X 與 Y 對調」，而且離成因很遠。

## 閘的形狀：存檔 → State → 存檔

`TestGameFileSurvivesAStateRoundTrip` 組一份**每個欄位都有值**的存檔，
讀進 `State`，再存出來，逐欄比對。掉的欄位當場現形，不需要有人事先想到它。

新增欄位不必登記任何清單——反射預設就會比它。

## 豁免只給快照指標

四個欄位進不了往返：`Combat`（要活的 `Battle`）、`Music`／`OneShotAudio`
（要音訊裝置）、`ECLSession`（要載入中的 ECL block）。

豁免不是白名單就算了：測試會檢查被豁免的欄位**型別是指標**。純量欄位沒有
「存不出來」這回事——它一定是被漏掉了。

## 順手釘住一個容易誤判的重複

時鐘存兩份：頂層的 `GameTime` 與 `Area.GameTime`。存檔時後者由 `s.gameClock`
覆寫，所以兩邊必須一致。**只填一邊會在往返之後對不上，而那不是掉欄位，
是這份重複本身**——測試的 fixture 兩邊都填，並在註解裡寫明原因，
免得下一個人把它當成 bug 去「修」。

## 與 `Fighter` 那道閘的分工

| 閘 | 守什麼 | 位置 |
|---|---|---|
| `TestEveryFighterFieldSurvivesASaveRoundTrip` | 戰鬥員的每個欄位 | `internal/combat` |
| `TestGameFileSurvivesAStateRoundTrip` | 整局存檔的每個欄位 | `internal/game` |

兩者都是反射掃描：**擋的是下一個新增的欄位**，不是已經想到的那些。
