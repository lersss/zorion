package generator

import (
	"context"
	"sync"
)

type JobType string

const (
	JobGenerateUniverse JobType = "generate_universe"
	JobGeneratePlanets  JobType = "generate_planets"
	JobGenerateFactions JobType = "generate_factions"
)

type JobStatus struct {
	mu         sync.RWMutex
	Total      int
	Processed  int
	Status     string // "idle", "running", "done", "error", "canceled"
	Error      string
	CancelFunc context.CancelFunc
}

type StatusManager struct {
	mu   sync.RWMutex
	jobs map[JobType]*JobStatus
}

func NewStatusManager() *StatusManager {
	return &StatusManager{
		jobs: make(map[JobType]*JobStatus),
	}
}

func (sm *StatusManager) Get(job JobType) *JobStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	if s, ok := sm.jobs[job]; ok {
		return s
	}
	return nil
}

func (sm *StatusManager) Start(job JobType, total int, cancel context.CancelFunc) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.jobs[job] = &JobStatus{
		Total:      total,
		Processed:  0,
		Status:     "running",
		Error:      "",
		CancelFunc: cancel,
	}
}

func (sm *StatusManager) Progress(job JobType, processed int) {
	sm.mu.RLock()
	s, ok := sm.jobs[job]
	sm.mu.RUnlock()
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Processed = processed
}

func (sm *StatusManager) Done(job JobType) {
	sm.mu.RLock()
	s, ok := sm.jobs[job]
	sm.mu.RUnlock()
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = "done"
	s.CancelFunc = nil
}

func (sm *StatusManager) Fail(job JobType, err string) {
	sm.mu.RLock()
	s, ok := sm.jobs[job]
	sm.mu.RUnlock()
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = "error"
	s.Error = err
	s.CancelFunc = nil
}

func (sm *StatusManager) Cancel(job JobType) {
	sm.mu.RLock()
	s, ok := sm.jobs[job]
	sm.mu.RUnlock()
	if !ok {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Status != "running" {
		return
	}
	if s.CancelFunc != nil {
		s.CancelFunc()
	}
	s.Status = "canceled"
	s.CancelFunc = nil
}

func (sm *StatusManager) GetStatus(job JobType) (total, processed int, status, err string) {
	s := sm.Get(job)
	if s == nil {
		return 0, 0, "idle", ""
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Total, s.Processed, s.Status, s.Error
}