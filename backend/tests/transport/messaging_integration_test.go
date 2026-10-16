package transporttest

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/racecontrol/backend/internal/domain"
	"github.com/racecontrol/backend/tests/testdb"
)

// messagingFixture is a team of its own with an administrator, so the team side
// of every count is only what the test writes. FIA administrators take part in
// every conversation in the database, so their checks look at one conversation.
type messagingFixture struct {
	team          int64
	otherTeam     int64
	teamAdmin     *http.Client
	teamAdminName string
	fiaAdmin      *http.Client
	fiaAdminName  string
}

func (h *harness) newMessagingFixture() messagingFixture {
	h.t.Helper()

	championship := testdb.NewChampionship(h.t, h.pool)
	fixture := messagingFixture{team: championship.AddTeam("Equipo A"), otherTeam: championship.AddTeam("Equipo B")}
	fixture.teamAdmin, fixture.teamAdminName = h.signedIn(domain.RoleTeamAdmin, fixture.team)
	fixture.fiaAdmin, fixture.fiaAdminName = h.signedIn(domain.RoleFIAAdmin, 0)
	return fixture
}

func (h *harness) startConversation(client *http.Client, payload map[string]any) response {
	h.t.Helper()

	started := h.do(client, http.MethodPost, "/conversations", payload, nil)
	expectStatus(h.t, started, http.StatusCreated)
	return started
}

func (h *harness) mailbox(client *http.Client, box string) response {
	h.t.Helper()

	got := h.do(client, http.MethodGet, "/conversations?box="+box, nil, nil)
	expectStatus(h.t, got, http.StatusOK)
	return got
}

func (h *harness) openConversation(client *http.Client, id float64) []map[string]any {
	h.t.Helper()

	opened := h.do(client, http.MethodGet, conversationPath(id), nil, nil)
	expectStatus(h.t, opened, http.StatusOK)
	return listOf(h.t, opened.body, "messages")
}

func conversationPath(id float64) string {
	return fmt.Sprintf("/conversations/%.0f", id)
}

func conversationWithID(conversations []map[string]any, id float64) map[string]any {
	for _, conversation := range conversations {
		if conversation["id"] == id {
			return conversation
		}
	}
	return nil
}

func expectMessage(t *testing.T, message map[string]any, body string, isOwn, isUnread bool) {
	t.Helper()

	if message["body"] != body || message["isOwn"] != isOwn || message["isUnread"] != isUnread {
		t.Fatalf("message = %v, want %q with isOwn=%t isUnread=%t", message, body, isOwn, isUnread)
	}
}

func TestMessagingRequiresASession(t *testing.T) {
	h := newHarness(t)
	anonymous := h.browser()

	for _, request := range []struct{ method, path string }{
		{http.MethodGet, "/conversations"},
		{http.MethodPost, "/conversations"},
		{http.MethodGet, "/conversations/1"},
		{http.MethodPost, "/conversations/1/messages"},
	} {
		got := h.do(anonymous, request.method, request.path, map[string]any{"subject": "Hola", "body": "Hola"}, nil)
		expectStatus(t, got, http.StatusUnauthorized)
	}
}

