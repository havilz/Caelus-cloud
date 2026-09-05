package automation

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

type EventSubscriber func(ctx context.Context, event domain.SystemEvent) error

type EventDispatcher interface {
	Publish(ctx context.Context, event domain.SystemEvent)

	Subscribe(subscriber EventSubscriber)
}

type CentralEventDispatcher struct {
	subscribers []EventSubscriber
	mu          sync.RWMutex
	eventChan   chan domain.SystemEvent
	stopChan    chan struct{}
	wg          sync.WaitGroup
	isClosed    bool
}

func NewCentralEventDispatcher() *CentralEventDispatcher {
	d := &CentralEventDispatcher{
		subscribers: make([]EventSubscriber, 0),
		eventChan:   make(chan domain.SystemEvent, 500),
		stopChan:    make(chan struct{}),
	}
	d.start()
	return d
}

func (d *CentralEventDispatcher) Subscribe(subscriber EventSubscriber) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.subscribers = append(d.subscribers, subscriber)
}

func (d *CentralEventDispatcher) Publish(ctx context.Context, event domain.SystemEvent) {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	select {
	case d.eventChan <- event:
		logger.Debug("System event published to dispatcher", "event_type", event.Type, "event_id", event.ID)
	default:
		logger.Warn("Event dispatcher queue full, broadcasting directly via goroutine", "event_type", event.Type)
		go d.broadcast(context.Background(), event)
	}
}

func (d *CentralEventDispatcher) start() {
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			select {
			case <-d.stopChan:
				return
			case event := <-d.eventChan:
				d.broadcast(context.Background(), event)
			}
		}
	}()
}

func (d *CentralEventDispatcher) broadcast(ctx context.Context, event domain.SystemEvent) {
	d.mu.RLock()
	subs := make([]EventSubscriber, len(d.subscribers))
	copy(subs, d.subscribers)
	d.mu.RUnlock()

	for _, sub := range subs {
		if sub != nil {
			go func(s EventSubscriber) {
				if err := s(ctx, event); err != nil {
					logger.Error("Error in event dispatcher subscriber", "event_type", event.Type, "error", err)
				}
			}(sub)
		}
	}
}

func (d *CentralEventDispatcher) Stop() {
	d.mu.Lock()
	if d.isClosed {
		d.mu.Unlock()
		return
	}
	d.isClosed = true
	close(d.stopChan)
	d.mu.Unlock()

	d.wg.Wait()
}
