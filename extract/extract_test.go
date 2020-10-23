package extract

import (
	"bytes"
	"os"
	"path"
	"strings"
	"testing"
)

var expectedOutput = `// Code generated by 'yaegi extract guthib.com/baz'. DO NOT EDIT.

package bar

import (
	"guthib.com/baz"
	"reflect"
)

func init() {
	Symbols["guthib.com/baz"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Hello": reflect.ValueOf(baz.Hello),
	}
}
`

func TestPackages(t *testing.T) {
	testCases := []struct {
		desc       string
		moduleOn   string
		wd         string
		arg        string
		importPath string
		expected   string
		contains   string
		dest       string
	}{
		{
			desc: "stdlib math pkg, using go/importer",
			dest: "math",
			arg:  "math",
			// We check this one because it shows both defects when we break it: the value
			// gets corrupted, and the type becomes token.INT
			// TODO(mpl): if the ident between key and value becomes annoying, be smarter about it.
			contains: `"MaxFloat64":             reflect.ValueOf(constant.MakeFromLiteral("179769313486231570814527423731704356798100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000", token.FLOAT, 0)),`,
		},
		{
			desc:     "using relative path, using go.mod",
			wd:       "./testdata/1/src/guthib.com/bar",
			arg:      "../baz",
			expected: expectedOutput,
		},
		{
			desc:       "using relative path, manual import path",
			wd:         "./testdata/2/src/guthib.com/bar",
			arg:        "../baz",
			importPath: "guthib.com/baz",
			expected:   expectedOutput,
		},
		{
			desc:       "using relative path, go.mod is ignored, because manual path",
			wd:         "./testdata/3/src/guthib.com/bar",
			arg:        "../baz",
			importPath: "guthib.com/baz",
			expected:   expectedOutput,
		},
		{
			desc:     "using relative path, dep in vendor, using go.mod",
			wd:       "./testdata/4/src/guthib.com/bar",
			arg:      "./vendor/guthib.com/baz",
			expected: expectedOutput,
		},
		{
			desc:       "using relative path, dep in vendor, manual import path",
			wd:         "./testdata/5/src/guthib.com/bar",
			arg:        "./vendor/guthib.com/baz",
			importPath: "guthib.com/baz",
			expected:   expectedOutput,
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			cwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			wd := test.wd
			if wd == "" {
				wd = cwd
			} else {
				if err := os.Chdir(wd); err != nil {
					t.Fatal(err)
				}
				defer func() {
					if err := os.Chdir(cwd); err != nil {
						t.Fatal(err)
					}
				}()
			}

			dest := path.Base(wd)
			if test.dest != "" {
				dest = test.dest
			}
			ext := Extractor{
				Dest: dest,
			}

			var out bytes.Buffer
			if _, err := ext.Extract(test.arg, test.importPath, &out); err != nil {
				t.Fatal(err)
			}

			if test.expected != "" {
				if out.String() != test.expected {
					t.Fatalf("\nGot:\n%q\nWant: \n%q", out.String(), test.expected)
				}
			}

			if test.contains != "" {
				if !strings.Contains(out.String(), test.contains) {
					t.Fatalf("Missing expected part: %s in %s", test.contains, out.String())
				}
			}
		})
	}
}