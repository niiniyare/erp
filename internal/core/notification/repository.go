package notification

<<<<<<< HEAD
//go:generate go run go.uber.org/mock/mockgen -source=repository.go -destination=mock.go -package=notification

=======
//go:generate sh -c "mockgen -source=$GOFILE -destination=$(echo $GOFILE | sed 's/\\.go$//')_mock.go -package=$GOPACKAGE"
>>>>>>> ft/ffg

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	db "github.com/niiniyare/erp/db/sqlc"
)

// Repository defines the interface for notification data persistence,
// primarily for managing user notification preferences.
type Repository interface {
	GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error)
	UpdateUserNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs *NotificationPreferences) error
}

type repository struct {
	store db.Store
}

// NewRepository creates a new notification repository.
func NewRepository(store db.Store) Repository {
	return &repository{
		store: store,
	}
}

// GetUserNotificationPreferences retrieves a user's notification preferences.
// If preferences don't exist, it creates and returns default preferences.
func (r *repository) GetUserNotificationPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error) {
	prefs, err := r.store.GetUserNotificationPreferences(ctx, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Preferences not found, create default ones
			return r.createDefaultPreferences(ctx, userID)
		}
		return nil, err
	}
	return fromSQLC(prefs)
}

// UpdateUserNotificationPreferences updates a user's notification preferences.
func (r *repository) UpdateUserNotificationPreferences(ctx context.Context, userID uuid.UUID, prefs *NotificationPreferences) error {
	params, err := toSQLCUpdate(prefs)
	if err != nil {
		return err
	}
	_, err = r.store.UpdateUserNotificationPreferences(ctx, params)
	return err
}

func (r *repository) createDefaultPreferences(ctx context.Context, userID uuid.UUID) (*NotificationPreferences, error) {
	defaultPrefs := defaultPreferences(userID)
	params, err := toSQLCCreate(defaultPrefs)
	if err != nil {
		return nil, err
	}

	created, err := r.store.CreateUserNotificationPreferences(ctx, params)
	if err != nil {
		return nil, err
	}

	return fromSQLC(created)
}

// --- Conversion Helpers ---

func fromSQLC(p *db.NotificationPreference) (*NotificationPreferences, error) {
	var nt map[NotificationType]bool
	if err := json.Unmarshal(p.NotificationTypes, &nt); err != nil {
		return nil, err
	}

	var pc []NotificationChannel
	if err := json.Unmarshal(p.PreferredChannels, &pc); err != nil {
		return nil, err
	}

	var qh *QuietHours
	if p.QuietHours != nil {
		if err := json.Unmarshal(p.QuietHours, &qh); err != nil {
			return nil, err
		}
	}

	return &NotificationPreferences{
		UserID:             p.UserID,
		EmailNotifications: p.EmailNotifications,
		InAppNotifications: p.InAppNotifications,
		SlackNotifications: p.SlackNotifications,
		NotificationTypes:  nt,
		PreferredChannels:  pc,
		QuietHours:         qh,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}, nil
}

func toSQLCCreate(p *NotificationPreferences) (db.CreateUserNotificationPreferencesParams, error) {
	nt, err := json.Marshal(p.NotificationTypes)
	if err != nil {
		return db.CreateUserNotificationPreferencesParams{}, err
	}
	pc, err := json.Marshal(p.PreferredChannels)
	if err != nil {
		return db.CreateUserNotificationPreferencesParams{}, err
	}
	qh, err := json.Marshal(p.QuietHours)
	if err != nil {
		return db.CreateUserNotificationPreferencesParams{}, err
	}

	return db.CreateUserNotificationPreferencesParams{
		UserID:             p.UserID,
		EmailNotifications: p.EmailNotifications,
		InAppNotifications: p.InAppNotifications,
		SlackNotifications: p.SlackNotifications,
		NotificationTypes:  nt,
		PreferredChannels:  pc,
		QuietHours:         qh,
	}, nil
}

func toSQLCUpdate(p *NotificationPreferences) (db.UpdateUserNotificationPreferencesParams, error) {
	nt, err := json.Marshal(p.NotificationTypes)
	if err != nil {
		return db.UpdateUserNotificationPreferencesParams{}, err
	}
	pc, err := json.Marshal(p.PreferredChannels)
	if err != nil {
		return db.UpdateUserNotificationPreferencesParams{}, err
	}
	qh, err := json.Marshal(p.QuietHours)
	if err != nil {
		return db.UpdateUserNotificationPreferencesParams{}, err
	}

	return db.UpdateUserNotificationPreferencesParams{
		UserID:             p.UserID,
		EmailNotifications: p.EmailNotifications,
		InAppNotifications: p.InAppNotifications,
		SlackNotifications: p.SlackNotifications,
		NotificationTypes:  nt,
		PreferredChannels:  pc,
		QuietHours:         qh,
	}, nil
}

func defaultPreferences(userID uuid.UUID) *NotificationPreferences {
	return &NotificationPreferences{
		UserID:             userID,
		EmailNotifications: true,
		InAppNotifications: true,
		SlackNotifications: false,
		NotificationTypes: map[NotificationType]bool{
			NotificationTypeAccessRequestCreated:  true,
			NotificationTypeAccessRequestApproved: true,
			NotificationTypeAccessRequestRejected: true,
			NotificationTypeGeneric:               true,
		},
		PreferredChannels: []NotificationChannel{NotificationChannelEmail, NotificationChannelInApp},
		QuietHours:        &QuietHours{Enabled: false},
	}
}
