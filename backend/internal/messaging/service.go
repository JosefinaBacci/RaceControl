package messaging

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/racecontrol/backend/internal/db"
	"github.com/racecontrol/backend/internal/db/dbgen"
	"github.com/racecontrol/backend/internal/domain"
)

// Transactor opens a conversation and writes its first message in one transaction.
type Transactor interface {
	InTx(ctx context.Context, work func(queries dbgen.Querier) error) error
}

// Participants and unread counts are derived for institutional conversations.
type Service struct {
	queries    dbgen.Querier
	transactor Transactor
}

func NewService(queries dbgen.Querier, transactor Transactor) *Service {
	return &Service{queries: queries, transactor: transactor}
}

type Team struct {
	ID   int64
	Code string
	Name string
}

type Author struct {
	Username string
	Role     domain.Role
}

// Message is a message as one reader sees it: IsOwn when they wrote it, IsUnread
// when someone else did and the reader had not opened it yet.
type Message struct {
	ID                int64
	Body              string
	SentAt            time.Time
	Author            Author
	IsOwn             bool
	IsUnread          bool
	IsReadByRecipient bool
}

type Conversation struct {
	ID            int64
	Subject       string
	Team          Team
	OtherTeamID   *int64
	OtherTeamName string
	CreatedAt     time.Time
	Messages      []Message
}

type ConversationSummary struct {
	ID             int64
	Subject        string
	Team           Team
	OtherTeamID    *int64
	OtherTeamName  string
	CreatedAt      time.Time
	LastMessage    Message
	UnreadCount    int
	HasOwnMessages bool
}

// Mailbox is one box of the reader's conversations, newest activity first.
// UnreadCount covers every conversation, whichever box is shown, because it is
// what the badge displays.
type Mailbox struct {
	Conversations []ConversationSummary
	UnreadCount   int
}

// TeamID is the FIA recipient. RecipientTeamID selects another team for a team
// sender; omitting it keeps the existing team-to-FIA contract.
type NewConversation struct {
	TeamID          *int64
	RecipientTeamID *int64
	Subject         string
	Body            string
}

const participantGuard = "RC007"

// Conversations lists the box of the reader's conversations.
func (s *Service) Conversations(ctx context.Context, reader domain.Identity, box Box) (Mailbox, error) {
	box, err := box.resolve()
	if err != nil {
		return Mailbox{}, err
	}
	_, err = teamScope(reader)
	if err != nil {
		return Mailbox{}, err
	}
	rows, err := s.queries.ListConversations(ctx, reader.UserID)
	if err != nil {
		return Mailbox{}, fmt.Errorf("list conversations of user %d: %w", reader.UserID, err)
	}

	mailbox := Mailbox{Conversations: make([]ConversationSummary, 0, len(rows))}
	for _, row := range rows {
		summary := summaryFromRow(row, reader.UserID)
		mailbox.UnreadCount += summary.UnreadCount
		if box.includes(summary) {
			mailbox.Conversations = append(mailbox.Conversations, summary)
		}
	}
	return mailbox, nil
}

// Start opens a conversation with its first message.
func (s *Service) Start(ctx context.Context, author domain.Identity, input NewConversation) (Conversation, error) {
	subject, err := subjectRule.normalize(input.Subject)
	if err != nil {
		return Conversation{}, err
	}
	body, err := bodyRule.normalize(input.Body)
	if err != nil {
		return Conversation{}, err
	}
	teamID, err := s.conversationTeam(ctx, author, input.TeamID)
	if err != nil {
		return Conversation{}, err
	}

	otherTeam, err := s.recipientTeam(ctx, author, input.RecipientTeamID)
	if err != nil {
		return Conversation{}, err
	}
	var conversationID int64
	err = s.transactor.InTx(ctx, func(queries dbgen.Querier) error {
		id, createErr := queries.CreateConversation(ctx, dbgen.CreateConversationParams{
			TeamID: teamID, OtherTeamID: otherTeam, Subject: subject, CreatedBy: author.UserID,
		})
		if createErr != nil {
			return createErr
		}
		conversationID = id
		_, addErr := queries.AddMessage(ctx, dbgen.AddMessageParams{ConversationID: id, AuthorID: author.UserID, Body: body})
		return addErr
	})
	if err != nil {
		return Conversation{}, translateWriteError(err, fmt.Sprintf("start conversation with team %d", teamID))
	}
	return s.conversation(ctx, author, conversationID)
}

