# 第一百零五輪：DOS `.FX` combat projection

狀態：`READY`（限 Bless／Curse 的 unconditional attack modifier）

依 CoAB combat rules reference：

- Bless（effect kind `0x01`）使 THAC0 `-1`，等價於 fighter `AttackBonus +1`。
- Curse（effect kind `0x02`）使 THAC0 `+1`，等價於 fighter `AttackBonus -1`。

`party.Character.Fighter()` 現在只對 `Active` effect 做這兩個 projection；因為 `FighterWithEquipment` 以它為基底，DOS-imported effects 會同時影響有／無 ITEMS catalog 的角色。

本輪刻意沒有套用 Bestow Curse、Prayer、Protection、Blind、Haste 或 morale／saving throw effects：它們需要 target、alignment、combat phase 或其他 state，不能由單一 party fighter projection 安全推導。未來 rules layer 應在 action／target 邊界處理。

格式與規則參考：[CoAB creature effects list](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)。
