package calendar

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// The calendar export is a plain table, so the PDF is written here instead of
// pulling a library: the layout is ours and the format is a handful of objects.
const (
	pdfPageWidth  = 595
	pdfPageHeight = 842

	pdfLeftMargin = 40
	pdfRightEdge  = 555

	pdfTitleY    = 800
	pdfSubtitleY = 782
	pdfHeaderY   = 758
	pdfRuleY     = 752
	pdfFirstRowY = 736
	pdfRowHeight = 15
	pdfFooterY   = 32

	pdfTitleSize  = 16
	pdfBodySize   = 9
	pdfFooterSize = 8

	pdfGray = "0.45 0.45 0.5"

	pdfRowsPerPage = (pdfFirstRowY - pdfFooterY - 20) / pdfRowHeight
)

// pdfColumn is one table column: where its text starts and how much fits.
type pdfColumn struct {
	title string
	x     int
	width int
}

var (
	pdfDateColumn     = pdfColumn{title: "Fecha", x: 40, width: 62}
	pdfTimeColumn     = pdfColumn{title: "Hora", x: 104, width: 38}
	pdfEventColumn    = pdfColumn{title: "Evento", x: 145, width: 270}
	pdfCategoryColumn = pdfColumn{title: "Categoría", x: 418, width: 62}
	pdfStatusColumn   = pdfColumn{title: "Estado", x: 483, width: 72}

	pdfColumns = []pdfColumn{pdfDateColumn, pdfTimeColumn, pdfEventColumn, pdfCategoryColumn, pdfStatusColumn}
)

var pdfStatusLabel = map[string]string{
	StatusScheduled: "Programado",
	StatusFinished:  "Finalizado",
}

// BuildPDF renders the events as a table that stays readable once downloaded.
func BuildPDF(events []Event, now time.Time) []byte {
	pages := pdfPagesOf(events, now)

	// Objects 1 and 2 are the catalogue and the page tree and 3 and 4 the fonts;
	// after them come one object per page and one per content stream.
	const firstPage = 5
	firstStream := firstPage + len(pages)

	objects := make([]string, 0, 2*len(pages)+4)
	objects = append(objects,
		"<</Type/Catalog/Pages 2 0 R>>",
		fmt.Sprintf("<</Type/Pages/Kids[%s]/Count %d>>", pdfPageReferences(firstPage, len(pages)), len(pages)),
		"<</Type/Font/Subtype/Type1/BaseFont/Helvetica/Encoding/WinAnsiEncoding>>",
		"<</Type/Font/Subtype/Type1/BaseFont/Helvetica-Bold/Encoding/WinAnsiEncoding>>",
	)
	for index := range pages {
		objects = append(objects, fmt.Sprintf(
			"<</Type/Page/Parent 2 0 R/MediaBox[0 0 %d %d]/Resources<</Font<</F1 3 0 R/F2 4 0 R>>>>/Contents %d 0 R>>",
			pdfPageWidth, pdfPageHeight, firstStream+index))
	}
	for _, page := range pages {
		objects = append(objects, fmt.Sprintf("<</Length %d>>\nstream\n%s\nendstream", len(page.content), page.content))
	}

	return assemblePDF(objects)
}

type pdfPage struct {
	content string
}

func pdfPagesOf(events []Event, now time.Time) []pdfPage {
	rows := make([][]string, 0, len(events))
	for _, event := range events {
		rows = append(rows, pdfRow(event, now))
	}

	chunks := chunkRows(rows, pdfRowsPerPage)
	if len(chunks) == 0 {
		// The calendar can be empty: the file still says what it is.
		chunks = [][][]string{nil}
	}

	pages := make([]pdfPage, 0, len(chunks))
	for index, chunk := range chunks {
		pages = append(pages, pdfPage{content: pdfPageContent(chunk, index+1, len(chunks), len(events), now)})
	}
	return pages
}

func pdfPageContent(rows [][]string, number, total, eventCount int, now time.Time) string {
	lines := []string{
		pdfText("Calendario RaceControl", "F2", pdfTitleSize, pdfLeftMargin, pdfTitleY, false),
		pdfText(pdfSubtitle(now, eventCount), "F1", pdfBodySize, pdfLeftMargin, pdfSubtitleY, true),
	}
	for _, column := range pdfColumns {
		lines = append(lines, pdfText(column.title, "F2", pdfBodySize, column.x, pdfHeaderY, true))
	}
	lines = append(lines, fmt.Sprintf("%s RG 0.5 w %d %d m %d %d l S", pdfGray, pdfLeftMargin, pdfRuleY, pdfRightEdge, pdfRuleY))

	for rowIndex, row := range rows {
		y := pdfFirstRowY - rowIndex*pdfRowHeight
		for columnIndex, column := range pdfColumns {
			lines = append(lines, pdfText(row[columnIndex], "F1", pdfBodySize, column.x, y, false))
		}
	}

	lines = append(lines, pdfText(fmt.Sprintf("Página %d de %d", number, total), "F1", pdfFooterSize, pdfLeftMargin, pdfFooterY, true))
	return strings.Join(lines, "\n")
}

