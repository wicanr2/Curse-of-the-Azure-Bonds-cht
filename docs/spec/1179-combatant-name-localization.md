# 1179 戰鬥員名字：68 種全部有繁中，而且由測試守著

狀態：`READY`

## 結論

原版六章 `MON*CHA.DAX` 一共 **68 種相異的戰鬥員名字**。remake 的翻譯管道是
game pack 的 `combatant_name_rules`（原始名 → message id）＋兩份語系檔
（message id → 譯名），入口是 engine 的 `LocalizeCombatantName(source, locale)`。

本輪之前只宣告了 **23 種**，其餘 45 種在戰鬥選單、命中訊息、目標清單裡都是英文
（`THRI-KREEN`、`NEO-OTYUGH`、`MARGOYLE`…）。現在 68 種全部有 `zh-TW` 與 `en`。

## 為什麼要有閘門

`LocalizeCombatantName` 查不到宣告時**原樣回傳原始名**，`found` 是 false 而呼叫端
（`combat_state.go`、`creation.go`）只在 true 時覆寫。所以漏一種的後果是

- 沒有錯誤訊息、沒有 log、測試不會紅；
- 玩家在戰鬥畫面直接看到大寫英文。

這正是「查詢回空 ≠ 不存在」那一類：要用**正對照**回答，不能靠翻閱。

## 閘門長什麼樣

`cmd/combatant-name-audit` 產 `docs/audit/combatant-names.md`，並附三條測試：

| 測試 | 擋住什麼 |
|---|---|
| `TestEveryOriginalCombatantNameIsLocalized` | 原版有、pack 沒宣告或沒翻譯 |
| `TestEveryDeclaredSourceExistsInTheOriginalData` | 宣告的 `source` 拼錯 ⇒ 那條規則永遠不生效 |
| `TestCombatantTranslationsAreDistinct` | 兩種生物撞同一個譯名 ⇒ 玩家分不出誰是誰 |

⚠ 第一條走的是 pack 的**執行時查詢**，不是比對兩份 JSON 的鍵：宣告與翻譯各自
對得起來、但查詢路徑串不起來的話，比 JSON 一樣會過。

突變驗過：把一條規則的 `source` 從 `THRI-KREEN` 改成 `THRI KREEN`，第一條與第二條
同時紅，而且各自從不同方向指出問題。

## 譯名的來源

`docs/knowledge/coab-glossary.md` 是專有名詞的權威表，而它明文規定「上場戰鬥的
怪物名不寫在表裡，唯一來源是 `combatant_name_rules`，稽核時自動併入」。所以
新增的 45 種要與表裡**已經有的敘述用譯名**一致——`cmd/glossary-audit` 在本輪
一口氣抓到 7 條不一致：

| 原文 | 本輪初稿 | 表裡既有（採用） | 出處 |
|---|---|---|---|
| `DISPLACER BEAST` | 位移獸 | **移位獸** | 手札 58 |
| `OTYUGH` | 歐提烏格 | **奧提尤格** | 手札 21 |
| `THRI-KREEN` | 螳螂人 | **斯瑞克林** | 手札 45 |

## 原文別名

另外三條是 `duplicate_rendering`：`AKABAR BEL AKAS`／`Akabar Bel Akash`、
`FIRE KNIFE`／`Fire Knives`、`WORG`／`warg`。**那不是重複**——遊戲資料裡的名字
常常是同一個詞的截短或單複數寫法（`MON*CHA` 的名字欄有長度上限），譯名相同是
正確的。

所以譯名表加了「原文別名」：備註欄寫「原文另作 \`X\`」，稽核就把 `X` 折進正式
寫法那一組。折疊之後這三對不但不再誤報，還**多了一道一致性檢查**——哪天有人
把其中一邊改掉，`conflicting_rendering` 會報。

## 順帶修掉的

- `item_name_E8`（`Efreeti`）本輪初稿寫「火巨靈」，與 `combatant.efreeti` 的
  「伊弗利特」不一致，已對齊（spec 1178 那張表的來源不同，所以先前沒被抓到）。
- 五個測試在斷言**顯示名**（`fighter.Name == "FIRE KNIFE"`）。顯示名是會被翻譯的
  欄位，身分要看 `SourceName`——原版名字在整條鏈上都保留著。改成 `SourceName`
  之後這些測試對「有沒有翻譯」免疫，而它們本來就不是在測翻譯。

## 回歸

`docs/audit/combatant-names.md`：**68 種、沒宣告 0、缺譯 0**。
`docs/audit/glossary.md`：詞條 138、不一致 **0**。

## 不宣稱

- 那 68 種以外的名字（劇情文字裡提到但不上場的生物）是否都在譯名表裡——
  那是 `glossary-audit` 既有的職責，本輪沒有擴張它的掃描範圍。
- 戰鬥員名字在原版是不是還有第二個來源（例如 ECL 直接指定的臨時名字）。
