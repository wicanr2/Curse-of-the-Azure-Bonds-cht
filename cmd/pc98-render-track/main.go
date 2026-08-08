// Command pc98-render-track renders a user-supplied CoAB MSCDRV.EXE selector
// to a local WAV without embedding commercial music data in the repository.
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/pc98music"
	"github.com/wicanr2/Curse-of-the-Azure-Bonds-cht/internal/tooltext"
)

const sampleRate = 44_100

func main() {
	seconds := flag.Duration("duration", 10*time.Second, tooltext.Text("pc98_render_track.duration"))
	gameTransition := flag.Bool(
		"game-transition",
		false,
		tooltext.Text("pc98_render_track.transition"),
	)
	flag.Parse()
	if flag.NArg() != 3 {
		fmt.Fprintln(
			os.Stderr,
			tooltext.Text("pc98_render_track.usage"),
		)
		os.Exit(2)
	}
	driver, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	var selector int
	if _, err := fmt.Sscanf(flag.Arg(1), "%d", &selector); err != nil {
		fatal(tooltext.Errorf("pc98_render_track.selector_parse", err))
	}
	if *seconds <= 0 {
		fatal(tooltext.Error("pc98_render_track.duration_positive"))
	}
	frames := int64(*seconds) * sampleRate / int64(time.Second)
	if frames <= 0 || frames > int64(^uint32(0))/4 {
		fatal(tooltext.Error("pc98_render_track.wav_range"))
	}
	var stream *pc98music.TrackPCMStream
	if *gameTransition {
		stream, err = pc98music.NewGameTrackPCMStream(driver, selector, sampleRate)
	} else {
		stream, err = pc98music.NewTrackPCMStream(driver, selector, sampleRate)
	}
	if err != nil {
		fatal(err)
	}
	defer stream.Close()
	pcm := make([]byte, frames*4)
	if _, err := io.ReadFull(stream, pcm); err != nil {
		fatal(err)
	}
	if err := writeWAV(flag.Arg(2), pcm); err != nil {
		fatal(err)
	}
}

func writeWAV(path string, pcm []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(36+len(pcm))); err != nil {
		return err
	}
	if _, err := file.Write([]byte("WAVEfmt ")); err != nil {
		return err
	}
	format := []uint32{16}
	for _, value := range format {
		if err := binary.Write(file, binary.LittleEndian, value); err != nil {
			return err
		}
	}
	for _, value := range []uint16{1, 2} {
		if err := binary.Write(file, binary.LittleEndian, value); err != nil {
			return err
		}
	}
	for _, value := range []uint32{sampleRate, sampleRate * 4} {
		if err := binary.Write(file, binary.LittleEndian, value); err != nil {
			return err
		}
	}
	for _, value := range []uint16{4, 16} {
		if err := binary.Write(file, binary.LittleEndian, value); err != nil {
			return err
		}
	}
	if _, err := file.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, uint32(len(pcm))); err != nil {
		return err
	}
	_, err = file.Write(pcm)
	return err
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
