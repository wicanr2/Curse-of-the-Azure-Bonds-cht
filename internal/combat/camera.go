package combat

import engineviewport "github.com/wicanr2/golden-box-remake-engine/viewport"

// CombatCamera 把戰鬥地圖格換成戰鬥視窗的格座標。平移本身在共用 engine 的
// `viewport.Camera`；這裡只保留 `TilePoint` 的轉接。
type CombatCamera struct {
	camera engineviewport.Camera
}

// NewCombatCamera 讓目前行動的格落在視窗的指定位置。沒有已知的行動格時回零值
// 相機，也就是保留絕對座標的 fallback。
func NewCombatCamera(active, viewportCenter TilePoint, ok bool) CombatCamera {
	if !ok {
		return CombatCamera{}
	}
	return CombatCamera{camera: engineviewport.NewCamera(active.X, active.Y,
		viewportCenter.X, viewportCenter.Y)}
}

// Origin 是相機的平移量，也就是視窗左上角對應的地圖格。
// 反查（由畫面格回推地圖格）需要它。
func (c CombatCamera) Origin() TilePoint {
	return TilePoint{X: c.camera.OriginX, Y: c.camera.OriginY}
}

// Apply 把一格戰鬥地圖座標換成視窗相對座標。
func (c CombatCamera) Apply(point TilePoint) TilePoint {
	x, y := c.camera.Apply(point.X, point.Y)
	return TilePoint{X: x, Y: y}
}
