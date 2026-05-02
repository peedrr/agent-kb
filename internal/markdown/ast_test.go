package markdown

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestFlatten(t *testing.T) {
	source := []byte(`# Heading One

Some intro text.

## Heading Two

Here is a [regular link](http://example.com) and a [[wikilink]].

` + "```" + `go
func main() {}
` + "```" + `
`)

	md := goldmark.New()
	p := md.Parser()
	reader := text.NewReader(source)
	doc := p.Parse(reader)

	t.Run("headings", func(t *testing.T) {
		headings := FlattenHeadings(doc, source)
		if len(headings) != 2 {
			t.Fatalf("expected 2 headings, got %d", len(headings))
		}
		if headings[0].Level != 1 || headings[0].Text != "Heading One" {
			t.Errorf("heading[0] = %+v, want Level=1 Text='Heading One'", headings[0])
		}
		if headings[0].Line != 1 {
			t.Errorf("heading[0].Line = %d, want 1", headings[0].Line)
		}
		if headings[1].Level != 2 || headings[1].Text != "Heading Two" {
			t.Errorf("heading[1] = %+v, want Level=2 Text='Heading Two'", headings[1])
		}
		if headings[1].Line != 5 {
			t.Errorf("heading[1].Line = %d, want 5", headings[1].Line)
		}
	})

	t.Run("links", func(t *testing.T) {
		links := FlattenLinks(doc, source)
		if len(links) != 1 {
			t.Fatalf("expected 1 link, got %d", len(links))
		}
		link := links[0]
		if link.Target != "http://example.com" {
			t.Errorf("link.Target = %q, want %q", link.Target, "http://example.com")
		}
		if link.Text != "regular link" {
			t.Errorf("link.Text = %q, want %q", link.Text, "regular link")
		}
		if link.IsWikilink {
			t.Errorf("link.IsWikilink = true, want false for regular markdown link")
		}
		if link.Line != 7 {
			t.Errorf("link.Line = %d, want 7", link.Line)
		}
	})

	t.Run("code blocks", func(t *testing.T) {
		blocks := FlattenCodeBlocks(doc, source)
		if len(blocks) != 1 {
			t.Fatalf("expected 1 code block, got %d", len(blocks))
		}
		if blocks[0].Language != "go" {
			t.Errorf("block.Language = %q, want %q", blocks[0].Language, "go")
		}
		if blocks[0].Line != 9 {
			t.Errorf("block.Line = %d, want 9", blocks[0].Line)
		}
	})

	t.Run("wikilink detection", func(t *testing.T) {
		wikiSource := []byte("[[target|display]]")
		doc := ast.NewDocument()
		link := ast.NewLink()
		link.Destination = []byte("target")

		txt := ast.NewText()
		txt.Segment = text.NewSegment(2, 8)
		link.AppendChild(link, txt)

		link.SetPos(0)
		doc.AppendChild(doc, link)

		links := FlattenLinks(doc, wikiSource)
		if len(links) != 1 {
			t.Fatalf("expected 1 link, got %d", len(links))
		}
		if !links[0].IsWikilink {
			t.Errorf("expected IsWikilink=true for [[...]] source")
		}
		if links[0].Target != "target" {
			t.Errorf("link.Target = %q, want %q", links[0].Target, "target")
		}
		if links[0].Text != "target" {
			t.Errorf("link.Text = %q, want %q", links[0].Text, "target")
		}
	})
}
