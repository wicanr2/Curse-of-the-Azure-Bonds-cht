# 第 429 輪：PC-98 正傷害施法中斷與法術格消耗

狀態：`READY`（限 `PUTDAMAGE` 正傷害的 pending spell transaction）

## 結論

PC-98 `PUTDAMAGE` 只有在最終套用傷害大於零時，才進入戰鬥施法中斷分支。
若目標 `Action+00h` 有 pending spell，原作會顯示「已無法繼續吟唱法術」、
清除第一個 matching memorized spell byte，再清掉 pending spell。miss、命中但
零傷害、魔法抗性或元素防護使最終傷害為零時，均不走這個分支。

這不是死亡限定，也不是 saving throw 本身觸發；判斷點位於所有傷害修正之後
的 positive applied damage。Action delay 不在此分支清除，角色仍依原本同輪
scheduler 位置繼續。

## 非破壞性輸入

| 輸入 | SHA-256 | 用途 | 等級 |
|---|---|---|---|
| `PC98-GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | Borland symbol、typed stubs | `exact` |
| overlay 23 | `a3ea0d9528be57a92c33fc345baa3e27eef375c84822afba0cfbb141c2faabc9` | `PUTDAMAGE` 與中斷 consumer | `exact` |
| overlay 24 | `0b5d1bebb367414fac93287449a230a36355c075efdc30640a4994bc11d5cd5f` | memorized slot consumer | `exact` |
| `pc98_cast_interruption_audit.idc` | `d6a7969ee2e1518275decdf343c7fcf3e8ed80905a52b3c2adb119f931845415` | 附加指令 ledger | 工具 |

IDA Pro 9.4 只讀掛載原始 overlays；database 與完整 ledger 只寫
`/tmp/coab-ida-429`。腳本不 rename、不 patch，也不覆寫原始位置。下列語意
全部是附加註解，不能取代原始 local offset 與 bytes。

## Exact 指令與資料鏈

Borland symbol `PUTDAMAGE 013E:1FFDh` 對應 overlay 23 local `1FFDh`。

- `204Ch..2053h` 比較最終 `DS:A02Eh`：`<=0` 直接跳到 `22DFh`，略過中斷。
- `219Ah..219Fh` 只在 combat mode `DS:7F27h==5` 進入 Action 分支。
- `21A1h..21A9h` 先清 `Action+01h`；該 raw byte的完整語意仍未知。
- `21B1h..21BAh` 要求 `Action+00h > 0`，即 pending spell 存在。
- overlay 23 Pascal short string `1F76h` 的 CP932 bytes 解碼為
  `は呪文を唱えられなくなった。`，語意是「已無法繼續吟唱法術」。
- `21E6h..21EFh` 把 `Action+00h` spell ID 傳給 far `014A:0070h`。
- typed TPOV resolver 將 `014A:0070h` 解析為 overlay 24 entry 16、local
  `1739h`。`1743h..1772h` 掃描 index `0..53h`，比較
  `Player+index+1Eh`；第一個 matching byte 在 `1766h` 清零，並把 index
  設成 `54h` 結束搜尋。這正是 84-byte memorized spell 區域。
- `21F4h..21FCh` 最後把 `Action+00h` 清零。

因此「正傷害 → 訊息 → 消耗第一個 matching memorized slot → 清 pending
spell」是 `exact data/control flow`。哪個現代 typed target 欄位對應原作
`Action+01h` 仍是 `unknown`；remake 依 transaction invariant 同步清除
`TargetID／TargetPoint`，標為 typed projection，不聲稱 raw layout exact。

## remake mapping

- engine `combat/action.InterruptSpell` 原子清 spell 與兩種 target，保留 delay；
  它不決定何種事件構成中斷。
- CoAB Battle 的單一 positive-damage helper 在實際扣 HP 時建立
  `SpellInterruption{FighterID,SpellID}`；近戰、追加傷害、Magic Missile、
  Fireball、Lightning Bolt／line spell、Cause Light Wounds 共用此邊界。
- CoAB State 在交回下一個玩家控制或結束戰鬥前，依 stable fighter ID 從正式
  roster 移除第一個 matching slot。測試不複製顯示文字，而從正式 locale 的
  `combat_spell_interrupted` stable ID 取得期望內容。
- 中斷沒有 matching roster slot 時仍清 Action，但不虛構另一筆 slot；這保留
  NPC／monster 與不完整匯入資料的 fail-closed 邊界。

## 驗證與未完成邊界

- engine regression：point target 中斷後 spell／target 全清，delay 保留。
- Battle regression：正傷害產生一筆事件；命中零傷害保留 pending spell。
- State regression：`[Bless,Curse,Bless]` 只移除第一個 Bless，訊息由正式
  locale stable ID 解析。
- 第 428 輪手動 Bless／Fireball focused regressions 保留。

尚未證明並因此未擴張的範圍：Cloudkill 低 HD 直接死亡 setter 是否仍通過
`PUTDAMAGE`、沉默／麻痺／石化／睡眠等非傷害中斷、monster memorized slot
raw writeback、原版中斷訊息的畫面停留時間與聲音。完整正常玩家路徑仍需在
可穩定控制 initiative 與敵方命中的原版／remake 場景取得動態 oracle。
