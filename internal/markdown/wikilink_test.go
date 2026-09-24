package markdown

import (
	"strings"
	"testing"
)

func TestParseWikilinks(t *testing.T) {
	t.Run("simple wikilink [[target]]", func(t *testing.T) {
		content := "See [[my-page]] for details."
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "my-page" {
			t.Errorf("Target = %q, want %q", wl.Target, "my-page")
		}
		if wl.Display != "my-page" {
			t.Errorf("Display = %q, want %q", wl.Display, "my-page")
		}
		if wl.HasHeading {
			t.Error("HasHeading = true, want false")
		}
		if wl.Heading != "" {
			t.Errorf("Heading = %q, want empty string", wl.Heading)
		}
	})

	t.Run("wikilink with display text [[target|display]]", func(t *testing.T) {
		content := "Check out [[my-page|My Page]] for more."
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "my-page" {
			t.Errorf("Target = %q, want %q", wl.Target, "my-page")
		}
		if wl.Display != "My Page" {
			t.Errorf("Display = %q, want %q", wl.Display, "My Page")
		}
		if wl.HasHeading {
			t.Error("HasHeading = true, want false")
		}
	})

	t.Run("wikilink with heading [[target#heading]]", func(t *testing.T) {
		content := "See [[my-page#Introduction]] for details."
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "my-page" {
			t.Errorf("Target = %q, want %q", wl.Target, "my-page")
		}
		if !wl.HasHeading {
			t.Error("HasHeading = false, want true")
		}
		if wl.Heading != "Introduction" {
			t.Errorf("Heading = %q, want %q", wl.Heading, "Introduction")
		}
		if wl.Display != "Introduction" {
			t.Errorf("Display = %q, want %q", wl.Display, "Introduction")
		}
	})

	t.Run("wikilink with heading and display [[target#heading|display]]", func(t *testing.T) {
		content := "See [[my-page#Introduction|intro]] for details."
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "my-page" {
			t.Errorf("Target = %q, want %q", wl.Target, "my-page")
		}
		if !wl.HasHeading {
			t.Error("HasHeading = false, want true")
		}
		if wl.Heading != "Introduction" {
			t.Errorf("Heading = %q, want %q", wl.Heading, "Introduction")
		}
		if wl.Display != "intro" {
			t.Errorf("Display = %q, want %q", wl.Display, "intro")
		}
	})

	t.Run("skip wikilinks inside fenced code blocks", func(t *testing.T) {
		content := "Some text\n```\n[[should-not-match]]\n```\nBut [[should-match]] here."
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "should-match" {
			t.Errorf("Target = %q, want %q", links[0].Target, "should-match")
		}
	})

	t.Run("skip wikilinks inside fenced code blocks with language hint", func(t *testing.T) {
		content := "```go\n[[ignored]]\n```\n[[valid]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "valid" {
			t.Errorf("Target = %q, want %q", links[0].Target, "valid")
		}
	})

	t.Run("skip wikilinks inside inline code", func(t *testing.T) {
		content := "Use `[[ignored]]` to link, but [[valid]] works."
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "valid" {
			t.Errorf("Target = %q, want %q", links[0].Target, "valid")
		}
	})

	t.Run("skip wikilinks inside HTML comments", func(t *testing.T) {
		content := "<!-- [[ignored]] -->\n[[valid]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "valid" {
			t.Errorf("Target = %q, want %q", links[0].Target, "valid")
		}
	})

	t.Run("multiple wikilinks on same line", func(t *testing.T) {
		content := "See [[a]] and [[b|B]] and [[c#section]] too."
		links := ParseWikilinks(content)
		if len(links) != 3 {
			t.Fatalf("len(links) = %d, want 3", len(links))
		}
		if links[0].Target != "a" {
			t.Errorf("links[0].Target = %q, want %q", links[0].Target, "a")
		}
		if links[0].Display != "a" {
			t.Errorf("links[0].Display = %q, want %q", links[0].Display, "a")
		}
		if links[1].Target != "b" {
			t.Errorf("links[1].Target = %q, want %q", links[1].Target, "b")
		}
		if links[1].Display != "B" {
			t.Errorf("links[1].Display = %q, want %q", links[1].Display, "B")
		}
		if links[2].Target != "c" {
			t.Errorf("links[2].Target = %q, want %q", links[2].Target, "c")
		}
		if !links[2].HasHeading {
			t.Error("links[2].HasHeading = false, want true")
		}
		if links[2].Heading != "section" {
			t.Errorf("links[2].Heading = %q, want %q", links[2].Heading, "section")
		}
	})

	t.Run("empty content returns empty slice", func(t *testing.T) {
		links := ParseWikilinks("")
		if len(links) != 0 {
			t.Errorf("len(links) = %d, want 0", len(links))
		}
	})

	t.Run("content with no wikilinks returns empty slice", func(t *testing.T) {
		content := "Just some plain text\nwith no links at all."
		links := ParseWikilinks(content)
		if len(links) != 0 {
			t.Errorf("len(links) = %d, want 0", len(links))
		}
	})

	t.Run("nested code blocks do not interfere", func(t *testing.T) {
		content := "```\n```\n[[valid]]\n```\n"
		links := ParseWikilinks(content)
		// The first ``` opens, the second ``` closes, so [[valid]] is outside
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "valid" {
			t.Errorf("Target = %q, want %q", links[0].Target, "valid")
		}
	})

	t.Run("multiline fenced code block", func(t *testing.T) {
		content := "Before\n```\n[[a]]\n[[b]]\n```\nAfter [[c]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "c" {
			t.Errorf("Target = %q, want %q", links[0].Target, "c")
		}
	})

	t.Run("inline code with multiple backticks", func(t *testing.T) {
		content := "``[[ignored]]`` but [[valid]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "valid" {
			t.Errorf("Target = %q, want %q", links[0].Target, "valid")
		}
	})

	t.Run("multiline HTML comment", func(t *testing.T) {
		content := "<!--\n[[ignored1]]\n[[ignored2]]\n-->\n[[valid]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "valid" {
			t.Errorf("Target = %q, want %q", links[0].Target, "valid")
		}
	})

	t.Run("path with slashes", func(t *testing.T) {
		content := "[[notes/daily/2024-01-01]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/daily/2024-01-01" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/daily/2024-01-01")
		}
	})

	t.Run("heading with special characters", func(t *testing.T) {
		content := "[[page#some-heading-name]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Heading != "some-heading-name" {
			t.Errorf("Heading = %q, want %q", links[0].Heading, "some-heading-name")
		}
	})

	t.Run("display text takes priority over heading for display", func(t *testing.T) {
		content := "[[page#section|Custom Display]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Display != "Custom Display" {
			t.Errorf("Display = %q, want %q", links[0].Display, "Custom Display")
		}
	})

	t.Run("heading-only wikilink [[#heading]]", func(t *testing.T) {
		content := "[[#Introduction]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "" {
			t.Errorf("Target = %q, want empty string", wl.Target)
		}
		if !wl.HasHeading {
			t.Error("HasHeading = false, want true")
		}
		if wl.Heading != "Introduction" {
			t.Errorf("Heading = %q, want %q", wl.Heading, "Introduction")
		}
		if wl.Display != "Introduction" {
			t.Errorf("Display = %q, want %q", wl.Display, "Introduction")
		}
	})
}

