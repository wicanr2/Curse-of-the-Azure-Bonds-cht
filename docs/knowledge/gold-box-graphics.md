# Gold Box 圖像格式知識庫

## 共用處理順序

1. 從原始 ZIP 讀取 DAX member。
2. 用 DAX index 取得 block payload。
3. 依 picture header 解析 indexed pixels；SSI picture 的 width 是 `widthUnits × 8`，height 是 rows。
4. masked graphics 將 palette index `0` 映射成 renderer 內部的透明 sentinel `16`。
5. 單層 picture 轉 EGA16 RGBA；多層 icon 先在 indexed domain 合成，再轉 RGBA。

## 戰鬥小人分類

- `CPIC*`／`COMSPR`：可直接顯示的 combat icon／static sprite。
- `SPRIT*`：以 frame count、delay 與 picture frame 組成的動畫 stream。
- `CHEAD`：玩家頭部 layer，通常比 body canvas 矮，左上對齊合成。
- `CBODY`：玩家身體／武器 layer，通常提供 24×24 的 destination canvas。

- `HEAD<area>`／`BODY<area>`：一般場景人物 layer；reference 使用 unmasked DAX pictures，BODY 相對 HEAD 向下 5 rows。

原作另以同一 layer block 加 `0x80` 取得 attack state；direction `> 3` 使用水平翻轉版本。

SPRIT frame 的 `x/y` 是 icon canvas 內的 placement metadata；應在 indexed frame 載入後保留，繪製時再依 renderer scale 套用，不能把它誤當成地圖座標。

PIC／FINAL 的後續 frame 不是獨立 pixels，而是相對於第一幀的 packed-byte XOR；SPRIT 不能套用這個規則。共用 parser 必須以來源 family 明確選擇 full-frame 或 XOR-from-first mode。

ECL 的 `PICTURE` block 是事件層 request，不應在 VM 內直接依賴 Ebiten；目前 `RunResult.PictureRequested`／`PictureBlock` 保持 renderer-neutral，game state 再把它轉成可恢復的 localized event screen。

PICTURE 的 block threshold 是重要的 family dispatch：`block < 0x78` 使用 `PIC<area>.DAX` animation，`block >= 0x78` 使用 `BIGPIC<area>.DAX` 的 unmasked static picture。兩者不能只靠同一個 filename key 或同一種透明設定處理。

PICTURE 還有第二個 dispatch：Area2 `HeadBlockId == 0xFF` 才使用上述 PIC/BIGPIC；若 head block 存在，body block 由 PICTURE operand 提供，改用 `HEAD<area>`＋`BODY<area>` scene composite。

目前可重用的資料邊界是 Area2 raw record `0x5C2` → `area.State.HeadBlockID` → `game.State.SceneHeadBlock`；這讓 renderer 不需要直接讀 DOS record，也讓後續 Gold Box 遊戲可替換自己的 Area state codec。

戰鬥圖示的方向與位置也應分開：direction 是 0–7 的 facing，tile position 是 combat map 座標，screen position 是 camera transform 後的結果。不要用 fighter list ordinal 取代真實 map position；目前 ordinal 只作 deterministic fallback。

共用 fighter adapter 應保存 `pos` 與 `size`，而不是只保存最後的 screen pixel；這樣 camera、occupied tiles、碰撞與不同 Gold Box 戰場尺寸可以在上層替換。

目前已知的 reference position 生成公式可由 `combat.ReferencePlacement` 重用，但它依賴 team／candidate／occupancy 資料；在資料未解出前，不能把 deterministic formation 宣稱為原版 placement。

新建玩家的初始 icon 欄位不是依隊伍 slot 變化：`head_icon=0`、`weapon_icon=0`；種族只先決定 `icon_size`，small races 是 dwarf／gnome／halfling，其餘是 normal。

## 合成規則

對 destination `a` 與 source `b` 的每個 indexed pixel：

```text
a = transparent, b = transparent → transparent
a = transparent                  → b
b = transparent                  → a
a、b 都不透明                   → a OR b
```

這是目前從參考 Gold Box 重寫程式確認的 `MergeIcon` 語意。不要先把透明色畫成黑色，也不要用一般 alpha overwrite 取代 indexed merge。

## 跨遊戲重用介面

後續遊戲可沿用 `internal/dax`、`internal/gfx.Picture`、`gfx.MergePictures`、`gfx.MergePicturesAt`、`gfx.Picture.FlipHorizontal` 與 preview generator，只需建立每款遊戲的 DAX member 名稱、mask color、palette 與 icon layer mapping。
