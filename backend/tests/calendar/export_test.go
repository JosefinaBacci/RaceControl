package calendartest

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/racecontrol/backend/internal/calendar"
)

var exportNow = time.Date(2026, 10, 15, 12, 30, 0, 0, time.UTC)

func raceEvent(id int64) calendar.Event {
	return calendar.Event{
		ID:           id,
		Kind:         calendar.KindRace,
		StartsAt:     time.Date(2026, 11, 15, 15, 0, 0, 0, time.UTC),
		EndsAt:       time.Date(2026, 11, 15, 17, 0, 0, 0, time.UTC),
		GPName:       pointer("Gran Premio de São Paulo"),
		RoundNumber:  pointer(int32(19)),
		CategoryName: "F1",
		CircuitName:  pointer("Interlagos"),
	}
}

func TestBuildICSWrapsEveryEventInItsOwnBlock(t *testing.T) {
	document := string(calendar.BuildICS([]calendar.Event{raceEvent(1)}, exportNow))

	if !strings.HasPrefix(document, "BEGIN:VCALENDAR\r\n") {
		t.Fatalf("the calendar does not open with the standard header: %q", document)
	}
	if !strings.HasSuffix(document, "END:VCALENDAR\r\n") {
		t.Fatalf("the calendar does not close: %q", document)
	}
	if strings.Count(document, "BEGIN:VEVENT") != 1 || strings.Count(document, "END:VEVENT") != 1 {
		t.Fatalf("want exactly one event block, got %q", document)
	}

	wanted := []string{
		"UID:event-1@racecontrol",
		"DTSTAMP:20261015T123000Z",
		"DTSTART:20261115T150000Z",
		"DTEND:20261115T170000Z",
		"SUMMARY:Gran Premio de São Paulo · Ronda 19",
		"CATEGORIES:F1",
		"LOCATION:Interlagos",
	}
	for _, line := range wanted {
		if !strings.Contains(document, line+"\r\n") {
			t.Fatalf("the calendar is missing %q:\n%s", line, document)
		}
	}
}

func TestBuildICSEscapesReservedCharacters(t *testing.T) {
	document := string(calendar.BuildICS([]calendar.Event{{
		ID:           7,
		Kind:         calendar.KindTechnicalControl,
		StartsAt:     exportNow,
		EndsAt:       exportNow.Add(time.Hour),
		ControlType:  pointer(`Verificación de pesos, 800 kg; revisión \ total`),
		CategoryName: "F1",
	}}, exportNow))

	want := `SUMMARY:Verificación de pesos\, 800 kg\; revisión \\ total`
	if !strings.Contains(document, want) {
		t.Fatalf("the summary is not escaped:\ngot  %q\nwant %q", document, want)
	}
}

func TestBuildICSFoldsLongLinesAtSeventyFiveOctets(t *testing.T) {
	controlType := strings.Repeat("aclaración de la normativa técnica ", 4)
	document := string(calendar.BuildICS([]calendar.Event{{
		ID:           7,
		Kind:         calendar.KindTechnicalControl,
		StartsAt:     exportNow,
		EndsAt:       exportNow.Add(time.Hour),
		ControlType:  &controlType,
		CategoryName: "F1",
	}}, exportNow))

	if !strings.Contains(document, "\r\n ") {
		t.Fatalf("a long line was printed without a continuation line:\n%s", document)
	}
	for _, line := range strings.Split(document, "\r\n") {
		if len(line) > 75 {
			t.Fatalf("line of %d octets was not folded: %q", len(line), line)
		}
		// The fold is counted in octets, so it must not cut a rune in half.
		if !utf8.ValidString(line) {
			t.Fatalf("the fold split a multi-byte character: %q", line)
		}
	}
}

func TestBuildICSWithoutEventsIsStillAValidCalendar(t *testing.T) {
	document := string(calendar.BuildICS(nil, exportNow))

	if strings.Contains(document, "BEGIN:VEVENT") {
		t.Fatalf("an empty calendar must not carry events: %q", document)
	}
	if !strings.HasSuffix(document, "END:VCALENDAR\r\n") {
		t.Fatalf("an empty calendar must still close: %q", document)
	}
}

