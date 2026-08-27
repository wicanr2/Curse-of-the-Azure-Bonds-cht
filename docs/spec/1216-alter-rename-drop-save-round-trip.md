# 1216 — ALTER 改名／移除的 remake 存檔往返

狀態：`READY`（限 remake JSON 隊伍存檔與既有 `SAVGAM` 改名證據）

## 問題

`docs/audit/remake-status.md` 原本把角色改名與移除一併列成「尚未 round-trip」。
這已經落後於實作與 spec 254：ALTER 兩條 UI 路徑都存在，`SAVGAM` 改名也已有
保留角色 ID、檔名基底與未知位元組的回歸。缺的是一條把兩個操作串起來，再經過
remake `SavePartyFile`／`LoadPartyFile` 的直接證據。

## 回歸契約

`TestCampAlterRenameAndDropSurvivePartySaveRoundTrip` 建立兩名角色，依正常狀態 API：

1. 把第二名角色改名為「新乙」；
2. 移除第一名角色；
3. 寫出 remake JSON 隊伍存檔；
4. 用新的 `State` 讀回；
5. 驗證 roster 與戰鬥投影都只剩同一個角色 ID，且兩邊名字均為「新乙」。

這條回歸與既有 ALTER 改名同步、移除確認測試一起通過。

## 證據邊界

- 已證實：remake JSON 的改名／移除可通過 `SavePartyFile` → `LoadPartyFile` 往返。
- 已證實：spec 254 所述 `SAVGAM` 改名保留 ID、檔名基底與未知位元組。
- 未證實：`MOVEPARTY` 跨 Gold Box 遊戲轉移。
- 未證實：DOS 角色檔刪除、相關 sidecar 清理與檔案系統語意。

因此狀態報表可以移除「角色刪除／改名尚未 round-trip」這個過時總括句，但不能
把整個 DOS 角色管理或跨遊戲轉移標成完成。
