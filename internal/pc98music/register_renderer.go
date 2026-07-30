package pc98music

import "fmt"

// YM2203EventRenderer expands the title-specific Sound BIOS intents in a
// TrackPlayback stream into the exact register order consumed by a reusable
// YM2203 synthesizer.
type YM2203EventRenderer struct {
	blocks     []FMParameterBlock
	parameters [3]int
	have       [3]bool
}

// NewYM2203EventRenderer parses the embedded, verified parameter bank from
// the user's local driver.
func NewYM2203EventRenderer(driver []byte) (*YM2203EventRenderer, error) {
	blocks, err := EmbeddedFMParameterBlocks(driver)
	if err != nil {
		return nil, err
	}
	return newYM2203EventRenderer(blocks), nil
}

func newYM2203EventRenderer(blocks []FMParameterBlock) *YM2203EventRenderer {
	renderer := &YM2203EventRenderer{
		blocks: append([]FMParameterBlock(nil), blocks...),
	}
	for channel := range renderer.parameters {
		renderer.parameters[channel] = -1
	}
	return renderer
}

// Render converts one ordered event batch. Unknown high parameter indices are
// retained as state but do not fabricate a tone. In the verified CoAB tracks
// these descriptor-only loads are replaced by an embedded 0..19 parameter
// before the first key-on.
func (renderer *YM2203EventRenderer) Render(
	events []MusicEvent,
) ([]YM2203RegisterWrite, error) {
	if renderer == nil {
		return nil, fmt.Errorf("YM2203 event renderer is nil")
	}
	var writes []YM2203RegisterWrite
	for _, event := range events {
		switch event.Kind {
		case EventRegisterWrite:
			writes = append(writes, YM2203RegisterWrite{
				Register: event.Register,
				Value:    event.Value,
			})
		case EventSetParameterBlock:
			if event.Channel < 0 || event.Channel >= len(renderer.parameters) {
				return nil, fmt.Errorf(
					"parameter event channel %d is outside 0..2",
					event.Channel,
				)
			}
			index := int(event.Parameter)
			renderer.parameters[event.Channel] = index
			renderer.have[event.Channel] = index >= 0 && index < len(renderer.blocks)
			if !renderer.have[event.Channel] {
				continue
			}
			levels, err := renderer.blocks[index].YM2203LevelSequence(0)
			if err != nil {
				return nil, err
			}
			tone, err := renderer.blocks[index].YM2203ToneWrites(
				event.Channel, levels[0],
			)
			if err != nil {
				return nil, err
			}
			writes = append(writes, tone...)
		case EventSetVolume:
			if event.Channel < 0 || event.Channel >= len(renderer.parameters) {
				return nil, fmt.Errorf(
					"volume event channel %d is outside 0..2",
					event.Channel,
				)
			}
			if !renderer.have[event.Channel] {
				continue
			}
			block := renderer.blocks[renderer.parameters[event.Channel]]
			levels, err := block.YM2203LevelSequence(event.Value)
			if err != nil {
				return nil, err
			}
			// levels[0] was already emitted by SETPARABLOCK. Each remaining
			// entry is the complete redraw after one carrier update.
			for _, level := range levels[1:] {
				tone, err := block.YM2203ToneWrites(event.Channel, level)
				if err != nil {
					return nil, err
				}
				writes = append(writes, tone...)
			}
		default:
			return nil, fmt.Errorf("unknown music event kind %q", event.Kind)
		}
	}
	return writes, nil
}