func TestEventStatusFollowsTheEndOfTheEvent(t *testing.T) {
	cases := []struct {
		name string
		end  time.Time
		want string
	}{
		{name: "an event that has not ended is scheduled", end: exportNow.Add(time.Hour), want: calendar.StatusScheduled},
		{name: "an event that already ended is finished", end: exportNow.Add(-time.Hour), want: calendar.StatusFinished},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			event := calendar.Event{EndsAt: c.end}
			if got := event.Status(exportNow); got != c.want {
				t.Fatalf("Status() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestBuildPDFIsAReadableDocument(t *testing.T) {
	document := calendar.BuildPDF([]calendar.Event{raceEvent(1)}, exportNow)

	assertPDFStructure(t, document)

	content := string(document)
	if !strings.Contains(content, "/Type/Catalog") || !strings.Contains(content, "/Root 1 0 R") {
		t.Fatalf("the document has no catalogue:\n%s", content)
	}
	if !strings.Contains(content, "15/11/2026") || !strings.Contains(content, "Gran Premio de S\xe3o Paulo") {
		t.Fatalf("the table does not carry the event:\n%s", content)
	}
	if !strings.Contains(content, "Programado") {
		t.Fatalf("the table does not carry the derived status:\n%s", content)
	}
}

func TestBuildPDFWritesAccentsInWinAnsi(t *testing.T) {
	document := calendar.BuildPDF([]calendar.Event{raceEvent(1)}, exportNow)

	// The base fonts read WinAnsi: a UTF-8 rune would print as two glyphs.
	if bytes.Contains(document, []byte("São")) {
		t.Fatal("the document kept the text in UTF-8")
	}
	if !bytes.Contains(document, []byte("S\xe3o Paulo")) {
		t.Fatalf("the ã does not travel in WinAnsi:\n%s", document)
	}
}

func TestBuildPDFPaginates(t *testing.T) {
	events := make([]calendar.Event, 0, 90)
	for id := int64(1); id <= 90; id++ {
		events = append(events, raceEvent(id))
	}

	document := calendar.BuildPDF(events, exportNow)
	assertPDFStructure(t, document)

	pages := strings.Count(string(document), "/Type/Page/")
	if pages < 2 {
		t.Fatalf("90 events fit in %d page(s), want more than one", pages)
	}
	if !strings.Contains(string(document), fmt.Sprintf("/Count %d", pages)) {
		t.Fatalf("the page tree does not count the %d pages", pages)
	}
	// The accent travels as a WinAnsi byte, so the needle skips it.
	if !strings.Contains(string(document), fmt.Sprintf("P\xe1gina 1 de %d", pages)) {
		t.Fatal("the footer does not number the pages")
	}
}

func TestBuildPDFWithoutEventsStillOpens(t *testing.T) {
	assertPDFStructure(t, calendar.BuildPDF(nil, exportNow))
}

// firstLine cuts the value at the end of its line.
func firstLine(value string) string {
	line, _, _ := strings.Cut(value, "\n")
	return line
}

// assertPDFStructure follows the cross-reference table the way a reader does,
// which is the only way to tell a complete file from a truncated one.
func assertPDFStructure(t *testing.T, document []byte) {
	t.Helper()

	if !bytes.HasPrefix(document, []byte("%PDF-1.4\n")) {
		t.Fatalf("the file does not open with the PDF header: %q", document)
	}
	if !bytes.HasSuffix(document, []byte("%%EOF\n")) {
		t.Fatalf("the file does not close with the end marker: %q", document)
	}

	text := string(document)
	marker := strings.LastIndex(text, "startxref\n")
	if marker < 0 {
		t.Fatalf("the file has no startxref:\n%s", text)
	}

	start, err := strconv.Atoi(firstLine(text[marker+len("startxref\n"):]))
	if err != nil {
		t.Fatalf("startxref does not point at a byte offset: %v", err)
	}
	if start >= len(text) || !strings.HasPrefix(text[start:], "xref\n") {
		t.Fatalf("startxref points at %q, not at the table", text[start:])
	}

	lines := strings.Split(text[start:], "\n")
	count, err := strconv.Atoi(strings.TrimPrefix(lines[1], "0 "))
	if err != nil {
		t.Fatalf("the table does not declare its entry count: %q", lines[1])
	}

	// Every entry is exactly 20 bytes long and the first one is the free head.
	entries := text[start+len("xref\n")+len(lines[1])+1:]
	for object := 1; object < count; object++ {
		entry := entries[object*20 : (object+1)*20]
		offset, err := strconv.Atoi(strings.TrimSpace(entry[:10]))
		if err != nil {
			t.Fatalf("entry %d of the table is malformed: %q", object, entry)
		}
		want := fmt.Sprintf("%d 0 obj\n", object)
		if !strings.HasPrefix(text[offset:], want) {
			t.Fatalf("entry %d points at %q, want %q", object, text[offset:offset+len(want)], want)
		}
	}
}
