package transporttest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/racecontrol/backend/internal/domain"
)

// sendRaw posts the body exactly as written, so a test can send what a JSON
// encoder never produces.
func (h *harness) sendRaw(client *http.Client, path, body string) response {
	h.t.Helper()

	result, err := client.Post(h.server.URL+path, "application/json", strings.NewReader(body))
	if err != nil {
		h.t.Fatalf("POST %s: %v", path, err)
	}
	defer result.Body.Close()
	raw, err := io.ReadAll(result.Body)
	if err != nil {
		h.t.Fatalf("read body: %v", err)
	}
	decoded := map[string]any{}
	if err := json.Unmarshal(raw, &decoded); err != nil {
		h.t.Fatalf("decode body %q: %v", raw, err)
	}
	return response{status: result.StatusCode, body: decoded, raw: string(raw)}
}

func TestAWrongJSONTypeIsReportedOnItsField(t *testing.T) {
	f := newControlFixture(t)
	path := fmt.Sprintf("/technical-controls/%d/results", f.eventID)

	cases := map[string]string{
		"passed": fmt.Sprintf(`{"teamId":%d,"conductedOn":"%s","passed":"true"}`, f.teamA, controlDay),
		"teamId": fmt.Sprintf(`{"teamId":"%d","conductedOn":"%s","passed":true}`, f.teamA, controlDay),
		"notes":  fmt.Sprintf(`{"teamId":%d,"conductedOn":"%s","passed":true,"notes":7}`, f.teamA, controlDay),
	}
	for field, body := range cases {
		got := f.h.sendRaw(f.fia, path, body)
		expectField(t, got, field)
		if !strings.Contains(got.body["message"].(string), "tipo") {
			t.Errorf("%s: message = %q, want it to name the wrong type", field, got.body["message"])
		}
	}
	expectField(t, f.h.sendRaw(&http.Client{}, "/auth/login", `{"username":5,"password":"x"}`), "username")
}

func TestABodyWithMoreThanOneJSONValueIsRefused(t *testing.T) {
	f := newControlFixture(t)
	path := fmt.Sprintf("/technical-controls/%d/results", f.eventID)
	valid := fmt.Sprintf(`{"teamId":%d,"conductedOn":"%s","passed":true}`, f.teamA, controlDay)

	for _, trailer := range []string{"{}", `{"teamId":1}`, "x", "[]"} {
		expectField(t, f.h.sendRaw(f.fia, path, valid+trailer), "body")
	}
	if len(f.history("")) != 0 {
		t.Fatal("a body with trailing data was stored")
	}
	expectStatus(t, f.h.sendRaw(f.fia, path, valid+"\n"), http.StatusCreated)
}

func TestLoginRefusesTrailingData(t *testing.T) {
	h := newHarness(t)
	username := h.createUser(domain.RoleFIAAdmin, 0)
	body := fmt.Sprintf(`{"username":%q,"password":%q}{}`, username, testPassword)

	expectField(t, h.sendRaw(&http.Client{}, "/auth/login", body), "body")
}
