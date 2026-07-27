# 第二百八十二輪：ECL code memory 與 Windlord's Inn

狀態：READY

## Reference evidence

`ovr008.vm_GetMemoryValueType` 將 `0x8000..0x9DFF` 分類為目前 ECL block memory；
`vm_GetMemoryValue` 直接讀取 `ecl_ptr[address]`。因此這一區不是一般零值 word map：
載入／切換 ECL block 時必須用 decoded payload 覆蓋 code window。

ECL2 block `0x01` SearchLocation entry 位於 `+0x0286`：

1. `AND 0x7F, C04F → 7F7B`；
2. `DIVIDE ...` 與 `GETTABLE 9DB8,... → 7F7A`；
3. `ON GOTO` 依 GEO selector 派送地點。

`CMD_AndOr` 在寫回 destination 前另呼叫 `compare_variables(result, 0)`。這個隱含
compare side effect 是一次性事件 utility `+0x1CDE` 的必要條件，不可只實作位元運算。

GEO2 block `1` 正式起點 `(7,13)` 往西可達 `(6,13)`；該 cell 的 `x2/C04F=0x86`。
SearchLocation dispatch 到 `+0x04D5`，顯示 PICTURE `3` 與 Windlord's Inn 老闆娘，
第二個 Continue boundary 記錄 Adventurer's Journal Entry 31。PICTURE 發生當下
`Area2.HeadBlockId @ 0x7EE1` 為 `3`，script 隨後立刻改回 `0xFF`；renderer signal
必須在 opcode 當下保存 selector，不能等整段 ECL 結束後再讀。

## Remake transaction

- `BlockSession` 建立、Reset、NEWECL Switch 時重載 `0x8000..0x9DFF` code window；
  其他 Area／player memory 保持共享。
- standalone runner 也為缺少的 code addresses 提供 payload bytes。
- AND／OR 更新 `=, <>, <, >, <=, >=` compare flags（result 對零）。
- 正式 dungeon lifecycle 由 GEO cell 同步 `C04B..C04F`，不硬編旅店座標分支。
- PICTURE signal 保存 opcode 當下的 HeadBlockId，讓 HEAD3／BODY3 原始人物素材在
  640×480 畫面以整數倍放大，繁中文字以 24px 高解析字型獨立繪製。
- 旅店兩段文字與 Continue option 顯示繁中；事件引用 Entry 31 時才將中文全文附加到
  遊戲內冒險手札，未觸發前不提前暴雷；手札頁使用 22 字寬、最多七行自動換行。
- 最後一個 Continue EXIT 回到 `(6,13)` 的 `ModeDungeon`。

## Regression

- synthetic GETTABLE 從 ECL `0x8010` code byte 讀值；
- NEWECL 只保留 code window 以外的 shared memory；
- AND zero result 驅動後續 `IF <>`；
- real image：正式序幕 → GEO west path → PICTURE 3 → 兩段繁中 → Entry 31 unlock →
  手札 UI 可翻到新增頁 → 返回同一地城格；
- 原有安全／危險 CAMP lifecycle 仍在同一真實流程後通過。
