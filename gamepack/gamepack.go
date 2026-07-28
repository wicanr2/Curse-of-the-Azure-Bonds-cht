package gamepack

import (
	"embed"
	"fmt"
	"sync"

	goldenbox "github.com/wicanr2/golden-box-remake-engine/engine"
)

// Files is kept title-local: the reusable engine never imports CoAB data.
//
//go:embed events/*.json
var Files embed.FS

var (
	defaultOnce sync.Once
	defaultPack *goldenbox.Pack
	defaultErr  error
)

func Default() (*goldenbox.Pack, error) {
	defaultOnce.Do(func() {
		data, err := Files.ReadFile("events/pit-of-moander.json")
		if err != nil {
			defaultErr = fmt.Errorf("read embedded CoAB game pack: %w", err)
			return
		}
		defaultPack, defaultErr = goldenbox.LoadPackBytes(data)
	})
	return defaultPack, defaultErr
}
