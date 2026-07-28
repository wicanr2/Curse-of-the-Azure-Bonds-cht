# 第三百三十六輪：摩安德之坑下層與散塔林命令

狀態：`READY`

## 原版資料證據

摩安德之坑兩層共用 GEO3 block `0x11`，以不同 ECL block 解讀同一組 terrain：

- ECL3 block `0x11` 的 terrain selector `0x06` 顯示向南下行的樓梯；
- YES 執行 `NEWECL 0x12`；
- ECL3 block `0x12` initial entry 在由上層進入時，將位置設為
  `(15,14)`、direction `2`（State half-direction `4`）；
- 下層同一格 terrain `0x81` 在 block `0x12` 顯示北牆向上的樓梯；
- 相鄰 `(14,15)` terrain `0x11` dispatch 到 `+0x0116`，顯示死去的
  ZHENTRIM fighter；
- `EXAMINE CORPSE` 分支發現帶有 Zhentil 印璽的正式命令，解鎖
  Journal Entry 46；
- plot bit 完成後重跑 SearchLocation 直接 EXIT，不得重播事件。

## 引擎與中文契約

- `NEWECL 0x11→0x12` 後，作品 adapter 必須套用下層 initial entry 的落點，
  不可沿用上層樓梯 `(15,11)`。
- ECL `0x11` 與 `0x12` 都保持 `GeoMapSet=3`、`GeoMapBlock=0x11`；不可因
  script block 改變而猜成不存在的 GEO block `0x12`。
- 保留原版樓梯選單順序 `NO / YES`。
- `EXAMINE CORPSE` 顯示為「檢查屍體」，屍體、印璽與手札 46 提示完整繁中。
- 本輪只證明命令已抄入手札 46，不從事件短文虛構命令全文。

## 可沿用的 Gold Box 知識

ECL block 與 GEO block 不保證一對一。不同樓層可共用同一張 GEO grid，以
NEWECL 改變 terrain dispatch 語意；跨作品 adapter 必須保存
`(ECL set/block, GEO set/block)` 兩組 identity，不能只維護單一 map ID。

## 驗收

- 延續同一 real-session，由 Pit level 1 terrain `0x86` 選 YES。
- 驗證切至 ECL `0x12`，仍使用 GEO3 block `0x11`，落點為 `(15,14,4)`。
- 在下層樓梯選 NO，再走到 `(14,15)`。
- 驗證屍體選單、`EXAMINE CORPSE`、Zhentil 印璽與 Journal Entry 46 提示繁中。
- 返回同格重跑 lifecycle，事件不得重播。
