# 第三十二輪：ECL monster spawn descriptors

狀態：`READY`（限 ECL command descriptor）

以 ECL1 block `0x51` payload `+0x1293` trace 驗證到第一組接近 `COMBAT` 的 encounter：

```text
+0x1293 SETUP MONSTER 4, 2, 4
+0x129A CLEARMONSTERS
+0x129B LOAD MONSTER 0x59, 4, 0x20
+0x12A2 LOAD MONSTER 0x56, 10, 0x56
+0x12A9 LOAD MONSTER 0x57, 10, 0x57
+0x12B0 COMBAT
```

本輪新增：

- `DecodeMonsterSetup` 解析 sprite、最大遭遇距離與 picture block。
- `DecodeMonsterSpawn` 解析 monster ID、copy count 與 icon block。
- 非 literal operand 會拒絕，避免把 memory pointer 誤當成怪物 ID。
- CLI `-scan-opcode` 可掃描候選 command；CLI trace 現在也輸出完整 operands。

這些 descriptor 尚未包含 `MON*CHA` 的 HP／AC／攻擊資料；那些資料必須由 monster DAX record loader 解碼後才能轉成 `combat.Fighter`。

- [x] 驗證真實 ECL spawn sequence。
- [x] 建立安全 descriptor decoder 與 regression。
- [x] 找到下一個 COMBAT 控制轉移。
- [ ] 解碼 `MON1CHA` 等 monster records。
- [ ] 將 descriptor 與 monster stats 建立 ECL-to-combat adapter。
