package feedback

import (
	"context"
	"database/sql"
	"time"

	"gitlab.roy9.ru/roy9/backend/core/feedbackservice/internal/domain/feedback"

	sq "github.com/Masterminds/squirrel"
)

const (
	queryTimeout = 7 * time.Second
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	repo := &Repository{
		db: db,
	}

	return repo
}

func (r *Repository) Create(ctx context.Context, fb feedback.Feedback) (int64, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var id int64

	err := sq.
		Insert("feedback").
		Columns("score", "text", "rent_id").
		Values(fb.Score, fb.Text, fb.RentID).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		Suffix("RETURNING id").
		QueryRowContext(c).
		Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) HasFeedbackForRent(ctx context.Context, rentID string) (bool, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var exists bool

	err := sq.Select("1").
		Prefix("SELECT EXISTS (").
		From("feedback").
		Where(sq.Eq{"rent_id": rentID}).
		Suffix(")").
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		QueryRowContext(c).
		Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repository) CreateFeedbackLocal(ctx context.Context, fl feedback.FeedbackLocal) (int64, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var id int64

	err := sq.
		Insert("feedback_local").
		Columns("user_id", "type", "text").
		Values(fl.UserID, fl.Type, fl.Text).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		Suffix("RETURNING id").
		QueryRowContext(c).
		Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *Repository) CreateFeedbackPartnership(ctx context.Context, fp feedback.FeedbackPartnership) (int64, error) {
	c, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	var id int64

	err := sq.
		Insert("feedback_partnership").
		Columns("user_id", "contact_name", "company_name", "email", "phone_num", "comment").
		Values(fp.UserID, fp.ContactName, fp.CompanyName, fp.Email, fp.PhoneNum, fp.Comment).
		PlaceholderFormat(sq.Dollar).
		RunWith(r.db).
		Suffix("RETURNING id").
		QueryRowContext(c).
		Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
