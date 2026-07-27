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
| `2E` | DAMAGE | 5 | 未實作 |
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

ECL1–ECL6 entry smoke 已實際遇到 `0x2D CALL`／`0x2F AND`。本輪只釋放有足夠
operand evidence 的 `AND`／`OR` bounded memory operation；`CALL` 留在 unsupported
boundary，直到能從原始反組譯或跨作品對照確認 return／context semantics。
