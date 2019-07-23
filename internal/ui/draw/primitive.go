package draw

import "github.com/loov/allocview/internal/ui/g"

func (list *List) BeginCommand() {
	if list.CurrentCommand != nil &&
		list.CurrentCommand.Count == 0 &&
		list.CurrentCommand.Callback == nil {
		*list.CurrentCommand = Command{
			Clip:    list.CurrentClip,
			Texture: list.CurrentTexture,
		}
		return
	}

	list.Commands = append(list.Commands, Command{
		Clip:    list.CurrentClip,
		Texture: list.CurrentTexture,
	})
	list.CurrentCommand = &list.Commands[len(list.Commands)-1]
}

func (list *List) Primitive_Ensure(index_count, vertex_count int) {
	if list.CurrentCommand.Count+Index(index_count) > CommandSplitThreshold {
		list.BeginCommand()
	}
}

func (list *List) Primitive_Reserve(index_count, vertex_count int) {
	list.Primitive_Ensure(index_count, vertex_count)
	list.CurrentCommand.Count += Index(index_count)
}

func (list *List) Primitive_Rect(r *g.Rect, color g.Color) {
	a, b, c, d := r.Corners()

	base := Index(len(list.Vertices))
	list.Indicies = append(list.Indicies,
		base+0, base+1, base+2,
		base+0, base+2, base+3,
	)
	list.Vertices = append(list.Vertices,
		Vertex{a, NoUV, color},
		Vertex{b, NoUV, color},
		Vertex{c, NoUV, color},
		Vertex{d, NoUV, color},
	)
}

func (list *List) Primitive_RectUV(r *g.Rect, uv *g.Rect, color g.Color) {
	a, b, c, d := r.Corners()
	uv_a, uv_b, uv_c, uv_d := uv.Corners()

	base := Index(len(list.Vertices))
	list.Indicies = append(list.Indicies,
		base+0, base+1, base+2,
		base+0, base+2, base+3,
	)
	list.Vertices = append(list.Vertices,
		Vertex{a, uv_a, color},
		Vertex{b, uv_b, color},
		Vertex{c, uv_c, color},
		Vertex{d, uv_d, color},
	)
}

func (list *List) Primitive_Quad(a, b, c, d g.Vector, color g.Color) {
	base := Index(len(list.Vertices))
	list.Indicies = append(list.Indicies,
		base+0, base+1, base+2,
		base+0, base+2, base+3,
	)
	list.Vertices = append(list.Vertices,
		Vertex{a, NoUV, color},
		Vertex{b, NoUV, color},
		Vertex{c, NoUV, color},
		Vertex{d, NoUV, color},
	)
}

func (list *List) Primitive_Tri(a, b, c g.Vector, color g.Color) {
	base := Index(len(list.Vertices))
	list.Indicies = append(list.Indicies,
		base+0, base+1, base+2,
	)
	list.Vertices = append(list.Vertices,
		Vertex{a, NoUV, color},
		Vertex{b, NoUV, color},
		Vertex{c, NoUV, color},
	)
}

func (list *List) Primitive_QuadUV(q *[4]g.Vector, uv *g.Rect, color g.Color) {
	a, b, c, d := q[0], q[1], q[2], q[3]
	uv_a, uv_b, uv_c, uv_d := uv.Corners()

	base := Index(len(list.Vertices))
	list.Indicies = append(list.Indicies,
		base+0, base+1, base+2,
		base+0, base+2, base+3,
	)
	list.Vertices = append(list.Vertices,
		Vertex{a, uv_a, color},
		Vertex{b, uv_b, color},
		Vertex{c, uv_c, color},
		Vertex{d, uv_d, color},
	)
}

func (list *List) Primitive_QuadColor(a, b, c, d g.Vector, acolor, bcolor, ccolor, dcolor g.Color) {
	base := Index(len(list.Vertices))
	list.Indicies = append(list.Indicies,
		base+0, base+1, base+2,
		base+0, base+2, base+3,
	)
	list.Vertices = append(list.Vertices,
		Vertex{a, NoUV, acolor},
		Vertex{b, NoUV, bcolor},
		Vertex{c, NoUV, ccolor},
		Vertex{d, NoUV, dcolor},
	)
}

func (list *List) Primitive_TriColor(a, b, c g.Vector, acolor, bcolor, ccolor g.Color) {
	base := Index(len(list.Vertices))
	list.Indicies = append(list.Indicies,
		base+0, base+1, base+2,
	)
	list.Vertices = append(list.Vertices,
		Vertex{a, NoUV, acolor},
		Vertex{b, NoUV, bcolor},
		Vertex{c, NoUV, ccolor},
	)
}
