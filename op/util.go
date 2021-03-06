package op

import (
	"bytes"
	"fmt"
	"github.com/emicklei/melrose/core"
	"io"
	"strconv"
	"strings"
	"text/scanner"
)

func appendStorexList(b *bytes.Buffer, isFirstParameter bool, list []core.Sequenceable) {
	if len(list) == 0 {
		return
	}
	if !isFirstParameter {
		fmt.Fprintf(b, ",")
	}
	for i, each := range list {
		if s, ok := each.(core.Storable); !ok {
			fmt.Fprintf(b, "nil")
		} else {
			fmt.Fprintf(b, "%s", s.Storex())
		}
		if i < len(list)-1 {
			io.WriteString(b, ",")
		}
	}
}

// "1 (4 5 6) 2 (4 5 6) 3 (4 5 6) 2 (4 5 6)"
func parseIndices(src string) [][]int {
	ii := [][]int{}
	var s scanner.Scanner
	s.Init(strings.NewReader(src))
	container := []int{}
	ingroup := false
	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		switch tok {
		case '(':
			if len(container) > 0 {
				ii = append(ii, container)
			}
			container = []int{}
			ingroup = true
		case ')':
			if len(container) > 0 {
				ii = append(ii, container)
			}
			container = []int{}
			ingroup = false
		default:
			i, err := strconv.Atoi(s.TokenText())
			if err != nil {
				i = 0 // set to invalid one
			} else {
				if ingroup {
					container = append(container, i)
				} else {
					ii = append(ii, []int{i})
				}
			}
		}
	}
	if len(container) > 0 {
		ii = append(ii, container)
	}
	return ii
}

func replacedAll(target []core.Sequenceable, from, to core.Sequenceable) []core.Sequenceable {
	newTarget := []core.Sequenceable{}
	for _, each := range target {
		if core.IsIdenticalTo(each, from) {
			newTarget = append(newTarget, to)
		} else {
			if other, ok := each.(core.Replaceable); ok {
				newTarget = append(newTarget, other.Replaced(from, to))
			} else {
				newTarget = append(newTarget, each)
			}
		}
	}
	return newTarget
}

// "1 (4 5 6) 2 (4 5 6) 3 (4 5 6) 2 (4 5 6)"
func formatIndices(src [][]int) string {
	var b bytes.Buffer
	for _, each := range src {
		if len(each) == 1 {
			fmt.Fprintf(&b, "%d ", each[0])
		} else {
			fmt.Fprintf(&b, "(")
			for _, other := range each {
				fmt.Fprintf(&b, "%d ", other)
			}
			fmt.Fprintf(&b, ") ")
		}
	}
	return b.String()
}
