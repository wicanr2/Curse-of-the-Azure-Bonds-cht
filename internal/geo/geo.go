// Package geo preserves the original CoAB import path while the reusable GEO
// implementation lives in golden-box-remake-engine/geometry.
package geo

import enginegeo "github.com/wicanr2/golden-box-remake-engine/geometry"

const (
	Width       = enginegeo.Width
	Height      = enginegeo.Height
	PayloadSize = enginegeo.PayloadSize
	BlockSize   = enginegeo.BlockSize
)

type Cell = enginegeo.Cell
type Grid = enginegeo.Grid

var Parse = enginegeo.Parse
var WrapCoordinate = enginegeo.WrapCoordinate
