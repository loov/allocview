package draw

import (
	"image"
	"io/ioutil"
	"math"

	"github.com/loov/allocview/internal/ui/g"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Font struct {
	Context *freetype.Context
	TTF     *truetype.Font
	Face    font.Face

	Rendered map[rune]Glyph
	Atlas    *Texture
	Image    *image.RGBA

	CursorX       int
	CursorY       int
	MaxGlyphInRow int
	DrawPadding   float32

	MaxBounds  fixed.Rectangle26_6
	LineHeight float32

	Dirty bool
}

type Glyph struct {
	Rune    rune
	Loc     image.Rectangle     // absolute location on image atlas
	RelLoc  g.Rect              // relative location on image atlas
	Bounds  fixed.Rectangle26_6 // such that point + bounds, gives image bounds where glyph should be drawn
	Advance fixed.Int26_6       // advance from point, to the next glyph
}

const (
	glyphMargin  = 2
	glyphPadding = 1
)

func LoadTTF(filename string, dpi, fontSize float64) (*Font, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ttf, err := truetype.Parse(content)
	if err != nil {
		return nil, err
	}

	return NewTTF(ttf, dpi, fontSize)
}

func NewTTF(ttf *truetype.Font, dpi, fontSize float64) (*Font, error) {
	atlas := &Font{}

	atlas.TTF = ttf

	atlas.Rendered = make(map[rune]Glyph, 256)

	atlas.DrawPadding = float32(fontSize * 0.5)
	atlas.LineHeight = float32(fontSize * 1.2)

	atlas.Image = image.NewRGBA(image.Rect(0, 0, 1024, 1024))

	atlas.Context = freetype.NewContext()
	atlas.Context.SetDPI(dpi)

	atlas.Context.SetFont(atlas.TTF)
	atlas.Context.SetFontSize(fontSize)

	atlas.Context.SetClip(atlas.Image.Bounds())
	atlas.Context.SetSrc(image.White)
	atlas.Context.SetDst(atlas.Image)

	atlas.MaxBounds = atlas.TTF.Bounds(fixed.I(int(fontSize)))

	opts := &truetype.Options{}
	opts.Size = fontSize
	opts.Hinting = font.HintingFull

	atlas.Face = truetype.NewFace(atlas.TTF, opts)
	return atlas, nil
}

func ceilPx(i fixed.Int26_6) int {
	const ceiling = 1<<6 - 1
	return int(i+ceiling) >> 6
}

func ceilPxf(i fixed.Int26_6) float32 {
	const div = 1 << 6
	return float32(i) / div
}

func (atlas *Font) loadGlyph(r rune) {
	if _, ok := atlas.Rendered[r]; ok {
		return
	}
	atlas.Dirty = true

	glyph := Glyph{}
	glyph.Rune = r

	bounds, advance, _ := atlas.Face.GlyphBounds(r)
	glyph.Bounds = bounds
	glyph.Advance = advance

	width := ceilPx(bounds.Max.X-bounds.Min.X) + glyphPadding*2
	height := ceilPx(bounds.Max.Y-bounds.Min.Y) + glyphPadding*2

	if atlas.CursorX+glyphMargin+width+glyphMargin > atlas.Image.Bounds().Dx() {
		atlas.CursorX = 0
		atlas.CursorY += glyphMargin + atlas.MaxGlyphInRow
	}

	x := atlas.CursorX + glyphMargin
	y := atlas.CursorY + glyphMargin

	glyph.Loc = image.Rect(x, y, x+width, y+height)
	glyph.RelLoc = RelBounds(glyph.Loc, atlas.Image.Bounds())

	pt := fixed.P(x+glyphPadding, y+glyphPadding).Sub(bounds.Min)
	atlas.Context.DrawString(string(r), pt)

	if height > atlas.MaxGlyphInRow {
		atlas.MaxGlyphInRow = height
	}
	atlas.CursorX += glyphMargin + width + glyphMargin

	atlas.Rendered[r] = glyph
}

func (atlas *Font) LoadAscii() {
	for r := rune(0); r < 128; r++ {
		atlas.loadGlyph(r)
	}
}

func (atlas *Font) LoadExtendedAscii() {
	for r := rune(0); r < 256; r++ {
		atlas.loadGlyph(r)
	}
}

func (atlas *Font) LoadGlyphs(text string) {
	for _, r := range text {
		atlas.loadGlyph(r)
	}
}

func (atlas *Font) Draw(list *List, text string, bounds g.Rect, color g.Color) {
	atlas.LoadGlyphs(text)

	textureID := list.IncludeTexture(atlas.Image, atlas.Dirty)
	atlas.Dirty = false

	list.PushTexture(textureID)
	defer list.PopTexture()

	x := bounds.Min.X + atlas.DrawPadding
	y := (bounds.Max.Y+bounds.Min.Y)/2 + (ceilPxf(atlas.MaxBounds.Min.Y)+ceilPxf(atlas.MaxBounds.Max.Y))/2

	lastRune := rune(0)
	for _, r := range text {
		glyph := atlas.Rendered[r]

		dx := float32(glyph.Loc.Dx())
		dy := float32(glyph.Loc.Dy())

		px := x + ceilPxf(glyph.Bounds.Min.X) - glyphPadding
		py := y + ceilPxf(glyph.Bounds.Min.Y) - glyphPadding

		// this is not the ideal way of positioning the letters
		// will create positioning artifacts
		px = float32(math.Trunc(float64(px)))
		py = float32(math.Trunc(float64(py)))

		list.RectUV(
			&g.Rect{
				Min: g.V(px, py),
				Max: g.V(px+dx, py+dy),
			},
			&glyph.RelLoc,
			color,
		)

		k := atlas.Face.Kern(lastRune, r)
		lastRune = r
		x += ceilPxf(glyph.Advance + k)
	}
}

func RelBounds(r, b image.Rectangle) (n g.Rect) {
	n.Min.X = float32(r.Min.X-b.Min.X) / float32(b.Dx())
	n.Min.Y = float32(r.Min.Y-b.Min.Y) / float32(b.Dy())
	n.Max.X = float32(r.Max.X-b.Min.X) / float32(b.Dx())
	n.Max.Y = float32(r.Max.Y-b.Min.Y) / float32(b.Dy())
	return n
}
