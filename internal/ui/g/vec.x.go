package g

var Rect01 = Rect{V0, V1}

func (a Vector) Rotate() Vector { return Vector{-a.Y, a.X} }

func (a Vector) ScaleTo(size float32) Vector {
	ilen := a.Len()
	if ilen > 0 {
		ilen = size / ilen
	}
	return a.Scale(ilen)
}

func (a Vector) Inflate(r Vector) Rect {
	return Rect{
		Min: a.Sub(r),
		Max: a.Add(r),
	}
}

func (a Vector) Rect() Rect { return Rect{a, a} }

func SegmentNormal(a, b Vector) Vector {
	return b.Sub(a).Rotate()
}

func (r Rect) AsInt32() (x, y, w, h int32) {
	x = int32(r.Min.X)
	y = int32(r.Min.Y)
	w = int32(r.Max.X - r.Min.X)
	h = int32(r.Max.Y - r.Min.Y)
	return
}

// Corners returns top-left, top-right, bottom-right, bottom-left vectors
func (r Rect) Corners() (tl, tr, br, bl Vector) {
	tl = r.TopLeft()
	tr = r.TopRight()
	br = r.BottomRight()
	bl = r.BottomLeft()
	return
}

func (r Rect) Clip(clip Rect) {
	if r.Min.X < clip.Min.X {
		r.Min.X = clip.Min.X
	}
	if r.Min.Y < clip.Min.Y {
		r.Min.Y = clip.Min.Y
	}
	if r.Max.X > clip.Max.X {
		r.Max.X = clip.Max.X
	}
	if r.Max.Y > clip.Max.Y {
		r.Max.Y = clip.Max.Y
	}
}

func (r Rect) Floor() {
	r.Min.X = (float32)((int)(r.Min.X))
	r.Min.Y = (float32)((int)(r.Min.Y))
	r.Max.X = (float32)((int)(r.Max.X))
	r.Max.Y = (float32)((int)(r.Max.Y))
}

func (r Rect) ClosestPoint(p Vector) Vector {
	if p.X > r.Max.X {
		p.X = r.Max.X
	} else if p.X < r.Min.X {
		p.X = r.Min.X
	}

	if p.Y > r.Max.Y {
		p.Y = r.Max.Y
	} else if p.Y < r.Min.Y {
		p.Y = r.Min.Y
	}

	return p
}

func (r Rect) ToGlobal(p Vector) Vector {
	return Vector{
		X: Lerp(p.X, r.Min.X, r.Max.X),
		Y: Lerp(p.Y, r.Min.Y, r.Max.Y),
	}
}

func (r Rect) ToRelative(p Vector) Vector {
	return Vector{
		X: InverseLerp(p.X, r.Min.X, r.Max.X),
		Y: InverseLerp(p.Y, r.Min.Y, r.Max.Y),
	}
}

func (r Rect) Subset(rel Rect) Rect {
	return Rect{
		Min: Vector{
			Lerp(rel.Min.X, r.Min.X, r.Max.X),
			Lerp(rel.Min.Y, r.Min.Y, r.Max.Y),
		},
		Max: Vector{
			X: Lerp(rel.Max.X, r.Min.X, r.Max.X),
			Y: Lerp(rel.Max.Y, r.Min.Y, r.Max.Y),
		},
	}
}

func (r Rect) Dx() float32 { return r.Max.X - r.Min.X }
func (r Rect) Dy() float32 { return r.Max.Y - r.Min.Y }

func (r Rect) TopLeft() Vector     { return r.Min }
func (r Rect) TopRight() Vector    { return Vector{r.Max.X, r.Min.Y} }
func (r Rect) BottomLeft() Vector  { return Vector{r.Min.X, r.Max.Y} }
func (r Rect) BottomRight() Vector { return r.Max }

func (r Rect) LeftCenter() Vector   { return Vector{r.Min.X, (r.Min.Y + r.Max.Y) / 2} }
func (r Rect) TopCenter() Vector    { return Vector{(r.Min.X + r.Max.X) / 2, r.Min.Y} }
func (r Rect) RightCenter() Vector  { return Vector{r.Max.X, (r.Min.Y + r.Max.Y) / 2} }
func (r Rect) BottomCenter() Vector { return Vector{(r.Min.X + r.Max.X) / 2, r.Max.Y} }

func (r Rect) VerticalLine(x, radius float32) Rect {
	r.Min.X = x - radius
	r.Max.X = x + radius
	return r
}

type Hit uint8

const (
	Inside = Hit(1 << iota)
	Left
	Top
	Right
	Bottom
)

func (h Hit) Contains(sub Hit) bool { return h&sub == sub }

func (r Rect) Test(p Vector, rad float32) Hit {
	var hit Hit
	if !r.Inflate(Vector{rad, rad}).Contains(p) {
		return hit
	}

	hit |= Inside
	if Abs(r.Min.X-p.X) <= rad {
		hit |= Left
	}
	if Abs(r.Min.Y-p.Y) <= rad {
		hit |= Top
	}
	if Abs(r.Max.X-p.X) <= rad {
		hit |= Right
	}
	if Abs(r.Max.Y-p.Y) <= rad {
		hit |= Bottom
	}

	return hit
}
