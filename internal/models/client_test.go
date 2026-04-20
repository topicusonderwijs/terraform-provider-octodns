package models

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

type savedCommit struct {
	zone    string
	comment string
}

// newBatchingTestClient returns a GitHubClient wired up for batching tests:
// a default scope is registered, SaveZoneFn records every commit into the
// returned slice under the shared mutex, and BatchWindow is short so tests
// stay fast.
func newBatchingTestClient(t *testing.T) (*GitHubClient, *[]savedCommit, *sync.Mutex) {
	t.Helper()

	var commitsMu sync.Mutex
	commits := []savedCommit{}

	client := &GitHubClient{
		Scopes:        map[string]Scope{},
		Zones:         map[string]*Zone{},
		dirtyZones:    map[string]*Zone{},
		dirtyComments: map[string][]string{},
		BatchWindow:   5 * time.Millisecond,
	}
	client.SaveZoneFn = func(z *Zone, c string) error {
		commitsMu.Lock()
		defer commitsMu.Unlock()
		commits = append(commits, savedCommit{zone: z.name, comment: c})
		return nil
	}

	if err := client.AddScope("default", "zones", "main", "yaml"); err != nil {
		t.Fatalf("AddScope failed: %s", err)
	}

	return client, &commits, &commitsMu
}

// runOperation simulates the caller-side pattern used by Create/Update/Delete
// in record_resource.go: bump InFlight before taking the lock, mark dirty,
// then FlushIfLast which owns the matching decrement.
func runOperation(client *GitHubClient, zoneName, comment string) error {
	client.InFlight.Add(1)
	client.Mutex.Lock()
	defer client.Mutex.Unlock()

	zone := &Zone{name: zoneName, scope: "default"}
	client.MarkZoneDirty(zone, comment)
	return client.FlushIfLast()
}

func TestBatching_SingleOperation(t *testing.T) {
	client, commits, mu := newBatchingTestClient(t)

	if err := runOperation(client, "example.com", "chore(default/example.com): create A record for www"); err != nil {
		t.Fatalf("runOperation failed: %s", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(*commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(*commits))
	}
	if (*commits)[0].comment != "chore(default/example.com): create A record for www" {
		t.Errorf("unexpected comment: %q", (*commits)[0].comment)
	}
}

func TestBatching_MultipleOperationsSameZone(t *testing.T) {
	client, commits, mu := newBatchingTestClient(t)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			comment := fmt.Sprintf("chore(default/example.com): update A record for www%d", i)
			if err := runOperation(client, "example.com", comment); err != nil {
				t.Errorf("runOperation failed: %s", err)
			}
		}(i)
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(*commits) != 1 {
		t.Fatalf("expected 1 batched commit, got %d", len(*commits))
	}
	want := "chore(default/example.com): 5 changes (5 updates)"
	if (*commits)[0].comment != want {
		t.Errorf("expected summary %q, got %q", want, (*commits)[0].comment)
	}
}

func TestBatching_MixedActionsSummary(t *testing.T) {
	client, commits, mu := newBatchingTestClient(t)

	ops := []string{
		"chore(default/example.com): create A record for a1",
		"chore(default/example.com): create A record for a2",
		"chore(default/example.com): create A record for a3",
		"chore(default/example.com): update A record for b1",
		"chore(default/example.com): update A record for b2",
		"chore(default/example.com): delete A record for c1",
	}

	var wg sync.WaitGroup
	for _, comment := range ops {
		wg.Add(1)
		go func(c string) {
			defer wg.Done()
			if err := runOperation(client, "example.com", c); err != nil {
				t.Errorf("runOperation failed: %s", err)
			}
		}(comment)
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(*commits) != 1 {
		t.Fatalf("expected 1 batched commit, got %d", len(*commits))
	}
	want := "chore(default/example.com): 6 changes (3 creates, 2 updates, 1 deletes)"
	if (*commits)[0].comment != want {
		t.Errorf("expected summary %q, got %q", want, (*commits)[0].comment)
	}
}

func TestBatching_MultipleZones(t *testing.T) {
	client, commits, mu := newBatchingTestClient(t)

	zones := []string{"zone-a.com", "zone-b.com", "zone-a.com", "zone-b.com"}

	var wg sync.WaitGroup
	for i, z := range zones {
		wg.Add(1)
		go func(i int, zoneName string) {
			defer wg.Done()
			comment := fmt.Sprintf("chore(default/%s): create A record for r%d", zoneName, i)
			if err := runOperation(client, zoneName, comment); err != nil {
				t.Errorf("runOperation failed: %s", err)
			}
		}(i, z)
	}
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if len(*commits) != 2 {
		t.Fatalf("expected 2 commits (one per zone), got %d", len(*commits))
	}
	seen := map[string]bool{}
	for _, c := range *commits {
		seen[c.zone] = true
	}
	if !seen["zone-a.com"] || !seen["zone-b.com"] {
		t.Errorf("expected commits for both zones, got: %+v", *commits)
	}
}

func TestFlushIfLast_NotLastNoFlush(t *testing.T) {
	client, commits, mu := newBatchingTestClient(t)

	// Three goroutines queue up (InFlight == 3), but only two reach FlushIfLast.
	// FlushIfLast decrements: 3 -> 2 (no flush), 2 -> 1 (no flush). Dirty zone
	// remains queued for the third goroutine (which this test doesn't run).
	client.InFlight.Add(3)
	client.Mutex.Lock()
	client.MarkZoneDirty(&Zone{name: "example.com", scope: "default"}, "chore(default/example.com): create A record for x")
	if err := client.FlushIfLast(); err != nil {
		t.Fatalf("FlushIfLast failed: %s", err)
	}
	client.Mutex.Unlock()

	client.Mutex.Lock()
	client.MarkZoneDirty(&Zone{name: "example.com", scope: "default"}, "chore(default/example.com): create A record for y")
	if err := client.FlushIfLast(); err != nil {
		t.Fatalf("FlushIfLast failed: %s", err)
	}
	client.Mutex.Unlock()

	mu.Lock()
	defer mu.Unlock()
	if len(*commits) != 0 {
		t.Fatalf("expected 0 commits while InFlight > 0, got %d", len(*commits))
	}
	if len(client.dirtyZones) == 0 {
		t.Errorf("expected dirty zones to remain queued")
	}
}

func TestFlushIfLast_EmptyDirtyNoOp(t *testing.T) {
	client, commits, mu := newBatchingTestClient(t)

	client.InFlight.Add(1)
	client.Mutex.Lock()
	// No MarkZoneDirty — dirty map stays empty.
	if err := client.FlushIfLast(); err != nil {
		t.Fatalf("FlushIfLast failed: %s", err)
	}
	client.Mutex.Unlock()

	mu.Lock()
	defer mu.Unlock()
	if len(*commits) != 0 {
		t.Fatalf("expected 0 commits when nothing dirty, got %d", len(*commits))
	}
}
