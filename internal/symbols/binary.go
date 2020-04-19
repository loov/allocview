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
	Data *dwarf.Data

	SymTable *gosym.Table

	Offset int64
}

func Load(path string) (*Binary, error) {
	data, symtab, err := loadDwarfData(path)
	if err != nil {
		return nil, fmt.Errorf("unable to load dwarf data: %w", err)
	}

	return &Binary{
		Data: data,

		SymTable: symtab,
	}, nil
}

func (bin *Binary) UpdateOffset(funcname string, funcaddr uintptr) {
	// TODO: handle errors
	sym := bin.SymTable.LookupFunc(funcname)
	if sym == nil {
		return
	}
	bin.Offset = int64(sym.Entry) - int64(funcaddr)
}

func loadDwarfData(path string) (*dwarf.Data, *gosym.Table, error) {
	{ // try elf
		f, err := elf.Open(path)
		if err == nil {
			d, err := f.DWARF()
			if err != nil {
				return nil, nil, fmt.Errorf("elf %q: unable to read: %w", path, err)
			}

			symtab, err := elfLoadTable(f)
			if err != nil {
				return nil, nil, fmt.Errorf("pe %q: unable to read symtab: %w", path, err)
			}

			return d, symtab, nil
		}
	}

	{ // try macho
		f, err := macho.Open(path)
		if err == nil {
			d, err := f.DWARF()
			if err != nil {
				return nil, nil, fmt.Errorf("macho %q: unable to read: %w", path, err)
			}

			symtab, err := machoLoadTable(f)
			if err != nil {
				return nil, nil, fmt.Errorf("pe %q: unable to read symtab: %w", path, err)
			}

			return d, symtab, nil
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

func elfLoadTable(f *elf.File) (*gosym.Table, error) {
	section := f.Section(".gopclntab")
	if section == nil {
		return nil, fmt.Errorf("unable to find pclntab")
	}

	data, err := section.Data()
	if err != nil {
		return nil, fmt.Errorf("unable to read pclntab data: %w", err)
	}

	lineTable := gosym.NewLineTable(data, f.Section(".text").Addr)
	symTable, err := gosym.NewTable(nil, lineTable)
	if err != nil {
		return nil, fmt.Errorf("unable to create gosym table: %w", err)
	}
	return symTable, nil
}

func machoLoadTable(f *macho.File) (*gosym.Table, error) {
	section := f.Section("__gopclntab")
	if section == nil {
		return nil, fmt.Errorf("unable to find pclntab")
	}

	data, err := section.Data()
	if err != nil {
		return nil, fmt.Errorf("unable to read pclntab data: %w", err)
	}

	lineTable := gosym.NewLineTable(data, f.Section("__text").Addr)
	symTable, err := gosym.NewTable(nil, lineTable)
	if err != nil {
		return nil, fmt.Errorf("unable to create gosym table: %w", err)
	}
	return symTable, nil
}

func peLoadTable(f *pe.File) (*gosym.Table, error) {
	pclntab, err := peLoadRange(f, "runtime.pclntab", "runtime.epclntab")
	if err != nil {
		return nil, fmt.Errorf("unable to read pclntab: %w", err)
	}

	symtab, err := peLoadRange(f, "runtime.symtab", "runtime.esymtab")
	if err != nil {
		return nil, fmt.Errorf("unable to read symtab: %w", err)
	}

	lineTable := gosym.NewLineTable(pclntab, 0)
	symTable, err := gosym.NewTable(symtab, lineTable)
	if err != nil {
		return nil, fmt.Errorf("unable to create gosym table: %w", err)
	}
	return symTable, nil
}

func peLoadRange(f *pe.File, sname, ename string) ([]byte, error) {
	start := peFindSymbol(f, sname)
	if start == nil {
		return nil, fmt.Errorf("unable to find start symbol %q", sname)
	}

	end := peFindSymbol(f, ename)
	if end == nil {
		return nil, fmt.Errorf("unable to find end symbol %q", ename)
	}

	if start.SectionNumber != end.SectionNumber {
		return nil, fmt.Errorf("start %q and end %q section do not match", sname, ename)
	}

	sect := f.Sections[start.SectionNumber-1]
	data, err := sect.Data()
	if err != nil {
		return nil, fmt.Errorf("start %q and end %q section failed to load", sname, ename)
	}

	return data[start.Value:end.Value], nil
}

func peFindSymbol(f *pe.File, name string) *pe.Symbol {
	for _, sym := range f.Symbols {
		if sym.Name == name {
			return sym
		}
	}
	return nil
}
