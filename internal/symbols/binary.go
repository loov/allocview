package symbols

import (
	"debug/dwarf"
	"debug/elf"
	"debug/gosym"
	"debug/macho"
	"debug/pe"
	"fmt"
)

// Binary handles symbol lookup based on stack frames.
type Binary struct {
	Data  *dwarf.Data
	Lines *Lines

	SymTable *gosym.Table
}

func Load(path string) (*Binary, error) {
	data, symtab, err := loadDwarfData(path)
	if err != nil {
		return nil, fmt.Errorf("unable to load dwarf data: %w", err)
	}

	lines, err := NewLines(data)
	if err != nil {
		return nil, fmt.Errorf("unable to load lines table: %w", err)
	}

	return &Binary{
		Data:  data,
		Lines: lines,

		SymTable: symtab,
	}, nil
}

func loadDwarfData(path string) (*dwarf.Data, *gosym.Table, error) {
	{ // try elf
		f, err := elf.Open(path)
		if err == nil {
			d, err := f.DWARF()
			if err != nil {
				return nil, nil, fmt.Errorf("elf %q: unable to read: %w", path, err)
			}
			return d, nil, nil
		}
	}

	{ // try macho
		f, err := macho.Open(path)
		if err == nil {
			d, err := f.DWARF()
			if err != nil {
				return nil, nil, fmt.Errorf("macho %q: unable to read: %w", path, err)
			}
			return d, nil, nil
		}
	}

	{ // try pe
		f, err := pe.Open(path)
		if err == nil {
			d, err := f.DWARF()
			if err != nil {
				return nil, nil, fmt.Errorf("pe %q: unable to read: %w", path, err)
			}

			symtab, err := peLoadTable(f)
			if err != nil {
				return nil, nil, fmt.Errorf("pe %q: unable to read symtab: %w", path, err)
			}

			return d, symtab, nil
		}
	}

	return nil, nil, fmt.Errorf("%q has unknown format", path)
}
