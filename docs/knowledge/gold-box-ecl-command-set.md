# Gold Box ECL command-set knowledge base

這份文件是後續 SSI Gold Box 作品共用的 ECL 反組譯入口。它把「已知
opcode 名稱／arity」與「已證實 runtime semantics」分開；command table 本身
只能證明 instruction cursor 要吃幾個 operand，不能自動證明 engine side effect。

## 已整理的 command table

| opcode | command | operands | 目前狀態 |
|---|---|---:|---|
| `00`–`02` | EXIT／GOTO／GOSUB | 0/1/1 | bounded control flow |
| `03`–`08` | COMPARE／ADD／SUBTRACT／DIVIDE／MULTIPLY／RANDOM | 2/3/3/3/3/2 | bounded VM |
| `09`–`14` | SAVE／LOAD CHARACTER／LOAD MONSTER／SETUP MONSTER／APPROACH／PICTURE／…／COMPARE AND | 2/1/3/3/0/1/…/4 | partial signal／bounded |
| `15`–`1F` | menus／IF variants／CLEARMONSTERS／party checks | variable／fixed | partial |
| `20`–`2C` | NEWECL／LOAD FILES／surprise／COMBAT／ON branches／treasure／menus／PARLAY | variable | partial signal／bounded |
| `2D` | CALL | 1 | **未實作；需確認 external dispatch 或 code call** |
| `2E` | DAMAGE | 5 | bounded raw signal（flags／dice／bonus／save flags） |
| `2F`–`30` | AND／OR | 3/3 | bounded 16-bit memory destination |
| `31`–`40` | sprite／item／clock／save table／NPC／pieces／PROGRAM／WHO／delay／spell／protection／… | variable | partial signal／bounded |

完整 arity source of truth 是 `internal/ecl/operand.go` 的 `KnownCommands`；
不要在這份文件另抄一份可漂移的完整 table。每個真正加入 VM 的 opcode 都必須
先新增 `docs/spec/<round>-...md` 的 `READY` contract，再加 synthetic 與 real-image
regression。

## Operand contract

目前已由 parser／既有 arithmetic regression 證實：`0x00` 是 byte literal；
`0x01`、`0x03` 可讀 memory word；`0x02` 是 word literal；`0x80` 是 packed
text；`0x81` 是 string-memory word。三 operand arithmetic 與本輪 `AND`／`OR`
使用前兩個 value、第三個 word destination，但這個形狀不應推廣到其他 command。

## Cross-game reuse rule

後續 Pool、Secret、Savage Frontier 等作品應重用 parser、command metadata、
bounded runner 與 evidence report；作品特有的 block namespace、memory layout、
external `PROGRAM` routine、monster table 與 save side effect 必須由各作品 adapter
注入。若同一 opcode 在另一作品出現不同 arity 或 operand code，優先保留 raw trace，
不要覆寫本表的 CoAB assumption。

## Current evidence boundary

ECL1–ECL6 entry smoke 已實際遇到 `0x1D PARTYSTRENGTH`、`0x22 PARTY SURPRISE`、
`0x2D CALL`／`0x2F AND`。`AND`／`OR` 已是 bounded memory operation，`CALL` 已建立
external-call signal；本輪再加入兩個 party-rule destination signal。它們都不把缺少的
作品 party context 猜成 VM memory side effect。

最新 smoke evidence 顯示，補上 variable monster operands 後，ECL3 block 17／18、
ECL4 block 33／37 等真實 entries 已能抵達 COMBAT 並產生 spawn signal；這是 bounded
descriptor bridge，不是完整外部 routine 或玩家流程完成的證明。

`0x2D CALL` 現在也有 reusable external-call boundary：保存 observed word address
並 return 到下一個 ECL instruction。後續作品可重用 signal contract，再依各作品
的 recognized address table 注入 routine handler；不可把它直接改成 `GOSUB`。

`0x2E DAMAGE` 的公開 CoAB reference operand 順序已確認為
`flags, dice_count, dice_size, damage_bonus, save_flags`。VM 現在只保存這五個 raw
numeric values 並繼續 cursor；target selection、saving throw、signed bonus、random
roll 與 party HP mutation 仍由作品 adapter 實作，避免把 ECL flags 誤當成一般攻擊。

