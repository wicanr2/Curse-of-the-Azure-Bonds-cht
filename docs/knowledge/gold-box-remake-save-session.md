# Gold Box 重製的 ECL session 存檔原則

Gold Box 存檔不能只保存隊伍與座標。ECL interpreter 可能正停在文字、選單、
戰鬥、財寶或外部 routine boundary；共享 work memory、resume PC 與亂數串流
共同決定下一個結果。

## 最小一致 transaction

```text
原始 game image
  └─ ECL code bytes（玩家自備，載入時重建）

remake save
  ├─ current block + resume PC + stack
  ├─ mutable work/string/compare state
  ├─ input cursors + pending monster descriptors
  └─ PRNG seed + underlying-source draw count
```

只保存 seed 會讓讀檔後亂數回到開頭；只保存 RNG 而不保存 ECL memory，則會
讓同一亂數落入不同劇情條件。兩者都不是 faithful continuation。

## Active combat transaction

戰鬥存檔不能只保存 fighter array。可恢復的最小集合還包含 stable TeamList
order、Action、effect linked-list projection、round、dynamic scheduler entries／
current selection、持續區域、battle-scoped modifiers、待交付 interruption，以及
戰鬥自己的 PRNG seed＋底層 draw count。少掉 scheduler 或 PRNG，即使畫面與
HP 相同，下一次 tie roll、AI 選敵與傷害骰仍會漂移。

作品 adapter 再保存 turn cursor、target／spell／move selection 與尚未開始的
visual transaction。renderer callback、map buffers 與原始素材由載入時重建，
不能序列化函式或複製商業資料。若 wall-clock elapsed 不屬於 State，就應拒絕
mid-animation save，而不是把動畫偷偷從頭播放並聲稱無縫 continuation。

這項 contract 先由 CoAB Sleep 的自然到期／受傷喚醒雙分支證明；再由真實
Standing Stone→Myth Drannor 紅網路徑，在四蜘蛛第一戰 save／全新 State load
後完成同 ECL session 的羅剎妖第二戰與 completion flag。這證明 campaign
combat handoff，不代表其他 encounter 類型或跨作品已驗證；共用仍待第二款
Gold Box。

混合陣營 encounter 還有第二個所有權邊界：ECL 可把 monster combat slot 暫時
設為 party team／QuickFight，但永久 save roster 不能因此新增角色。CoAB
提爾雪雅事件已證明 version 7 應把 `TemporaryAlly` 完整保存在 Battle snapshot，
而 `Characters` 仍只保存冒險隊伍；戰後 continuation 再從 runtime party 移除
臨時 Fighter。若只從 roster 重建 Battle，盟友會消失；若把 Battle party side
全部寫回 roster，則會污染永久隊伍。兩層 snapshot 必須保持分離。

## 商業資料與向下相容

code window 只保存相對於玩家自備 block 的 runtime differences，不複製完整
ECL bytes。舊 save 沒有某欄位時應採版本化相容預設，文件必須如實標成「未曾
保存」，不能根據目前 seed 猜造過去的 draw position。

## 仍需第二款作品驗證

目前 contract 已由 CoAB 真實 Burial Glen 隨機事件驗證；在第二款 Gold Box
採用前仍是 cross-title candidate。下一款需交叉驗證跨 block、選單 pause、
戰鬥 handoff、存讀檔後的下一批 RANDOM，以及長局 replay 上限。
