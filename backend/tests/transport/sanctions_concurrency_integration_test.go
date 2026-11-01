package transporttest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"

	"github.com/racecontrol/backend/internal/domain"
)

const concurrentRequests = 8

// statusesOf sends the same request concurrently and answers each status. It
// never fails the test from a goroutine: the caller judges the statuses.
func (h *harness) statusesOf(client *http.Client, method, path string, payload any) []int {
	h.t.Helper()

	encoded, err := json.Marshal(payload)
	if err != nil {
		h.t.Fatalf("marshal: %v", err)
	}
	statuses := make([]int, concurrentRequests)
	var wait sync.WaitGroup
	for index := range statuses {
		wait.Add(1)
		go func() {
			defer wait.Done()
			statuses[index] = sendOnce(client, method, h.server.URL+path, encoded)
		}()
	}
	wait.Wait()
	return statuses
}

func sendOnce(client *http.Client, method, url string, body []byte) int {
	request, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return 0
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return 0
	}
	defer response.Body.Close()
	return response.StatusCode
}

func countStatus(statuses []int, want int) int {
	count := 0
	for _, status := range statuses {
		if status == want {
			count++
		}
	}
	return count
}

func TestConcurrentRevocationsRecordOne(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	created := fixture.issueOnAlonso()

	statuses := h.statusesOf(fixture.fiaAdmin, http.MethodPost, sanctionPath(created, "/revocation"), map[string]any{"note": "Apelación"})
	if countStatus(statuses, http.StatusOK) != 1 || countStatus(statuses, http.StatusConflict) != concurrentRequests-1 {
		t.Fatalf("statuses = %v, want one 200 and the rest 409", statuses)
	}
	detail := h.do(h.browser(), http.MethodGet, sanctionPath(created, ""), nil, nil)
	if versions := listOf(t, detail.body, "versions"); len(versions) != 2 {
		t.Fatalf("versions = %d, want the issue and one revocation", len(versions))
	}
}

func TestConcurrentAmendmentsAreAllKeptInOrder(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	created := fixture.issueOnAlonso()

	statuses := h.statusesOf(fixture.fiaAdmin, http.MethodPut, sanctionPath(created, ""), fixture.amendment("Corrección simultánea"))
	if countStatus(statuses, http.StatusOK) != concurrentRequests {
		t.Fatalf("statuses = %v, want every amendment answered", statuses)
	}
	detail := h.do(h.browser(), http.MethodGet, sanctionPath(created, ""), nil, nil)
	if versions := listOf(t, detail.body, "versions"); len(versions) != 2 {
		t.Fatalf("versions = %d, want one amendment: the others submit the terms that already stand", len(versions))
	}
}

func TestConcurrentAcknowledgementsKeepOneConfirmation(t *testing.T) {
	h := newHarness(t)
	fixture := h.newSanctionFixture()
	teamA, _ := h.signedIn(domain.RoleTeamAdmin, fixture.teamA)
	created := fixture.issueOnAlonso()
	versionID := currentVersionID(t, created.body)

	statuses := h.statusesOf(teamA, http.MethodPost, sanctionAcknowledgementPath(versionID), nil)
	if countStatus(statuses, http.StatusOK) != concurrentRequests {
		t.Fatalf("statuses = %v, want every confirmation answered", statuses)
	}
	var rows int
	err := h.pool.QueryRow(context.Background(),
		`SELECT count(*) FROM sanction_acknowledgements WHERE sanction_version_id = $1 AND team_id = $2`,
		int64(versionID), fixture.teamA).Scan(&rows)
	if err != nil || rows != 1 {
		t.Fatalf("acknowledgements = %d (%v), want exactly one", rows, err)
	}
}
