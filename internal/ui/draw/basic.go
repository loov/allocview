package draw

import (
	"github.com/loov/allocview/internal/ui/g"
)

func (list *List) StrokeLine(points []g.Vector, thickness float32, color g.Color) {
	if len(points) < 2 || color.Transparent() || thickness == 0 {
		return
	}

	// TODO: optimize for thin line
	startIndexCount := len(list.Indicies)

	R := g.Abs(thickness / 2.0)
	R2 := R * R
	s2R2 := g.Sqrt2 * R2

	// draw each segment, where
	// x1-------^--------a1-------^---------b1
	// |        | xn     |        | abn      |
	// | - - - - - - - - a - - - - - - - - - b
	// |                 |                   |
	// x2----------------a2-----------------b2
	// x1, x2, xn are the previous segments end corners and normal
	a, b := points[0], points[1]
	xn := g.SegmentNormal(a, b).ScaleTo(R)
	x1, x2 := a.Add(xn), a.Sub(xn)

	for _, b := range points[1:] {
		abn := g.SegmentNormal(a, b).ScaleTo(R)

		dot := xn.Dot(abn)
		if dot == 0 { // straight segment
			b1, b2 := b.Add(xn), b.Sub(xn)
			list.Primitive_Quad(x1, b1, b2, x2, color)
			x1, x2 = b1, b2
		} else {
			scale := s2R2 / g.Sqrt(xn.Dot(abn)+R2)
			if scale < 2*R {
				// corner without chamfer
				xabn := xn.Add(abn)
				pbcn := xabn.ScaleTo(scale)
				b1, b2 := a.Add(pbcn), a.Sub(pbcn)
				list.Primitive_Quad(x1, b1, b2, x2, color)
				x1, x2 = b1, b2
			} else {
				// corner with chamfer and overlap
				b1, b2 := a.Add(xn), a.Sub(xn)
				list.Primitive_Quad(x1, b1, b2, x2, color)

				x1, x2 = a.Add(abn), a.Sub(abn)

				dot := xn.Rotate().Dot(abn)
				if dot < 0 {
					list.Primitive_Tri(b1, x1, a, color)
				} else if dot > 0 {
					list.Primitive_Tri(b2, a, x2, color)
				}
			}
		}

		a, xn = b, abn
	}

	a, b = points[len(points)-2], points[len(points)-1]
	b1, b2 := b.Add(xn), b.Sub(xn)
	list.Primitive_Quad(x1, b1, b2, x2, color)

	list.CurrentCommand.Count += Index(len(list.Indicies) - startIndexCount)
}

func (list *List) StrokeColoredLine(points []g.Vector, thickness float32, colors []g.Color) {
	if len(points) < 2 || thickness == 0 || len(colors) != len(points) {
		return
	}

	// TODO: optimize for thin line
	startIndexCount := len(list.Indicies)

	R := g.Abs(thickness / 2.0)
	R2 := R * R
	s2R2 := g.Sqrt2 * R2

	// draw each segment, where
	// ---------^--------x1-------^---------b1
	// |        | xn     |        | abn      |
	// | - - - - - - - - a - - - - - - - - - b
	// |                 |                   |
	// ------------------x2-----------------b2
	// x1, x2, xn are the previous segments end corners and normal
	x, b := points[0], points[1]
	acolor := colors[0]
	xn := g.SegmentNormal(x, b).ScaleTo(R)
	x1, x2 := x.Add(xn), x.Sub(xn)

	for i, b := range points[1:] {
		bcolor := colors[i+1]
		abn := g.SegmentNormal(x, b).ScaleTo(R)

		dot := xn.Dot(abn)
		if dot == 0 { // straight segment
			b1, b2 := b.Add(xn), b.Sub(xn)
			list.Primitive_QuadColor(x1, b1, b2, x2, acolor, bcolor, bcolor, acolor)
			x1, x2 = b1, b2
		} else {
			scale := s2R2 / g.Sqrt(xn.Dot(abn)+R2)
			if scale < 2*R {
				// corner without chamfer
				xabn := xn.Add(abn)
				pbcn := xabn.ScaleTo(scale)
				b1, b2 := x.Add(pbcn), x.Sub(pbcn)
				list.Primitive_QuadColor(x1, b1, b2, x2, acolor, bcolor, bcolor, acolor)
				x1, x2 = b1, b2
			} else {
				// corner with chamfer and overlap
				b1, b2 := x.Add(xn), x.Sub(xn)
				list.Primitive_QuadColor(x1, b1, b2, x2, acolor, bcolor, bcolor, acolor)

				x1, x2 = x.Add(abn), x.Sub(abn)

				dot := xn.Rotate().Dot(abn)
				if dot < 0 {
					list.Primitive_Tri(b1, x1, x, bcolor)
				} else if dot > 0 {
					list.Primitive_Tri(b2, x, x2, bcolor)
				}
			}
		}

		x, xn, acolor = b, abn, bcolor
	}

	xcolor := colors[len(colors)-2]
	b1, b2 := x.Add(xn), x.Sub(xn)
	list.Primitive_QuadColor(x1, b1, b2, x2, xcolor, acolor, acolor, xcolor)

	list.CurrentCommand.Count += Index(len(list.Indicies) - startIndexCount)
}