func pdfSubtitle(now time.Time, eventCount int) string {
	return fmt.Sprintf("Descargado el %s UTC · %s · Horarios en UTC",
		now.UTC().Format("02/01/2006 15:04"), pdfEventCount(eventCount))
}

func pdfEventCount(count int) string {
	if count == 1 {
		return "1 evento"
	}
	return fmt.Sprintf("%d eventos", count)
}

func pdfRow(event Event, now time.Time) []string {
	return []string{
		pdfFit(event.StartsAt.UTC().Format("02/01/2006"), pdfDateColumn.width, pdfBodySize),
		pdfFit(event.StartsAt.UTC().Format("15:04"), pdfTimeColumn.width, pdfBodySize),
		pdfFit(event.Name(), pdfEventColumn.width, pdfBodySize),
		pdfFit(event.CategoryName, pdfCategoryColumn.width, pdfBodySize),
		pdfFit(pdfStatusLabel[event.Status(now)], pdfStatusColumn.width, pdfBodySize),
	}
}

// pdfFit trims the text to what the column holds: a PDF has no automatic layout.
func pdfFit(text string, width, size int) string {
	const (
		averageGlyphWidth = 0.52
		ellipsis          = "..."
	)
	maxRunes := int(float64(width) / (float64(size) * averageGlyphWidth))

	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes-len(ellipsis)]) + ellipsis
}

func chunkRows(rows [][]string, size int) [][][]string {
	chunks := make([][][]string, 0, (len(rows)+size-1)/size)
	for start := 0; start < len(rows); start += size {
		chunks = append(chunks, rows[start:min(start+size, len(rows))])
	}
	return chunks
}

// pdfText positions one line absolutely, which is all the table needs.
func pdfText(text, font string, size, x, y int, gray bool) string {
	color := "0 0 0"
	if gray {
		color = pdfGray
	}
	return fmt.Sprintf("%s rg\nBT /%s %d Tf 1 0 0 1 %d %d Tm (%s) Tj ET",
		color, font, size, x, y, escapePDFText(text))
}

func pdfPageReferences(firstPage, count int) string {
	references := make([]string, 0, count)
	for page := firstPage; page < firstPage+count; page++ {
		references = append(references, fmt.Sprintf("%d 0 R", page))
	}
	return strings.Join(references, " ")
}

// assemblePDF writes the header, the objects and the cross-reference table that
// lets any reader open the file.
func assemblePDF(objects []string) []byte {
	var document bytes.Buffer
	document.WriteString("%PDF-1.4\n")

	offsets := make([]int, len(objects)+1)
	for index, body := range objects {
		offsets[index+1] = document.Len()
		fmt.Fprintf(&document, "%d 0 obj\n%s\nendobj\n", index+1, body)
	}

	startxref := document.Len()
	writeXref(&document, offsets)
	fmt.Fprintf(&document, "trailer\n<</Size %d/Root 1 0 R>>\nstartxref\n%d\n%%%%EOF\n",
		len(objects)+1, startxref)
	return document.Bytes()
}

// writeXref emits the table whose entries are fixed at 20 bytes.
func writeXref(document *bytes.Buffer, offsets []int) {
	fmt.Fprintf(document, "xref\n0 %d\n", len(offsets))
	document.WriteString("0000000000 65535 f \n")
	for object := 1; object < len(offsets); object++ {
		fmt.Fprintf(document, "%010d 00000 n \n", offsets[object])
	}
}

// escapePDFText writes the value as bytes in WinAnsi encoding, the one the
// standard fonts expect, and escapes the three characters a string cannot hold.
func escapePDFText(text string) string {
	var encoded bytes.Buffer
	for _, character := range text {
		switch character {
		case '(', ')', '\\':
			encoded.WriteByte('\\')
			encoded.WriteByte(byte(character))
		default:
			encoded.WriteByte(winAnsi(character))
		}
	}
	return encoded.String()
}

func winAnsi(character rune) byte {
	switch {
	case character == '\n', character == '\r', character == '\t':
		return ' '
	case character >= ' ' && character <= '~':
		return byte(character)
	case character >= 0xA0 && character <= 0xFF:
		// WinAnsi keeps Latin-1 where it differs from the 0x80-0x9F range.
		return byte(character)
	}
	if mapped, ok := winAnsiExtras[character]; ok {
		return mapped
	}
	return '?'
}

// winAnsiExtras covers CP1252 in the range where it leaves Latin-1, which is
// where the typographic marks the labels use live.
var winAnsiExtras = map[rune]byte{
	'€': 0x80, '‚': 0x82, 'ƒ': 0x83, '„': 0x84, '…': 0x85, '†': 0x86,
	'‡': 0x87, 'ˆ': 0x88, '‰': 0x89, 'Š': 0x8A, '‹': 0x8B, 'Œ': 0x8C,
	'Ž': 0x8E, '‘': 0x91, '’': 0x92, '“': 0x93, '”': 0x94, '•': 0x95,
	'–': 0x96, '—': 0x97, '˜': 0x98, '™': 0x99, 'š': 0x9A, '›': 0x9B,
	'œ': 0x9C, 'ž': 0x9E, 'Ÿ': 0x9F,
}
