package game

import "github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/segment"

// EnterSegment 依註冊表的宣告直接進入一段主線，讓分段驗收不必每次從頭跑。
//
// ⚠ 它只到「段的入口」：專用旗標（`-lava-tube` 之類）通常還會把隊伍走到段內
// 某一格、或先打完某一場戰鬥，那些是段內的檢查點，不是段的入口。
func (s *State) EnterSegment(seg segment.Segment) error {
	if seg.EnterFrom == 0x00 && !seg.Overland {
		// LastECL ＝ 0 是引擎的「全新開局」條件（spec 1141）：主迴圈在這個
		// 條件下載入 block 1。remake 的對應入口就是 BeginAdventure，它自己
		// 會切到 0x01 並跑完新遊戲進場，不需要另外設 LastECL。
		return s.BeginAdventure()
	}
	if err := s.StartStorySegment(seg.Block, seg.EnterFrom, seg.GameArea, !seg.Overland); err != nil {
		return err
	}
	if seg.Overland && s.Mode == ModeWilderness && len(s.Choices) == 0 {
		// 世界地圖上的段沒有 GEO 幾何可以站，而 initial lifecycle 也沒有留下
		// 選單。用開場的立石當確定性的落點，畫面才有東西可看。
		s.PrepareWorldMapPreview()
	}
	return nil
}
