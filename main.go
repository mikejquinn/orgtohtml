package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/niklasfasching/go-org/org"
	"github.com/yosssi/gohtml"
)

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.New("").Funcs(tmplFuncs).Parse(tmplStr))
}

func main() {
	doc := org.New().Silent().Parse(os.Stdin, "")
	if doc.Error != nil {
		log.Fatalln("Error parsing org doc:", doc.Error)
	}
	fmt.Fprintln(os.Stdout, style)
	if err := renderHTML(os.Stdout, doc); err != nil {
		log.Fatalln("Error rendering HTML:", err)
	}
}

func renderHTML(w io.Writer, doc *org.Document) error {
	s, err := execTmpl("ul", doc)
	if err != nil {
		return err
	}
	s = gohtml.Format(s)
	s += "\n"
	_, err = w.Write([]byte(s))
	return err
}

func execTmpl(name string, v interface{}) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, v); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var tmplFuncs = map[string]interface{}{
	"children":   children,
	"isListItem": isListItem,
	"li":         li,
}

func children(n interface{}) []org.Node {
	switch n := n.(type) {
	case *org.Document:
		return n.Nodes
	case org.Headline:
		return n.Children
	case org.List:
		return n.Items
	case org.ListItem:
		return n.Children
	case []org.Node:
		return n
	}
	panic(fmt.Sprintf("children called on node of type %T", n))
}

func isListItem(n org.Node) bool {
	switch n.(type) {
	case org.Headline, org.ListItem:
		return true
	}
	return false
}

func liHeadline(hl org.Headline) (string, error) {
	var s strings.Builder
	if len(hl.Title) > 0 {
		t := hl.Title[0] // TODO: How could a headline have more than one title?
		if hl.Status != "" {
			s.WriteString("<strong>")
			s.WriteString(hl.Status)
			s.WriteString("</strong>")
		}
		s.WriteString(" ")
		s.WriteString(t.String())
	}
	i := 0
loop:
	for i < len(hl.Children) {
		switch n := hl.Children[i].(type) {
		case org.Headline:
			// Headlines will always be last -- process them below.
			break loop
		case org.Paragraph:
			s.WriteString("<br />")
			s.WriteString(liParagraph(n))
		case org.List:
			s.WriteString("\n")
			s1, err := liList(n)
			if err != nil {
				return "", err
			}
			s.WriteString(s1)
		}
		i++
	}

	if i < len(hl.Children) {
		hls := hl.Children[i:]
		s1, err := execTmpl("ul", hls)
		if err != nil {
			return "", err
		}
		s.WriteString(s1)
	}

	return s.String(), nil
}

func liListItem(l org.ListItem) (string, error) {
	var s strings.Builder
	for _, c := range l.Children {
		switch c := c.(type) {
		case org.Paragraph:
			s.WriteString(liParagraph(c))
		default:
			panic(fmt.Sprintf("child is of type %T", c))
		}
	}
	return s.String(), nil
}

func li(n org.Node) (string, error) {
	switch n := n.(type) {
	case org.Headline:
		return liHeadline(n)
	case org.ListItem:
		return liListItem(n)
	}
	panic(fmt.Sprintf("li called with node of type %T", n))
}

func liParagraph(p org.Paragraph) string {
	var s strings.Builder
	for _, c := range p.Children {
		switch c := c.(type) {
		case org.Text:
			s.WriteString(c.Content)
		case org.LineBreak:
			s.WriteString("<br />")
		default:
			log.Printf("liParagraph: ignoring type %T", c)
		}
	}
	return s.String()
}

func liList(l org.List) (string, error) {
	switch l.Kind {
	case "unordered":
		return execTmpl("ul", l)
	case "ordered":
		return execTmpl("ol", l)
	}
	panic("unrecognized kind: " + l.Kind)
}

const tmplStr = `
{{define "ul" -}}
<ul>
{{range $li := children . -}}
  {{if isListItem $li -}}
  <li>
  {{- li . -}}
  </li>
  {{end -}}
{{end -}}
</ul>
{{end -}}

{{define "ol" -}}
<ol>
{{range $li := children . -}}
  {{if isListItem $li -}}
  <li>
  {{- li . -}}
  </li>
  {{end -}}
{{end -}}
</ol>
{{end -}}
`

const style = `
<style>
body {
  max-width: 1000px;
  margin: 0 auto;
  font-family: Helvetica, arial, sans-serif;
  font-size: 14px;
  line-height: 1.6;
  padding-top: 10px;
  padding-bottom: 10px;
  background-color: white;
  padding: 30px; }

body > *:first-child {
  margin-top: 0 !important; }
body > *:last-child {
  margin-bottom: 0 !important; }

a {
  color: #4183C4; }
a.absent {
  color: #cc0000; }
a.anchor {
  display: block;
  padding-left: 30px;
  margin-left: -30px;
  cursor: pointer;
  position: absolute;
  top: 0;
  left: 0;
  bottom: 0; }

p, blockquote, ul, ol, dl, li, table, pre {
  margin: 0px 0; }

li p.first {
  display: inline-block; }

ul, ol {
  padding-left: 30px; }

ul :first-child, ol :first-child {
  margin-top: 0; }

ul :last-child, ol :last-child {
  margin-bottom: 0; }


table {
  padding: 0; }
  table tr {
    border-top: 1px solid #cccccc;
    background-color: white;
    margin: 0;
    padding: 0; }
    table tr:nth-child(2n) {
      background-color: #f8f8f8; }
    table tr th {
      font-weight: bold;
      border: 1px solid #cccccc;
      text-align: left;
      margin: 0;
      padding: 6px 13px; }
    table tr td {
      border: 1px solid #cccccc;
      text-align: left;
      margin: 0;
      padding: 6px 13px; }
    table tr th :first-child, table tr td :first-child {
      margin-top: 0; }
    table tr th :last-child, table tr td :last-child {
      margin-bottom: 0; }
</style>
`
