package blackfriday

import (
    "testing"
)

func runTerminalMarkdownBlock(input string, flags int) string {
    flags |= TERM_NO_HEADER_FOOTER
    // flags |= TERM_DEBUG_LOGGING
    renderer := TerminalRenderer(flags)
    extensions := 0
    extensions |= EXTENSION_NO_INTRA_EMPHASIS
    extensions |= EXTENSION_TABLES
    extensions |= EXTENSION_FENCED_CODE
    extensions |= EXTENSION_AUTOLINK
    return string(Markdown([]byte(input), renderer, extensions))
}

func createTerminal(flags int) *Terminal {
    flags |= TERM_NO_HEADER_FOOTER
    return NewTerminal(flags)
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

func doTerminalRuneTest(t *testing.T, r []rune, width int, expected int) {
    term := NewTerminal(0)
    actual := term.runesInWidth(r, width)
    if actual != expected {
        t.Errorf("\nInput   [%#v]\nExpected[%#v]\nActual  [%#v]",
            string(r), expected, actual)
    }
}

func TestTerminalRunes(t *testing.T) {
    doTerminalRuneTest(t, []rune("whatever"), 10, 8)
    doTerminalRuneTest(t, []rune("whatever this is"), 10, 10)
    doTerminalRuneTest(t, []rune("こんにちは。 こんにちは。"), 10, 5)
    doTerminalRuneTest(t, []rune("こんにちは。こんにちは。"), 10, 5)
    doTerminalRuneTest(t, []rune("こんにちは。 こんにちは。 こんにちは。 こんにちは。"), 20, 10)
    // doTerminalRuneTest(t, []rune("whatever"), -10, 0)
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
        "This is a wrap test. Wrap on.\n",
        //                  ^-- 20 at period
        "\nThis is a wrap test.\nWrap on.\n",

        "This is a wrap test. Wrap on. Wrap off. The wrapper.\n",
        //                  ^-- 20 at period    ^-- 40 here
        "\nThis is a wrap test.\nWrap on. Wrap off.\nThe wrapper.\n",

        "1 3 5 7 9 1 3 5 7 9 1 3 5 7 9 1\n",
        "\n1 3 5 7 9 1 3 5 7 9\n1 3 5 7 9 1\n",

        "1 3 5 7 9 *1x3* 5 7 9 1 3 5 7 9 1\n",
        "\n1 3 5 7 9 \x1b[4m1x3\x1b[0m 5 7 9\n1 3 5 7 9 1\n",

        "This is a **wrap** test. Wrap on.\n",
        "\nThis is a \x1b[1mwrap\x1b[0m test.\nWrap on.\n",

        "This is a ***wrap*** test. Wrap on.\n",
        "\nThis is a \x1b[7mwrap\x1b[0m test.\nWrap on.\n",

        " This is a *wrapper* test. Wrap on.\n",
        //                     ^-- 20 at 'e'
        "\nThis is a \x1b[4mwrapper\x1b[0m\ntest. Wrap on.\n",

        "123456789012345678901234567890\n",
        "\n12345678901234567890\n1234567890\n",

        "こんにちは。 This is a wrap test.\n",
        //                  ^-- 20 at 's'
        "\nこんにちは。 This is\na wrap test.\n",

        "こんにちは。 こんにちは。\n",
        //                  ^-- 20
        "\nこんにちは。\nこんにちは。\n",

        // BUG: Should break at space
        "&lt;&copy;&yen;&amp;&cent;&pound;&yen;&euro;&copy;&reg;&gt; &lt;&amp;&cent;&pound;&yen;&euro;&copy;&gt; &lt;&copy;&gt;\n",
        "\n<©¥&¢£¥€©®> <&¢£¥€©\n> <©>\n",
        //                    ^-- 20 at '>'
        // "\n<©¥&¢£¥€©®> <&¢£¥€©>\n<©>\n",

        "&lt;&copy;&yen;&amp;&cent;&pound;&yen;&euro;&copy;&reg;&gt; a\n",
        "\n<©¥&¢£¥€©®> a\n",

        "こんにちは。 んこにちは。 ちこんには。 はこんにち。\n",
        //2345678901234567890
        //                  ^-- 20
        "\nこんにちは。\nんこにちは。\nちこんには。\nはこんにち。\n",

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

func TestTerminalLists(t *testing.T) {
    var tests = []string{
        "1. one\n3. two\n",
        "\n 1. one\n 2. two\n",

        "* one\n* two\n",
        "\n  \u2022 one\n  \u2022 two\n",

        " - one\n - two\n",
        "\n  \u2022 one\n  \u2022 two\n",

        "- 3456 89 1 3456 8901. 4567 90.\n- 345\n",
        //                  ^-- 20 is here
        "\n  \u2022 3456 89 1 3456\n    8901. 4567 90.\n  \u2022 345\n",
    }

    flags := TERM_FIXED_WIDTH_20
    doTerminalTests(t, tests, flags)
}

func TestTerminalFencedCodeBlock(t *testing.T) {
    var tests = []string{
        "``` go\nfunc() bool {\n\treturn true;\n}\n```\n",
        "func() bool {\n\treturn true;\n}\n",

    }

    flags := TERM_FIXED_WIDTH_20
    doTerminalTests(t, tests, flags)
}

func TestTerminalCodeSpan(t *testing.T) {
    var tests = []string{
        "this is `source code`\n",
        "\nthis is source code\n",

    }

    flags := TERM_FIXED_WIDTH_20
    doTerminalTests(t, tests, flags)
}


func TestTerminalEntities(t *testing.T) {
    var tests = []string{
        "copy symbol entity: &copy;\n",
        "\ncopy symbol entity: ©\n",

        "ene con tilde: n&#771;\n",
        "\nene con tilde: ñ\n",

        "euro symbol: &euro;\n",
        "\neuro symbol: €\n",

        "&nbsp;&lt;&gt;&amp;&cent;&pound;&yen;&euro;&copy;&reg;\n",
        "\n\u00a0<>&¢£¥€©®\n",

    }

    doTerminalTests(t, tests, 0)
}