func TestParseWikilinksExplicitDestination(t *testing.T) {
	t.Run("[[display]](dest) uses the dest as target", func(t *testing.T) {
		links := ParseWikilinks("See [[My Page]](pages/my-page).")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "pages/my-page" {
			t.Errorf("Target = %q, want %q", wl.Target, "pages/my-page")
		}
		if wl.Display != "My Page" {
			t.Errorf("Display = %q, want %q", wl.Display, "My Page")
		}
		if wl.Destination != "pages/my-page" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "pages/my-page")
		}
		if wl.HasHeading {
			t.Error("HasHeading = true, want false")
		}
	})

	t.Run("pipe display wins over bracket target with explicit dest", func(t *testing.T) {
		links := ParseWikilinks("[[a|b]](dest)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "dest" {
			t.Errorf("Target = %q, want %q", wl.Target, "dest")
		}
		if wl.Display != "b" {
			t.Errorf("Display = %q, want %q", wl.Display, "b")
		}
	})

	t.Run("pipe display survives heading clearing with explicit dest", func(t *testing.T) {
		links := ParseWikilinks("[[a#h|b]](dest)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "dest" {
			t.Errorf("Target = %q, want %q", wl.Target, "dest")
		}
		if wl.Destination != "dest" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "dest")
		}
		if wl.Display != "b" {
			t.Errorf("Display = %q, want %q", wl.Display, "b")
		}
		if wl.HasHeading {
			t.Error("HasHeading = true, want false (the heading is discarded)")
		}
		if wl.Heading != "" {
			t.Errorf("Heading = %q, want empty string", wl.Heading)
		}
	})

	t.Run("explicit dest wins over a heading in the bracket part", func(t *testing.T) {
		links := ParseWikilinks("[[label#section]](dest)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "dest" {
			t.Errorf("Target = %q, want %q", wl.Target, "dest")
		}
		if wl.Destination != "dest" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "dest")
		}
		if wl.HasHeading {
			t.Error("HasHeading = true, want false (the heading is discarded)")
		}
		if wl.Heading != "" {
			t.Errorf("Heading = %q, want empty string", wl.Heading)
		}
		if wl.Display != "label" {
			t.Errorf("Display = %q, want %q (the heading-derived display falls back to the bracket label)", wl.Display, "label")
		}
	})

	t.Run("hash after the pipe is display text, not a heading", func(t *testing.T) {
		// The pipe split runs before the heading split, so in [[a|b#h]] the "#h"
		// belongs to the pipe display and survives the explicit-dest override.
		links := ParseWikilinks("[[a|b#h]](dest)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "dest" {
			t.Errorf("Target = %q, want %q", wl.Target, "dest")
		}
		if wl.Display != "b#h" {
			t.Errorf("Display = %q, want %q", wl.Display, "b#h")
		}
		if wl.HasHeading {
			t.Error("HasHeading = true, want false")
		}
	})

	t.Run("adjacent parens are required", func(t *testing.T) {
		content := "[[Paris]] (the city)"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "Paris" {
			t.Errorf("Target = %q, want %q", wl.Target, "Paris")
		}
		if wl.Destination != "" {
			t.Errorf("Destination = %q, want empty string", wl.Destination)
		}
		if wantEnd := len("[[Paris]]"); wl.End != wantEnd {
			t.Errorf("End = %d, want %d", wl.End, wantEnd)
		}
	})

	t.Run("invalid dest with a space leaves a plain wikilink", func(t *testing.T) {
		content := "[[Paris]](the city)"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "Paris" {
			t.Errorf("Target = %q, want %q", wl.Target, "Paris")
		}
		if wl.Display != "Paris" {
			t.Errorf("Display = %q, want %q", wl.Display, "Paris")
		}
		if wl.Destination != "" {
			t.Errorf("Destination = %q, want empty string", wl.Destination)
		}
		if wantEnd := len("[[Paris]]"); wl.End != wantEnd {
			t.Errorf("End = %d, want %d (an invalid dest is not part of the token)", wl.End, wantEnd)
		}
	})

	t.Run("angle brackets around the dest are stripped", func(t *testing.T) {
		links := ParseWikilinks("[[My Page]](<my page.md>)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "my page" {
			t.Errorf("Target = %q, want %q", wl.Target, "my page")
		}
		if wl.Display != "My Page" {
			t.Errorf("Display = %q, want %q", wl.Display, "My Page")
		}
	})

	t.Run("angle-bracket dest with spaces is a valid explicit dest", func(t *testing.T) {
		links := ParseWikilinks("[[a]](<my page.md>)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "my page" {
			t.Errorf("Target = %q, want %q", wl.Target, "my page")
		}
		if wl.Destination != "my page" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "my page")
		}
	})

	t.Run("whitespace inside the angle brackets is trimmed", func(t *testing.T) {
		links := ParseWikilinks("[[a]](<foo >)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "foo" {
			t.Errorf("Target = %q, want %q", wl.Target, "foo")
		}
		if wl.Destination != "foo" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "foo")
		}
	})

	t.Run("angle-bracket whitespace around a path dest is trimmed", func(t *testing.T) {
		links := ParseWikilinks("[[a]](< ./notes/x.md >)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("double-quoted title after the dest is stripped", func(t *testing.T) {
		content := `[[Guide]](notes/x.md "title")`
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "notes/x" {
			t.Errorf("Target = %q, want %q", wl.Target, "notes/x")
		}
		if wl.Destination != "notes/x" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "notes/x")
		}
		if wl.Display != "Guide" {
			t.Errorf("Display = %q, want %q", wl.Display, "Guide")
		}
		if wl.End != len(content) {
			t.Errorf("End = %d, want %d (the title is part of the token)", wl.End, len(content))
		}
	})

	t.Run("single-quoted title after the dest is stripped", func(t *testing.T) {
		links := ParseWikilinks("[[a]](notes/x.md 'title')")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("parenthesized title after the dest is stripped", func(t *testing.T) {
		links := ParseWikilinks("[[a]](notes/x.md (title))")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("title may contain an unbalanced closing paren", func(t *testing.T) {
		links := ParseWikilinks(`[[a]](notes/x.md "ti)tle")`)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("angle-bracketed dest with a title", func(t *testing.T) {
		links := ParseWikilinks(`[[a]](<notes/x.md> "title")`)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("title directly after an angle-bracketed dest", func(t *testing.T) {
		links := ParseWikilinks(`[[a]](<notes/x.md>"title")`)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("unclosed title leaves a plain wikilink", func(t *testing.T) {
		content := `[[a]](notes/x.md "title)`
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "a" || wl.Destination != "" {
			t.Errorf("Target = %q, Destination = %q, want %q and empty string", wl.Target, wl.Destination, "a")
		}
		if wantEnd := len("[[a]]"); wl.End != wantEnd {
			t.Errorf("End = %d, want %d (an invalid dest is not part of the token)", wl.End, wantEnd)
		}
	})

	t.Run("text after the title leaves a plain wikilink", func(t *testing.T) {
		content := `[[a]](notes/x.md "title" extra)`
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "a" || wl.Destination != "" {
			t.Errorf("Target = %q, Destination = %q, want %q and empty string", wl.Target, wl.Destination, "a")
		}
	})

	t.Run("empty parens leave a plain wikilink", func(t *testing.T) {
		content := "[[a]]()"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "a" || wl.Destination != "" {
			t.Errorf("Target = %q, Destination = %q, want %q and empty string", wl.Target, wl.Destination, "a")
		}
		if wantEnd := len("[[a]]"); wl.End != wantEnd {
			t.Errorf("End = %d, want %d (ignored parens are not part of the token)", wl.End, wantEnd)
		}
	})

	t.Run("path dest with .md suffix keeps the path", func(t *testing.T) {
		links := ParseWikilinks("[[Guide]](concepts/guide.md)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "concepts/guide" {
			t.Errorf("Target = %q, want %q", wl.Target, "concepts/guide")
		}
	})

	t.Run("path dest without whitespace still parses", func(t *testing.T) {
		links := ParseWikilinks("[[Guide]](notes/daily/2024-01-01)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "notes/daily/2024-01-01" {
			t.Errorf("Target = %q, want %q", wl.Target, "notes/daily/2024-01-01")
		}
		if wl.Destination != "notes/daily/2024-01-01" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "notes/daily/2024-01-01")
		}
	})

	t.Run("balanced nested parens in the dest are accepted", func(t *testing.T) {
		links := ParseWikilinks("[[a]](foo(bar))")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "foo(bar)" {
			t.Errorf("Target = %q, want %q", wl.Target, "foo(bar)")
		}
		if wl.Destination != "foo(bar)" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "foo(bar)")
		}
	})

	t.Run("leading ./ is stripped from the dest", func(t *testing.T) {
		links := ParseWikilinks("[[Guide]](./notes/daily/2024-01-01.md)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/daily/2024-01-01" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/daily/2024-01-01")
		}
	})

	t.Run("repeated ./ and .md affixes collapse", func(t *testing.T) {
		links := ParseWikilinks("[[a]](././a.md.md)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "a" {
			t.Errorf("Target = %q, want %q", wl.Target, "a")
		}
		if wl.Destination != "a" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "a")
		}
	})

	t.Run("repeated .md affixes collapse", func(t *testing.T) {
		links := ParseWikilinks("[[a]](notes/x.md.md)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("a dest that normalizes to empty stays a plain wikilink", func(t *testing.T) {
		content := "[[a]](.md)"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "a" || wl.Destination != "" {
			t.Errorf("Target = %q, Destination = %q, want %q and empty string", wl.Target, wl.Destination, "a")
		}
		if wantEnd := len("[[a]]"); wl.End != wantEnd {
			t.Errorf("End = %d, want %d (an empty dest is not part of the token)", wl.End, wantEnd)
		}
	})

	t.Run("spans cover the whole token including the dest", func(t *testing.T) {
		content := "before [[a]](dest) after"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		wantStart := strings.Index(content, "[[a]](dest)")
		wantEnd := wantStart + len("[[a]](dest)")
		if wl.Start != wantStart || wl.End != wantEnd {
			t.Errorf("span = [%d,%d), want [%d,%d)", wl.Start, wl.End, wantStart, wantEnd)
		}
	})

	t.Run("spans cover only the brackets for a plain wikilink", func(t *testing.T) {
		content := "before [[a]] after"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		wantStart := strings.Index(content, "[[a]]")
		wantEnd := wantStart + len("[[a]]")
		if wl.Start != wantStart || wl.End != wantEnd {
			t.Errorf("span = [%d,%d), want [%d,%d)", wl.Start, wl.End, wantStart, wantEnd)
		}
	})

	t.Run("unclosed dest parens leave a plain wikilink", func(t *testing.T) {
		links := ParseWikilinks("[[a]](unclosed")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "a" || wl.Destination != "" {
			t.Errorf("Target = %q, Destination = %q, want %q and empty string", wl.Target, wl.Destination, "a")
		}
	})

	t.Run("unclosed dest scanning into a fenced code block leaves a plain wikilink", func(t *testing.T) {
		content := "[[a]](see below\n```\n)\n```\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "a" || wl.Destination != "" {
			t.Errorf("Target = %q, Destination = %q, want %q and empty string", wl.Target, wl.Destination, "a")
		}
		if wantEnd := len("[[a]]"); wl.End != wantEnd {
			t.Errorf("End = %d, want %d (the ) found inside the code block is not part of the token)", wl.End, wantEnd)
		}
	})

	t.Run("explicit dest inside inline code is still excluded", func(t *testing.T) {
		content := "Use `[[a]](dest)` to link, but [[valid]] works."
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "valid" {
			t.Errorf("Target = %q, want %q", links[0].Target, "valid")
		}
	})

	t.Run("explicit dest inside a fenced code block is still excluded", func(t *testing.T) {
		content := "```\n[[a]](dest)\n```\n[[valid]]"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "valid" {
			t.Errorf("Target = %q, want %q", links[0].Target, "valid")
		}
	})
}

// Triple and deeper bracket runs are one balanced token, so the inner text is
// never a bracket fragment such as "[a".
func TestParseWikilinksDegenerateBracketRuns(t *testing.T) {
	t.Run("triple brackets yield the inner target", func(t *testing.T) {
		links := ParseWikilinks("[[[a]]]")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "a" {
			t.Errorf("Target = %q, want %q — never the bracket fragment %q", wl.Target, "a", "[a")
		}
		if wl.Display != "a" {
			t.Errorf("Display = %q, want %q", wl.Display, "a")
		}
		if wl.Destination != "" {
			t.Errorf("Destination = %q, want empty string", wl.Destination)
		}
		if wl.Start != 0 || wl.End != len("[[[a]]]") {
			t.Errorf("span = [%d,%d), want [0,%d)", wl.Start, wl.End, len("[[[a]]]"))
		}
	})

	t.Run("triple brackets with an explicit dest use the dest as target", func(t *testing.T) {
		content := "[[[a]]](notes/x.md)"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "notes/x" {
			t.Errorf("Target = %q, want %q", wl.Target, "notes/x")
		}
		if wl.Destination != "notes/x" {
			t.Errorf("Destination = %q, want %q", wl.Destination, "notes/x")
		}
		if wl.Display != "a" {
			t.Errorf("Display = %q, want %q", wl.Display, "a")
		}
		if wl.Start != 0 || wl.End != len(content) {
			t.Errorf("span = [%d,%d), want [0,%d) (the whole token including the dest)", wl.Start, wl.End, len(content))
		}
	})
}

// Spec-literal display outcomes for the explicit-destination form. Both are
// decided corner cases, not defects: the destination is the target, and the
// display falls back to the bracket label.
func TestParseWikilinksExplicitDestinationDisplayFallbacks(t *testing.T) {
	t.Run("empty bracket label [[#h]](dest) has an empty display", func(t *testing.T) {
		// The bracket part holds only a heading, so the heading-derived display is
		// the empty bracket label. The pipe fallback does not apply because no pipe
		// display was written.
		links := ParseWikilinks("[[#h]](dest)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "dest" || wl.Destination != "dest" {
			t.Errorf("Target = %q, Destination = %q, want %q for both", wl.Target, wl.Destination, "dest")
		}
		if wl.Display != "" {
			t.Errorf("Display = %q, want empty string", wl.Display)
		}
		if wl.HasHeading || wl.Heading != "" {
			t.Errorf("HasHeading = %v, Heading = %q, want false and empty string (the dest discards the heading)", wl.HasHeading, wl.Heading)
		}
	})

	t.Run("empty pipe display [[a#h|]](dest) falls back to the bracket label", func(t *testing.T) {
		// The pipe is present but carries no display text, so it does not count as
		// a pipe display: the heading-derived display resets to the bracket label
		// "a" rather than the heading "h".
		links := ParseWikilinks("[[a#h|]](dest)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "dest" || wl.Destination != "dest" {
			t.Errorf("Target = %q, Destination = %q, want %q for both", wl.Target, wl.Destination, "dest")
		}
		if wl.Display != "a" {
			t.Errorf("Display = %q, want %q", wl.Display, "a")
		}
		if wl.HasHeading || wl.Heading != "" {
			t.Errorf("HasHeading = %v, Heading = %q, want false and empty string (the dest discards the heading)", wl.HasHeading, wl.Heading)
		}
	})
}

// Degenerate bracket runs at offset 0 and four brackets deep: the parser token
// starts at the same offset goldmark gives its Link.Pos(), so the two readers
// describe one token.
func TestParseWikilinksDegenerateBracketRunsAtOffsetZero(t *testing.T) {
	t.Run("deeper nesting starts at offset 0", func(t *testing.T) {
		content := "[[[[a]]]](x.md)"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "x" || wl.Destination != "x" {
			t.Errorf("Target = %q, Destination = %q, want %q for both", wl.Target, wl.Destination, "x")
		}
		if wl.Display != "a" {
			t.Errorf("Display = %q, want %q", wl.Display, "a")
		}
		if wl.Start != 0 || wl.End != len(content) {
			t.Errorf("span = [%d,%d), want [0,%d) (the whole token including the dest)", wl.Start, wl.End, len(content))
		}
	})
}

// A backslash-escaped bracket makes the token literal text in CommonMark, so
// neither the link graph nor CEL records it as a link.
func TestParseWikilinksEscapedBrackets(t *testing.T) {
	t.Run("an escaped opening bracket excludes the token", func(t *testing.T) {
		content := `The literal \[[alpha]](notes/x.md) stays text, but [[real]] is a link.`
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "real" {
			t.Errorf("Target = %q, want %q", links[0].Target, "real")
		}
	})

	t.Run("an escaped closing bracket excludes the token", func(t *testing.T) {
		content := `The literal [[alpha\]](notes/x.md) stays text, but [[real]] is a link.`
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "real" {
			t.Errorf("Target = %q, want %q", links[0].Target, "real")
		}
	})

	t.Run("an escaped backslash does not escape the bracket", func(t *testing.T) {
		content := `A bare backslash \\[[alpha]] still links.`
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "alpha" {
			t.Errorf("Target = %q, want %q", links[0].Target, "alpha")
		}
	})
}

// An indented (four-space or tab) block is code in CommonMark, so
// wikilink-shaped tokens inside one stay literal text in both readers.
func TestParseWikilinksIndentedCodeBlocks(t *testing.T) {
	t.Run("a four-space indented block is excluded", func(t *testing.T) {
		content := "A paragraph.\n\n    [[indented]](notes/x.md)\n\n[[real]]\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "real" {
			t.Errorf("Target = %q, want %q", links[0].Target, "real")
		}
	})

	t.Run("a tab-indented block is excluded", func(t *testing.T) {
		content := "A paragraph.\n\n\t[[indented]]\n\n[[real]]\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "real" {
			t.Errorf("Target = %q, want %q", links[0].Target, "real")
		}
	})

	t.Run("a three-space indent is prose, not code", func(t *testing.T) {
		content := "A paragraph.\n\n   [[real]]\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "real" {
			t.Errorf("Target = %q, want %q", links[0].Target, "real")
		}
	})

	t.Run("an indented line after a paragraph is a lazy continuation", func(t *testing.T) {
		content := "A paragraph.\n    [[lazy]]\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "lazy" {
			t.Errorf("Target = %q, want %q", links[0].Target, "lazy")
		}
	})
}

