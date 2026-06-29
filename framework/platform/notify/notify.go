// Package notify provides outbound notification dispatch for tenants.
//
// Notifications are enqueued in the database and dispatched by a background
// worker. This package provides the enqueue API and the Dispatcher interface
// that transport implementations must satisfy.
package notify

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Channel is the delivery mechanism for a notification.
type Channel string

const (
	ChannelEmail   Channel = "email"
	ChannelSMS     Channel = "sms"
	ChannelPush    Channel = "push"
	ChannelWebhook Channel = "webhook"
)

// Status is the delivery state of a notification.
type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

// Notification is a single outbound message.
type Notification struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	RecipientID string     `json:"recipient_id"`
	Channel     Channel    `json:"channel"`
	Subject     string     `json:"subject"`
	Body        string     `json:"body"`
	Status      Status     `json:"status"`
	ErrorMsg    string     `json:"error_msg,omitempty"`
	SentAt      *time.Time `json:"sent_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Store is the persistence interface for notifications.
type Store interface {
	Create(ctx context.Context, n *Notification) error
	ListPending(ctx context.Context, limit int) ([]*Notification, error)
	MarkSent(ctx context.Context, id uuid.UUID) error
	MarkFailed(ctx context.Context, id uuid.UUID, errMsg string) error
}

// Dispatcher sends a notification over its channel.
type Dispatcher interface {
	Dispatch(ctx context.Context, n *Notification) error
}

// Service enqueues notifications and drives dispatch.
type Service struct {
	store       Store
	dispatchers map[Channel]Dispatcher
}

func NewService(store Store) *Service {
	return &Service{
		store:       store,
		dispatchers: make(map[Channel]Dispatcher),
	}
}

// Register adds a transport implementation for channel.
func (s *Service) Register(ch Channel, d Dispatcher) {
	s.dispatchers[ch] = d
}

// Send enqueues a notification for async delivery.
func (s *Service) Send(ctx context.Context, n *Notification) error {
	return s.store.Create(ctx, n)
}

// ProcessPending fetches up to limit pending notifications and dispatches them.
// Call this from a background worker or Temporal activity.
func (s *Service) ProcessPending(ctx context.Context, limit int) error {
	pending, err := s.store.ListPending(ctx, limit)
	if err != nil {
		return err
	}
	for _, n := range pending {
		d, ok := s.dispatchers[n.Channel]
		if !ok {
			if markErr := s.store.MarkFailed(ctx, n.ID, "no dispatcher registered for channel "+string(n.Channel)); markErr != nil {
				return markErr
			}
			continue
		}
		if dispErr := d.Dispatch(ctx, n); dispErr != nil {
			if markErr := s.store.MarkFailed(ctx, n.ID, dispErr.Error()); markErr != nil {
				return markErr
			}
			continue
		}
		if markErr := s.store.MarkSent(ctx, n.ID); markErr != nil {
			return markErr
		}
	}
	return nil
}
