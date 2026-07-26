# 格式規格

本目錄只收錄由原始映像、執行觀察、反組譯或可重現工具支持的規格。

狀態定義：

- `DRAFT`：有初步觀察，但仍有未驗證的關鍵假設。
- `READY`：可供實作，欄位、邊界、錯誤行為與驗證案例已足夠明確。

目前規格：

- [第一輪素材與格式盤點](./round-01-inventory.md)（`DRAFT`）
- [第二輪 DAX 容器與 ECL 文字取樣](./round-02-dax-container.md)（`DRAFT`）
- [第三輪 DOS loader 與 GAME.OVR](./round-03-loader-and-overlay.md)（`DRAFT`）
- [第四輪 Go 核心層與繁中資源](./round-04-go-core.md)（`READY`：限 DAX container／locale）
- [第五輪 ECL operand framing](./round-05-ecl-operands.md)（`DRAFT`）
- [第六輪 ECL 安全 trace walker](./round-06-ecl-trace.md)（`DRAFT`）
- [第七輪 ECL packed text extraction](./round-07-ecl-text.md)（`DRAFT`）
- [第八輪繁中開場狀態核心](./round-08-localized-state.md)（`DRAFT`）
- [第九輪 Ebiten opening prototype](./round-09-ebiten-opening.md)（`DRAFT`）
- [第十輪 ECL data-driven opening](./round-10-data-driven-opening.md)（`DRAFT`）
- [第十一輪 ECL branch target graph](./round-11-ecl-branch-graph.md)（`DRAFT`）
- [第二十六輪 ECL5 real NEWECL regression](./round-26-ecl5-newecl-regression.md)（`READY`：限已定位 transition entry）
- [第二十七輪 Shadowdale 荒野入口](./round-27-shadowdale-map-entry.md)（`READY`：限入口狀態與輸入 contract）
- [第二十八輪 Shadowdale 場所 menu](./round-28-shadowdale-place-menu.md)（`READY`：限選項與 state-level event contract）
- [第二十九輪 TREASURE bounded prefix](./round-29-treasure-prefix.md)（`READY`：限 operand framing 與安全前綴）
- [第三十輪 COMBAT request signal](./round-30-combat-request.md)（`READY`：限控制轉移 signal）
- [第三十一輪 AD&D combat core](./round-31-combat-core.md)（`READY`：限可注入骰點的核心規則）
- [第三十二輪 ECL monster spawn descriptors](./round-32-ecl-monster-spawn.md)（`READY`：限 ECL command descriptor）
- [第三十三輪 MON*CHA monster record](./round-33-monster-cha-record.md)（`READY`：限固定 record offsets 與 raw combat fields）
- [第三十四輪 MON*ITM／MON*SPC records](./round-34-monster-items-affects.md)（`READY`：限 raw record parser）
- [第三十五輪 monster item／effect 繁中名稱](./round-35-monster-item-localization.md)（`READY`：限本輪實際觀察 IDs）
- [第三十六輪 ECL-to-enemy encounter adapter](./round-36-ecl-encounter-adapter.md)（`READY`：限 enemy fighter 建立）
