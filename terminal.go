//
// Blackfriday Markdown Processor (forked)
// Available at http://github.com/grymoire7/blackfriday
//
// Copyright © 2014
// Distributed under the Simplified BSD License.
// See README.md for details.
//

//
//
// Terminal rendering backend
//
//

package blackfriday

import (
    "bytes"
    "errors"
    "fmt"
    "html"
    "io/ioutil"
    "log"
    "regexp"
    "runtime"
    "strings"
    "syscall"
    "unicode"
    "unsafe"
)

const (
    keyEscape = 27
)

const (
    COLOR_BLACK = 1 << iota
    COLOR_RED
    COLOR_GREEN
    COLOR_YELLOW
    COLOR_BLUE
    COLOR_MAGENTA
    COLOR_CYAN
    COLOR_WHITE
)

// option flags
const (
    TERM_NO_HEADER_FOOTER = 1 << iota
    TERM_FIXED_WIDTH_20
    TERM_DEBUG_LOGGING
)

type CharStyle struct {
    Bold  bool
    Underline bool
    Inverse bool
    FGColor int
    BGColor int
}

var defaultCharStyle = CharStyle{
    Bold: false,
    Underline: false,
    Inverse: false,
    FGColor: 0,
    BGColor: 0,
}

// EscapeCodes contains escape sequences that can be written to the terminal in
// order to achieve different styles of text.
type EscapeCodes struct {
    // Foreground colors
    Black, Red, Green, Yellow, Blue, Magenta, Cyan, White []byte

    // Effects
    Bold, Underline, Inverse []byte

    // Reset all attributes
    Reset []byte
}

var vt100EscapeCodes = EscapeCodes{
    Black:   []byte{keyEscape, '[', '3', '0', 'm'},
    Red:     []byte{keyEscape, '[', '3', '1', 'm'},
    Green:   []byte{keyEscape, '[', '3', '2', 'm'},
    Yellow:  []byte{keyEscape, '[', '3', '3', 'm'},
    Blue:    []byte{keyEscape, '[', '3', '4', 'm'},
    Magenta: []byte{keyEscape, '[', '3', '5', 'm'},
    Cyan:    []byte{keyEscape, '[', '3', '6', 'm'},
    White:   []byte{keyEscape, '[', '3', '7', 'm'},

    Reset:     []byte{keyEscape, '[', '0', 'm'},
    Bold:      []byte{keyEscape, '[', '1', 'm'},
    Underline: []byte{keyEscape, '[', '4', 'm'},
    Inverse:   []byte{keyEscape, '[', '7', 'm'},
}

// Terminal is a type that implements the Renderer interface for terminal
// output.
//
// Do not create this directly, instead use the TerminalRenderer function.
type Terminal struct {
    escape     *EscapeCodes
    flags      int
    termWidth  int
    xpos       int
    charstyle  CharStyle
    styleStack []CharStyle
    listCount  int
    whitespace *regexp.Regexp
    logging    bool
}

// TerminalRenderer creates and configures a Terminal object, which
// satisfies the Renderer interface.
//
// flags is a set of TERMINAL_* options ORed together (currently no such options
// are defined).
func TerminalRenderer(flags int) Renderer {
    width, err := getTerminalSize()
    if err != nil {
        width = 80
    }

    // for unit testing
    if flags&TERM_FIXED_WIDTH_20 != 0 {
        width = 20
    }

    logging := true
    if flags&TERM_DEBUG_LOGGING == 0 {
        logging = false
        log.SetOutput(ioutil.Discard)
    }
    log.Println("Width:", width)

    return &Terminal{
        escape:     &vt100EscapeCodes,
        flags:      flags,
        termWidth:  width,
        xpos:       0,
        charstyle:  defaultCharStyle,
        listCount:  0,
        whitespace: regexp.MustCompile(`\s+`),
        logging:    logging,
    }
}

