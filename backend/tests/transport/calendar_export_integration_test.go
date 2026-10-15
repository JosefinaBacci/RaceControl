package transporttest

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// download fetches a file endpoint: its body is a document, not JSON, so the
// shared do helper cannot decode it.
func (h *harness) download(path string) (*http.Response, []byte) {
	h.t.Helper()

	request, err := http.NewRequest(http.MethodGet, h.server.URL+path, nil)
	if err != nil {
		h.t.Fatalf("new request: %v", err)
	}

	result, err := h.browser().Do(request)
	if err != nil {
		h.t.Fatalf("GET %s: %v", path, err)
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		h.t.Fatalf("read body: %v", err)
	}
	return result, body
}

func TestCalendarDownloadsArePublic(t *testing.T) {
	h := newHarness(t)

	ics, _ := h.download("/calendar.ics")
	if ics.StatusCode != http.StatusOK {
		t.Fatalf("the calendar is not downloadable without a session: %d", ics.StatusCode)
	}
	pdf, _ := h.download("/calendar.pdf")
	if pdf.StatusCode != http.StatusOK {
		t.Fatalf("the PDF is not downloadable without a session: %d", pdf.StatusCode)
	}
}

func TestCalendarICSDescribesTheSeededCalendar(t *testing.T) {
	h := newHarness(t)

	result, body := h.download("/calendar.ics")
	if result.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", result.StatusCode)
	}
	if type_ := result.Header.Get("Content-Type"); type_ != "text/calendar; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want the calendar type", type_)
	}
	if disposition := result.Header.Get("Content-Disposition"); !strings.HasPrefix(disposition, "attachment;") {
		t.Fatalf("Content-Disposition = %q, want an attachment", disposition)
	}

	document := string(body)
	if !strings.HasPrefix(document, "BEGIN:VCALENDAR\r\n") || !strings.HasSuffix(document, "END:VCALENDAR\r\n") {
		t.Fatalf("the document is not a calendar:\n%s", document)
	}
	if !strings.Contains(document, "\r\n") {
		t.Fatal("the document does not use the CRLF line endings the standard requires")
	}
	for _, wanted := range []string{"BEGIN:VEVENT", "DTSTART:", "SUMMARY:", "UID:event-"} {
		if !strings.Contains(document, wanted) {
			t.Fatalf("the calendar is missing %q:\n%s", wanted, document)
		}
	}
}

func TestCalendarICSHonoursTheCategoryFilter(t *testing.T) {
	h := newHarness(t)

	_, everything := h.download("/calendar.ics")
	_, f1 := h.download("/calendar.ics?category=f1")

	if len(f1) > len(everything) {
		t.Fatalf("the filtered calendar is longer than the whole one: %d > %d", len(f1), len(everything))
	}
	if !strings.Contains(string(f1), "CATEGORIES:Formula 1\r\n") {
		t.Fatalf("the filtered calendar does not carry the F1 events:\n%s", f1)
	}
	if strings.Contains(string(f1), "CATEGORIES:Formula 2\r\n") || strings.Contains(string(f1), "CATEGORIES:Formula 3\r\n") {
		t.Fatalf("the filtered calendar carries other categories:\n%s", f1)
	}
}

func TestCalendarPDFCarriesTheSeededCalendar(t *testing.T) {
	h := newHarness(t)

	result, body := h.download("/calendar.pdf")
	if result.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", result.StatusCode)
	}
	if type_ := result.Header.Get("Content-Type"); type_ != "application/pdf" {
		t.Fatalf("Content-Type = %q, want a PDF", type_)
	}

	if !bytes.HasPrefix(body, []byte("%PDF-1.4\n")) || !strings.HasSuffix(string(body), "%%EOF\n") {
		t.Fatalf("the document is not a PDF: %q", body[:min(len(body), 32)])
	}
	document := string(body)
	if !strings.Contains(document, "/Type/Catalog") || !strings.Contains(document, "Calendario RaceControl") {
		t.Fatal("the document is not the calendar table")
	}
	if !strings.Contains(document, fmt.Sprintf("/Count %d", strings.Count(document, "/Type/Page/"))) {
		t.Fatal("the page tree does not count its pages")
	}
}

func TestCalendarDownloadsRejectAnUnknownFilter(t *testing.T) {
	h := newHarness(t)

	result, _ := h.download("/calendar.ics?kind=friendly")
	if result.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", result.StatusCode)
	}
}
