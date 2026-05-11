package main

import (
	"fmt"

	"git.sr.ht/~ser/godjot/v2/djot_html"
	"git.sr.ht/~ser/godjot/v2/djot_parser"
)

func ConvertDjot(content []byte) (html string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if line := locateDjotFailureLine(content); line > 0 {
				err = fmt.Errorf("godjot failed near line %d: %v", line, r)
			} else {
				err = fmt.Errorf("godjot failed: %v", r)
			}
		}
	}()

	return convertDjot(content), nil
}

func convertDjot(content []byte) string {
	ast := djot_parser.BuildDjotAst(content)
	return djot_html.New().ConvertDjot(&djot_html.HtmlWriter{}, ast...).String()
}

func locateDjotFailureLine(content []byte) int {
	return firstFailingLine(content, func(prefix []byte) bool {
		return panics(func() {
			convertDjot(prefix)
		})
	})
}

func firstFailingLine(content []byte, fails func([]byte) bool) int {
	for line, end := 1, 0; end < len(content); line++ {
		for end < len(content) && content[end] != '\n' {
			end++
		}
		if end < len(content) {
			end++
		}
		if fails(content[:end]) {
			return line
		}
	}

	if len(content) == 0 {
		return 0
	}
	if fails(content) {
		return 1
	}
	return 0
}

func panics(f func()) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = true
		}
	}()
	f()
	return false
}