// GetSize returns the dimensions of the given terminal.
func getTerminalSize() (width int, err error) {
    // Dimensions: Row, Col, XPixel, YPixel
    var dimensions [4]uint16
    const (
        TIOCGWINSZ_OSX = 1074295912
    )

    tio := syscall.TIOCGWINSZ
    if runtime.GOOS == "darwin" {
        tio = TIOCGWINSZ_OSX
    }

    r1, _, _ := syscall.Syscall(
        syscall.SYS_IOCTL,
        uintptr(syscall.Stdout),
        uintptr(tio),
        uintptr(unsafe.Pointer(&dimensions)),
    )
    if int(r1) == -1 {
        r1, _, _ = syscall.Syscall(
            syscall.SYS_IOCTL,
            uintptr(syscall.Stdin),
            uintptr(tio),
            uintptr(unsafe.Pointer(&dimensions)),
        )
    }
    if int(r1) == -1 {
        r1, _, _ = syscall.Syscall(
            syscall.SYS_IOCTL,
            uintptr(syscall.Stderr),
            uintptr(tio),
            uintptr(unsafe.Pointer(&dimensions)),
        )
    }
    if int(r1) == -1 {
        return 0, errors.New("GetWinsize error")
    }

    return int(dimensions[1]), err
}

func (t *Terminal) pushStyle() {
    t.styleStack = append(t.styleStack, t.charstyle)
}

func (t *Terminal) popStyle(out *bytes.Buffer) CharStyle {
    if len(t.styleStack) > 0 {
        t.charstyle, t.styleStack = t.styleStack[len(t.styleStack)-1], t.styleStack[:len(t.styleStack)-1]
    } else {
        t.charstyle = defaultCharStyle
    }

    // Restore styles to terminal
    out.Write(t.escape.Reset)
    t.setFGColor(out, t.charstyle.FGColor)
    if (t.charstyle.Bold) {
        out.Write(t.escape.Bold)
    }
    if (t.charstyle.Underline) {
        out.Write(t.escape.Underline)
    }
    if (t.charstyle.Inverse) {
        out.Write(t.escape.Inverse)
    }

    return t.charstyle
}

func (t *Terminal) setFGColor(out *bytes.Buffer, c int) {

    t.charstyle.FGColor = c

    switch c {
    case COLOR_BLACK:
        out.Write(t.escape.Black)
    case COLOR_RED:
        out.Write(t.escape.Red)
    case COLOR_GREEN:
        out.Write(t.escape.Green)
    case COLOR_YELLOW:
        out.Write(t.escape.Yellow)
    case COLOR_BLUE:
        out.Write(t.escape.Blue)
    case COLOR_MAGENTA:
        out.Write(t.escape.Magenta)
    case COLOR_CYAN:
        out.Write(t.escape.Cyan)
    case COLOR_WHITE:
        out.Write(t.escape.White)
    }

}

/* Some runes (e.g. こんにちは。) take up two cell widths in
 * the terminal instead of one.  This complicates line wrapping
 * a bit.
 *
 * Until the unicode package offers support of telling the difference
 * we need our own function to do it.  For more information see:
 *
 * https://groups.google.com/forum/#!topic/golang-dev/oX7BHEdceis
 */
func (t *Terminal) runeWidth(r rune) int {
    if r >= 0x1100 &&
            (r <= 0x115f || r == 0x2329 || r == 0x232a ||
                    (r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) ||
                    (r >= 0xac00 && r <= 0xd7a3) ||
                    (r >= 0xf900 && r <= 0xfaff) ||
                    (r >= 0xfe30 && r <= 0xfe6f) ||
                    (r >= 0xff00 && r <= 0xff60) ||
                    (r >= 0xffe0 && r <= 0xffe6) ||
                    (r >= 0x20000 && r <= 0x2fffd) ||
                    (r >= 0x30000 && r <= 0x3fffd)) {
            return 2
    }
    return 1 
}

/* Calculates the number of terminal cells the given array of
 * runes will require in the terminal.
 */
func (t *Terminal) runesCellLen(ra []rune) int {
    cells := 0;
    for _, r := range ra {
        cells += t.runeWidth(r)
    }
    return cells
}