// Open answers the conversation with what was unread until now and moves the
// reader's cursor to its last message.
func (s *Service) Open(ctx context.Context, reader domain.Identity, conversationID int64) (Conversation, error) {
	conversation, err := s.conversation(ctx, reader, conversationID)
	if err != nil {
		return Conversation{}, err
	}
	if len(conversation.Messages) == 0 {
		return conversation, nil
	}
	err = s.queries.MarkConversationRead(ctx, dbgen.MarkConversationReadParams{
		ConversationID:    conversationID,
		UserID:            reader.UserID,
		LastReadMessageID: conversation.Messages[len(conversation.Messages)-1].ID,
	})
	if err != nil {
		return Conversation{}, fmt.Errorf("mark conversation %d read for user %d: %w", conversationID, reader.UserID, err)
	}
	return conversation, nil
}

// Reply adds a message to a conversation the author takes part in.
func (s *Service) Reply(ctx context.Context, author domain.Identity, conversationID int64, rawBody string) (Message, error) {
	body, err := bodyRule.normalize(rawBody)
	if err != nil {
		return Message{}, err
	}
	if _, err := s.header(ctx, author, conversationID); err != nil {
		return Message{}, err
	}
	row, err := s.queries.AddMessage(ctx, dbgen.AddMessageParams{ConversationID: conversationID, AuthorID: author.UserID, Body: body})
	if err != nil {
		return Message{}, translateWriteError(err, fmt.Sprintf("reply to conversation %d", conversationID))
	}
	return Message{
		ID:     row.ID,
		Body:   body,
		SentAt: row.CreatedAt.Time,
		Author: Author{Username: author.Username, Role: author.Role},
		IsOwn:  true,
	}, nil
}

// teamScope validates the role and derives the sender team. SQL independently
// checks participation for every conversation read (000029).
func teamScope(identity domain.Identity) (pgtype.Int8, error) {
	switch {
	case identity.Role == domain.RoleFIAAdmin:
		return pgtype.Int8{}, nil
	case identity.Role == domain.RoleTeamAdmin && identity.TeamID != nil:
		return db.Int8(*identity.TeamID), nil
	}
	return pgtype.Int8{}, domain.ErrForbidden
}

// conversationTeam is the team a new conversation is with. A team administrator's
// is their own, whatever team they sent, so they can never reach another team.
func (s *Service) conversationTeam(ctx context.Context, author domain.Identity, requested *int64) (int64, error) {
	scope, err := teamScope(author)
	if err != nil {
		return 0, err
	}
	if scope.Valid {
		return scope.Int64, nil
	}
	if requested == nil {
		return 0, domain.NewValidationError("teamId", "elegí la escudería")
	}
	exists, err := s.queries.TeamExists(ctx, *requested)
	if err != nil {
		return 0, fmt.Errorf("check team %d exists: %w", *requested, err)
	}
	if !exists {
		return 0, domain.NewValidationError("teamId", "la escudería no existe")
	}
	return *requested, nil
}

