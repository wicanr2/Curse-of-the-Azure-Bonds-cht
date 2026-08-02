# 第四百二十六輪：PC-98 Quick Cure 指定目標 handoff（READY）

## 範圍與結論

本輪接通 Quick AI 的全域法術 `03h` Cure Light Wounds：原版 selector 先找
可治療的鄰近同隊目標，把 far pointer 帶過 `CASTCOMBATSPELL` 的非即時
action handoff，重新取得施法者行動後才結算並消耗 slot。這不代表所有 Quick
法術、手動 CAST 延遲、完整相鄰格 tie order 或施法中斷均已完成。

## 非破壞性與位址空間

- `GAME.EXE` SHA-256：
  `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0`。
- `GAME.OVR` SHA-256：
  `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a`。
- 原始 executable／overlay 只讀；IDA database 與報告只寫入
  `/tmp/coab-ida-426`。IDC 只附加 research scope，不覆蓋原始位址、bytes 或
  Borland symbols。
- caller `overlay 09:+041Dh` 的 far `00B8:0075h`，經 typed TPOV resolver
  解析為「overlay 13、entry 17、handler `+1E30h`」。其中 `entry=17` 是 entry
  編號，不是 overlay 17；兩種位址空間不可混讀。

## 原版控制流

### suitability

overlay 09 `+03D3h..+04C9h`：

1. `03E8h..03EFh` 先比較 spell record `+0Dh Priority`。
2. spell `03h` 在 `0412h..041Dh` 傳入 caster far pointer 與輸出 target far
   pointer，呼叫 `00B8:0075h`。
3. handler 回傳非零才在 `0426h` 把 suitability 設為 1。

這段 opcode、參數與 branch 是 `exact`；`03h` 身分由全域 spell table 與
既有 Cure catalog 交叉驗證，為 `material-exact`。

### Cure target selector

overlay 13 `+1E30h..+2046h`：

- 以施法者位置周圍九格建立候選，要求目標 team byte 與 caster 相同。
- 一般 combatant 必須 `current HP(+1A5h) < max HP(+78h)`。
- selector 保存目前 HP 較低者；施法者自身低於半血另有優先分支。
- 遇到 `Tile_DownPlayer=1Fh` 時掃描 down-player 表，以座標比對倒地角色。
- 一般候選目前 HP 小於 8 時保留一般候選；否則可由符合條件的倒地候選取代。
- `1FFDh..2026h` 把選中 far pointer寫入 caller 的輸出；`202Ah..2046h`
  依指標是否為 null 回傳布林值。

上述欄位比較、常數、pointer writer 與回傳是 `exact control flow`。九格實際
掃描順序、相同 HP tie order 與 down-player status set 已由 READY spec 427
supersede；本輪原先的 `strong inference` 不再是目前權威狀態。

overlay 09 `+04CCh..+0624h` 是通用 Quick target helper；它沿 Player
`+14Eh` linked list、經 `CHECKTARGET` 與 suitability 保存候選。尾端
`00FA:0048h` 的 TPOV entry handler 是特殊值 `FFFFh`，因此只記錄為
resident／runtime pointer operation，不能擅自命名為 gameplay target setter。

## remake 契約與驗證

- engine `combat/action.State.TargetID` 是 opaque stable identity；
  `BeginTargetedSpell／TakeTargetedSpell` 與 Battle wrapper 保證 spell＋target
  同一 transaction 清除。作品規則不進 engine。
- CoAB `quickCureTarget` 只考慮施法者 3×3 內、未死亡且受傷的隊友；目前以
  stable fighter order 解 equal-HP tie，並依原版 8 HP 門檻處理倒地候選。
- slot 在 pending action 重新入列並實際結算時才消耗；排程期間保留 target
  ID，不借用可被 UI 改動的全域 cursor index。
- focused regression 證明鄰近 3 HP 隊友勝過遠方 1 HP 隊友，target 經 delay
  保存且治療後 slot 消耗。
- 真實資料正常路徑由 Standing Stone 進 Myth Drannor；Red Plume ECL 兩次
  箭傷建立受傷目標，隨後七敵戰鬥由 ALT+M＋Quick 選中 `03h`、保存 hero
  target、完成 pending handoff 並消耗 slot。敵方回合會在同一 headless
  `CombatAct` 內續跑，因此正常路徑不以淨 HP 增量冒充單獨治療證據。

## 尚未完成

- 九格順序、相同 HP tie、施法者半血分支與倒地 status set 已由 spec 427
  關閉；只剩同格多 corpse table ordering 缺 runtime fixture。
- Cure 對 unconscious／dying／animated 的完整原版 placement 與 health-status
  轉移；死亡目標仍必須 fail-closed。
- 手動 CAST 的 casting-time scheduler、施法被攻擊中斷、slot refund／loss。
- 其餘 Quick 法術、非零 MinRange area safety 與完整 AI target pointer。
