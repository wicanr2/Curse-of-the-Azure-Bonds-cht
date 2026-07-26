package combat

// CombatCamera converts CombatMap tile coordinates into the tile coordinates
// used by the combat viewport. The original combat screen follows the active
// character; keeping this transform in the platform-neutral core lets Ebiten
// and later Gold Box front ends share the same map/camera contract.
type CombatCamera struct {
	Origin TilePoint
}

// NewCombatCamera centers the active tile at the requested viewport tile.
// Callers that do not have a decoded active position should use the zero
// camera, which preserves the existing absolute-coordinate fallback.
func NewCombatCamera(active, viewportCenter TilePoint, ok bool) CombatCamera {
	if !ok {
		return CombatCamera{}
	}
	return CombatCamera{Origin: TilePoint{
		X: active.X - viewportCenter.X,
		Y: active.Y - viewportCenter.Y,
	}}
}

// Apply translates one CombatMap tile into viewport-relative coordinates.
func (c CombatCamera) Apply(point TilePoint) TilePoint {
	return TilePoint{X: point.X - c.Origin.X, Y: point.Y - c.Origin.Y}
}
