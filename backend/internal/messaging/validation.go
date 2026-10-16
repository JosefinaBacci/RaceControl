package messaging

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/racecontrol/backend/internal/domain"
)

// textRule is how a free-text field is cleaned and bounded. The bounds match the
// CHECK constraints of 000028, so a refused value is a 400 and never a 500.
type textRule struct {
	field     string
	label     string
	maxLength int
}

var (
	subjectRule = textRule{field: "subject", label: "el asunto", maxLength: 120}
	bodyRule    = textRule{field: "body", label: "el mensaje", maxLength: 4000}
)

// normalize trims the value and checks it. A NUL character is refused because
// PostgreSQL cannot store it in text.
func (rule textRule) normalize(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", domain.NewValidationError(rule.field, rule.label+" es obligatorio")
	}
	if utf8.RuneCountInString(value) > rule.maxLength {
		return "", domain.NewValidationError(rule.field,
			fmt.Sprintf("%s no puede superar los %d caracteres", rule.label, rule.maxLength))
	}
	if strings.ContainsRune(value, 0) {
		return "", domain.NewValidationError(rule.field, rule.label+" contiene caracteres no válidos")
	}
	return value, nil
}

// Box picks which conversations a list shows.
type Box string

const (
	BoxInbox Box = "inbox"
	BoxSent  Box = "sent"
)

// resolve answers an empty box as the inbox and refuses one it does not know.
func (box Box) resolve() (Box, error) {
	switch box {
	case "":
		return BoxInbox, nil
	case BoxInbox, BoxSent:
		return box, nil
	}
	return "", domain.NewValidationError("box", "la bandeja debe ser inbox o sent")
}

// includes tells whether the conversation belongs in the box: the inbox holds
// every conversation of the user, the sent box those they wrote in.
func (box Box) includes(summary ConversationSummary) bool {
	return box == BoxInbox || summary.HasOwnMessages
}
