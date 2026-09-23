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

	t.Run("leading ./ is stripped from the dest", func(t *testing.T) {
		links := ParseWikilinks("[[Guide]](./notes/daily/2024-01-01.md)")
		if len(links) != 1 {
			t.Fatalf("len(links) = %d, want 1", len(links))
		}
		if links[0].Target != "notes/daily/2024-01-01" {
			t.Errorf("Target = %q, want %q", links[0].Target, "notes/daily/2024-01-01")
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