func TestTeamAdminAndFIAExchangeMessages(t *testing.T) {
	h := newHarness(t)
	fixture := h.newMessagingFixture()

	started := h.startConversation(fixture.teamAdmin, map[string]any{
		"teamId": fixture.otherTeam, "subject": "  Consulta técnica  ", "body": "¿Cuándo es el control?",
	})
	id := started.body["id"].(float64)
	if team := objectOf(t, started.body, "team"); team["id"] != float64(fixture.team) || started.body["subject"] != "Consulta técnica" {
		t.Fatalf("conversation = %v, want the trimmed subject with the admin's own team", started.body)
	}
	expectMessage(t, listOf(t, started.body, "messages")[0], "¿Cuándo es el control?", true, false)

	summary := conversationWithID(listOf(t, h.mailbox(fixture.fiaAdmin, "inbox").body, "conversations"), id)
	if summary == nil || summary["unreadCount"] != float64(1) {
		t.Fatalf("FIA summary = %v, want the conversation with one unread message", summary)
	}
	expectMessage(t, objectOf(t, summary, "lastMessage"), "¿Cuándo es el control?", false, true)

	expectMessage(t, h.openConversation(fixture.fiaAdmin, id)[0], "¿Cuándo es el control?", false, true)
	expectMessage(t, h.openConversation(fixture.fiaAdmin, id)[0], "¿Cuándo es el control?", false, false)
	if summary := conversationWithID(listOf(t, h.mailbox(fixture.fiaAdmin, "").body, "conversations"), id); summary["unreadCount"] != float64(0) {
		t.Fatalf("FIA summary after opening = %v, want nothing unread", summary)
	}

	reply := h.do(fixture.fiaAdmin, http.MethodPost, conversationPath(id)+"/messages", map[string]any{"body": "El martes."}, nil)
	expectStatus(t, reply, http.StatusCreated)
	if author := objectOf(t, reply.body, "author"); author["username"] != fixture.fiaAdminName || author["role"] != "fia_admin" {
		t.Fatalf("reply = %v, want it written by %s", reply.body, fixture.fiaAdminName)
	}

	if inbox := h.mailbox(fixture.teamAdmin, "inbox"); inbox.body["unreadCount"] != float64(1) {
		t.Fatalf("team unreadCount = %v, want only the FIA reply", inbox.body["unreadCount"])
	}
	messages := h.openConversation(fixture.teamAdmin, id)
	expectMessage(t, messages[0], "¿Cuándo es el control?", true, false)
	expectMessage(t, messages[1], "El martes.", false, true)
	if inbox := h.mailbox(fixture.teamAdmin, "inbox"); inbox.body["unreadCount"] != float64(0) {
		t.Fatalf("team unreadCount after opening = %v, want 0", inbox.body["unreadCount"])
	}
}

func TestEveryTeamAdminTakesPartInTheTeamsConversations(t *testing.T) {
	h := newHarness(t)
	fixture := h.newMessagingFixture()
	id := h.startConversation(fixture.teamAdmin, map[string]any{"subject": "Acreditaciones", "body": "Necesitamos dos pases."}).body["id"].(float64)
	colleague, _ := h.signedIn(domain.RoleTeamAdmin, fixture.team)

	inbox := h.mailbox(colleague, "inbox")
	if inbox.body["unreadCount"] != float64(1) || conversationWithID(listOf(t, inbox.body, "conversations"), id) == nil {
		t.Fatalf("colleague inbox = %v, want the conversation with one unread message", inbox.body)
	}
	expectMessage(t, h.openConversation(colleague, id)[0], "Necesitamos dos pases.", false, true)
}

func TestFIAStartsAConversationWithATeam(t *testing.T) {
	h := newHarness(t)
	fixture := h.newMessagingFixture()

	started := h.startConversation(fixture.fiaAdmin, map[string]any{"teamId": fixture.team, "subject": "Citación", "body": "Presentarse el viernes."})
	id := started.body["id"].(float64)

	inbox := h.mailbox(fixture.teamAdmin, "inbox")
	summary := conversationWithID(listOf(t, inbox.body, "conversations"), id)
	if summary == nil || summary["unreadCount"] != float64(1) || inbox.body["unreadCount"] != float64(1) {
		t.Fatalf("team inbox = %v, want the FIA conversation unread", inbox.body)
	}
	reply := h.do(fixture.teamAdmin, http.MethodPost, conversationPath(id)+"/messages", map[string]any{"body": "Recibido."}, nil)
	expectStatus(t, reply, http.StatusCreated)
	expectMessage(t, reply.body, "Recibido.", true, false)
}

func TestFIAMustPickAnExistingTeam(t *testing.T) {
	h := newHarness(t)
	fixture := h.newMessagingFixture()

	for _, teamID := range []any{nil, unknownTeamID} {
		got := h.do(fixture.fiaAdmin, http.MethodPost, "/conversations", map[string]any{"teamId": teamID, "subject": "Hola", "body": "Hola"}, nil)
		expectStatus(t, got, http.StatusBadRequest)
		if got.body["field"] != "teamId" {
			t.Fatalf("error = %v, want it on teamId", got.body)
		}
	}
}

