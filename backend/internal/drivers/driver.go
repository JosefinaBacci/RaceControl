package drivers

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
)

// Driver is a person on a team's lineup: the person facts plus the membership
// facts that vary with time (the starter/reserve role and the season number).
type Driver struct {
	ID            int64
	FirstName     string
	LastName      string
	LicenseNumber string
	CountryCode   string
	IsStarter     bool
	Number        *int32
}

// NewDriver is a create request. The number is optional: a rookie can join the
// team before the championship assigns one.
type NewDriver struct {
	FirstName     string
	LastName      string
	LicenseNumber string
	CountryCode   string
	IsStarter     bool
	Number        *int32
}

// DriverChanges carries the fields an update may touch. A nil field is left as it
// was and no field can be cleared.
type DriverChanges struct {
	FirstName     *string
	LastName      *string
	LicenseNumber *string
	CountryCode   *string
	IsStarter     *bool
	Number        *int32
}

func driverFromRow(id int64, firstName, lastName, licenseNumber, countryCode string, isStarter bool, number pgtype.Int4) Driver {
	return Driver{
		ID:            id,
		FirstName:     firstName,
		LastName:      lastName,
		LicenseNumber: licenseNumber,
		CountryCode:   countryCode,
		IsStarter:     isStarter,
		Number:        db.Int32Pointer(number),
	}
}

func presentTeam(teamID int64) pgtype.Int8 {
	return pgtype.Int8{Int64: teamID, Valid: true}
}
