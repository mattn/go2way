package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
)

type Token int

const (
	ILLEGAL Token = iota
	SPACE
	STRING
	COMP
	TOKEN
	QUOTE
	OPERATOR
	COMMENT
)

type State int

type Scanner struct {
	last    Token
	curr    Token
	scan    *bufio.Scanner
	quote   bool
	comment bool
}

func (s *Scanner) classOf(r rune) Token {
	if r == ' ' || r == '\t' || r == '\r' || r == '\n' {
		return SPACE
	}
	if r == '=' || r == '<' || r == '>' || r == '!' {
		return COMP
	}
	if r == '+' || r == '-' || r == '*' || r == '/' {
		return OPERATOR
	}
	if '0' <= r && r <= '9' || ('a' <= r && r <= 'z') || ('A' <= r && r <= 'A') {
		return TOKEN
	}
	if r == '\'' {
		return QUOTE
	}
	return ILLEGAL
}

func (s *Scanner) splitToken(data []byte, atEOF bool) (int, []byte, error) {
	bpos := 0
	b := data
	s.curr = s.last

	var clazz Token
	for {
		r, i := utf8.DecodeRune(b)
		if i == 0 {
			break
		}
		if len(b) > 2 && ((r == '/' && b[1] == '*') || (r == '*' && b[1] == '/')) {
			clazz = COMMENT
			i++
		} else {
			clazz = s.classOf(r)
		}

		if s.comment {
			if bpos == 0 && clazz == COMMENT {
				s.comment = false
			}
		} else if clazz == COMMENT {
			s.comment = true
		}

		if s.quote {
			if bpos > 0 && clazz == QUOTE {
				s.quote = false
			} else {
				clazz = QUOTE
			}
		} else if clazz == QUOTE {
			s.quote = true
		}

		if clazz != s.last {
			s.last = clazz
			break
		}
		bpos += i
		b = b[i:]
	}
	var err error
	if atEOF {
		err = io.EOF
	}
	return bpos, data[:bpos], err
}

func NewScanner(r io.Reader) *Scanner {
	s := bufio.NewScanner(r)
	scan := &Scanner{
		scan: s,
		curr: ILLEGAL,
		last: SPACE,
	}
	s.Split(scan.splitToken)
	return scan
}

func (s *Scanner) Text() string {
	return s.scan.Text()
}

func (s *Scanner) InComment() bool {
	return s.comment
}

func (s *Scanner) Token() Token {
	return s.curr
}

func (s *Scanner) Scan() bool {
	return s.scan.Scan()
}

func main() {
	s := `
	select * from foo where id = /*a*/'foo bar'
	`

	out := colorable.NewColorableStdout()
	scan := NewScanner(strings.NewReader(s))
	for scan.Scan() {
		switch scan.Token() {
		case COMMENT:
			fmt.Fprint(out, color.BlueString(scan.Text()))
		case OPERATOR:
			fmt.Fprint(out, color.MagentaString(scan.Text()))
		case TOKEN:
			if scan.InComment() {
				fmt.Fprint(out, color.YellowString(scan.Text()))
			} else {
				fmt.Fprint(out, color.GreenString(scan.Text()))
			}
		case QUOTE:
			fmt.Fprint(out, color.RedString(scan.Text()))
		default:
			fmt.Fprint(out, color.WhiteString(scan.Text()))
		}
	}
}