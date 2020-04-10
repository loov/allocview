package symbols

import "debug/dwarf"

type Lines struct {
	data *dwarf.Data
}

func NewLines(data *dwarf.Data) (*Lines, error) {
	lines := &Lines{data: data}
	return lines, nil
}
