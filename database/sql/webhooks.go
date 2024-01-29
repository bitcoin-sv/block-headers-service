package sql

import (
	"context"
	"database/sql"
	"time"

	"github.com/bitcoin-sv/pulse/repository/dto"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	sqlInsertWebhook = `
	INSERT INTO webhooks(url, token_header, token, created_at)
	VALUES(:url, :token_header, :token, :created_at)
	`

	sqlGetWebhookByUrl = ` 
	SELECT url, token_header, token, created_at, last_emit_status, last_emit_timestamp, errors_count, is_active
	FROM webhooks
	WHERE url = ?
	`

	sqlGetAllWebhooks = `
	SELECT url, token_header, token, created_at, last_emit_status, last_emit_timestamp, errors_count, is_active
	FROM webhooks
	`

	sqlDeleteWebhookByUrl = `
	DELETE FROM webhooks
	WHERE url = ?
	`

	sqlUpdateWebhook = `
	UPDATE webhooks
	SET last_emit_status = ?, last_emit_timestamp = ?, errors_count = ?, is_active = ?
	WHERE url IN (?)
	`
)

// CreateWebhook method will add new webhook into db.
func (h *HeadersDb) CreateWebhook(ctx context.Context, rWebhook *dto.DbWebhook) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.NamedExecContext(ctx, h.db.Rebind(sqlInsertWebhook), *rWebhook); err != nil {
		return errors.Wrap(err, "failed to insert webhook")
	}
	return errors.Wrap(tx.Commit(), "failed to commit tx")
}

// GetWebhookByUrl method will search and return webhook by url.
func (h *HeadersDb) GetWebhookByUrl(ctx context.Context, url string) (*dto.DbWebhook, error) {
	var rWebhook dto.DbWebhook
	if err := h.db.GetContext(ctx, &rWebhook, h.db.Rebind(sqlGetWebhookByUrl), url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find webhook")
		}
		return nil, errors.Wrapf(err, "failed to get webhook using url %s", url)
	}
	return &rWebhook, nil
}

// GetAllWebhooks method will return all webhooks from db.
func (h *HeadersDb) GetAllWebhooks(ctx context.Context) ([]*dto.DbWebhook, error) {
	var rWebhooks []*dto.DbWebhook
	if err := h.db.SelectContext(ctx, &rWebhooks, h.db.Rebind(sqlGetAllWebhooks)); err != nil {
		return nil, errors.Wrap(err, "failed to get all webhooks")
	}
	return rWebhooks, nil
}

// DeleteWebhookByUrl method will delete webhook by url from db.
func (h *HeadersDb) DeleteWebhookByUrl(ctx context.Context, url string) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := h.db.Prepare(sqlDeleteWebhookByUrl)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:all

	if _, err = stmt.Exec(url); err != nil {
		return errors.Wrap(err, "failed to delete webhook")
	}
	return errors.Wrap(tx.Commit(), "failed to commit tx")
}

// UpdateWebhook method will update webhook in db.
func (h *HeadersDb) UpdateWebhook(ctx context.Context, url string, lastEmitTimestamp time.Time, lastEmitStatus string, errorsCount int, active bool) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query, args, err := sqlx.In(sqlUpdateWebhook, lastEmitStatus, lastEmitTimestamp, errorsCount, active, url)
	if err != nil {
		return errors.Wrapf(err, "failed to update webhook with url %s", url)
	}
	if _, err := tx.ExecContext(ctx, h.db.Rebind(query), args...); err != nil {
		return errors.Wrapf(err, "failed to update webhook with name %s", url)
	}
	return errors.Wrap(tx.Commit(), "failed to commit tx")

}
