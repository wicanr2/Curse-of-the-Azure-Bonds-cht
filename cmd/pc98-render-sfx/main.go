// Command pc98-render-sfx renders one exact GAME.EXE SOUNDFX selector to a
// local WAV without embedding commercial sound-program data in the repository.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98sfx"
)

const sampleRate = 44_100

func main() {
	clockHz := flag.Uint64("clock", 8_000_000, "重建使用的 PC-98 CPU clock")
	flag.Parse()
	if flag.NArg() != 3 {
		fmt.Fprintln(
			os.Stderr,
			"用法：pc98-render-sfx [-clock 8000000] GAME.EXE SELECTOR OUTPUT.wav",
		)
		os.Exit(2)
	}
	game, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	var selector int
	if _, err := fmt.Sscanf(flag.Arg(1), "%d", &selector); err != nil {
		fatal(fmt.Errorf("selector：%w", err))
	}
	effects, err := pc98sfx.Import(game)
	if err != nil {
		fatal(err)
	}
	if selector < 0 || selector >= len(effects) {
		fatal(fmt.Errorf("selector 必須介於 0 與 %d", len(effects)-1))
	}
	mono, err := pc98sfx.RenderPCM(
		effects[selector],
		pc98sfx.V30PrefetchedProfile(*clockHz),
		sampleRate,
	)
	if err != nil {
		fatal(err)
	}
	pcm := stereoPCM(mono)
	if err := writeWAV(flag.Arg(2), pcm); err != nil {
		fatal(err)
	}
	fmt.Printf(
		"selector=%d symbol=%s event=%s clock_hz=%d frames=%d\n",
		selector, effects[selector].Symbol, effects[selector].Event,
		*clockHz, len(mono),
	)
}

func stereoPCM(mono []int16) []byte {
	output := make([]byte, len(mono)*4)
	for index, sample := range mono {
		binary.LittleEndian.PutUint16(output[index*4:], uint16(sample))
		binary.LittleEndian.PutUint16(output[index*4+2:], uint16(sample))
	}
	return output
}

func writeWAV(path string, pcm []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, value := range []any{
		[]byte("RIFF"), uint32(36 + len(pcm)), []byte("WAVEfmt "),
		uint32(16), uint16(1), uint16(2), uint32(sampleRate),
		uint32(sampleRate * 4), uint16(4), uint16(16),
		[]byte("data"), uint32(len(pcm)), pcm,
	} {
		if err := binary.Write(file, binary.LittleEndian, value); err != nil {
			return err
		}
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
