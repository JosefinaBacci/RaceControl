package sanctions

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// NewSanction is what the FIA submits to issue a sanction.
type NewSanction struct {
	Target Target
	Terms  Terms
}

// Amendment replaces the terms of a sanction; Note explains why.
type Amendment struct {
	Terms Terms
	Note  string
}

// Issue records a sanction and its first version, attributed to the actor. The
// date is the server's: the record says when the FIA issued it.
func (s *Service) Issue(ctx context.Context, actorUserID int64, input NewSanction) (Detail, error) {
	if err := validateTarget(ctx, s.queries, input.Target); err != nil {
		return Detail{}, err
	}
	terms, err := resolveTerms(ctx, s.queries, input.Terms)
	if err != nil {
		return Detail{}, err
	}

	var sanctionID int64
	err = s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		created, err := queries.CreateSanction(ctx, dbgen.CreateSanctionParams{
			DriverID: optionalInt8(input.Target.DriverID),
			TeamID:   optionalInt8(input.Target.TeamID),
		})
		if err != nil {
			return fmt.Errorf("create sanction: %w", err)
		}
		sanctionID = created.ID
		if err := validateEvent(ctx, queries, terms.eventID, created.AccountableTeamID); err != nil {
			return err
		}
		return appendVersion(ctx, queries, versionDraft{sanctionID: created.ID, action: ActionIssued, terms: terms, actorUserID: actorUserID})
	})
	if err != nil {
		return Detail{}, translateWriteError(err)
	}
	return s.Get(ctx, sanctionID)
}

// Amend appends a version with the new terms. Submitting the terms that already
// stand records nothing, so the teams are not asked to confirm a change that did
// not happen.
func (s *Service) Amend(ctx context.Context, actorUserID, sanctionID int64, amendment Amendment) (Detail, error) {
	terms, err := resolveTerms(ctx, s.queries, amendment.Terms)
	if err != nil {
		return Detail{}, err
	}
	note, err := normalizeNote(amendment.Note)
	if err != nil {
		return Detail{}, err
	}

	err = s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		current, err := lockStandingVersion(ctx, queries, sanctionID)
		if err != nil {
			return err
		}
		if err := validateEvent(ctx, queries, terms.eventID, current.AccountableTeamID); err != nil {
			return err
		}
		if terms.equals(termsOf(current)) {
			return nil
		}
		return appendVersion(ctx, queries, versionDraft{
			sanctionID: sanctionID, action: ActionAmended, terms: terms, note: note, actorUserID: actorUserID,
		})
	})
	if err != nil {
		return Detail{}, translateWriteError(err)
	}
	return s.Get(ctx, sanctionID)
}

// Revoke appends the revocation, which keeps the terms it annuls so the record
// still says what was revoked.
func (s *Service) Revoke(ctx context.Context, actorUserID, sanctionID int64, rawNote string) (Detail, error) {
	note, err := normalizeNote(rawNote)
	if err != nil {
		return Detail{}, err
	}

	err = s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		current, err := lockStandingVersion(ctx, queries, sanctionID)
		if err != nil {
			return err
		}
		return appendVersion(ctx, queries, versionDraft{
			sanctionID: sanctionID, action: ActionRevoked, terms: termsOf(current), note: note, actorUserID: actorUserID,
		})
	})
	if err != nil {
		return Detail{}, translateWriteError(err)
	}
	return s.Get(ctx, sanctionID)
}

// lockStandingVersion takes the sanction's lock and answers its current version,
// refusing a sanction that is unknown or already revoked.
func lockStandingVersion(ctx context.Context, queries dbgen.Querier, sanctionID int64) (dbgen.GetCurrentSanctionVersionRow, error) {
	if _, err := queries.LockSanction(ctx, sanctionID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dbgen.GetCurrentSanctionVersionRow{}, domain.ErrNotFound
		}
		return dbgen.GetCurrentSanctionVersionRow{}, fmt.Errorf("lock sanction %d: %w", sanctionID, err)
	}
	current, err := queries.GetCurrentSanctionVersion(ctx, sanctionID)
	if err != nil {
		return dbgen.GetCurrentSanctionVersionRow{}, fmt.Errorf("load current version of sanction %d: %w", sanctionID, err)
	}
	if Action(current.Action) == ActionRevoked {
		return dbgen.GetCurrentSanctionVersionRow{}, domain.NewConflictError("id", "la sanción ya fue revocada")
	}
	return current, nil
}

func termsOf(current dbgen.GetCurrentSanctionVersionRow) resolvedTerms {
	return resolvedTerms{
		penaltyTypeID: current.PenaltyTypeID,
		penaltyValue:  db.Int32Pointer(current.PenaltyValue),
		reason:        current.Reason,
		eventID:       db.Int64Pointer(current.EventID),
	}
}

type versionDraft struct {
	sanctionID  int64
	action      Action
	terms       resolvedTerms
	note        string
	actorUserID int64
}

func appendVersion(ctx context.Context, queries dbgen.Querier, draft versionDraft) error {
	_, err := queries.CreateSanctionVersion(ctx, dbgen.CreateSanctionVersionParams{
		SanctionID:    draft.sanctionID,
		Action:        string(draft.action),
		PenaltyTypeID: draft.terms.penaltyTypeID,
		PenaltyValue:  optionalInt4(draft.terms.penaltyValue),
		Reason:        draft.terms.reason,
		EventID:       optionalInt8(draft.terms.eventID),
		Note:          optionalNote(draft.note),
		ActorUserID:   draft.actorUserID,
	})
	if err != nil {
		return fmt.Errorf("append %s version to sanction %d: %w", draft.action, draft.sanctionID, err)
	}
	return nil
}
