# 第二百五十輪：半獸人 race／DOS parser

## 狀態

READY

## 證據

reference `Gbl.RaceClasses` 明確列出半獸人的 cleric、fighter、thief 與 multi-class
組合；`race_ages` 的 half-orc row 提供 cleric `20 + 1d4`、ranger-like slot
`13 + 1d4`、thief `20 + 2d4`，但 fighter／paladin／magic-user entry 為零。`Limits`
也提供 half-orc age brackets 與共用 age effects。

## 實作邊界

- `party.RaceHalfOrc` 與 raw DOS race `6` 已可解析、驗證、序列化並使用正常 combat icon size。
- 建立選單接入半獸人 cleric／thief；fighter 的空白 age entry 不被假設成可生成。
- half-orc multi-class、fighter age policy、alignment／level limits 與完整原版建立流程仍待接續。

## 驗證

- `TestHalfOrcSingleClassEvidence` 驗證 age table 與 class restriction。
- `TestParseDOSHalfOrcRace` 驗證 raw race `6` 進入 DOS player／Character。
