package transport

import (
	"context"
	"net/http"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/drivers"
)

type DriverManager interface {
	History(ctx context.Context, reader domain.Identity, teamID *int64) ([]drivers.Assignment, error)
	Lineup(ctx context.Context, teamID int64) ([]drivers.Driver, error)
	Get(ctx context.Context, teamID, id int64) (drivers.Driver, error)
	Create(ctx context.Context, teamID int64, input drivers.NewDriver) (drivers.Driver, error)
	Update(ctx context.Context, teamID, id int64, changes drivers.DriverChanges) (drivers.Driver, error)
	Release(ctx context.Context, teamID, id int64) error
	Directory(ctx context.Context, filter drivers.PublicFilter) ([]drivers.PublicDriver, error)
	PublicDriver(ctx context.Context, id int64) (drivers.PublicDriver, error)
}

type driverHandlers struct {
	drivers DriverManager
}

type driverResponse struct {
	ID            int64  `json:"id"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	LicenseNumber string `json:"licenseNumber"`
	CountryCode   string `json:"countryCode"`
	IsStarter     bool   `json:"isStarter"`
	Number        *int32 `json:"number"`
}

type driverListResponse struct {
	Drivers []driverResponse `json:"drivers"`
}

type publicDriverResponse struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	CountryCode  string `json:"countryCode"`
	IsStarter    bool   `json:"isStarter"`
	Number       *int32 `json:"number"`
	TeamID       int64  `json:"teamId"`
	TeamName     string `json:"teamName"`
	CategoryCode string `json:"categoryCode"`
	CategoryName string `json:"categoryName"`
}

type publicDriverListResponse struct {
	Drivers []publicDriverResponse `json:"drivers"`
}

type createDriverRequest struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	LicenseNumber string `json:"licenseNumber"`
	CountryCode   string `json:"countryCode"`
	IsStarter     bool   `json:"isStarter"`
	Number        *int32 `json:"number"`
}

type updateDriverRequest struct {
	FirstName     *string `json:"firstName"`
	LastName      *string `json:"lastName"`
	LicenseNumber *string `json:"licenseNumber"`
	CountryCode   *string `json:"countryCode"`
	IsStarter     *bool   `json:"isStarter"`
	Number        *int32  `json:"number"`
}

func (h *driverHandlers) list(w http.ResponseWriter, r *http.Request) {
	teamID, err := teamIDFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	lineup, err := h.drivers.Lineup(r.Context(), teamID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := driverListResponse{Drivers: make([]driverResponse, 0, len(lineup))}
	for _, driver := range lineup {
		response.Drivers = append(response.Drivers, toDriverResponse(driver))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *driverHandlers) get(w http.ResponseWriter, r *http.Request) {
	h.respondWithDriver(w, r, http.StatusOK, func(teamID, id int64) (drivers.Driver, error) {
		return h.drivers.Get(r.Context(), teamID, id)
	})
}

func (h *driverHandlers) create(w http.ResponseWriter, r *http.Request) {
	teamID, err := teamIDFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var request createDriverRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	driver, err := h.drivers.Create(r.Context(), teamID, drivers.NewDriver(request))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toDriverResponse(driver))
}

func (h *driverHandlers) update(w http.ResponseWriter, r *http.Request) {
	var request updateDriverRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	h.respondWithDriver(w, r, http.StatusOK, func(teamID, id int64) (drivers.Driver, error) {
		return h.drivers.Update(r.Context(), teamID, id, drivers.DriverChanges(request))
	})
}

func (h *driverHandlers) release(w http.ResponseWriter, r *http.Request) {
	teamID, err := teamIDFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	if err := h.drivers.Release(r.Context(), teamID, id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// directory and show are the public reads of US12: anyone, with or without a
// session, sees the current lineup.
func (h *driverHandlers) directory(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	lineup, err := h.drivers.Directory(r.Context(), drivers.PublicFilter{
		Category: query.Get("category"),
		Search:   query.Get("q"),
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := publicDriverListResponse{Drivers: make([]publicDriverResponse, 0, len(lineup))}
	for _, driver := range lineup {
		response.Drivers = append(response.Drivers, toPublicDriverResponse(driver))
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *driverHandlers) show(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	driver, err := h.drivers.PublicDriver(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toPublicDriverResponse(driver))
}

func (h *driverHandlers) respondWithDriver(w http.ResponseWriter, r *http.Request, status int,
	operation func(teamID, id int64) (drivers.Driver, error)) {
	teamID, err := teamIDFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	driver, err := operation(teamID, id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, status, toDriverResponse(driver))
}

// teamIDFrom resolves the team a scoped handler works on. A team administrator
// always carries one; anyone else reaching here is forbidden.
func teamIDFrom(r *http.Request) (int64, error) {
	identity, err := identityFrom(r)
	if err != nil {
		return 0, err
	}
	if identity.TeamID == nil {
		return 0, domain.ErrForbidden
	}
	return *identity.TeamID, nil
}

func toDriverResponse(driver drivers.Driver) driverResponse {
	return driverResponse{
		ID:            driver.ID,
		FirstName:     driver.FirstName,
		LastName:      driver.LastName,
		LicenseNumber: driver.LicenseNumber,
		CountryCode:   driver.CountryCode,
		IsStarter:     driver.IsStarter,
		Number:        driver.Number,
	}
}

func toPublicDriverResponse(driver drivers.PublicDriver) publicDriverResponse {
	return publicDriverResponse{
		ID:           driver.ID,
		FirstName:    driver.FirstName,
		LastName:     driver.LastName,
		CountryCode:  driver.CountryCode,
		IsStarter:    driver.IsStarter,
		Number:       driver.Number,
		TeamID:       driver.TeamID,
		TeamName:     driver.TeamName,
		CategoryCode: driver.CategoryCode,
		CategoryName: driver.CategoryName,
	}
}
