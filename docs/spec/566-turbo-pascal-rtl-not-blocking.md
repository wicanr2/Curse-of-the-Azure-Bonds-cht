# 第五百六十六輪：Turbo Pascal RTL 與編譯器輔助函式判為不阻塞

狀態：`READY`（限本輪列出的判定規則與例外清單）。日期：2026-08-14

## 結論先行

DOS resident `START.EXE` 有 139 個函式帶 Borland mangled 名稱——那是 IDA 依
Borland 簽章還原的 Turbo Pascal RTL 與編譯器輔助碼。本輪把其中 **133 個標為
`不阻塞`**，並把 **5 個影響玩家可見結果的留在 `待解讀`**。

| 類別 | 例子 | 判定 |
|---|---|---|
| `System`／`Dos`／`Crt`／`Overlay` 單元 | `@CLRSCR$qv`、`@MSDOS$qm9REGISTERS`、`@GetMem$q…`、`@OVRINIT$q6String` | 不阻塞 |
| 編譯器運算子輔助 | `@$basg$qm6Stringt1`（指派）、`@$brmul$q7Longintt1`（乘）、`@Set@MemberOf$q4Byte` | 不阻塞 |
| 實數運算輔助 | `__RealAdd`、`__RealMul`、`__RealTrunc` | 不阻塞 |
| **亂數** | `@Random$q4Word`、`@Randomize$qv` | **待解讀** |
| **PC 喇叭與時序** | `@SOUND$q4WORD`、`@NOSOUND$qv`、`@DELAY$q4WORD` | **待解讀** |
| 進入點 | `PROGRAM`（DOS）、`start`（PC-98） | 待解讀 |

## 判定依據與可推翻性

依據是**具體證據**，不是「看起來不重要」：IDA 從 Borland 簽章還原出的
mangled 名稱本身就標明了單元與型別簽名。要推翻只需指出

1. 某個名稱其實不是 Turbo Pascal RTL，或
2. 某個 RTL 函式確實決定了玩家可見的規則、畫面、聲音或存檔結果。

第 2 點正是例外清單的由來。原版亂數是戰鬥命中、遭遇與寶物的來源，專案明確
禁止用 remake 的 PRNG 冒稱原版；`SOUND`／`NOSOUND`／`DELAY` 玩家直接聽得到、
看得到。這三類即使屬 RTL 也不得標成不阻塞。

## 重生

```sh
python3 scripts/rtl_ledger_batch.py            # 預覽
python3 scripts/rtl_ledger_batch.py --write    # 寫入台帳後重跑 cmd/re-ledger
```

判定寫在腳本裡，台帳是產物。日後發現某個函式其實會影響玩家結果，**改台帳
即可，不要改腳本的判定方式去遷就個案**。

## 這份規格明確不宣稱

- **PC-98 resident 沒有這批名稱**。`PC98-GAME.EXE` 只有 1 個具名函式（`start`），
  333 個函式仍全部 `待解讀`；不得因為 DOS 側標了不阻塞就推論 PC-98 對應位置
  也是 RTL。兩者的 RTL 位置不同，也沒有簽章可依。
- **overlay 內沒有 RTL**。本輪標記全部落在 DOS resident，overlay 側一個都沒有
  ——這符合 Turbo Pascal 把 RTL 放 resident 的配置，但也表示 overlay 的 2,400
  多個函式無法用這條規則減少。
- **IDA 標為 library 的 182 個函式不全在本輪範圍**。只有「有名稱」的才被判定；
  無名稱但被標 library 的仍是 `待解讀`，因為沒有可引用的證據。
