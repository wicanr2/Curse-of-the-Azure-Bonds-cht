# 原作的硬換行落在哪裡（`33h PRINT RETURN`）

由 `cmd/ecl-print-return-audit` 產生，不要手改。語意見 spec 1147：`65A0h := 1`、`inc 65A1h` ⇒ 欄歸位、列前進。

| 項目 | 數 |
|---|---:|
| 走得到的 `33h` 指令 | 120 |
| 其中連著兩條以上的段落（**會空行**）| 10 |
| 換行段落合計 | 110 |

⇒ 有 10 個段落會空行。那幾處 remake 目前會把空行擠掉。

| ECL/區塊 | 位移 | 連續幾條 | 位元組上的前一條 | 後一條 |
|---|---:|---:|---|---|
| ECL2/0x01 | `1D6Dh` | 2 | PRINTCLEAR | PRINT |
| ECL2/0x02 | `0CD4h` | 2 | EXIT | PRINT |
| ECL3/0x10 | `09EEh` | 2 | PRINT | GOSUB |
| ECL3/0x10 | `0B01h` | 2 | PRINT | PRINT |
| ECL3/0x11 | `03CBh` | 2 | PRINTCLEAR | PRINT |
| ECL3/0x12 | `0FCDh` | 2 | PRINTCLEAR | PRINT |
| ECL4/0x20 | `040Bh` | 2 | PRINTCLEAR | PRINT |
| ECL4/0x20 | `0F2Fh` | 2 | PRINTCLEAR | PRINT |
| ECL4/0x21 | `0DBAh` | 2 | PRINTCLEAR | PRINT |
| ECL4/0x22 | `13A4h` | 2 | PRINT | PRINT |

空行落在哪兩句之間（位元組上最近的兩段文字）：

- **ECL2/0x01 `1D6Dh`**
  - 前：YOU SEE A SIGN OVERHEAD
  - 後：—
- **ECL2/0x02 `0CD4h`**
  - 前：—
  - 後：THEY GROWL LOUDLY AS YOU ESCAPE.
- **ECL3/0x10 `09EEh`**
  - 前：OTHER GUARDS GATHER BEHIND HIM.
  - 後：YOU HAVE BEEN LED IN TO SEE THE RED PLUME COMMANDER.
- **ECL3/0x10 `0B01h`**
  - 前：BUSINESS IN YULASH.
  - 後：HOW DO YOU RESPOND?
- **ECL3/0x11 `03CBh`**
  - 前：WE MUST LEAVE YOU NOW.
  - 後：ALIAS THANKS YOU FOR YOUR HELP.  SHE WISHES YOU
- **ECL3/0x12 `0FCDh`**
  - 前：OK, IT'S YOUR DEAD BODY.
  - 後：ALIAS AND DRAGONBAIT LEAVE.
- **ECL4/0x20 `040Bh`**
  - 前：THE GORGE AND GROG SHOP
  - 後：ARENA COMBAT BETTING
- **ECL4/0x20 `0F2Fh`**
  - 前：'YOU KILLED FRITZ!'
  - 後：HE STARTS GRUNTING INCOHERANTLY AND PUSHING HIMSELF DOWN
- **ECL4/0x21 `0DBAh`**
  - 前：YOU NOTICE A SMALL DOOR BELOW THE ALTAR.
  - 後：WHAT DO YOU DO?
- **ECL4/0x22 `13A4h`**
  - 前：HAS SOME QUESTIONS FOR YOU.'
  - 後：OLIVE AND DIMSWART LEAVE.


⚠ 走訪跟不到的碼不在分母裡，而且「連續」只認**位元組相鄰**（中間夾著跳躍、執行序上仍相鄰的算不到）⇒ 兩個方向都是**下界**。
