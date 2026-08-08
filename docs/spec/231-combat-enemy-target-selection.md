# 第二百三十一輪：combat enemy target selection（READY）

## Reference evidence

CoAB reference `engine/ovr014.cs` 的 `find_target` 會呼叫
`ovr025.BuildNearTargets(max_range, player)`，由 `ovr032.Rebuild_SortedCombatantList`
建立對立隊伍的可達目標清單；在沒有既有 target、且未納入本輪未解出的 visibility／wall
條件時，最多 20 次從清單中擲骰選取目標。`engine/ovr010.cs` 的 monster-turn path
會先呼叫 `find_target(false, 1, 0xff, player)`，再執行攻擊或 guarding。

## 本輪 contract

- `combat.Battle.SelectCombatTarget(attackerID, targetSide)` 只從存活、指定 side 的
  fighter 中選一個目標。
- candidate 在完整 `LegacyObjectID` projection 存在時依一基底原始 combat-object
  身分排序；projection 不完整時才以 fighter ID 作 deterministic fallback，使用
  Battle 既有 seeded RNG 選取，並避免 Go map iteration 造成非 deterministic 結果。
  這項 bounded tie projection 由第 507 輪補上，不等於完整原始 candidate list。
- State 的 enemy physical attack 與 multi-attack sequence 在該 enemy turn 開始時選一次
  target；同一回合的剩餘攻擊維持同一 target，target 倒下才沿用既有 next-living
  cursor fallback。

## 明確 boundary

本輪不宣稱已解出 `Rebuild_SortedCombatantList` 的牆面／路徑距離、visibility、AI
spell priority、guarding、flee、persistent `Action.target` 或完整 monster turn。
這些欄位仍需各自的 reference evidence 與 adapter；本輪只取代 remake 目前明確錯誤的
「敵方固定攻擊 party[0]」行為。

## Verification

- combat unit test：同一 seed 與相同 candidate order 得到相同 target；不同 candidate
  可被選到。
- game regression：enemy turn 不再固定選 party roster 第一位，且 target 倒下後
  multi-attack 仍切換至下一個存活 party。
