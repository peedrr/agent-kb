package markdown

import (
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
