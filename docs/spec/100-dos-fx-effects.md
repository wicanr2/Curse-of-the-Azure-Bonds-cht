# 第一百輪：DOS `.FX` effects 匯入

狀態：`READY`（限 9-byte effects stream 的保存與繁中名稱）

公開 CoAB PC creature effects format 將角色 `.FX` 與 monster special effects 表示為連續的 9-byte records：第一 byte 是 effect kind，第二／三 byte 是 little-endian remaining minutes，第四 byte 是 active effect strength，其餘四 bytes 是 effect-specific data。

本輪新增：

- `DOSPlayerRecord.ApplyEffects`：把外部 `.FX` stream 解成 `monster.AffectRecord`。
- `Character.ApplyDOSEffects`：保存 imported effects，並由 `DOSPlayerRecord.Character()` 帶入 party JSON／runtime model。
- 擴充 `monster.ChineseAffectName` 的常見效果繁中名稱。

parser 嚴格要求資料長度是 9 的倍數；空 stream 合法。這輪只保存 effects，不直接修改 HP、AC、能力或戰鬥狀態，避免把 effect strength／duration 的規則猜進共用 parser。pointer address-space、FX slot lifecycle、effect tick／解除與完整 save container 仍待後續反組譯。

格式參考：[Curse of the Azure Bonds creature effects list](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)。
