package cmd_test

import (
	"testing"

	"github.com/marstr/goalias/goalias/cmd"
)

func TestExpandPackageName(t *testing.T) {
	testCases := []struct {
		string
		expected string
	}{}

	for _, tc := range testCases {
		t.Run(tc.string, func(t *testing.T) {
			got, err := cmd.ExpandPackageName(tc.string)
			if err != nil {
				t.Error(err)
				return
			}

			if got != tc.expected {
				t.Logf("got:\n\t%q\nwant:\n\t%q", got, tc.expected)
				t.Fail()
			}
		})
	}
}
