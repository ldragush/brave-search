package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/haltman-io/brave-search/internal/cli"
)

func GatherQueries(opts cli.Options, stdin io.Reader) ([]string, error) {
	var raw []string

	for _, v := range opts.SearchQueries {
		raw = append(raw, splitCommaList(v)...)
	}

	for _, v := range opts.SearchQueryFiles {
		files := splitCommaList(v)
		for _, f := range files {
			lines, err := readLinesFile(f)
			if err != nil {
				return nil, err
			}
			raw = append(raw, lines...)
		}
	}

	if opts.UseStdin {
		lines, err := readLinesReader(stdin)
		if err != nil {
			return nil, err
		}
		raw = append(raw, lines...)
	}

	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, q := range raw {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		if _, ok := seen[q]; ok {
			continue
		}
		seen[q] = struct{}{}
		out = append(out, q)
	}
	return out, nil
}

func splitCommaList(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func readLinesFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open query file %q: %w", path, err)
	}
	defer f.Close()
	return readLinesReader(f)
}

func readLinesReader(r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1024), 1024*1024)

	var out []string
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("failed to read lines: %w", err)
	}
	return out, nil
}
