package markdown

import (
	"github.com/peedrr/agent-kb/internal/cel"
	"github.com/yuin/goldmark/ast"
)

// FlattenHeadings extracts all headings from a Goldmark AST.
func FlattenHeadings(doc ast.Node, source []byte) []cel.Heading {
	var headings []cel.Heading
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		headings = append(headings, cel.Heading{
			Level: h.Level,
			Text:  extractText(h, source),
			Line:  offsetToLine(source, h.Pos()),
		})
		return ast.WalkContinue, nil
	})
	return headings
}

// FlattenLinks extracts all links from a Goldmark AST.
// Wikilinks are detected by checking whether the raw source begins with "[[".
func FlattenLinks(doc ast.Node, source []byte) []cel.Link {
	var links []cel.Link
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		l, ok := n.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}
		pos := l.Pos()
		isWikilink := pos+1 < len(source) && source[pos] == '[' && source[pos+1] == '['
		links = append(links, cel.Link{
			Target:     string(l.Destination),
			Text:       extractText(l, source),
			IsWikilink: isWikilink,
			Line:       offsetToLine(source, pos),
		})
		return ast.WalkContinue, nil
	})
	return links
}

// FlattenCodeBlocks extracts all fenced code blocks from a Goldmark AST.
func FlattenCodeBlocks(doc ast.Node, source []byte) []cel.CodeBlock {
	var blocks []cel.CodeBlock
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		cb, ok := n.(*ast.FencedCodeBlock)
		if !ok {
			return ast.WalkContinue, nil
		}
		lang := ""
		if cb.Info != nil {
			lang = string(cb.Info.Value(source))
		}
		blocks = append(blocks, cel.CodeBlock{
			Language: lang,
			Line:     offsetToLine(source, cb.Pos()),
		})
		return ast.WalkContinue, nil
	})
	return blocks
}

func extractText(n ast.Node, source []byte) string {
	var text []byte
	ast.Walk(n, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := child.(*ast.Text); ok {
			text = append(text, t.Value(source)...)
		}
		return ast.WalkContinue, nil
	})
	return string(text)
}