func TestAnotherTeamCannotReachTheConversation(t *testing.T) {
	h := newHarness(t)
	fixture := h.newMessagingFixture()
	id := h.startConversation(fixture.teamAdmin, map[string]any{"subject": "Privado", "body": "Solo para la FIA."}).body["id"].(float64)
	outsider, _ := h.signedIn(domain.RoleTeamAdmin, fixture.otherTeam)

	for _, box := range []string{"inbox", "sent"} {
		if conversationWithID(listOf(t, h.mailbox(outsider, box).body, "conversations"), id) != nil {
			t.Fatalf("the %s box of another team lists the conversation", box)
		}
	}
	expectStatus(t, h.do(outsider, http.MethodGet, conversationPath(id), nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(outsider, http.MethodPost, conversationPath(id)+"/messages", map[string]any{"body": "Hola"}, nil), http.StatusNotFound)
	expectStatus(t, h.do(outsider, http.MethodGet, "/conversations/999999999", nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(outsider, http.MethodGet, "/conversations/not-a-number", nil, nil), http.StatusNotFound)
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPost, "/conversations/999999999/messages", map[string]any{"body": "Hola"}, nil), http.StatusNotFound)

	if messages := h.openConversation(fixture.teamAdmin, id); len(messages) != 1 {
		t.Fatalf("messages = %v, want the refused reply not to be stored", messages)
	}
}

func TestSentBoxHoldsTheConversationsTheUserWroteIn(t *testing.T) {
	h := newHarness(t)
	fixture := h.newMessagingFixture()
	id := h.startConversation(fixture.teamAdmin, map[string]any{"subject": "Neumáticos", "body": "¿Hay prueba?"}).body["id"].(float64)

	if conversationWithID(listOf(t, h.mailbox(fixture.teamAdmin, "sent").body, "conversations"), id) == nil {
		t.Fatal("the author's sent box does not list the conversation they started")
	}
	if conversationWithID(listOf(t, h.mailbox(fixture.fiaAdmin, "sent").body, "conversations"), id) != nil {
		t.Fatal("the FIA sent box lists a conversation the FIA administrator never wrote in")
	}
	expectStatus(t, h.do(fixture.fiaAdmin, http.MethodPost, conversationPath(id)+"/messages", map[string]any{"body": "Sí."}, nil), http.StatusCreated)
	if conversationWithID(listOf(t, h.mailbox(fixture.fiaAdmin, "sent").body, "conversations"), id) == nil {
		t.Fatal("the FIA sent box does not list the conversation after replying")
	}
}

func TestMessagingValidatesItsInput(t *testing.T) {
	h := newHarness(t)
	fixture := h.newMessagingFixture()
	id := h.startConversation(fixture.teamAdmin, map[string]any{"subject": strings.Repeat("á", 120), "body": strings.Repeat("é", 4000)}).body["id"].(float64)

	for _, invalid := range []struct {
		name    string
		payload map[string]any
		field   string
	}{
		{"blank subject", map[string]any{"subject": "   ", "body": "Hola"}, "subject"},
		{"long subject", map[string]any{"subject": strings.Repeat("a", 121), "body": "Hola"}, "subject"},
		{"blank body", map[string]any{"subject": "Hola", "body": "\n\t "}, "body"},
		{"long body", map[string]any{"subject": "Hola", "body": strings.Repeat("a", 4001)}, "body"},
		{"NUL in body", map[string]any{"subject": "Hola", "body": "a\u0000b"}, "body"},
		{"unknown field", map[string]any{"subject": "Hola", "body": "Hola", "html": true}, "body"},
	} {
		got := h.do(fixture.teamAdmin, http.MethodPost, "/conversations", invalid.payload, nil)
		expectStatus(t, got, http.StatusBadRequest)
		if got.body["field"] != invalid.field {
			t.Fatalf("%s: error = %v, want it on %s", invalid.name, got.body, invalid.field)
		}
	}

	blankReply := h.do(fixture.teamAdmin, http.MethodPost, conversationPath(id)+"/messages", map[string]any{"body": " "}, nil)
	expectStatus(t, blankReply, http.StatusBadRequest)
	expectStatus(t, h.do(fixture.teamAdmin, http.MethodGet, "/conversations?box=trash", nil, nil), http.StatusBadRequest)
}

func TestMessagingIsServedWithStrictTransportSecurity(t *testing.T) {
	h := newHarness(t)
	fixture := h.newMessagingFixture()

	request, err := http.NewRequest(http.MethodGet, h.server.URL+"/conversations", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	result, err := fixture.teamAdmin.Do(request)
	if err != nil {
		t.Fatalf("GET /conversations: %v", err)
	}
	defer result.Body.Close()
	if hsts := result.Header.Get("Strict-Transport-Security"); !strings.HasPrefix(hsts, "max-age=") {
		t.Fatalf("Strict-Transport-Security = %q, want HSTS on messaging", hsts)
	}
}
