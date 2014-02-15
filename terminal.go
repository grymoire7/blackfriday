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
    "log"
    "syscall"
    "unsafe"
)

const (
    keyEscape = 27
)

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
    escape    *EscapeCodes
    termWidth int
}

// TerminalRenderer creates and configures a Terminal object, which
// satisfies the Renderer interface.
//
// flags is a set of TERMINAL_* options ORed together (currently no such options
// are defined).
func TerminalRenderer(flags int) Renderer {
    width, _, err := getTerminalSize(0)
    if err != nil {
        width = 80
    }
    log.Println("width: ", width)

    return &Terminal{
        escape:    &vt100EscapeCodes,
        termWidth: width,
    }
}

// GetSize returns the dimensions of the given terminal.
func getTerminalSize(fd int) (width, height int, err error) {
    var dimensions [4]uint16

    if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
        return -1, -1, err
    }
    return int(dimensions[1]), int(dimensions[0]), nil
}

// render code chunks using verbatim, or listings if we have a language
func (t *Terminal) BlockCode(out *bytes.Buffer, text []byte, lang string) {
    out.Write(text)
}

func (t *Terminal) BlockQuote(out *bytes.Buffer, text []byte) {
    out.Write(text)
}

func (t *Terminal) BlockHtml(out *bytes.Buffer, text []byte) {
    out.Write(text)
}

func (t *Terminal) Header(out *bytes.Buffer, text func() bool, level int) {
    marker := out.Len()
    out.WriteString("\n")

    switch level {
    case 1: // #
        out.Write(t.escape.Red)
        out.Write(t.escape.Bold)
    case 2: // ##
        out.Write(t.escape.Yellow)
        out.Write(t.escape.Bold)
    case 3: // ###
        out.Write(t.escape.Green)
        out.Write(t.escape.Bold)
    case 4: // ####
        out.Write(t.escape.Blue)
        out.Write(t.escape.Bold)
    case 5: // #####
        out.Write(t.escape.Magenta)
        out.Write(t.escape.Bold)
    case 6: // ######
        out.Write(t.escape.Cyan)
        out.Write(t.escape.Bold)
    }

    if !text() {
        out.Truncate(marker)
        return
    }

    out.Write(t.escape.Reset)
    out.WriteString("\n")
}

func (t *Terminal) HRule(out *bytes.Buffer) {
    out.WriteString("\n--------------------------------\n")
}

func (t *Terminal) List(out *bytes.Buffer, text func() bool, flags int) {
    marker := out.Len()
    /*
    	if flags&LIST_TYPE_ORDERED != 0 {
    		out.WriteString("\n a ")
    	} else {
    		out.WriteString("\n  b ")
    	}
    */
    if !text() {
        out.Truncate(marker)
        return
    }
    if flags&LIST_TYPE_ORDERED != 0 {
        out.WriteString("\n")
    } else {
        out.WriteString("\n")
    }
}

func (t *Terminal) ListItem(out *bytes.Buffer, text []byte, flags int) {
    out.WriteString("\n\u2022 ")
    out.Write(text)
}

func (t *Terminal) Paragraph(out *bytes.Buffer, text func() bool) {
    marker := out.Len()
    out.WriteString("\n")
    if !text() {
        out.Truncate(marker)
        return
    }
    out.WriteString("\n")
}

// It might be better to turn this extension off and present as text unless
// we can reliably use ansi box drawing characters.
func (t *Terminal) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
    /*
    	out.WriteString("\n\\begin{tabular}{")
    	for _, elt := range columnData {
    		switch elt {
    		case TABLE_ALIGNMENT_LEFT:
    			out.WriteByte('l')
    		case TABLE_ALIGNMENT_RIGHT:
    			out.WriteByte('r')
    		default:
    			out.WriteByte('c')
    		}
    	}
    */
    out.Write(header)
    out.Write(body)
}

func (t *Terminal) TableRow(out *bytes.Buffer, text []byte) {
    if out.Len() > 0 {
        out.WriteString("|\n")
    }
    out.Write(text)
}

func (t *Terminal) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
    if out.Len() > 0 {
        out.WriteString(" & ")
    }
    out.Write(text)
}

func (t *Terminal) TableCell(out *bytes.Buffer, text []byte, align int) {
    if out.Len() > 0 {
        out.WriteString(" & ")
    }
    out.Write(text)
}

// TODO: this
func (t *Terminal) Footnotes(out *bytes.Buffer, text func() bool) {

}

func (t *Terminal) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {

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
    out.WriteString("\n")
    escapeSpecialChars(out, text)
    out.WriteString("\n")
}

// bold
func (t *Terminal) DoubleEmphasis(out *bytes.Buffer, text []byte) {
    out.WriteString("\033[1m")
    out.Write(text)
    out.WriteString("\033[0m")
}

// italic -> underline
func (t *Terminal) Emphasis(out *bytes.Buffer, text []byte) {
    out.WriteString("\033[0;4m")
    out.Write(text)
    out.WriteString("\033[0m")
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
    out.WriteString("\n")
}

func (t *Terminal) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
    out.WriteString("[")
    out.Write(link)
    out.WriteString("][")
    out.Write(content)
    out.WriteString("]")
}

func (t *Terminal) RawHtmlTag(out *bytes.Buffer, tag []byte) {
}

func (t *Terminal) TripleEmphasis(out *bytes.Buffer, text []byte) {
    out.WriteString("\033[7m")
    out.Write(text)
    out.WriteString("\033[0m")
}

func (t *Terminal) StrikeThrough(out *bytes.Buffer, text []byte) {
    out.WriteString("--")
    out.Write(text)
    out.WriteString("--")
}

// TODO: this
func (t *Terminal) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {

}

func (t *Terminal) Entity(out *bytes.Buffer, entity []byte) {
    // TODO: convert this into a unicode character or something
    out.Write(entity)
}

func (t *Terminal) NormalText(out *bytes.Buffer, text []byte) {
    out.Write(text)
}

// header and footer
func (t *Terminal) DocumentHeader(out *bytes.Buffer) {
    // out.WriteString("GMAN(1) Version ")
    // out.WriteString(VERSION)
    // out.WriteString("\n")
}

func (t *Terminal) DocumentFooter(out *bytes.Buffer) {
    out.WriteString("\nGMAN(1) Version ")
    out.WriteString(VERSION)
}
