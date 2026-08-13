# 第五百八十一輪：`PUTDAMAGE` 傷害管線

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-23:1FFDh`／DOS `overlay-23:1FDFh`（787 bytes）。

這是傷害從「數值」變成「角色狀態改變 ＋ 玩家看到的訊息」的整條路徑。

```text
PUTDAMAGE(has_save, save_kind, damage, target)
    DS:A02Eh := damage
    CHECKFX(06h, target)                      ← 可改寫 DS:A02Eh
    if has_save <> 0 then
        if save_kind = 1 then DS:A02Eh := 0                 ← 豁免＝完全免傷
        if save_kind = 2 then DS:A02Eh := DS:A02Eh div 2    ← 豁免＝減半（有號除法）
    else
        CHECKFX(14h, target)                  ← 沒有豁免時走另一個 effect 檢查

    if DS:A02Eh <= 0 then
        <顯示>「は、ダメージを受けなかった。」／`is Unaffected`
    else
        msg := Str(DS:A02Eh) + 「ポイントのダメージを受けた。」／` points of damage `
        case DS:A02Fh and 0F7h of                ← 先遮掉 bit 3
            01h: 前綴「は火によって」    ／`from Fire`
            02h: 前綴「は冷気によって」  ／`from Cold`
            04h: 前綴「は電撃によって」  ／`from Electricity`
            10h: 前綴「は酸によって」    ／`from Acid`
        if DS:A02Fh and 8 = DS:A02Fh then        ← 只有 bit 3、其餘全 0
            前綴「は魔法によって」      ／`from Magic`
        <顯示>(msg, target)
        <扣血>(DS:A02Eh, target)

        if DS:7F27h = 5 then                     ← 只在這個模式下失去法術
            p := target^[18Eh]
            p^[1] := 0
            if p^[0] > 0 then
                <顯示>「は呪文を唱えられなくなった。」／`lost a spell`
                <處理>(p^[0], target) ; p^[0] := 0

        if target^[197h] = 0 then
            msg := 「は倒れた。」／`Goes Down`
            if target^[196h] = 5 then msg := msg + 「そして、瀕死の状態だ。」／`, and is Dying`
            if target^[196h] in {6, 7, 8} then msg := 「は死んだ。」／`is killed`
            <顯示>(msg, delay := DS:9637h + 1, target)
            if DS:7F27h = 5 then
                REMOVEFX(target) ; CHECKFX(0Dh, target)
                if target^[197h] = 0 then <far 0189:0013>(target)
                else                          <far 0418:14AA>()
            else <far 0418:14AA>()
    <far 014A:0089>()
```

## 傷害屬性 `DS:A02Fh` 的位元對應

訊息前綴把每個位元的名字釘死了：

| 位元 | 值 | 屬性 |
|---:|---:|---|
| 0 | `01h` | Fire |
| 1 | `02h` | Cold |
| 2 | `04h` | Electricity |
| 3 | `08h` | Magic |
| 4 | `10h` | Acid |

⚠ **選訊息用的是完整值比較，不是位元測試。** `and 0F7h` 之後要**剛好等於**
`01h`／`02h`／`04h`／`10h` 才有前綴；`Magic` 更嚴，要求 `DS:A02Fh` 只有 bit 3
而其餘全 0。所以**混合屬性的傷害不會顯示任何前綴**——這是原版行為。
免疫／抗性那一側則是逐位元 `and`（[spec 573](573-effprocs-effect-handlers-first-batch.md)），
兩邊的判定方式不同，不要互相套用。

## 幾個容易做錯的細節

- **豁免只有兩種強度**：`save_kind = 1` 歸零、`= 2` 減半，其餘值不做事。
- **減半是有號除法**（`cwd` ＋ `idiv`），所以負傷害（治療？）會往零收斂。
- **沒有豁免時會多走一次 `CHECKFX(14h)`**，有豁免時不走。
- 死亡訊息是**覆寫**不是附加：`is killed` 會把 `Goes Down, and is Dying`
  整段換掉。
- 訊息停留時間取自全域 `DS:9637h + 1`，不是常數。
- `{6, 7, 8}` 這個死亡狀態集合，與 `KILLDUDE` 不再覆寫的集合
  （[spec 579](579-character-status-fields.md)）是同一組，兩處各自的 32 bytes
  set 常數逐位元組相同——互相印證。

## 明確不宣稱

- `<扣血>`（`014A:00AC`）內部怎麼把 `DS:A02Eh` 套進 `+1A5h`。
- `target^[18Eh]` 指向的結構（只知道 `+0` 非零代表有法術可失去、`+1` 會被清零）。
- `DS:7F27h = 5` 是什麼模式（戰鬥？），`DS:9637h` 是什麼設定。