// A list item opens its content at the marker's content column, so a line
// indented four columns past that column is code inside the item while item
// content indented four spaces past the marker's own column is prose. The
// exclusion for an indented block is measured from the innermost open item.
func TestParseWikilinksIndentedCodeBlocksInLists(t *testing.T) {
	t.Run("item content indented four spaces past the marker is prose", func(t *testing.T) {
		content := "- item\n\n    [[a]](notes/x.md)\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("ordered item content is measured from the marker", func(t *testing.T) {
		content := "1. item\n\n    [[a]](notes/x.md)\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("nested item content is measured from the innermost marker", func(t *testing.T) {
		content := "- outer\n\n  - inner\n\n      [[a]](notes/x.md)\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/x")
		}
	})

	t.Run("content four columns past the item content column is code", func(t *testing.T) {
		content := "- item\n\n      [[a]](notes/x.md)\n\n[[real]]\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "real" {
			t.Errorf("Target = %q, want %q", links[0].Target, "real")
		}
	})
}

// The indented-code exclusion is line-structured while a token's span can
// cross lines: a quoted link title may contain a newline. Only a token whose
// bracket part lies inside an indented block is excluded; a span that merely
// crosses an indented line is still recorded.
func TestParseWikilinksIndentedCodeBlockSpanBoundary(t *testing.T) {
	t.Run("a multiline title crossing an indented line keeps the token", func(t *testing.T) {
		content := "[[a]](notes/x.md \"line one\n    line two\")\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		wl := links[0]
		if wl.Target != "notes/x" {
			t.Errorf("Target = %q, want %q", wl.Target, "notes/x")
		}
		if wl.Display != "a" {
			t.Errorf("Display = %q, want %q", wl.Display, "a")
		}
	})

	t.Run("a token inside an indented block is excluded across its title", func(t *testing.T) {
		content := "A paragraph.\n\n    [[a]](notes/x.md \"title\nline two\")\n\n[[real]]\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "real" {
			t.Errorf("Target = %q, want %q", links[0].Target, "real")
		}
	})

	t.Run("an inner text crossing an indented boundary is kept", func(t *testing.T) {
		content := "[[a\n\n    b]](x.md)\n"
		links := ParseWikilinks(content)
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "x" {
			t.Errorf("Target = %q, want %q", links[0].Target, "x")
		}
	})
}

