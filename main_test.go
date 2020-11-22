package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/niklasfasching/go-org/org"
)

func TestParseTree(t *testing.T) {
	for _, tt := range []struct {
		name string
		org  string
		want string
	}{
		{
			"only headings",
			`
* H1
** H2
** H3
* H4`,
			`
<ul>
  <li>
    H1
    <ul>
      <li>
        H2
      </li>
      <li>
        H3
      </li>
    </ul>
  </li>
  <li>
    H4
  </li>
</ul>
`,
		},
		{
			"headings with text",
			`
* H1
This is some text.
** H2
More text in a sub list.
`,
			`
<ul>
  <li>
    H1
    <br />
    This is some text.
    <ul>
      <li>
        H2
        <br />
        More text in a sub list.
      </li>
    </ul>
  </li>
</ul>
`,
		},
		{
			`unordered list`,
			`
* H1
- one
- two
`,
			`
<ul>
  <li>
    H1
    <ul>
      <li>
        one
      </li>
      <li>
        two
      </li>
    </ul>
  </li>
</ul>
`,
		},
		{
			`unordered list with line break`,
			`
* H1
1. one
2. two
   Line break.
`,
			`
<ul>
  <li>
    H1
    <ol>
      <li>
        one
      </li>
      <li>
        two
        <br />
        Line break.
      </li>
    </ol>
  </li>
</ul>
`,
		},
		{
			`heading with status`,
			`
* TODO H1
`,
			`
<ul>
  <li>
    <strong>
      TODO
    </strong>
    H1
  </li>
</ul>
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc := org.New().Silent().Parse(strings.NewReader(tt.org[1:]), "")
			var w bytes.Buffer
			if err := renderHTML(&w, doc); err != nil {
				t.Error(err)
				return
			}
			got := w.String()
			if got != tt.want[1:] {
				t.Error(cmp.Diff(got, tt.want[1:]))
			}
		})
	}
}
