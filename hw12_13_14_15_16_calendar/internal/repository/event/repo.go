package event

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository incapsulate all work with event in database.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository return new SQL repository for events.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

// Create saves event in database.
func (r *Repository) Create(ctx context.Context, event Event) (Event, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		// log.Warn("can't acquire connection from pool", "err", err)
		return Event{}, fmt.Errorf("can't acquire connection: %w", err)
	}
	defer conn.Release()

	busy, err := r.isDateBusy(ctx, conn, event)
	if err != nil {
		return Event{}, err
	}
	if busy {
		return Event{}, ErrDateBusy
	}

	query := `
		INSERT INTO events 
		(id, title, scheduled_at, finished_at, description, user_id, notify_before_seconds) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = conn.Exec(
		ctx,
		query,
		event.ID,
		event.Title,
		event.ScheduledAt,
		event.FinishedAt,
		event.Description,
		event.UserID,
		event.NotifyBeforeSeconds,
	)
	if err != nil {
		// log.Warn("can't create user in database", "err", err)
		return Event{}, fmt.Errorf("can't create event: %w", err)
	}

	return event, nil
}

// Delete by id in database.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		// log.Warn("can't acquire connection from pool", "err", err)
		return fmt.Errorf("can't acquire connection: %w", err)
	}
	defer conn.Release()

	tag, err := conn.Exec(ctx, "DELETE FROM events WHERE id = ($1)", id)
	if err != nil {
		// log.Warn("can't execute request for deleting", "err", err)
		return fmt.Errorf("can't delete event: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ErrEventNotFound
	}

	return nil
}

// Update evnt in database.
func (r *Repository) Update(ctx context.Context, event Event) (Event, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		// log.Warn("can't acquire connection from pool", "err", err)
		return Event{}, fmt.Errorf("can't acquire connection: %w", err)
	}
	defer conn.Release()

	busy, err := r.isDateBusy(ctx, conn, event)
	if err != nil {
		return Event{}, err
	}
	if busy {
		return Event{}, ErrDateBusy
	}

	query := `
		UPDATE events 
		SET  title = $2, scheduled_at = $3, finished_at = $4, description = $5, user_id = $6, notify_before_seconds = $7
		WHERE id = $1
	`

	tag, err := conn.Exec(
		ctx,
		query,
		event.ID,
		event.Title,
		event.ScheduledAt,
		event.FinishedAt,
		event.Description,
		event.UserID,
		event.NotifyBeforeSeconds,
	)
	if err != nil {
		// log.Warn("can't update user in database", "event", event, "err", err)
		return Event{}, fmt.Errorf("can't update event: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return Event{}, ErrEventNotFound
	}
	return event, nil
}

// GetByDay get all events by day.
func (r *Repository) GetByDay(ctx context.Context, startOfDay, startOfNextDay time.Time) (
	[]Event, error,
) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		// log.Warn("can't acquire connection from pool", "err", err)
		return nil, fmt.Errorf("can't acquire connection: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(
		ctx,
		`
        SELECT id, title, scheduled_at, finished_at, description, user_id, notify_before_seconds   
				FROM events
        WHERE scheduled_at >= $1
        AND scheduled_at < $2
        ORDER BY scheduled_at
        `,
		startOfDay,
		startOfNextDay,
	)
	if err != nil {
		// log.Warn("can't execute query for find by day", "err", err)
		return nil, fmt.Errorf("can't execute select by day: %w", err)
	}

	// close rows automatically
	events, err := pgx.CollectRows(rows, pgx.RowToStructByName[Event])
	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetByWeek get all events by week.
func (r *Repository) GetByWeek(ctx context.Context, startOfWeek, endOfWeek time.Time) (
	[]Event, error,
) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		// log.Warn("can't acquire connection from pool", "err", err)
		return nil, fmt.Errorf("can't acquire connection: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(
		ctx,
		`
        SELECT id, title, scheduled_at, finished_at, description, user_id, notify_before_seconds   
				FROM events
        WHERE scheduled_at >= $1
        AND scheduled_at < $2
        ORDER BY scheduled_at
        `,
		startOfWeek,
		endOfWeek,
	)
	if err != nil {
		// log.Warn("can't execute query for find by week", "err", err)
		return nil, fmt.Errorf("can't execute select by week: %w", err)
	}

	// close rows automatically
	events, err := pgx.CollectRows(rows, pgx.RowToStructByName[Event])
	if err != nil {
		return nil, err
	}

	return events, nil
}

// GetByMonth get all events by month.
func (r *Repository) GetByMonth(ctx context.Context, startOfMonth, endOfMonth time.Time) ([]Event, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		// log.Warn("can't acquire connection from pool", "err", err)
		return nil, fmt.Errorf("can't acquire connection: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(
		ctx,
		`
        SELECT id, title, scheduled_at, finished_at, description, user_id, notify_before_seconds   
				FROM events
        WHERE scheduled_at >= $1
        AND scheduled_at < $2
        ORDER BY scheduled_at
        `,
		startOfMonth,
		endOfMonth,
	)
	if err != nil {
		// log.Warn("can't execute query for find by month", "err", err)
		return nil, fmt.Errorf("can't execute select by month: %w", err)
	}

	// close rows automatically
	events, err := pgx.CollectRows(rows, pgx.RowToStructByName[Event])
	if err != nil {
		return nil, err
	}

	return events, nil
}

type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *Repository) isDateBusy(ctx context.Context, q queryer, event Event) (bool, error) {
	const query = `
		SELECT EXISTS(
			SELECT 1
			FROM events
			WHERE user_id = $1
				AND id <> $2
				AND scheduled_at < $4
				AND finished_at > $3
		)
	`

	var busy bool
	if err := q.QueryRow(ctx, query, event.UserID, event.ID, event.ScheduledAt, event.FinishedAt).Scan(&busy); err != nil {
		return false, fmt.Errorf("can't check event date availability: %w", err)
	}
	return busy, nil
}