// Wraps text with given line prefix and writes to out buffer.
func (t *Terminal) wrapTextOut(out *bytes.Buffer, text []byte, prefix string) error {
    // fmt.Println(string(text))
    // escapeSpecialChars(out, text) ???
    // Normalize whitespace
    s := t.whitespace.ReplaceAll(text, []byte(" "))
    r := bytes.Runes(s)
    rpos := 0
    prefixLen := 0 // len(prefix)

    for rpos < len(r) {
        remainigCells := t.termWidth - t.xpos - prefixLen
        toolong := true

        // If we're at the beginning of a terminal line (t.xpos == 0)
        // then advance rpos past any whitespace.
        if t.xpos == 0 {
            for unicode.IsSpace(r[rpos]) {
                rpos++
            }
        }

        // if we don't need to wrap, then don't
        if t.runesCellLen(r[rpos:]) < remainigCells {
            out.WriteString(string(r[rpos:]))
            t.xpos += t.runesCellLen(r[rpos:])
            break
        }

        // search backward for a space to wrap at
        // TODO: this does not yet account for runes that require
        // more than one terminal cell
        for i := remainigCells; i > 0; i-- {
            if unicode.IsSpace(r[rpos+i]) {
                out.WriteString(string(r[rpos : rpos+i]))
                rpos += i + 1
                toolong = false
                break
            }
        }

        // If we have a run of text with no whitespace longer than the
        // remaining space availble and we're the start of a terminal
        // line (xpos == 0) then truncate the line.
        // TODO: this does not yet account for runes that require
        // more than one terminal cell
        if toolong && t.xpos == 0 {
            out.WriteString(string(r[rpos : rpos+t.termWidth-prefixLen]))
            rpos += t.termWidth - prefixLen
        }

        if rpos < len(r) {
            t.endLine(out)
            // out.WriteString(prefix)
        }
    }

    return nil
}

func (t *Terminal) endLine(out *bytes.Buffer) {
    t.xpos = 0
    out.WriteString("\n")
}

// render code chunks using verbatim, or listings if we have a language
// we currently ignore the language
func (t *Terminal) BlockCode(out *bytes.Buffer, text []byte, lang string) {
    out.Write(text)
}

func (t *Terminal) BlockQuote(out *bytes.Buffer, text []byte) {
    out.Write(text)
}

func (t *Terminal) BlockHtml(out *bytes.Buffer, text []byte) {
    log.Println("!!! BlockHtml is currently unsupported.")
    log.Println(string(text))
}

func (t *Terminal) Header(out *bytes.Buffer, text func() bool, level int) {
    marker := out.Len()
    t.endLine(out) // TODO: should not need this

    t.pushStyle()

    switch level {
    case 1: // #
        t.setFGColor(out, COLOR_RED)
    case 2: // ##
        t.setFGColor(out, COLOR_YELLOW)
    case 3: // ###
        t.setFGColor(out, COLOR_GREEN)
    case 4: // ####
        t.setFGColor(out, COLOR_BLUE)
    case 5: // #####
        t.setFGColor(out, COLOR_MAGENTA)
    case 6: // ######
        t.setFGColor(out, COLOR_CYAN)
    default:
        t.setFGColor(out, COLOR_CYAN)
    }

    t.charstyle.Bold = true
    out.Write(t.escape.Bold)

    if !text() {
        out.Truncate(marker)
        return
    }

    t.popStyle(out)
    t.endLine(out)
}

func (t *Terminal) HRule(out *bytes.Buffer) {
    hr := strings.Repeat("\u2500", t.termWidth)
    out.WriteString(hr)
    t.endLine(out)
}

func (t *Terminal) List(out *bytes.Buffer, text func() bool, flags int) {
    marker := out.Len()
    if flags&LIST_TYPE_ORDERED != 0 {
        t.listCount = 0
    }

    if !text() {
        out.Truncate(marker)
        return
    }
    if flags&LIST_TYPE_ORDERED != 0 {
        t.endLine(out)
    } else {
        t.endLine(out)
    }
}

func (t *Terminal) ListItem(out *bytes.Buffer, text []byte, flags int) {
    t.endLine(out)
    var prefix string
    if flags&LIST_TYPE_ORDERED != 0 {
        t.listCount++
        prefix = fmt.Sprintf("%d. ", t.listCount)
    } else {
        prefix = "\u2022 "
    }
    s := strings.TrimSpace( string(text) )
    // t.xpos = utf8.RuneCount([]byte(s)) + len(prefix)
    t.xpos = t.runesCellLen([]rune(s)) + len(prefix)
    out.WriteString(prefix)
    out.WriteString(s)
}