// assertWikilinkTargets parses content and fails unless the recorded wikilinks
// have exactly the wanted targets, in order.
func assertWikilinkTargets(t *testing.T, content string, want ...string) {
	t.Helper()
	links := ParseWikilinks(content)
	if len(links) != len(want) {
		t.Fatalf("ParseWikilinks(%q) recorded %d links, want %d", content, len(links), len(want))
	}
	for i, target := range want {
		if links[i].Target != target {
			t.Errorf("links[%d].Target = %q, want %q", i, links[i].Target, target)
		}
	}
}

// A line of three or more '-' or '*' characters separated only by spaces is a
// thematic break, and a run of ten or more digits keeps an ordered-marker line
// a paragraph, so neither line opens a list item.
func TestParseWikilinksThematicBreakAndOrderedMarkerDigitCap(t *testing.T) {
	t.Run("a '*' thematic break is not a list item", func(t *testing.T) {
		assertWikilinkTargets(t, "* * *\n\n    [[a]](notes/x.md)")
	})

	t.Run("a '-' thematic break is not a list item", func(t *testing.T) {
		assertWikilinkTargets(t, "- - -\n\n    [[a]](notes/x.md)")
	})

	t.Run("ten digits before '.' keep the line a paragraph", func(t *testing.T) {
		assertWikilinkTargets(t, "1234567890. item\n\n    [[a]](notes/x.md)")
	})

	t.Run("a single '-' bullet is still a list item", func(t *testing.T) {
		assertWikilinkTargets(t, "- item\n\n    [[a]](notes/x.md)", "notes/x")
	})

	t.Run("a single-digit ordered marker is still a list item", func(t *testing.T) {
		assertWikilinkTargets(t, "9. item\n\n    [[a]](notes/x.md)", "notes/x")
	})
}

