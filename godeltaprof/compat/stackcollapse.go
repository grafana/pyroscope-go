package compat

import (
	"bytes"
	"io"
	"regexp"
	"sort"
	"strings"
	"testing"

	gprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stack struct {
	funcs []string
	line  string
	value []int64
}

func expectEmptyProfile(t *testing.T, buffer io.Reader) {
	profile, err := gprofile.Parse(buffer)
	require.NoError(t, err)
	ls := stackCollapseProfile(t, profile)
	assert.Empty(t, ls)
}

func expectNoStackFrames(t *testing.T, buffer *bytes.Buffer, sfPattern string) {
	profile, err := gprofile.ParseData(buffer.Bytes())
	require.NoError(t, err)
	line := findStack(t, stackCollapseProfile(t, profile), sfPattern)
	assert.Nilf(t, line, "stack frame %s found", sfPattern)
}

func expectStackFrames(t *testing.T, buffer *bytes.Buffer, sfPattern string, values ...int64) {
	profile, err := gprofile.ParseData(buffer.Bytes())
	require.NoError(t, err)
	line := findStack(t, stackCollapseProfile(t, profile), sfPattern)
	assert.NotNilf(t, line, "stack frame %s not found", sfPattern)
	if line != nil {
		for i := range values {
			assert.Equalf(t, values[i], line.value[i], "expected %v got %v", values, line.value)
		}
	}
}

func findStack(t *testing.T, res []stack, re string) *stack {
	rr := regexp.MustCompile(re)
	for i, re := range res {
		if rr.MatchString(re.line) {
			return &res[i]
		}
	}
	return nil
}

func stackCollapseProfile(t testing.TB, p *gprofile.Profile) []stack {
	var ret []stack
	for _, s := range p.Sample {
		var funcs []string
		for i := range s.Location {

			loc := s.Location[i]
			for _, line := range loc.Line {
				f := line.Function
				//funcs = append(funcs, fmt.Sprintf("%s:%d", f.Name, line.Line))
				funcs = append(funcs, f.Name)
			}
		}
		for i := 0; i < len(funcs)/2; i++ {
			j := len(funcs) - i - 1
			funcs[i], funcs[j] = funcs[j], funcs[i]
		}

		ret = append(ret, stack{
			line:  strings.Join(funcs, ";"),
			funcs: funcs,
			value: s.Value,
		})
	}
	sort.Slice(ret, func(i, j int) bool {
		return strings.Compare(ret[i].line, ret[j].line) < 0
	})
	var unique []stack
	for _, s := range ret {
		if len(unique) == 0 {
			unique = append(unique, s)
			continue
		}
		if unique[len(unique)-1].line == s.line {
			for i := 0; i < len(s.value); i++ {
				unique[len(unique)-1].value[i] += s.value[i]
			}
			continue
		}
		unique = append(unique, s)

	}
	t.Log("============= stackCollapseProfile ================")
	for _, s := range unique {
		t.Log(s.line, s.value)
	}
	t.Log("===================================================")

	return unique
}