func (s *Service) conversation(ctx context.Context, reader domain.Identity, conversationID int64) (Conversation, error) {
	header, err := s.header(ctx, reader, conversationID)
	if err != nil {
		return Conversation{}, err
	}
	rows, err := s.queries.ListConversationMessages(ctx, dbgen.ListConversationMessagesParams{
		ConversationID: conversationID, ReaderID: reader.UserID,
	})
	if err != nil {
		return Conversation{}, fmt.Errorf("list messages of conversation %d: %w", conversationID, err)
	}

	conversation := Conversation{
		ID:          header.ID,
		Subject:     header.Subject,
		Team:        Team{ID: header.TeamID, Code: header.TeamCode, Name: header.TeamName},
		OtherTeamID: db.Int64Pointer(header.OtherTeamID), OtherTeamName: header.OtherTeamName,
		CreatedAt: header.CreatedAt.Time,
		Messages:  make([]Message, 0, len(rows)),
	}
	for _, row := range rows {
		conversation.Messages = append(conversation.Messages, Message{
			ID:                row.ID,
			Body:              row.Body,
			SentAt:            row.CreatedAt.Time,
			Author:            Author{Username: row.AuthorUsername, Role: domain.Role(row.AuthorRole)},
			IsOwn:             row.AuthorID == reader.UserID,
			IsUnread:          row.IsUnread,
			IsReadByRecipient: row.IsReadByRecipient,
		})
	}
	return conversation, nil
}

// header finds the conversation among the reader's. Another team's conversation
// is answered as not found, the same as a missing one, so its existence never
// leaks.
func (s *Service) header(ctx context.Context, reader domain.Identity, conversationID int64) (dbgen.GetConversationRow, error) {
	_, err := teamScope(reader)
	if err != nil {
		return dbgen.GetConversationRow{}, err
	}
	header, err := s.queries.GetConversation(ctx, dbgen.GetConversationParams{ConversationID: conversationID, ReaderID: reader.UserID})
	if errors.Is(err, pgx.ErrNoRows) {
		return dbgen.GetConversationRow{}, domain.ErrNotFound
	}
	if err != nil {
		return dbgen.GetConversationRow{}, fmt.Errorf("get conversation %d: %w", conversationID, err)
	}
	return header, nil
}

func summaryFromRow(row dbgen.ListConversationsRow, readerID int64) ConversationSummary {
	isOwn := row.LastMessageAuthorID == readerID
	return ConversationSummary{
		ID:          row.ID,
		Subject:     row.Subject,
		Team:        Team{ID: row.TeamID, Code: row.TeamCode, Name: row.TeamName},
		OtherTeamID: db.Int64Pointer(row.OtherTeamID), OtherTeamName: row.OtherTeamName,
		CreatedAt: row.CreatedAt.Time,
		LastMessage: Message{
			ID:     row.LastMessageID,
			Body:   row.LastMessageBody,
			SentAt: row.LastMessageCreatedAt.Time,
			Author: Author{Username: row.LastMessageAuthorUsername, Role: domain.Role(row.LastMessageAuthorRole)},
			IsOwn:  isOwn,
			// The newest message is past every cursor an unread message is past,
			// so it is unread exactly when someone else wrote it and anything is.
			IsUnread: !isOwn && row.UnreadCount > 0,
		},
		UnreadCount:    int(row.UnreadCount),
		HasOwnMessages: row.HasOwnMessages,
	}
}

// translateWriteError answers a missing conversation (no row to lock) and a
// non-participant refused by RC007 as not found, like a read would.
func translateWriteError(err error, action string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if code, isGuard := db.GuardViolation(err); isGuard && code == participantGuard {
		return domain.ErrNotFound
	}
	return fmt.Errorf("%s: %w", action, err)
}

func (s *Service) recipientTeam(ctx context.Context, author domain.Identity, requested *int64) (pgtype.Int8, error) {
	if requested == nil {
		return pgtype.Int8{}, nil
	}
	if author.Role != domain.RoleTeamAdmin || author.TeamID == nil || *requested == *author.TeamID {
		return pgtype.Int8{}, domain.NewValidationError("recipientTeamId", "elegí otra escudería o la FIA")
	}
	exists, err := s.queries.TeamExists(ctx, *requested)
	if err != nil {
		return pgtype.Int8{}, fmt.Errorf("check recipient team: %w", err)
	}
	if !exists {
		return pgtype.Int8{}, domain.NewValidationError("recipientTeamId", "la escudería no existe")
	}
	return db.Int8(*requested), nil
}
