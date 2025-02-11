package tap

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

type TAP struct {
	mu          sync.Mutex
	count       int
	failed      int
	planned     int
	indent      string
	planPrinted bool
}

func Run(fn func(t *TAP)) bool {
	t := &TAP{planned: -1, indent: ""}
	fn(t)
	if !t.summary() {
		return false
	}
	return true
}

func (t *TAP) Plan(n int) {
	t.planned = n
	t.printPlan()
}

func (t *TAP) printPlan() {
	if !t.planPrinted {
		fmt.Printf("%s1..%d\n", t.indent, max(t.planned, 0))
		t.planPrinted = true
	}
}

func (t *TAP) Pass(name string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.count++
	if t.count == 1 && t.planned == -1 {
		t.printPlan()
	}
	fmt.Printf("%sok %d - %s\n", t.indent, t.count, name)
	return true
}

func (t *TAP) Fail(name string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.count++
	t.failed++
	if t.count == 1 && t.planned == -1 {
		t.printPlan()
	}
	fmt.Printf("%snot ok %d - %s\n", t.indent, t.count, name)
	return false
}

func (t *TAP) Ok(cond bool, name string) bool {
	if cond {
		return t.Pass(name)
	}
	return t.Fail(name)
}

func (t *TAP) FOk(f func() bool, name string) bool {
	var (
		buf bytes.Buffer
		ok  bool
	)

	captureOutput(&buf, func() {
		defer func() {
			if r := recover(); r != nil {
				ok = false
			}
		}()
		ok = f()
	})

	ok = t.Ok(ok, name)
	for _, line := range strings.Split(buf.String(), "\n") {
		if line != "" {
			fmt.Printf("%s# %s\n", t.indent, line)
		}
	}
	return ok
}

func (t *TAP) Diag(msg string) {
	fmt.Printf("%s# %s\n", t.indent, msg)
}

func (t *TAP) Subtest(desc string, fn func(sub *TAP)) {
	fmt.Printf("%s# Subtest: %s\n", t.indent, desc)
	sub := &TAP{planned: -1, indent: t.indent + "  "}
	fn(sub)

	// Проверяем статус sub
	success := sub.summary()
	if !success {
		t.failed++ // Родительский тест должен считать sub как проваленный
		fmt.Printf("%snot ok %d - %s\n", t.indent, t.count+1, desc)
	} else {
		fmt.Printf("%sok %d - %s\n", t.indent, t.count+1, desc)
	}
	t.count++ // Не забываем увеличивать общий счётчик
}

func (t *TAP) summary() bool {
	ok := true
	if t.planned != -1 && t.count != t.planned {
		fmt.Printf("%s# Planned %d tests but ran %d\n", t.indent, t.planned, t.count)
		ok = false
	}
	if t.failed > 0 {
		fmt.Printf("%s# Tests failed: %d/%d\n", t.indent, t.failed, t.count)
		ok = false
	}
	return ok
}

func captureOutput(buf *bytes.Buffer, fn func()) {
	saveStdout := os.Stdout
	saveStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	fn()

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = saveStdout
	os.Stderr = saveStderr

	buf.Write(out)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
