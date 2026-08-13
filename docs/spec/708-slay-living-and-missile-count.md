# 第七百零八輪：豁免失敗即死，與「數量減半、傷害不減半」

狀態：`READY`。等級：`exact`（流程）／`strong inference`（法術身分）。
日期：2026-08-14
位置：DOS `overlay-22` 的 `45B5h`、`54EBh`。

## `45B5h`：三條路

```text
p := DS:7435h
DS:6F95h := 40h                                  ← 傷害旗標
DS:6F94h := 43h                                  ← 傷害值（67）
<overlay-23 entry#4>(p, 9)                       ← 這一支可能把 DS:6F94h 清成 0

if DS:6F94h = 0 then
    <overlay-24 entry#20>(p, 'is unaffected'（CS:45A7h）, 0Ah, 1)
elif <overlay-23 entry#8>(p, 4, 0) = 0 then
    <overlay-23 entry#1>(p, 6, 'is slain'（CS:459Eh）)
else
    DS:6F95h := 8                                ← 換一種傷害旗標
    d := ROLLDAMAGEDICE(2, 8) + 1                ← 2d8+1
    <overlay-23 entry#20>(d, 0, 0)
```

「豁免失敗即死，豁免成功承受 `2d8+1` 傷害」是 AD&D **Slay Living** 的招牌數字，
所以這支幾乎確定是它。不過本輪沒有從法術表反查編號，身分標為推論。

三條路的判別方式各不相同，值得分開記：

- 第一條靠**下游有沒有把全域清成 0**（`entry#4` 之後回頭看 `DS:6F94h`），
  和 spec 703 `4289h` 的哨兵是同一種手法。
- 第二條靠 `entry#8` 的回傳值。
- 兩種傷害旗標 `40h` 與 `8` 在同一支裡先後出現：即死那條路用 `40h`，
  傷害那條路改成 `8`。所以旗標是隨結果換的，不是進來就決定的。

`DS:6F94h` 先被填 `43h`（67）——遠超過任何一次 `2d8+1`。那個值是給即死路徑用的，
不是傷害路徑。

## `54EBh`：數量減半，但傷害用的是沒減半的值

```text
p     := DS:6506h
n     := <overlay-24 entry#36>(DS:6F97h)          ← 存在 bp−6
half  := (<overlay-24 entry#36>(DS:6F97h) + 1) div 2   ← 存在 bp−5，有號 idiv
if half < 1 then half := 1                        ← 無號 jnb，夾下限

a := <overlay-32 entry#15>(p)
b := <overlay-32 entry#16>(p)
<sub_175Bh>(a, b, DS:7559h, DS:755Ah, 2, half)    ← 用 half

d := ROLLDAMAGEDICE(n, 4) + n                     ← ⚠ 用 n，不是 half
<sub_F06h>(DS:6F97h, 0, 0, d, 0Ah, 空字串（CS:54EAh）)
```

`entry#36` 被呼叫**兩次**（參數相同），第一次的結果存成 `n`、第二次的結果
`+1` 之後除以 2 存成 `half`。兩個區域變數位移相鄰（`bp−6` 與 `bp−5`），
在反組譯裡極容易看成同一個。

**`half` 只餵給 `sub_175Bh`，傷害公式用的是 `n`。** 也就是「數量減半、每一份
的傷害不減半」。改寫時若只留一個變數，兩邊都會錯。

`(n + 1) div 2` 用的是 `cwd` ＋ `idiv 2`（往零取整），不是右移；下限夾在 1，
所以 `n = 0` 時仍會有 1 份。

## 明確不宣稱

- `overlay-23` entry#1／#4／#8／#20、`overlay-24 entry#36`、`sub_175Bh` 的行為。
- `entry#36` 回傳什麼（本輪只知道它被當成一個數量／等級用）。
- 傷害旗標 `40h`／`8`／`0Ch` 各代表哪一類。
- `DS:7559h`／`755Ah` 是什麼。
