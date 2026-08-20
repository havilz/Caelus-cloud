package automation

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/havilz/caelus-cloud/backend/internal/domain"
	"github.com/havilz/caelus-cloud/backend/pkg/logger"
)

// EventSubscriber mendefinisikan fungsi penerima kejadian sistem.
type EventSubscriber func(ctx context.Context, event domain.SystemEvent) error

// EventDispatcher mendefinisikan antarmuka penyebar kejadian sistem terpusat.
type EventDispatcher interface {
	// Publish mendistribusikan event sistem ke seluruh subscriber yang terdaftar.
	Publish(ctx context.Context, event domain.SystemEvent)

	// Subscribe mendaftarkan subscriber baru untuk menerima event sistem.
	Subscribe(subscriber EventSubscriber)
}

// CentralEventDispatcher mengimplementasikan EventDispatcher berbasis in-memory fan-out yang aman untuk goroutine.
type CentralEventDispatcher struct {
	subscribers []EventSubscriber
	mu          sync.RWMutex
	eventChan   chan domain.SystemEvent
	stopChan    chan struct{}
	wg          sync.WaitGroup
	isClosed    bool
}

// NewCentralEventDispatcher membuat dan menginisialisasi instance CentralEventDispatcher baru.
// Mengembalikan pointer *CentralEventDispatcher.
func NewCentralEventDispatcher() *CentralEventDispatcher {
	d := &CentralEventDispatcher{
		subscribers: make([]EventSubscriber, 0),
		eventChan:   make(chan domain.SystemEvent, 500),
		stopChan:    make(chan struct{}),
	}
	d.start()
	return d
}

// Subscribe mendaftarkan fungsi penangan kejadian sistem baru.
func (d *CentralEventDispatcher) Subscribe(subscriber EventSubscriber) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.subscribers = append(d.subscribers, subscriber)
}

// Publish memasukkan event baru ke antrean penyebaran asinkron.
func (d *CentralEventDispatcher) Publish(ctx context.Context, event domain.SystemEvent) {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	select {
	case d.eventChan <- event:
		logger.Debug("System event dipublikasikan ke dispatcher", "event_type", event.Type, "event_id", event.ID)
	default:
		logger.Warn("Antrean event dispatcher penuh, mendistribusikan langsung via goroutine", "event_type", event.Type)
		go d.broadcast(context.Background(), event)
	}
}

// start memulai loop penyebaran event di latar belakang.
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

// broadcast mengirimkan event ke seluruh subscriber yang terdaftar secara paralel.
func (d *CentralEventDispatcher) broadcast(ctx context.Context, event domain.SystemEvent) {
	d.mu.RLock()
	subs := make([]EventSubscriber, len(d.subscribers))
	copy(subs, d.subscribers)
	d.mu.RUnlock()

	for _, sub := range subs {
		if sub != nil {
			go func(s EventSubscriber) {
				if err := s(ctx, event); err != nil {
					logger.Error("Error pada subscriber event dispatcher", "event_type", event.Type, "error", err)
				}
			}(sub)
		}
	}
}

// Stop menghentikan dispatcher secara anggun.
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
