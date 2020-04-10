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

	LineTable *gosym.LineTable
	SymTable  *gosym.Table
}

func Load(path string) (*Binary, error) {
	data, pclntab, symtab, err := loadDwarfData(path)
	if err != nil {
		return nil, fmt.Errorf("unable to load dwarf data: %w", err)
	}

	lineTable := gosym.NewLineTable(pclntab, 0)
	symTable, err := gosym.NewTable(symtab, lineTable)
	if err != nil {
		return nil, fmt.Errorf("unable to create gosym table: %w", err)
	}

	lines, err := NewLines(data)
	if err != nil {
		return nil, fmt.Errorf("unable to load lines table: %w", err)
	}

	return &Binary{
		Data:  data,
		Lines: lines,

		LineTable: lineTable,
		SymTable:  symTable,
	}, nil
}

func loadDwarfData(path string) (_ *dwarf.Data, pclntab, symtab []byte, _ error) {
	{ // try elf
		f, err := elf.Open(path)
		if err == nil {
			d, err := f.DWARF()
			if err != nil {
				return nil, nil, nil, fmt.Errorf("elf %q: unable to read: %w", path, err)
			}
			return d, nil, nil, nil
		}
	}

	{ // try macho
		f, err := macho.Open(path)
		if err == nil {
			d, err := f.DWARF()
			if err != nil {
				return nil, nil, nil, fmt.Errorf("macho %q: unable to read: %w", path, err)
			}
			return d, nil, nil, nil
		}
	}

	{ // try pe
		f, err := pe.Open(path)
		if err == nil {
			d, err := f.DWARF()
			if err != nil {
				return nil, nil, nil, fmt.Errorf("pe %q: unable to read: %w", path, err)
			}

			pclntab, err := peLoadTable(f, "runtime.pclntab", "runtime.epclntab")
			if err != nil {
				return nil, nil, nil, fmt.Errorf("pe %q: unable to read pclntab: %w", path, err)
			}

			symtab, err := peLoadTable(f, "runtime.symtab", "runtime.esymtab")
			if err != nil {
				return nil, nil, nil, fmt.Errorf("pe %q: unable to read symtab: %w", path, err)
			}

			return d, pclntab, symtab, nil
		}
	}

	return nil, nil, nil, fmt.Errorf("%q has unknown format", path)
}

func peLoadTable(pefile *pe.File, sname, ename string) ([]byte, error) {
	start := peFindSymbol(pefile, sname)
	if start == nil {
		return nil, fmt.Errorf("unable to find start symbol %q", sname)
	}

	end := peFindSymbol(pefile, ename)
	if end == nil {
		return nil, fmt.Errorf("unable to find end symbol %q", ename)
	}

	if start.SectionNumber != end.SectionNumber {
		return nil, fmt.Errorf("start %q and end %q section do not match", sname, ename)
	}

	sect := pefile.Sections[start.SectionNumber-1]
	data, err := sect.Data()
	if err != nil {
		return nil, fmt.Errorf("start %q and end %q section failed to load", sname, ename)
	}

	return data[start.Value:end.Value], nil
}

func peFindSymbol(pefile *pe.File, name string) *pe.Symbol {
	for _, sym := range pefile.Symbols {
		if sym.Name == name {
			return sym
		}
	}
	return nil
}
