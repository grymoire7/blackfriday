package blackfriday

import (
    "testing"
)

func runTerminalMarkdownBlock(input string, flags int) string {
    flags |= TERM_NO_HEADER_FOOTER
	renderer := TerminalRenderer(flags)
    extensions := 0
    extensions |= EXTENSION_NO_INTRA_EMPHASIS
    extensions |= EXTENSION_TABLES
    extensions |= EXTENSION_FENCED_CODE
    extensions |= EXTENSION_AUTOLINK
	return string(Markdown([]byte(input), renderer, extensions))
}

func doTerminalTests(t *testing.T, tests []string, extensions int) {
	// catch and report panics
	var candidate string
	defer func() {
		if err := recover(); err != nil {
			t.Errorf("\npanic while processing [%#v]\n", candidate)
		}
	}()

	for i := 0; i+1 < len(tests); i += 2 {
		input := tests[i]
		candidate = input
		expected := tests[i+1]
		actual := runTerminalMarkdownBlock(candidate, extensions)
		if actual != expected {
			t.Errorf("\nInput   [%#v]\nExpected[%#v]\nActual  [%#v]",
				candidate, expected, actual)
		}
	}
}

func TestTerminalPrefixHeader(t *testing.T) {
	var tests = []string{
		"# Header 1\n",
		"\n\x1b[31m\x1b[1mHeader 1\x1b[0m\n",

		"## Header 2\n",
		"\n\x1b[33m\x1b[1mHeader 2\x1b[0m\n",

		"### Header 3\n",
		"\n\x1b[32m\x1b[1mHeader 3\x1b[0m\n",

		"#### Header 4\n",
		"\n\x1b[34m\x1b[1mHeader 4\x1b[0m\n",

		"##### Header 5\n",
		"\n\x1b[35m\x1b[1mHeader 5\x1b[0m\n",

		"###### Header 6\n",
		"\n\x1b[36m\x1b[1mHeader 6\x1b[0m\n",

		"####### Header 7\n",
		"\n\x1b[36m\x1b[1m# Header 7\x1b[0m\n",
	}

	doTerminalTests(t, tests, 0)
}

func TestTerminalUnderlineHeaders(t *testing.T) {
	var tests = []string{
		"Header 1\n========\n",
		"\n\x1b[31m\x1b[1mHeader 1\x1b[0m\n",

		"Header 2\n--------\n",
		"\n\x1b[33m\x1b[1mHeader 2\x1b[0m\n",

		"A\n=\n",
		"\n\x1b[31m\x1b[1mA\x1b[0m\n",

		"B\n-\n",
		"\n\x1b[33m\x1b[1mB\x1b[0m\n",

		"Paragraph\nHeader\n=\n",
		"\nParagraph\n\n\x1b[31m\x1b[1mHeader\x1b[0m\n",

		"Header\n===\nParagraph\n",
		"\n\x1b[31m\x1b[1mHeader\x1b[0m\n\nParagraph\n",

		"Header\n===\nAnother header\n---\n",
		"\n\x1b[31m\x1b[1mHeader\x1b[0m\n\n\x1b[33m\x1b[1mAnother header\x1b[0m\n",

		"   Header\n======\n",
		"\n\x1b[31m\x1b[1mHeader\x1b[0m\n",

		"Header with *inline*\n=====\n",
		"\n\x1b[31m\x1b[1mHeader with \x1b[4minline\x1b[0m\x1b[31m\x1b[1m\x1b[0m\n",

	}

	doTerminalTests(t, tests, 0)
}

func TestTerminalWrap(t *testing.T) {
	var tests = []string{
		" This is a wrap test. Wrap on.\n",
		"\nThis is a wrap test.\nWrap on.\n",

		"こんにちは。 This is a wrap test.\n",
		"\nこんにちは。 This is a\nwrap test.\n",

        // Uncomment when this kind of wrapping works.
		// "こんにちは。 こんにちは。\n",
		// "\nこんにちは。 \nこんにちは。\n",
	}

    flags := TERM_FIXED_WIDTH_20
	doTerminalTests(t, tests, flags)
}

func TestTerminalRules(t *testing.T) {
	var tests = []string{
		"- - -\n",
		"────────────────────\n",

		"* * *\n",
		"────────────────────\n",

		"-----------------------------\n",
		"────────────────────\n",

	}

    flags := TERM_FIXED_WIDTH_20
	doTerminalTests(t, tests, flags)
}

