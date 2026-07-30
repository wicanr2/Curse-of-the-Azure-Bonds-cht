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
)

const sampleRate = 44_100

func main() {
	seconds := flag.Duration("duration", 10*time.Second, "輸出長度")
	flag.Parse()
	if flag.NArg() != 3 {
		fmt.Fprintln(
			os.Stderr,
			"用法：pc98-render-track [-duration 10s] MSCDRV.EXE SELECTOR OUTPUT.wav",
		)
		os.Exit(2)
	}
	driver, err := os.ReadFile(flag.Arg(0))
	if err != nil {
		fatal(err)
	}
	var selector int
	if _, err := fmt.Sscanf(flag.Arg(1), "%d", &selector); err != nil {
		fatal(fmt.Errorf("selector：%w", err))
	}
	if *seconds <= 0 {
		fatal(fmt.Errorf("duration 必須大於零"))
	}
	frames := int64(*seconds) * sampleRate / int64(time.Second)
	if frames <= 0 || frames > int64(^uint32(0))/4 {
		fatal(fmt.Errorf("輸出長度超出 WAV 32-bit 範圍"))
	}
	stream, err := pc98music.NewTrackPCMStream(driver, selector, sampleRate)
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