`0x34 ECL CLOCK` 是 reference `vm_LoadCmdSets(2)` 的兩個 numeric operands：
`timeStep`、`timeSlot`。CoAB 的 bounded VM 現在輸出 raw `ClockRequest`，State 再呼叫
共用 `AdvanceGameTime(timeSlot, timeStep)`；這個「VM signal → 作品 time adapter」分層
可重用到其他 Gold Box 遊戲，但各作品仍須重新驗證 slot scale 與事件觸發規則。

`0x35 SAVE TABLE` 的 reference operand contract 是 `value, address, offset`：先把
前兩個 numeric operand 解成 value 與 destination base，再把第三個 numeric operand
加到 destination，將 value 寫入 `memory[address+offset]`。這與 `0x2A GETTABLE`
的 indexed read 成對；bounded VM 已實作兩者，並以 synthetic memory regression 保護
「operand 可是 literal 或 memory value、offset 可是 literal 或 memory value」的語意。
這是可跨 Gold Box 重用的 raw ECL memory primitive，作品專屬的 table schema 仍須另行
由反組譯證據建立。

`0x24 COMBAT` 也是一個 resumable engine boundary：VM 會把 PC 推進到 COMBAT 後的
instruction，再把控制權交給 battle loop。State 的 adapter 在 party victory 後以同一個
`RuntimeState` 續跑，因此可以接回 ECL 的 text、menu、PICTURE 或 `NEWECL`；direct-entry
戰鬥沒有 ECL session 時則維持一般結果畫面。這個 contract 對後續 Gold Box 作品比「戰鬥
結束就回地圖」更接近原版事件 continuation。

ECL event text 也採同一 evidence discipline：只有已由 raw image 解出的 segment
才進入作品 locale catalog，未知句子維持原文，避免跨作品誤套 CoAB 翻譯。

事件 segment 與 menu pause 是兩個不同 observable outputs；State 必須先保存 text，
再切換 choices／waiting state。這項 ordering 可供後續 Gold Box 遊戲共用。

`PRINT RETURN` 同樣只保存 renderer-facing cursor signal，不應誤當成 VM `RETURN`；
它不觸碰 ECL call stack。

`FIND ITEM`／`DESTROY ITEMS` 是另一組可跨作品重用的 inventory boundary；VM 只
保存 item IDs，不應在缺 party roster context 時自行改 inventory 或 compare flags。

`PARTYSTRENGTH (0x1D)` 與 `PARTY SURPRISE (0x22)` 現在由 bounded VM 保存已驗證的 word
destination request 並繼續 cursor；State 注入 `PartyContext` 時會依 reference 計算並寫回
shared ECL memory，沒有 roster context 的純 VM path 則維持 unresolved request，不寫入
猜測值。

`CHECKPARTY (0x1E)` 已釋放三個 reference branches：normalized selector `0xA5..0xAC`
查 thief skills、`0x9F` 查 movement、`8001` 查 active affect；四個 destination 的
min／max／average／found ordering 由 `PartyContext` 寫入。未知 selector 仍只保存 raw
request，這個規則可跨 Gold Box 重用，但各作品需重新驗證 selector table。

`WHO (0x39)` 是一個獨立的 character-selection boundary：它使用目前 prompt 呼叫
`selectAPlayer`，不是從 operand 產生選項。VM 現在保存 `WhoRequest.Prompt`，interactive
path 會 pause，State 以 roster UI 消費 selection，再由 shared `RuntimeState` resume；
selected player ID 已保存，但其他 global routine side effects 仍需各自驗證。

`LOAD CHARACTER (0x0A)` 與 WHO 不同：reference 從 operand value 取得 1-based player
selector，低 7 bits 對應 `TeamList[selector]`，bit 7 是 restore／party-summary redraw
flag。共用 VM 會同時保存 raw word address 與 `LoadCharacterRequest`；CoAB State 已將
有效 selector 接回 persistent roster 的 selected player，無效 selector 保留 not-found
狀態。`FreeCurrentPlayer`、external string context 與 redraw side effects 仍須各作品
依 reference 逐欄接線。

目前 CoAB 的 State adapter 已把 verified `DESTROY ITEMS` IDs 廣播到 persistent
party roster；這是 ECL effect 的明確 mutation，與玩家操作用、會保護 readied item
的 `Character.RemoveItem` 不同。後續 Gold Box 作品可沿用「VM signal → 作品 party
adapter」分層，但仍須各自驗證 item type namespace、compare result 與角色範圍。
