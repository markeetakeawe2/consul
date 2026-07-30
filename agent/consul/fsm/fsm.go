package fsm

import (
	"encoding/json"
	"io"
	"sync"

	"github.com/markeetakeawe2/consul/agent/consul/state"
)

type FSM struct {
	mu        sync.RWMutex
	state     *state.Store
	lastIndex uint64
}

func NewFSM(state *state.Store) *FSM {
	return &FSM{
		state: state,
	}
}

func (f *FSM) Apply(logIndex uint64, cmdType string, serviceID string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if logIndex <= f.lastIndex {
		return
	}

	f.lastIndex = logIndex
	if cmdType == "register" {
		f.state.Register(serviceID)
	} else if cmdType == "deregister" {
		f.state.Deregister(serviceID)
	}
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	var snapData struct {
		Index    uint64
		Services map[string]bool
	}

	if err := json.NewDecoder(rc).Decode(&snapData); err != nil {
		return err
	}

	f.state.Restore(snapData.Services)
	f.lastIndex = snapData.Index

	return nil
}

func (f *FSM) LastIndex() uint64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.lastIndex
}

func (f *FSM) State() *state.Store {
	return f.state
}
