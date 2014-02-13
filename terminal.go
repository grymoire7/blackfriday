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
)

// Terminal is a type that implements the Renderer interface for terminal
// output.
//
// Do not create this directly, instead use the TerminalRenderer function.
type Terminal struct {
}

// TerminalRenderer creates and configures a Terminal object, which
// satisfies the Renderer interface.
//
// flags is a set of TERMINAL_* options ORed together (currently no such options
// are defined).
func TerminalRenderer(flags int) Renderer {
	return &Terminal{}
}

// render code chunks using verbatim, or listings if we have a language
func (options *Terminal) BlockCode(out *bytes.Buffer, text []byte, lang string) {
	out.Write(text)
}

func (options *Terminal) BlockQuote(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (options *Terminal) BlockHtml(out *bytes.Buffer, text []byte) {
	out.Write(text)
}

func (options *Terminal) Header(out *bytes.Buffer, text func() bool, level int) {
	marker := out.Len()

	switch level {
	case 1:
        out.WriteString("\n\033[31m\033[1m") // #
	case 2:
        out.WriteString("\n\033[33m\033[1m") // ##
	case 3:
        out.WriteString("\n\033[32m\033[1m") // ###
	case 4:
        out.WriteString("\n\033[34m\033[1m") // ####
	case 5:
        out.WriteString("\n\033[35m\033[1m") // #####
	case 6:
        out.WriteString("\n\033[36m\033[1m") // ######
	}
	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("\033[0m\n")
}

func (options *Terminal) HRule(out *bytes.Buffer) {
	out.WriteString("\n--------------------------------\n")
}

func (options *Terminal) List(out *bytes.Buffer, text func() bool, flags int) {
	marker := out.Len()
	if flags&LIST_TYPE_ORDERED != 0 {
		out.WriteString("\n + ")
	} else {
		out.WriteString("\n  * ")
	}
	if !text() {
		out.Truncate(marker)
		return
	}
	if flags&LIST_TYPE_ORDERED != 0 {
		out.WriteString("\n + ")
	} else {
		out.WriteString("\n * ")
	}
}

func (options *Terminal) ListItem(out *bytes.Buffer, text []byte, flags int) {
	out.WriteString("\n- ")
	out.Write(text)
}

func (options *Terminal) Paragraph(out *bytes.Buffer, text func() bool) {
	marker := out.Len()
	out.WriteString("\n")
	if !text() {
		out.Truncate(marker)
		return
	}
	out.WriteString("\n")
}

func (options *Terminal) Table(out *bytes.Buffer, header []byte, body []byte, columnData []int) {
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
	out.WriteString("}\n")
	out.Write(header)
	out.WriteString(" \\\\\n\\hline\n")
	out.Write(body)
	out.WriteString("\n\\end{tabular}\n")
}

func (options *Terminal) TableRow(out *bytes.Buffer, text []byte) {
	if out.Len() > 0 {
		out.WriteString("|\n")
	}
	out.Write(text)
}

func (options *Terminal) TableHeaderCell(out *bytes.Buffer, text []byte, align int) {
	if out.Len() > 0 {
		out.WriteString(" & ")
	}
	out.Write(text)
}

func (options *Terminal) TableCell(out *bytes.Buffer, text []byte, align int) {
	if out.Len() > 0 {
		out.WriteString(" & ")
	}
	out.Write(text)
}

// TODO: this
func (options *Terminal) Footnotes(out *bytes.Buffer, text func() bool) {

}

func (options *Terminal) FootnoteItem(out *bytes.Buffer, name, text []byte, flags int) {

}

func (options *Terminal) AutoLink(out *bytes.Buffer, link []byte, kind int) {
	out.WriteString("href[")
	if kind == LINK_TYPE_EMAIL {
		out.WriteString("mailto:")
	}
	out.Write(link)
	out.WriteString("[]")
	out.Write(link)
	out.WriteString("]")
}

func (options *Terminal) CodeSpan(out *bytes.Buffer, text []byte) {
	out.WriteString("\n")
	escapeSpecialChars(out, text)
	out.WriteString("\n")
}

// bold
func (options *Terminal) DoubleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("\033[1m")
	out.Write(text)
	out.WriteString("\033[0m")
}

// italic -> underline
func (options *Terminal) Emphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("\033[0;4m")
	out.Write(text)
	out.WriteString("\033[0m")
}

func (options *Terminal) Image(out *bytes.Buffer, link []byte, title []byte, alt []byte) {
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

func (options *Terminal) LineBreak(out *bytes.Buffer) {
	out.WriteString("\n")
}

func (options *Terminal) Link(out *bytes.Buffer, link []byte, title []byte, content []byte) {
	out.WriteString("[")
	out.Write(link)
	out.WriteString("][")
	out.Write(content)
	out.WriteString("]")
}

func (options *Terminal) RawHtmlTag(out *bytes.Buffer, tag []byte) {
}

func (options *Terminal) TripleEmphasis(out *bytes.Buffer, text []byte) {
	out.WriteString("\033[7m")
	out.Write(text)
	out.WriteString("\033[0m")
}

func (options *Terminal) StrikeThrough(out *bytes.Buffer, text []byte) {
	out.WriteString("--")
	out.Write(text)
	out.WriteString("--")
}

// TODO: this
func (options *Terminal) FootnoteRef(out *bytes.Buffer, ref []byte, id int) {

}

func (options *Terminal) Entity(out *bytes.Buffer, entity []byte) {
	// TODO: convert this into a unicode character or something
	out.Write(entity)
}

func (options *Terminal) NormalText(out *bytes.Buffer, text []byte) {
    out.Write(text)
}

// header and footer
func (options *Terminal) DocumentHeader(out *bytes.Buffer) {
	// out.WriteString("GMAN(1) Version ")
	// out.WriteString(VERSION)
	// out.WriteString("\n")
}

func (options *Terminal) DocumentFooter(out *bytes.Buffer) {
	out.WriteString("\nGMAN(1) Version ")
	out.WriteString(VERSION)
}

