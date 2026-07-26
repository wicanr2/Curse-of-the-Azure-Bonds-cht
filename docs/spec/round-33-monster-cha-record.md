# 第三十三輪：MON*CHA monster record

狀態：`READY`（限固定 record offsets 與 raw combat fields）

`MON1CHA.DAX` 的每個 decoded block 是 `422 = 0x1A6` bytes，與 reference `Player.StructSize` 一致。依 reference `Player`／`DataType` offsets 建立 parser：

```text
0x00      Pascal name (15-byte field)
0x73      signed THAC0
0x78      max HP
0x124     base AC
0x126     monster mod ID
0x199     unsigned hit bonus (IByte)
0x19A     unsigned combat AC (IByte)
0x19E/1A0 attack 1 dice count/sides
0x1A2     signed attack 1 damage bonus
0x1A4     current HP
0x1A5     movement/initiative byte
```

實際驗證：

```text
MON1CHA block 0x56:
name=BUGBEAR mod=0xFF hp=24/24 ac=55 attack-bonus=44 damage=2d4+0 initiative=9
```

Parser 可將 record 轉成 `combat.Fighter`，但完整 ECL adapter 仍需處理 `MON*ITM` 裝備、`MON*SPC` effects、攻擊二、AC 顯示轉換與 party battle setup。

- [x] 驗證固定 record size。
- [x] 解析名稱、HP、AC、攻擊與 initiative 欄位。
- [x] 建立 `Record.Fighter` adapter。
- [ ] 接入 `MON*ITM`／`MON*SPC`。
- [ ] 將 ECL spawn sequence 實際建立成 Battle。