func (list *List) StrokeClosedLine(points []g.Vector, thickness float32, color g.Color) {
	if len(points) < 2 || color.Transparent() || thickness == 0 {
		return
	}
	if len(points) < 3 {
		list.StrokeLine(points, thickness, color)
		return
	}

	startIndexCount := len(list.Indicies)

	R := g.Abs(thickness / 2.0)
	R2 := R * R
	s2R2 := g.Sqrt2 * R2

	// draw each segment, where
	// x1-------^--------a1-------^---------b1
	// |        | xn     |        | abn      |
	// | - - - - - - - - a - - - - - - - - - b
	// |                 |                   |
	// x2----------------a2-----------------b2
	// x1, x2, xn are the previous segments end corners and normal
	var x1, x2 g.Vector

	x, a, b := points[len(points)-1], points[0], points[1]
	xn := g.SegmentNormal(x, a).ScaleTo(R)
	abn := g.SegmentNormal(a, b).ScaleTo(R)

	var t1, t2 g.Vector
	dot := xn.Dot(abn)
	if dot == 0 { // straight segment
		x1, x2 = a.Add(xn), a.Sub(xn)
		t1, t2 = x1, x2
	} else {
		scale := s2R2 / g.Sqrt(dot+R2)
		if scale < 2*R {
			// corner without chamfer
			xabn := xn.Add(abn)
			pbcn := xabn.ScaleTo(scale)
			x1, x2 = a.Add(pbcn), a.Sub(pbcn)
			t1, t2 = x1, x2
		} else {
			// corner with chamfer and overlap
			t1, t2 = a.Add(xn), a.Sub(xn)
			x1, x2 = a.Add(abn), a.Sub(abn)

			dot := xn.Rotate().Dot(abn)
			if dot < 0 {
				list.Primitive_Tri(t1, x1, a, color)
			} else if dot > 0 {
				list.Primitive_Tri(t2, a, x2, color)
			}
		}
	}

	for i := 1; i < len(points)+1; i++ {
		var b g.Vector
		if i >= len(points) {
			b = points[0]
		} else {
			b = points[i]
		}

		abn := g.SegmentNormal(a, b).ScaleTo(R)
		dot := xn.Dot(abn)
		if dot == 0 { // straight segment
			b1, b2 := b.Add(xn), b.Sub(xn)
			list.Primitive_Quad(x1, b1, b2, x2, color)
			x1, x2 = b1, b2
		} else {
			scale := s2R2 / g.Sqrt(dot+R2)
			if scale < 2*R {
				// corner without chamfer
				xabn := xn.Add(abn)
				pbcn := xabn.ScaleTo(scale)
				b1, b2 := a.Add(pbcn), a.Sub(pbcn)
				list.Primitive_Quad(x1, b1, b2, x2, color)
				x1, x2 = b1, b2
			} else {
				// corner with chamfer and overlap
				b1, b2 := a.Add(xn), a.Sub(xn)
				list.Primitive_Quad(x1, b1, b2, x2, color)

				x1, x2 = a.Add(abn), a.Sub(abn)

				dot := xn.Rotate().Dot(abn)
				if dot < 0 {
					list.Primitive_Tri(b1, x1, a, color)
				} else if dot > 0 {
					list.Primitive_Tri(b2, a, x2, color)
				}
			}
		}

		a, xn = b, abn
	}
	list.Primitive_Quad(x1, t1, t2, x2, color)

	list.CurrentCommand.Count += Index(len(list.Indicies) - startIndexCount)
}

const segmentsPerArc = 24

func (list *List) FillArc(center g.Vector, R float32, start, sweep float32, color g.Color) {
	if color.Transparent() || R == 0 {
		return
	}
	R = g.Abs(R)
	startIndexCount := len(list.Indicies)

	// N := sweep * R gives one segment per pixel
	N := Index(g.Clamp(g.Abs(sweep)*R/g.Tau, 3, segmentsPerArc))

	theta := sweep / float32(N)
	rots, rotc := g.Sincos(theta)
	dy, dx := g.Sincos(start)
	dy *= R
	dx *= R

	// add center point to the vertex buffer
	base := Index(len(list.Vertices))
	list.Vertices = append(list.Vertices, Vertex{center, NoUV, color})
	// add the first point the vertex buffer
	p := g.Vector{center.X + dx, center.Y + dy}
	list.Vertices = append(list.Vertices, Vertex{p, NoUV, color})
	// loop over rest of the points
	for i := Index(0); i < N; i++ {
		dx, dy = dx*rotc-dy*rots, dx*rots+dy*rotc
		p = g.Vector{center.X + dx, center.Y + dy}
		list.Vertices = append(list.Vertices, Vertex{p, NoUV, color})
		list.Indicies = append(list.Indicies, base, base+i+1, base+i+2)
	}

	list.CurrentCommand.Count += Index(len(list.Indicies) - startIndexCount)
}

func (list *List) FillCircle(center g.Vector, R float32, color g.Color) {
	if color.Transparent() || R == 0 {
		return
	}
	R = g.Abs(R)
	startIndexCount := len(list.Indicies)

	// N := 2 * PI * R gives one segment per pixel
	N := Index(g.Clamp(R, 3, segmentsPerArc))

	theta := g.Tau / float32(N)
	rots, rotc := g.Sincos(theta)

	dx, dy := R, float32(0)

	// add center point to the vertex buffer
	base := Index(len(list.Vertices))
	list.Vertices = append(list.Vertices, Vertex{center, NoUV, color})
	// add the first point the vertex buffer
	p := g.Vector{center.X + dx, center.Y + dy}
	list.Vertices = append(list.Vertices, Vertex{p, NoUV, color})

	// loop over rest of the points
	for i := Index(0); i < N; i++ {
		dx, dy = dx*rotc-dy*rots, dx*rots+dy*rotc
		p = g.Vector{center.X + dx, center.Y + dy}
		list.Vertices = append(list.Vertices, Vertex{p, NoUV, color})
		list.Indicies = append(list.Indicies, base, base+i+1, base+i+2)
	}

	list.CurrentCommand.Count += Index(len(list.Indicies) - startIndexCount)
}
