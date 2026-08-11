// Package etenfont adapts ETen's original 16x15 Big5 bitmap font to font.Face.
package etenfont

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/text/encoding/traditionalchinese"
)

const (
	glyphWidth      = 16
	glyphHeight     = 15
	glyphBytes      = 30
	asciiGlyphWidth = 8
	asciiGlyphBytes = 15
)

type cachedGlyph struct {
	mask    *image.Alpha
	advance int
}

// Face reads ETen's original STDFONT.15 Chinese glyphs, optional SPCFONT.15
// full-width symbols, and its matching ASCFONT.15 half-width glyphs.
// Unsupported runes still delegate to Fallback.
type Face struct {
	standard []byte
	symbols  []byte
	ascii    []byte
	Fallback font.Face
	Bold     bool
	mu       sync.Mutex
	cache    map[rune]cachedGlyph
}

// Load opens an ETen STDFONT.15 and optional SPCFONT.15.
func Load(standardPath, symbolPath string, fallback font.Face, bold bool) (*Face, error) {
	return LoadWithASCII(standardPath, symbolPath, "", fallback, bold)
}

// LoadWithASCII opens an ETen STDFONT.15, optional SPCFONT.15, and optional
// ASCFONT.15. ASCFONT.15 is the companion 8x15 raster used by the original
// ETen family for ASCII and ordinary punctuation. Keeping it separate lets a
// deployment without that local file retain the previous fallback behaviour.
func LoadWithASCII(standardPath, symbolPath, asciiPath string, fallback font.Face, bold bool) (*Face, error) {
	standard, err := os.ReadFile(standardPath)
	if err != nil {
		return nil, fmt.Errorf("read ETen standard font: %w", err)
	}
	if len(standard)%glyphBytes != 0 {
		return nil, fmt.Errorf("ETen standard font size %d is not divisible by %d", len(standard), glyphBytes)
	}
	var symbols []byte
	if symbolPath != "" {
		symbols, err = os.ReadFile(symbolPath)
		if err != nil {
			return nil, fmt.Errorf("read ETen symbol font: %w", err)
		}
		if len(symbols)%glyphBytes != 0 {
			return nil, fmt.Errorf("ETen symbol font size %d is not divisible by %d", len(symbols), glyphBytes)
		}
	}
	var ascii []byte
	if asciiPath != "" {
		ascii, err = os.ReadFile(asciiPath)
		if err != nil {
			return nil, fmt.Errorf("read ETen ASCII font: %w", err)
		}
		if len(ascii) < 256*asciiGlyphBytes || len(ascii)%asciiGlyphBytes != 0 {
			return nil, fmt.Errorf("ETen ASCII font size %d does not contain 256 8x15 glyphs", len(ascii))
		}
	}
	return &Face{standard: standard, symbols: symbols, ascii: ascii, Fallback: fallback, Bold: bold, cache: make(map[rune]cachedGlyph)}, nil
}

func (f *Face) Close() error { return nil }

func (f *Face) Metrics() font.Metrics {
	return font.Metrics{Height: fixed.I(glyphHeight), Ascent: fixed.I(14), Descent: fixed.I(1)}
}

func (f *Face) Kern(r0, r1 rune) fixed.Int26_6 { return 0 }

func (f *Face) GlyphAdvance(r rune) (fixed.Int26_6, bool) {
	if glyph, ok := f.glyph(r); ok {
		return fixed.I(glyph.advance), true
	}
	return f.Fallback.GlyphAdvance(r)
}

func (f *Face) GlyphBounds(r rune) (fixed.Rectangle26_6, fixed.Int26_6, bool) {
	if glyph, ok := f.glyph(r); ok {
		return fixed.R(0, -14, glyph.advance, 1), fixed.I(glyph.advance), true
	}
	return f.Fallback.GlyphBounds(r)
}

func (f *Face) Glyph(dot fixed.Point26_6, r rune) (image.Rectangle, image.Image, image.Point, fixed.Int26_6, bool) {
	glyph, ok := f.glyph(r)
	if !ok {
		return f.Fallback.Glyph(dot, r)
	}
	x, y := dot.X.Floor(), dot.Y.Floor()-14
	return image.Rect(x, y, x+glyph.advance, y+glyphHeight), glyph.mask, image.Point{}, fixed.I(glyph.advance), true
}

