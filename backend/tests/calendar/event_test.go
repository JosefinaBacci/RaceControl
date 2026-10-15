package calendartest

import (
	"testing"
	"time"

	"github.com/racecontrol/backend/internal/calendar"
)

func pointer[T any](value T) *T {
	return &value
}

func TestEventNamePerKind(t *testing.T) {
	cases := []struct {
		name  string
		event calendar.Event
		want  string
	}{
		{
			name:  "race names the Grand Prix and the round",
			event: calendar.Event{Kind: calendar.KindRace, GPName: pointer("Gran Premio de Singapur"), RoundNumber: pointer(int32(18))},
			want:  "Gran Premio de Singapur · Ronda 18",
		},
		{
			name:  "tyre test names its fixed text and date",
			event: calendar.Event{Kind: calendar.KindTyreTest, StartsAt: time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)},
			want:  "Prueba de neumáticos · 20/10/2026",
		},
		{
			name:  "technical control names its type",
			event: calendar.Event{Kind: calendar.KindTechnicalControl, ControlType: pointer("Verificación de pesos")},
			want:  "Verificación de pesos",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.event.Name(); got != c.want {
				t.Fatalf("Name() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestEventNameUsesTheUTCDate keeps the label of a tyre test from shifting with
// the server's timezone: the same event must read the same everywhere.
func TestEventNameUsesTheDateInUTC(t *testing.T) {
	event := calendar.Event{
		Kind:     calendar.KindTyreTest,
		StartsAt: time.Date(2026, 10, 20, 23, 30, 0, 0, time.FixedZone("ART", -3*3600)),
	}

	if got := event.Name(); got != "Prueba de neumáticos · 21/10/2026" {
		t.Fatalf("Name() = %q, want the UTC date 21/10/2026", got)
	}
}

func TestEventNameFallsBackWhenTheRaceHasNoRound(t *testing.T) {
	event := calendar.Event{Kind: calendar.KindRace, GPName: pointer("Gran Premio de Singapur")}

	if got := event.Name(); got != "Gran Premio de Singapur" {
		t.Fatalf("Name() = %q, want the Grand Prix name alone", got)
	}
}