func (t *Terminal) Paragraph(out *bytes.Buffer, text func() bool) {
    marker := out.Len()
    t.endLine(out)
    if !text() {
        out.Truncate(marker)
        return
    }
    t.endLine(out)
}

// It might be better to turn this extension off and present as text unless
// we can reliably use ansi box drawing characters.
func (t *Terminal) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
}

func (t *Terminal) TableRow(out *bytes.Buffer, text []byte) {
}

func (t *Terminal) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
}

func (t *Terminal) TableCell(out *bytes.Buffer, text []byte, align int) {
}

func (t *Terminal) Footnotes(out *bytes.Buffer, text func() bool) {
    log.Println("!!! Footnotes are currently unsupported.")
}

func (t *Terminal) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {
    log.Println("!!! Footnote items are currently unsupported.")
    log.Println(string(text))
}

func (t *Terminal) AutoLink(out *bytes.Buffer, link []byte, kind int) {
    out.WriteString("href[")
    if kind == LINK_TYPE_EMAIL {
        out.WriteString("mailto:")
    }
    out.Write(link)
    out.WriteString("[]")
    out.Write(link)
    out.WriteString("]")
}

func (t *Terminal) CodeSpan(out *bytes.Buffer, text []byte) {
    t.NormalText(out, text)
}

// italic -> underline
func (t *Terminal) Emphasis(out *bytes.Buffer, text []byte) {
    if len(text) == 0 {
        return
    }
    t.pushStyle()
    out.Write(t.escape.Underline)
    // We cannot call NormalText (with line wrap) since
    // the caller processes the `text` before calling this
    // method.
    // t.NormalText(out, text)
    out.Write(text)
    t.popStyle(out)
}

// bold
func (t *Terminal) DoubleEmphasis(out *bytes.Buffer, text []byte) {
    t.pushStyle()
    out.Write(t.escape.Bold)
    out.Write(text)
    t.popStyle(out)
}

func (t *Terminal) TripleEmphasis(out *bytes.Buffer, text []byte) {
    t.pushStyle()
    out.Write(t.escape.Inverse)
    out.Write(text)
    t.popStyle(out)
}

func (t *Terminal) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
    if bytes.HasPrefix(link, []byte("http://")) || bytes.HasPrefix(link, []byte("https://")) {
        // treat it like a link
        out.WriteString("href[")
        out.Write(link)
        out.WriteString("][")
        out.Write(alt)
        out.WriteString("]")
    } else {
        out.WriteString("[")
        out.Write(link)
        out.WriteString("]")
    }
}

func (t *Terminal) LineBreak(out *bytes.Buffer) {
    out.WriteString("\n!!! LineBreak was called. Amazing.\n")
}

func (t *Terminal) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
    t.Emphasis(out, link)
    t.NormalText(out, []byte("["))
    t.NormalText(out, content)
    t.NormalText(out, []byte("]"))
}

func (t *Terminal) RawHtmlTag(out *bytes.Buffer, tag []byte) {
    log.Println("!!! Raw HTML tags are unsupported.")
    log.Println(string(tag))
}

// Not widely supported in terminal
func (t *Terminal) StrikeThrough(out *bytes.Buffer, text []byte) {
    t.NormalText(out, []byte("~~"))
    t.NormalText(out, text)
    t.NormalText(out, []byte("~~"))
}

// TODO: this
func (t *Terminal) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {
    log.Println("!!! Footnote refs are currently unsupported.")
    log.Println(string(id) + ":" + string(ref))
}

func (t *Terminal) Entity(out *bytes.Buffer, entity []byte) {
    s := html.UnescapeString( string(entity) )
    out.WriteString(s)
}

func (t *Terminal) NormalText(out *bytes.Buffer, text []byte) {
    t.wrapTextOut(out, text, "")
}

// header and footer
func (t *Terminal) DocumentHeader(out *bytes.Buffer) {
    // out.WriteString("GMAN(1) Version ")
    // out.WriteString(VERSION)
    // out.WriteString("\n")
}

func (t *Terminal) DocumentFooter(out *bytes.Buffer) {
    if (t.flags & TERM_NO_HEADER_FOOTER) == 0 {
        out.WriteString("\nGMAN(1) Version ")
        out.WriteString(VERSION)
    }
}

