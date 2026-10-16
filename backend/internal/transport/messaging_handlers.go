package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/internal/messaging"
)

type Messenger interface {
	Conversations(ctx context.Context, reader domain.Identity, box messaging.Box) (messaging.Mailbox, error)
	Start(ctx context.Context, author domain.Identity, input messaging.NewConversation) (messaging.Conversation, error)
	Open(ctx context.Context, reader domain.Identity, conversationID int64) (messaging.Conversation, error)
	Reply(ctx context.Context, author domain.Identity, conversationID int64, body string) (messaging.Message, error)
}

type messagingHandlers struct {
	messenger Messenger
}

type conversationTeamResponse struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type messageAuthorResponse struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type messageResponse struct {
	ID       int64                 `json:"id"`
	Body     string                `json:"body"`
	SentAt   time.Time             `json:"sentAt"`
	Author   messageAuthorResponse `json:"author"`
	IsOwn    bool                  `json:"isOwn"`
	IsUnread bool                  `json:"isUnread"`
}

type conversationResponse struct {
	ID        int64                    `json:"id"`
	Subject   string                   `json:"subject"`
	Team      conversationTeamResponse `json:"team"`
	CreatedAt time.Time                `json:"createdAt"`
	Messages  []messageResponse        `json:"messages"`
}

type conversationSummaryResponse struct {
	ID          int64                    `json:"id"`
	Subject     string                   `json:"subject"`
	Team        conversationTeamResponse `json:"team"`
	CreatedAt   time.Time                `json:"createdAt"`
	LastMessage messageResponse          `json:"lastMessage"`
	UnreadCount int                      `json:"unreadCount"`
}

type mailboxResponse struct {
	UnreadCount   int                           `json:"unreadCount"`
	Conversations []conversationSummaryResponse `json:"conversations"`
}

type startConversationRequest struct {
	TeamID  *int64 `json:"teamId"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type replyRequest struct {
	Body string `json:"body"`
}

func (h *messagingHandlers) list(w http.ResponseWriter, r *http.Request) {
	identity, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	mailbox, err := h.messenger.Conversations(r.Context(), identity, messaging.Box(r.URL.Query().Get("box")))
	if err != nil {
		writeError(w, r, err)
		return
	}
	response := mailboxResponse{
		UnreadCount:   mailbox.UnreadCount,
		Conversations: make([]conversationSummaryResponse, 0, len(mailbox.Conversations)),
	}
	for _, summary := range mailbox.Conversations {
		response.Conversations = append(response.Conversations, conversationSummaryResponse{
			ID:          summary.ID,
			Subject:     summary.Subject,
			Team:        conversationTeamResponse(summary.Team),
			CreatedAt:   summary.CreatedAt,
			LastMessage: toMessageResponse(summary.LastMessage),
			UnreadCount: summary.UnreadCount,
		})
	}
	writeJSON(w, http.StatusOK, response)
}

// start ignores teamId for a team administrator: their conversations are always
// with their own team.
func (h *messagingHandlers) start(w http.ResponseWriter, r *http.Request) {
	identity, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var request startConversationRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	conversation, err := h.messenger.Start(r.Context(), identity, messaging.NewConversation(request))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toConversationResponse(conversation))
}

// open answers each message flagged unread as it was before this read, and marks
// the conversation read.
func (h *messagingHandlers) open(w http.ResponseWriter, r *http.Request) {
	identity, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	conversation, err := h.messenger.Open(r.Context(), identity, id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toConversationResponse(conversation))
}

func (h *messagingHandlers) reply(w http.ResponseWriter, r *http.Request) {
	identity, err := identityFrom(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	id, err := pathID(r)
	if err != nil {
		writeError(w, r, err)
		return
	}
	var request replyRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, r, err)
		return
	}
	message, err := h.messenger.Reply(r.Context(), identity, id, request.Body)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, toMessageResponse(message))
}

func toConversationResponse(conversation messaging.Conversation) conversationResponse {
	messages := make([]messageResponse, 0, len(conversation.Messages))
	for _, message := range conversation.Messages {
		messages = append(messages, toMessageResponse(message))
	}
	return conversationResponse{
		ID:        conversation.ID,
		Subject:   conversation.Subject,
		Team:      conversationTeamResponse(conversation.Team),
		CreatedAt: conversation.CreatedAt,
		Messages:  messages,
	}
}

func toMessageResponse(message messaging.Message) messageResponse {
	return messageResponse{
		ID:       message.ID,
		Body:     message.Body,
		SentAt:   message.SentAt,
		Author:   messageAuthorResponse{Username: message.Author.Username, Role: string(message.Author.Role)},
		IsOwn:    message.IsOwn,
		IsUnread: message.IsUnread,
	}
}
