package actionqueue

import (
	"sync"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
)

type ActionQueue interface {
	Enqueue(serverID uuid.UUID, action domain.AgentAction)
	PopAll(serverID uuid.UUID) []domain.AgentAction
	HasPendingAction(serverID uuid.UUID, actionType string, target string) bool
}

type memoryActionQueue struct {
	mu     sync.Mutex
	queues map[uuid.UUID][]domain.AgentAction
}

func NewActionQueue() ActionQueue {
	return &memoryActionQueue{
		queues: make(map[uuid.UUID][]domain.AgentAction),
	}
}

func (q *memoryActionQueue) Enqueue(serverID uuid.UUID, action domain.AgentAction) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queues[serverID] = append(q.queues[serverID], action)
}

func (q *memoryActionQueue) PopAll(serverID uuid.UUID) []domain.AgentAction {
	q.mu.Lock()
	defer q.mu.Unlock()

	actions, exists := q.queues[serverID]
	if !exists || len(actions) == 0 {
		return nil
	}

	delete(q.queues, serverID)
	return actions
}

func (q *memoryActionQueue) HasPendingAction(serverID uuid.UUID, actionType string, target string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	actions, exists := q.queues[serverID]
	if !exists {
		return false
	}

	for _, a := range actions {
		if a.Type == actionType && (a.Target == target || a.Target == "caelus-"+target || target == "caelus-"+a.Target) {
			return true
		}
	}
	return false
}
