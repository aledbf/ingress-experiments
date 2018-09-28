package main

import (
	"testing"
)

func TestParseOutput(t *testing.T) {
	testCases := map[string]struct {
		input   string
		isError bool
		output  *UnexpectedConfiguration
	}{
		"invalid configuration should return parse error": {
			input: `Error: exit status 1
			2018/09/27 17:00:00 [emerg] 198#198: "client_max_body_size" directive invalid value in /tmp/nginx-cfg353378898:425
			nginx: [emerg] "client_max_body_size" directive invalid value in /tmp/nginx-cfg353378898:425
			nginx: configuration file /tmp/nginx-cfg353378898 test failed
			`,
			output: &UnexpectedConfiguration{
				File:      "/tmp/nginx-cfg353378898",
				Line:      425,
				Directive: "\"client_max_body_size\"",
			},
		},
	}

	for title, testCase := range testCases {
		t.Run(title, func(t *testing.T) {
			po, err := parseOutput(testCase.input)
			if err != nil && !testCase.isError {
				t.Fatal("Expected error")
			}

			if po == nil {
				t.Error("Expected a parsed error but none returned")
			}

			if !testCase.output.Equal(po) {
				t.Errorf("Unexpected configuration error returned")
			}
		})
	}
}
