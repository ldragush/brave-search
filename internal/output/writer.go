package output

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

type Writer struct {
	mu     sync.Mutex
	silent bool
	w      *bufio.Writer
}

func NewWriter(file *os.File, silent bool) *Writer {
	return &Writer{
		silent: silent,
		w:      bufio.NewWriterSize(file, 64*1024),
	}
}

func (w *Writer) Println(s string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	fmt.Fprintln(w.w, s)
	_ = w.w.Flush()
}

func WriteLinesToFile(path string, lines []string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	bw := bufio.NewWriterSize(f, 64*1024)
	for _, l := range lines {
		fmt.Fprintln(bw, l)
	}
	return bw.Flush()
}
