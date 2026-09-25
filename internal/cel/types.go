// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cel

// Heading represents a Markdown heading element.
type Heading struct {
	Level int    `cel:"level"`
	Text  string `cel:"text"`
	Line  int    `cel:"line"`
}

// Link represents a Markdown link element.
type Link struct {
	Target     string `cel:"target"`
	Text       string `cel:"text"`
	IsWikilink bool   `cel:"is_wikilink"`
	Line       int    `cel:"line"`
}

// CodeBlock represents a Markdown code block element.
type CodeBlock struct {
	Language string `cel:"language"`
	Line     int    `cel:"line"`
}
