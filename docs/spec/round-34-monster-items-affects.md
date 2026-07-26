# 第三十四輪：MON*ITM／MON*SPC records

狀態：`READY`（限 raw record parser）

依 reference 固定大小建立解析器：

```text
Item.StructSize   = 0x3F (63 bytes)
Affect.StructSize = 9 bytes
```

`MON1ITM.DAX` block `0x59` 實際解出 5 筆 item records：type `0x1C`、`0x2E`、`0x23`、`0x3B`、`0x37`，並保留 readied、count、cursed、plus 與三個 affect IDs。`MON1SPC.DAX` block `0x35` 實際解出 2 筆 affect records。

Item 顯示名稱不是 raw `0x2A` name field；reference 以 type／三個 name numbers／hidden flags 組合 item name。因此本輪只保存 raw fields，不宣稱中文 item name 已完成。

- [x] 建立 63-byte item parser。
- [x] 建立 9-byte affect parser。
- [x] 以 MON1 實際 blocks 驗證。
- [ ] 建立 item type／name-number catalog 與繁中名稱。
- [ ] 將 item effects 合併進 monster Fighter／Battle。
