package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"my-go-app/internal/models"

	"github.com/lucsky/cuid"
	"gorm.io/gorm"
)

var ErrNotificationNotFound = errors.New("notification not found")

const defaultNotificationLimit = 50

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepositoryInterface {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) forUser(ctx context.Context, userID string) *gorm.DB {
	return r.db.WithContext(ctx).Model(&models.Notification{}).Where(`"UserId" = ?`, userID)
}

func (r *NotificationRepository) count(q *gorm.DB) (int64, error) {
	var n int64
	err := q.Count(&n).Error
	return n, err
}

func (r *NotificationRepository) ListForUser(ctx context.Context, userID string, filter models.NotificationFilter) (*models.NotificationList, error) {
	if userID == "" {
		return nil, errors.New("userID is required")
	}
	limit := filter.Limit
	if limit <= 0 || limit > defaultNotificationLimit {
		limit = defaultNotificationLimit
	}

	q := r.forUser(ctx, userID)
	switch filter.Filter {
	case "", "all":
	case "unread":
		q = q.Where(`"Read" = false`)
	default:
		q = q.Where(`"Type" = ?`, filter.Filter)
	}
	if s := strings.TrimSpace(filter.Search); s != "" {
		like := "%" + s + "%"
		q = q.Where(`("Title" ILIKE ? OR "Description" ILIKE ?)`, like, like)
	}

	list := &models.NotificationList{Notifications: []models.Notification{}}
	if err := q.Order(`"CreatedAt" DESC`).Limit(limit).Find(&list.Notifications).Error; err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}

	// Timestamps are stored as UTC wall-clock (timestamp without time zone),
	// as Prisma wrote them; "today" is the server's local day, as before.
	y, m, d := time.Now().Date()
	startOfDay := time.Date(y, m, d, 0, 0, 0, 0, time.Local).UTC()

	counts := []struct {
		dst   *int64
		where string
		args  []any
	}{
		{&list.Stats.Total, "", nil},
		{&list.Stats.Unread, `"Read" = false`, nil},
		{&list.Stats.ReadToday, `"Read" = true AND "UpdatedAt" >= ?`, []any{startOfDay}},
		{&list.FilterCounts.Messages, `"Type" = ?`, []any{"message"}},
		{&list.FilterCounts.Followers, `"Type" = ?`, []any{"user"}},
		{&list.FilterCounts.Alerts, `"Type" = ?`, []any{"alert"}},
		{&list.FilterCounts.Updates, `"Type" = ?`, []any{"settings"}},
	}
	for _, c := range counts {
		q := r.forUser(ctx, userID)
		if c.where != "" {
			q = q.Where(c.where, c.args...)
		}
		n, err := r.count(q)
		if err != nil {
			return nil, fmt.Errorf("count notifications: %w", err)
		}
		*c.dst = n
	}
	list.FilterCounts.All = list.Stats.Total
	list.FilterCounts.Unread = list.Stats.Unread
	return list, nil
}

func (r *NotificationRepository) SetRead(ctx context.Context, userID, notificationID string, read bool) (*models.Notification, error) {
	var n models.Notification
	err := r.forUser(ctx, userID).Where(`"Id" = ?`, notificationID).First(&n).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotificationNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get notification: %w", err)
	}
	n.Read = read
	n.UpdatedAt = time.Now().UTC()
	if err := r.db.WithContext(ctx).Model(&n).
		Updates(map[string]any{"Read": n.Read, "UpdatedAt": n.UpdatedAt}).Error; err != nil {
		return nil, fmt.Errorf("update notification: %w", err)
	}
	return &n, nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	return r.forUser(ctx, userID).Where(`"Read" = false`).
		Updates(map[string]any{"Read": true, "UpdatedAt": time.Now().UTC()}).Error
}

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	if n.UserId == "" || n.Type == "" || n.Title == "" {
		return errors.New("userId, type and title are required")
	}
	if n.Id == "" {
		n.Id = cuid.New()
	}
	now := time.Now().UTC()
	n.CreatedAt, n.UpdatedAt = now, now
	return r.db.WithContext(ctx).Create(n).Error
}
