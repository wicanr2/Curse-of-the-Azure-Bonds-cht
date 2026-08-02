# 第四百二十七輪：PC-98 COMPTARGCURE 順序與 CHARSTATUS（READY）

## 範圍與結論

本輪關閉第 426 輪保留的 Quick Cure 九格順序、相同 HP tie、施法者半血例外、
倒地 8 HP 門檻與不可治療 status 集合。CoAB adapter 已依原版
`COMPTARGCURE` 更新；這不代表 Cure 的 HP 骰、倒地後站起、手動 CAST 延遲或
施法中斷全部完成。

## 輸入與非破壞性

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbols／types／members、DXDIR／DYDIR | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay 13 `COMPTARGCURE` | `exact` |

原檔與抽出的 overlay 全程唯讀；IDA database／報告只在 `/tmp`。符號名稱與
型別是附加證據，原始位置及 bytes 未被修改。`pc98-symbol-audit -members`
只列印 legacy member table，不回寫 executable。

## Borland symbol 與位址橋接

- symbol `COMPTARGCURE = 00B8:1E30h`，與 typed TPOV
  `00B8:0075h → overlay 13 entry 17 → +1E30h` 及函式 prologue／`retf 8`
  交叉一致。
- `0189:004Dh／006Bh／0070h` 經 overlay 32 typed stubs 分別落到
  `FINDOBJECT +07F5h`、`FINDX +12EBh`、`FINDY +1313h`，並與 Borland
  `LOADTACMAP` symbols 一致。
- `DXDIR = 0C29:489Eh`、`DYDIR = 0C29:48A7h`。MZ file offsets
  `114CEh／114D7h` 的九 bytes 是：

```text
DXDIR: 00 01 01 01 00 FF FF FF 00
DYDIR: FF FF 00 01 01 01 00 FF 00
```

因此 exact 掃描順序是北、東北、東、東南、南、西南、西、西北、自身。

## CHARSTATUS 與倒地排除集合

Borland type 1386 是 `CHARSTATUS`，size 1。其連續 member 24–32 的 ordinal
由 member `type` 欄位保存：

| raw | member |
|---:|---|
| 0 | `OK` |
| 1 | `ANIMATED` |
| 2 | `TEMPGONE` |
| 3 | `RUNNING` |
| 4 | `UNCONC` |
| 5 | `DYING` |
| 6 | `DEAD` |
| 7 | `STONED` |
| 8 | `GONE` |

`COMPTARGCURE +1F9Eh` 將 down-player `+196h` status 傳給 set-membership
helper；set constant 位於 overlay 13 `+1E10h`，前兩 bytes `C0 01` 的 bits
是 `{6,7,8}`。`+1FAEh jnz` 排除集合成員，所以可選 `UNCONC／DYING`，不可選
`DEAD／STONED／GONE`。Remake `HealthStatus` 是簡化投影、ordinal 不同；adapter
必須依語意映射，不能直接比較 raw 數字。

## exact 選擇規則

1. 依九格順序呼叫 `FINDOBJECT`；active candidate 必須同 team 且
   `current HP(+1A5h) < max HP(+78h)`。
2. `+1EFFh..+1F3Bh` 只在 current HP **嚴格小於**目前最佳值時取代，所以
   equal HP 保留先掃到的方向。
3. 自身是最後一格；若自身 `current HP < max HP/2`，即使先前目標 HP 更低，
   `+1F04h..+1F3Bh` 仍改選施法者。
4. `Tile_DownPlayer=1Fh` 依 corpse table 座標找合法倒地者；最後若 active
   最佳 HP `< 8`，保留 active，否則合法 down-player 取代。沒有 down-player
   時仍使用 active。

上述 directions、strict branch、自身比較、常數 8、status set、pointer writer
與回傳均為 `exact control flow`。同一格若存在多筆重疊 corpse 時的 table
順序尚無 runtime fixture；目前保留 Battle fighter order，標為
`strong inference`，不可擴大宣稱該罕見 tie exact。

## remake 與驗證

- `quickCureTarget` 依原 `DXDIR／DYDIR` 走九格，不再把所有候選按 HP 排序。
- equal HP 選北不選東；自身低於半血可覆蓋 HP 更低的北側隊友。
- active HP 8 時合法倒地者取代，HP 7 時保留 active。
- `combatHealingTargets` 排除 remake 的 Dead 與 Stoned；Gone 尚無持久 enum
  投影，不能以不存在的值猜測。
- focused tests 覆蓋上述四個邊界；Standing Stone→Red Plume 真實箭傷玩家
  路徑仍是 Quick Cure scheduler／slot integration gate。

## 尚未完成

- 同格多 corpse 的原始 table ordering 與 runtime 畫面 oracle。
- raw `TEMPGONE／RUNNING／GONE` 的完整持久狀態投影與存檔 round-trip。
- Cure 對 unconscious／dying 的完整 health transition、戰鬥 placement 與
  skull／corpse 動態畫面。
- 手動 CAST casting-time、攻擊中斷、slot loss／refund 及其餘 Quick 法術。
