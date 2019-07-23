package draw

import (
	"github.com/loov/allocview/internal/ui/g"
	"github.com/loov/vsofont"
)

func (list *List) TextVector(text string, dot g.Vector, height, thickness float32, color g.Color) {
	if color.Transparent() || height == 0 || thickness == 0 {
		return
	}

	start := dot
	dot.Y -= height

	font := vsofont.Example
	for _, r := range text {
		switch r {
		case ' ':
			dot.X += height
		case '\t':
			dot.X += height * 4
		case '\n':
			dot.X = start.X
			dot.Y += height
		default:
			glyph, ok := font.Glyphs[string(r)]
			if !ok {
				glyph = font.Glyphs["?"]
			}

			for _, line := range glyph.Lines {
				list.StrokeLine([]g.Vector{
					dot.Add(g.Vector(line[0]).Scale(height)),
					dot.Add(g.Vector(line[1]).Scale(height)),
				}, thickness, color)
			}
			dot.X += height + font.Spacing*height
		}
	}
}
