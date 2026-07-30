package fsm

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"testing"

	"github.com/markeetakeawe2/consul/agent/consul/state"
)

func TestFSM_RestoreRace(t *testing.T) {
	store := state.NewStore()
	fsm := NewFSM(store)

	fsm.Apply(1, "register", "srv-1")

	snapData := struct {
		Index    uint64
		Services map[string]bool
	}{
		Index:    1,
		Services: store.Snapshot(),
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&snapData); err != nil {
		t.Fatalf("failed to encode snapshot: %v", err)
	}
	snapshotBytes := buf.Bytes()

	fsm.Apply(2, "deregister", "srv-1")
	fsm.Apply(3, "register", "srv-2")
	fsm.Apply(4, "deregister", "srv-2")
	fsm.Apply(5, "register", "srv-3")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		rc := io.NopCloser(bytes.NewReader(snapshotBytes))
		if err := fsm.Restore(rc); err != nil {
			t.Errorf("Restore failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		fsm.Apply(6, "register", "srv-4")
	}()

	wg.Wait()

	lastIndex := fsm.LastIndex()
	if lastIndex == 1 {
		if !store.IsRegistered("srv-1") {
			t.Errorf("Expected srv-1 to be registered at index 1")
		}
		if store.IsRegistered("srv-4") {
			t.Errorf("srv-4 should not be registered if Restore was last")
		}
	} else if lastIndex == 6 {
		if !store.IsRegistered("srv-4") {
			t.Errorf("Expected srv-4 to be registered at index 6")
		}
	} else {
		t.Errorf("Unexpected last index: %d", lastIndex)
	}

	if lastIndex == 1 {
		fsm.Apply(2, "deregister", "srv-1")
		if store.IsRegistered("srv-1") {
			t.Errorf("Expected srv-1 to be deregistered after replay")
		}
	}
}
