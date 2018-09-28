package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	out := ""
	pe, err := parseOutput(out)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pe)
}

// UnexpectedConfiguration holds information about invalid directives
// in a NGINX configuration file
type UnexpectedConfiguration struct {
	File      string
	Line      int
	Directive string
}

// Equal tests for equality between two UnexpectedConfiguration types
func (ce *UnexpectedConfiguration) Equal(ce2 *UnexpectedConfiguration) bool {
	if ce.File != ce2.File {
		return false
	}

	if ce.Line != ce2.Line {
		return false
	}

	if ce.Directive != ce2.Directive {
		return false
	}

	return true
}

var (
	directiveRegex = regexp.MustCompile(`: (.*) directive invalid value in (.*):(\d+)`)
)

func parseOutput(out string) (*UnexpectedConfiguration, error) {
	result := directiveRegex.FindStringSubmatch(out)
	if len(result) == 4 {
		line, _ := strconv.Atoi(strings.TrimSpace(result[3]))
		return &UnexpectedConfiguration{
			File:      result[2],
			Line:      line,
			Directive: result[1],
		}, nil
	}

	return nil, fmt.Errorf("It was not possible to parse NGINX error")
}
