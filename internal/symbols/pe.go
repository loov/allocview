package symbols

import (
	"debug/gosym"
	"debug/pe"
	"fmt"
)

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