// An indented line directly after a non-blank line is a lazy continuation of
// the preceding paragraph, so it opens no list item and closes none: a marker
// on such a line is literal paragraph text.
func TestParseWikilinksLazyContinuationListBookkeeping(t *testing.T) {
	t.Run("a demoted marker line opens no list item", func(t *testing.T) {
		assertWikilinkTargets(t, "paragraph\n    - item\n\n        [[a]](notes/x.md)")
	})

	t.Run("a demoted line of plain text stays prose", func(t *testing.T) {
		assertWikilinkTargets(t, "paragraph\n    [[a]](notes/x.md)", "notes/x")
	})

	t.Run("a lazy continuation inside an item keeps the item open", func(t *testing.T) {
		assertWikilinkTargets(t, "- item\n        [[a]](notes/x.md)", "notes/x")
	})
}

// Only the lines strictly inside a fenced code block are opaque to the list
// tracker: they neither open nor close items, and a blank line inside a fence
// is not a paragraph break. The delimiter lines are ordinary markdown, so a
// fence opened below an open item's content column closes that item while a
// fence indented to the item's content column leaves it open. A mid-line
// backtick pair that the fence detector pairs is a fence with no interior, so
// the line carrying it opens its item normally.
func TestParseWikilinksFencedCodeInListTracking(t *testing.T) {
	t.Run("a marker inside a fence opens no item", func(t *testing.T) {
		assertWikilinkTargets(t, "```\n- item\n  ```\n\n    [[a]](notes/x.md)")
	})

	t.Run("a fence with plain code keeps the indented block code", func(t *testing.T) {
		assertWikilinkTargets(t, "```\ncode\n  ```\n\n    [[a]](notes/x.md)")
	})

	t.Run("a real list still keeps its four-space item content prose", func(t *testing.T) {
		assertWikilinkTargets(t, "- item\n\n    [[a]](notes/x.md)", "notes/x")
	})

	t.Run("a column-zero fence closes the item above it", func(t *testing.T) {
		assertWikilinkTargets(t, "- item\n```\ncode\n```\n\n    [[b]](notes/x.md)")
	})

	t.Run("a fence at the item content column leaves the item open", func(t *testing.T) {
		assertWikilinkTargets(t, "- item\n  ```\n  code\n  ```\n\n  [[a]](notes/x.md)", "notes/x")
	})

	t.Run("inline code spans in a bullet open its item", func(t *testing.T) {
		assertWikilinkTargets(t, "- toggle ```a``` and ```b``` now\n\n    [[note]](x.md)", "x")
	})

	t.Run("a single inline code span in a bullet opens its item", func(t *testing.T) {
		assertWikilinkTargets(t, "- item ```code``` end\n\n    [[a]](notes/x.md)", "notes/x")
	})

	t.Run("a single-line fence still excludes its own backtick span", func(t *testing.T) {
		assertWikilinkTargets(t, "- toggle ```[[a]]``` now\n\n[[real]]", "real")
	})
}

