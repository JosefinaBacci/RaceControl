package calendar

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	// RFC 5545 caps a content line at 75 octets and folds it with CRLF plus a space.
	icsLineOctetLimit = 75
	icsTimestamp      = "20060102T150405Z"
	icsProductID      = "-//RaceControl//Calendario FIA//ES"
	icsCalendarName   = "Calendario RaceControl"
	icsUIDDomain      = "racecontrol"
)

// BuildICS renders the events as an RFC 5545 calendar, the standard file a
// calendar application opens without a connection.
func BuildICS(events []Event, now time.Time) []byte {
	var document strings.Builder
	for _, line := range icsLines(events, now) {
		document.WriteString(foldICSLine(line))
		document.WriteString("\r\n")
	}
	return []byte(document.String())
}

func icsLines(events []Event, now time.Time) []string {
	lines := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"PRODID:" + icsProductID,
		"CALSCALE:GREGORIAN",
		"METHOD:PUBLISH",
		"X-WR-CALNAME:" + icsCalendarName,
	}
	for _, event := range events {
		lines = append(lines, icsEventLines(event, now)...)
	}
	return append(lines, "END:VCALENDAR")
}

func icsEventLines(event Event, now time.Time) []string {
	lines := []string{
		"BEGIN:VEVENT",
		fmt.Sprintf("UID:event-%d@%s", event.ID, icsUIDDomain),
		"DTSTAMP:" + icsTimestampOf(now),
		"DTSTART:" + icsTimestampOf(event.StartsAt),
		"DTEND:" + icsTimestampOf(event.EndsAt),
		"SUMMARY:" + escapeICSText(event.Name()),
		"CATEGORIES:" + escapeICSText(event.CategoryName),
	}
	if location := textOf(event.CircuitName); location != "" {
		lines = append(lines, "LOCATION:"+escapeICSText(location))
	}
	return append(lines, "END:VEVENT")
}

func icsTimestampOf(value time.Time) string {
	return value.UTC().Format(icsTimestamp)
}

// escapeICSText escapes the characters RFC 5545 reserves in a text value. The
// backslash goes first so it does not escape the escapes.
func escapeICSText(value string) string {
	return strings.NewReplacer(
		"\\", "\\\\",
		";", "\\;",
		",", "\\,",
		"\r\n", "\\n",
		"\n", "\\n",
		"\r", "\\n",
	).Replace(value)
}

func foldICSLine(line string) string {
	var folded strings.Builder
	remaining, limit := line, icsLineOctetLimit
	for len(remaining) > limit {
		cut := runeStartAtOrBefore(remaining, limit)
		folded.WriteString(remaining[:cut])
		folded.WriteString("\r\n ")
		remaining = remaining[cut:]
		// A continuation line spends one of its octets on the leading space.
		limit = icsLineOctetLimit - 1
	}
	folded.WriteString(remaining)
	return folded.String()
}

// runeStartAtOrBefore moves the cut back to a rune boundary so a multi-byte
// character is never split between two lines.
func runeStartAtOrBefore(value string, limit int) int {
	cut := limit
	for cut > 0 && !utf8.RuneStart(value[cut]) {
		cut--
	}
	if cut == 0 {
		return limit
	}
	return cut
}
