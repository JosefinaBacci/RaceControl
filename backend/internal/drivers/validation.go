package drivers

import (
	"regexp"
	"strings"

	"github.com/racecontrol/backend/internal/domain"
)

const (
	licenseMinLength = 3
	licenseMaxLength = 32
	maxNumber        = 999
)

var countryCodePattern = regexp.MustCompile(`^[a-z]{2}$`)

// driverState is a driver reduced to the values a write needs, so create and
// update validate the very same shape.
type driverState struct {
	firstName     string
	lastName      string
	licenseNumber string
	countryCode   string
	isStarter     bool
	number        *int32
}

func newState(input NewDriver) driverState {
	return normalize(driverState{
		firstName:     input.FirstName,
		lastName:      input.LastName,
		licenseNumber: input.LicenseNumber,
		countryCode:   input.CountryCode,
		isStarter:     input.IsStarter,
		number:        input.Number,
	})
}

func (state driverState) withChanges(changes DriverChanges) driverState {
	if changes.FirstName != nil {
		state.firstName = *changes.FirstName
	}
	if changes.LastName != nil {
		state.lastName = *changes.LastName
	}
	if changes.LicenseNumber != nil {
		state.licenseNumber = *changes.LicenseNumber
	}
	if changes.CountryCode != nil {
		state.countryCode = *changes.CountryCode
	}
	if changes.IsStarter != nil {
		state.isStarter = *changes.IsStarter
	}
	if changes.Number != nil {
		state.number = changes.Number
	}
	return normalize(state)
}

func normalize(state driverState) driverState {
	state.firstName = strings.TrimSpace(state.firstName)
	state.lastName = strings.TrimSpace(state.lastName)
	state.licenseNumber = strings.TrimSpace(state.licenseNumber)
	state.countryCode = strings.ToLower(strings.TrimSpace(state.countryCode))
	return state
}

func validateState(state driverState) error {
	if state.firstName == "" {
		return domain.NewValidationError("firstName", "el nombre es obligatorio")
	}
	if state.lastName == "" {
		return domain.NewValidationError("lastName", "el apellido es obligatorio")
	}
	if length := len(state.licenseNumber); length < licenseMinLength || length > licenseMaxLength {
		return domain.NewValidationError("licenseNumber",
			"la licencia debe tener entre 3 y 32 caracteres")
	}
	if !countryCodePattern.MatchString(state.countryCode) {
		return domain.NewValidationError("countryCode", "el país va como dos letras, por ejemplo ar")
	}
	if state.number == nil {
		return domain.NewValidationError("number", "el número de temporada es obligatorio")
	}
	if *state.number < 0 || *state.number > maxNumber {
		return domain.NewValidationError("number", "el número debe estar entre 0 y 999")
	}
	return nil
}
