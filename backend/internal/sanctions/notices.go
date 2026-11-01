package sanctions

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// AcknowledgementStatus is pending until the team confirms it received the
// version. It mirrors the score notices of US24.
type AcknowledgementStatus string

const (
	AcknowledgementPending   AcknowledgementStatus = "pending"
	AcknowledgementConfirmed AcknowledgementStatus = "confirmed"
)

// Acknowledgement is who confirmed a version for a team, and when. Both are nil
// while it is pending.
type Acknowledgement struct {
	Status AcknowledgementStatus
	At     *time.Time
	By     *string
}

// Notice is a sanction as a team it was sent to sees it: the version that
// stands, whether the sanction is on the team itself, and its confirmation.
type Notice struct {
	Sanction        Sanction
	IsOwnTeam       bool
	Acknowledgement Acknowledgement
}

// NoticeBoard is a team's notices, latest change first, with how many await
// confirmation.
type NoticeBoard struct {
	Notices      []Notice
	PendingCount int
}

// TeamAcknowledgement is where one recipient team stands on a version.
type TeamAcknowledgement struct {
	Team            Team
	Acknowledgement Acknowledgement
}

// VersionFollowUp is a version as the FIA follows it: who wrote it, and every
// team it was sent to with its confirmation.
type VersionFollowUp struct {
	Version   Version
	ChangedBy string
	IsCurrent bool
	Teams     []TeamAcknowledgement
}

// FollowUp is a sanction with its versions, latest first.
type FollowUp struct {
	Sanction Sanction
	Versions []VersionFollowUp
}

// TeamNotices lists the sanctions sent to the team, as they stand.
func (s *Service) TeamNotices(ctx context.Context, teamID int64) (NoticeBoard, error) {
	notices, err := s.teamNotices(ctx, teamID, pgtype.Int8{})
	if err != nil {
		return NoticeBoard{}, err
	}
	return NoticeBoard{Notices: notices, PendingCount: countPending(notices)}, nil
}

// Acknowledge records that the caller's team received the version. The database
// takes the team from the user and refuses a version the team was not sent or
// that a later one replaced (000036); confirming again keeps the first
// confirmation.
func (s *Service) Acknowledge(ctx context.Context, identity domain.Identity, versionID int64) (Notice, error) {
	if identity.TeamID == nil {
		return Notice{}, domain.ErrForbidden
	}
	err := s.queries.AcknowledgeSanctionVersion(ctx, dbgen.AcknowledgeSanctionVersionParams{
		VersionID:      versionID,
		AcknowledgedBy: identity.UserID,
	})
	if err != nil {
		return Notice{}, translateAcknowledgeError(err, versionID)
	}
	return s.teamNotice(ctx, *identity.TeamID, versionID)
}

// translateAcknowledgeError answers a version the team was not sent as not
// found, so a team cannot probe another championship's versions.
func translateAcknowledgeError(err error, versionID int64) error {
	code, isGuard := db.GuardViolation(err)
	switch {
	case isGuard && code == guardNotRecipient:
		return domain.ErrNotFound
	case isGuard && code == guardVersionStale:
		return domain.NewConflictError("id", "la sanción cambió después de este aviso: confirmá la versión vigente")
	}
	return fmt.Errorf("acknowledge sanction version %d: %w", versionID, err)
}

func (s *Service) teamNotice(ctx context.Context, teamID, versionID int64) (Notice, error) {
	notices, err := s.teamNotices(ctx, teamID, db.Int8(versionID))
	if err != nil {
		return Notice{}, err
	}
	if len(notices) == 0 {
		return Notice{}, domain.ErrNotFound
	}
	return notices[0], nil
}

func (s *Service) teamNotices(ctx context.Context, teamID int64, versionID pgtype.Int8) ([]Notice, error) {
	rows, err := s.queries.ListTeamSanctionNotices(ctx, dbgen.ListTeamSanctionNoticesParams{TeamID: teamID, VersionID: versionID})
	if err != nil {
		return nil, fmt.Errorf("list sanction notices of team %d: %w", teamID, err)
	}
	notices := make([]Notice, 0, len(rows))
	for _, row := range rows {
		notices = append(notices, Notice{
			Sanction:        sanctionFromRow(row.SanctionVersionDetail),
			IsOwnTeam:       row.SanctionVersionDetail.AccountableTeamID == teamID,
			Acknowledgement: acknowledgementOf(row.AcknowledgedAt, row.AcknowledgedBy),
		})
	}
	return notices, nil
}

// FollowUps lists every sanction, newest issue first, with each version and the
// confirmation of every team it was sent to.
func (s *Service) FollowUps(ctx context.Context) ([]FollowUp, error) {
	rows, err := s.queries.ListSanctionAcknowledgements(ctx)
	if err != nil {
		return nil, fmt.Errorf("list sanction acknowledgements: %w", err)
	}
	followUps := make([]FollowUp, 0)
	for _, row := range rows {
		followUps = appendFollowUpRow(followUps, row)
	}
	return followUps, nil
}

// appendFollowUpRow folds one recipient row into its sanction and version. The
// rows come grouped by sanction and version, latest version first, so the first
// row of a sanction carries the version that stands.
func appendFollowUpRow(followUps []FollowUp, row dbgen.ListSanctionAcknowledgementsRow) []FollowUp {
	details := row.SanctionVersionDetail
	if len(followUps) == 0 || followUps[len(followUps)-1].Sanction.ID != details.SanctionID {
		followUps = append(followUps, FollowUp{Sanction: sanctionFromRow(details)})
	}
	followUp := &followUps[len(followUps)-1]
	if len(followUp.Versions) == 0 || followUp.Versions[len(followUp.Versions)-1].Version.ID != details.VersionID {
		followUp.Versions = append(followUp.Versions, VersionFollowUp{
			Version:   versionFromRow(details),
			ChangedBy: row.ChangedBy,
			IsCurrent: details.IsCurrent,
		})
	}
	version := &followUp.Versions[len(followUp.Versions)-1]
	version.Teams = append(version.Teams, TeamAcknowledgement{
		Team:            Team{ID: row.RecipientTeamID, Code: row.RecipientTeamCode, Name: row.RecipientTeamName},
		Acknowledgement: acknowledgementOf(row.AcknowledgedAt, row.AcknowledgedBy),
	})
	return followUps
}

func acknowledgementOf(at pgtype.Timestamptz, by pgtype.Text) Acknowledgement {
	if !at.Valid {
		return Acknowledgement{Status: AcknowledgementPending}
	}
	return Acknowledgement{Status: AcknowledgementConfirmed, At: db.TimePointer(at), By: db.TextPointer(by)}
}

func countPending(notices []Notice) int {
	pending := 0
	for _, notice := range notices {
		if notice.Acknowledgement.Status == AcknowledgementPending {
			pending++
		}
	}
	return pending
}