// A line indented four or more columns that starts below the innermost open
// item's content column closes the items that no longer contain it and is then
// judged against what remains: code when it reaches four columns past the
// remaining item's content column (or sits at the top level), prose otherwise.
func TestParseWikilinksIndentedCodeBelowItemContentColumn(t *testing.T) {
	t.Run("a three-space item closes under a four-space line", func(t *testing.T) {
		assertWikilinkTargets(t, "   - item\n\n    [[a]](notes/x.md)")
	})

	t.Run("a wide ordered marker closes under a four-space line", func(t *testing.T) {
		assertWikilinkTargets(t, "100. item\n\n    [[a]](notes/x.md)")
	})

	t.Run("four spaces after a marker push the content column past the line", func(t *testing.T) {
		assertWikilinkTargets(t, "-    item\n\n    [[a]](notes/x.md)")
	})

	t.Run("item content four spaces past the marker stays prose", func(t *testing.T) {
		assertWikilinkTargets(t, "- item\n\n    [[a]](notes/x.md)", "notes/x")
	})

	t.Run("item content five spaces past the marker stays prose", func(t *testing.T) {
		assertWikilinkTargets(t, "- item\n\n     [[a]](notes/x.md)", "notes/x")
	})

	t.Run("six spaces past the marker is code inside the item", func(t *testing.T) {
		assertWikilinkTargets(t, "- item\n\n      [[a]](notes/x.md)")
	})

	t.Run("a line inside the outer item stays the outer item's prose", func(t *testing.T) {
		assertWikilinkTargets(t, "- outer\n   - inner\n\n    [[a]](notes/x.md)", "notes/x")
	})

	t.Run("a line four columns past the remaining outer item is code", func(t *testing.T) {
		assertWikilinkTargets(t, "- outer\n      - inner\n\n      [[a]](notes/x.md)")
	})
}
