package symbols_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/loov/allocview/internal/symbols"
)

func TestSymbols(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempdir)

	binpath := filepath.Join(tempdir, "test.exe")

	build := exec.Command("go", "build", "-o", binpath, "./testdata")
	_, err = build.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	expectedLines := []int{23, 17, 12}

	out, err := exec.Command(binpath).CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}

	// binary outputs things in this order:

	// <symbol name>
	// <symbol address>
	// <caller> ...

	lines := strings.Split(string(out), "\n")
	symbolName := strings.TrimSpace(lines[0])
	symbolAddr, err := strconv.ParseUint(lines[1], 10, 64)
	if err != nil {
		t.Fatal(err)
	}

	callers := []uint64{}
	for _, frame := range lines[2:] {
		frame = strings.TrimSpace(frame)
		if frame == "" {
			continue
		}
		pc, err := strconv.ParseUint(frame, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		callers = append(callers, pc)
	}
	if len(callers) < 3 {
		t.Fatal("not enough callers")
	}

	bin, err := symbols.Load(binpath)
	if err != nil {
		t.Fatal(err)
	}

	sym := bin.SymTable.LookupFunc(symbolName)
	symbolOffset := sym.Entry - symbolAddr

	for i, expline := range expectedLines {
		pc := callers[i]
		_, line, _ := bin.SymTable.PCToLine(pc + symbolOffset - 1)
		if expline != line {
			t.Errorf("got line %d, expected %d", line, expline)
		}
	}
}
