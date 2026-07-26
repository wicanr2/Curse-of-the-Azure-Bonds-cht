# 第三十五輪：monster item／effect 繁中名稱

狀態：`READY`（限本輪實際觀察的 type／effect IDs）

依 `MON1ITM block 0x59` 與 `MON1SPC block 0x35` 的 raw IDs 接入繁中名稱：

```text
Item type 0x1C  -> 弩矢
Item type 0x2E  -> 輕弩
Item type 0x23  -> 闊劍
Item type 0x3B  -> 盾牌
Item type 0x37  -> 鏈甲
Affect 0x18     -> 偵測隱形
Affect 0x5A     -> 酸液吐息
```

弩矢的 count 會顯示為數量；未知 IDs 不會猜測翻譯，而是顯示 `未翻譯物品／效果(0xNN)`。

- [x] 已觀察 item type 的繁中顯示。
- [x] 已觀察 affect kind 的繁中顯示。
- [x] 未知 IDs 有安全 fallback。
- [ ] 擴充完整 item type／name-number catalog。
- [ ] 將 locale catalog 與 Battle equipment/effects 狀態整合。
