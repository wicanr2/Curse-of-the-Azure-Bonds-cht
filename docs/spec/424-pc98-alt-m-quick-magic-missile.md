# 第 424 輪：PC-98 ALT+M 與 Quick Magic Missile

狀態：`READY`（限 ALT+M gate、selector、全域 `0Fh` Magic Missile instant cast）

## 結論

PC-98 `ALT+M` 切換 `DS:A86Ch`；可控制 PC 只有旗標非零時才讓 Quick AI
考慮法術，`ControlMorale > 7Fh` 的 NPC 則略過此 gate。selector 從 Player
memorized slots 以全域 spell ID 執行有界 priority 搜尋。Magic Missile
`0Fh` priority 4、`CastOn=1`、`MinRange=0`，可走已實作的 instant spell
target／visual／damage／slot-consumption 路徑。

本輪不把 ALT+M 擴張成完整法術 AI。`MinRange>0` 的 area suitability 與
casting delay 尚未接完；抽到這類或其他未接法術時，remake 會明確收回玩家
控制，不偷偷改成物理攻擊或另一個法術。

## 非破壞性輸入

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbols、resident stubs、spell records | `exact` |
| `PC98-GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | typed TPOV resolution | `exact` |
| overlay 09 | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | Quick selector／suitability | `exact` |
| overlay 13 | `db1c03daaa984b94056e67d7b9e10fdc6bd3f2393f8a9e5ec720703f93835d4a` | `CASTCOMBATSPELL` | `exact` |

IDA Pro 9.4 只讀取 `workplace/ida406/pc98-overlays/`，資料庫與報告寫在
`/tmp/coab-ida-424`。versioned IDC 只附加研究 scope label；未 rename／patch
原始 image、overlay 或基準 database。

## selector 指令證據

overlay 09 local `0627h..0754h`：

1. `0631h..0679h` 走 Player spell list `+1Eh..+71h`，把每個非零 byte 原樣
   收進最多 `53h` 筆候選；沒有職業基底轉換。
2. `067Fh` 初始 priority 7；`0683h..068Eh` 擲 `1d7` tiers。
3. `069Eh..06AEh`：PC 需 `A86Ch!=0`，NPC bypass。
4. `06B0h..06C3h` 要求對方隊伍仍有人。
5. `06D3h..0729h` 每個 priority 恰抽三次 `1d spellCount`；呼叫 local
   `03D3h` suitability，失敗三次後 priority 減一。
6. `072Bh..0742h` 若選到 spell，呼叫 far `00B8:007Fh` 並回傳 result；否則
   回傳 0。

typed TPOV resolver 把 `00B8:007Fh` 落到 overlay 13 entry 19、handler
`27A1h`；Borland symbol 是 `CASTCOMBATSPELL`。該 routine 讀 record
`+0Bh WhereCast`、`+0Ch CastingTime`；`CastingTime/3 == 0` 時呼叫
`CASTSPELL`，否則寫 pending casting action。這證明 selector 成功不等於所有
法術都可同步立即結算。

## suitability 與推論等級

- `03D3h..04C9h` 讀 record `+0Dh Priority`、`+0Eh CastOn`、`+0Fh
  MinRange`：`exact`。
- Magic Missile record `0Fh @ file 012ED4h` 是
  `02 01 06 04 00 00 04 00 00 04 00 01 01 04 01 00`；priority 4、
  `CastOn=1`、`MinRange=0`：`exact`。
- `02D3h..03D0h` 的非零 MinRange helper 會建立落點範圍，逐 candidate 比較
  team，讀 spell record `+08h／+09h` 並呼叫 effect／save predicate；它不是
  單純幾何距離：`exact control flow`。完整 helper 名稱與所有回傳條件仍是
  `hypothesis`，故本輪 fail-closed。
- Magic Missile 的敵方目標目前沿用 Battle 同一 PRNG 的存活敵人抽選；與
  原始 `CASTSPELL／GETSPELLTARGETS` 的完整 target ordering 尚未逐指令關閉，
  標為 `strong inference`，不可宣稱 exact tie order。

## remake 與驗證

- engine `combat/quickspell` 保存 `1d7`／三次候選／priority 遞減與重複 slot
  權重；CoAB JSON `combat_ai_spells` 保存原始 metadata。
- `ALT+M` 在每場 combat initialization 重設為 off；Ebiten 先辨識 Alt+M，
  不與一般 `M` 移動衝突。
- State 只在 Quick party fighter 入口呼叫 selector，使用 Battle 持續 PRNG；
  Magic Missile 經既有 Begin／Cast、視覺與聲音 handoff。
- engine selector／schema、CoAB game-pack anchor、toggle reset、global `0Fh`
  slot consumption 與 visual event 均有 deterministic tests。
- 正常玩家路徑從 Standing Stone 旅行至 Myth Drannor、沿 GEO6 到紅網四蜘蛛
  戰，啟用 ALT+M＋ALT+Q，實際消耗 `0Fh` slot 後續跑至羅剎妖揭露。

尚未完成：完整 nonzero MinRange area safety、Cure special target、所有法術
casting delay、NPC 未支援 spell handoff、原始 target tie order 與 ALT+M
DOS／PC-98 畫面文字的 runtime 截圖。
