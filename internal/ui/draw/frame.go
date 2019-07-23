package draw

type Frame struct {
	Textures *Textures
	Lists    []*List
	lists    []*List
}

func (frame *Frame) Reset() {
	if frame.Textures == nil {
		frame.Textures = NewTextures()
	}
	frame.lists = frame.Lists
	frame.Lists = nil
}

func (frame *Frame) Layer() *List {
	var list *List

	if len(frame.lists) > 0 {
		list = frame.lists[0]
		frame.lists = frame.lists[1:]
	} else {
		list = NewList(frame.Textures)
	}
	list.Reset(frame.Textures)
	frame.Lists = append(frame.Lists, list)
	return list
}