// bitmap is retained for the focused raster tests. Production callers should
// use font.Face methods, which also preserve each glyph's native advance.
func (f *Face) bitmap(r rune) (*image.Alpha, bool) {
	glyph, ok := f.glyph(r)
	if !ok {
		return nil, false
	}
	return glyph.mask, true
}

func (f *Face) glyph(r rune) (cachedGlyph, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if cached, ok := f.cache[r]; ok {
		return cached, true
	}
	if raw, ok := f.standardGlyph(r); ok {
		glyph := cachedGlyph{mask: rasterGlyph(raw, glyphWidth, glyphBytes/glyphHeight, f.Bold), advance: glyphWidth}
		f.cache[r] = glyph
		return glyph, true
	}
	if raw, ok := f.asciiGlyph(r); ok {
		glyph := cachedGlyph{mask: rasterGlyph(raw, asciiGlyphWidth, 1, f.Bold), advance: asciiGlyphWidth}
		f.cache[r] = glyph
		return glyph, true
	}
	return cachedGlyph{}, false
}

func rasterGlyph(raw []byte, width, bytesPerRow int, bold bool) *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, width, glyphHeight))
	for y := 0; y < glyphHeight; y++ {
		for x := 0; x < width; x++ {
			on := raw[y*bytesPerRow+x/8]&(0x80>>uint(x&7)) != 0
			if bold && x > 0 {
				on = on || raw[y*bytesPerRow+(x-1)/8]&(0x80>>uint((x-1)&7)) != 0
			}
			if on {
				mask.SetAlpha(x, y, color.Alpha{A: 0xff})
			}
		}
	}
	return mask
}

func (f *Face) standardGlyph(r rune) ([]byte, bool) {
	encoded, err := traditionalchinese.Big5.NewEncoder().Bytes([]byte(string(r)))
	if err != nil || len(encoded) != 2 {
		return nil, false
	}
	raw := rawBig5(int(encoded[0]), int(encoded[1]))
	lastSymbol := rawBig5(0xa3, 0xbf)
	if raw <= lastSymbol {
		return glyphAt(f.symbols, raw)
	}
	const commonCount = 5401
	var index int
	if raw <= rawBig5(0xc6, 0x7e) {
		index = raw - rawBig5(0xa4, 0x40)
	} else {
		index = commonCount + raw - rawBig5(0xc9, 0x40)
	}
	return glyphAt(f.standard, index)
}

func (f *Face) asciiGlyph(r rune) ([]byte, bool) {
	if len(f.ascii) == 0 {
		return nil, false
	}
	value, ok := etenASCIICode(r)
	if !ok {
		return nil, false
	}
	offset := int(value) * asciiGlyphBytes
	if offset+asciiGlyphBytes > len(f.ascii) {
		return nil, false
	}
	return f.ascii[offset : offset+asciiGlyphBytes], true
}

// ETen's ASCFONT has the half-width forms used by the game UI. Map only
// typographic aliases; translation text remains data-driven and no game term
// is encoded here.
func etenASCIICode(r rune) (byte, bool) {
	if r >= 0 && r <= 0xff {
		return byte(r), true
	}
	aliases := map[rune]byte{
		'　': ' ', '，': ',', '。': '.', '、': ',', '：': ':', '；': ';',
		'！': '!', '？': '?', '（': '(', '）': ')', '［': '[', '］': ']',
		'｛': '{', '｝': '}', '／': '/', '－': '-', '＋': '+', '＝': '=',
		'％': '%', '＆': '&', '＊': '*', '＜': '<', '＞': '>', '｜': '|',
		'「': '"', '」': '"', '『': '"', '』': '"', '〈': '<', '〉': '>',
		'《': '<', '》': '>', '【': '[', '】': ']', '〔': '[', '〕': ']',
	}
	value, ok := aliases[r]
	return value, ok
}

func rawBig5(high, low int) int {
	trail := low - 0x40
	if low >= 0x7f {
		trail = low - 0x62
	}
	return (high-0xa1)*157 + trail
}

func glyphAt(data []byte, index int) ([]byte, bool) {
	offset := index * glyphBytes
	if index < 0 || offset+glyphBytes > len(data) {
		return nil, false
	}
	return data[offset : offset+glyphBytes], true
}
