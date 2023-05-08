package sql

import (
	"context"
	"database/sql"

	"github.com/libsv/bitcoin-hc/repository/dto"
	"github.com/pkg/errors"
)

const (
	sqlInsertWebhook = `
	INSERT INTO webhooks(name, url, tokenHeader, token, createdAt)
	VALUES(:name, :url, :tokenHeader, :token, :createdAt)
	ON CONFLICT DO NOTHING
	`

	sqlGetWebhookByName = ` 
	SELECT name, url, tokenHeader, token, createdAt, lastEmitStatus, lastEmitTimestamp, errorsCount
	FROM webhooks
	WHERE name = ?
	`

	sqlGetWebhookByUrl = ` 
	SELECT name, url, tokenHeader, token, createdAt, lastEmitStatus, lastEmitTimestamp, errorsCount
	FROM webhooks
	WHERE url = ?
	`

	sqlDeleteWebhookByName = `
	DELETE FROM webhooks
	WHERE name = ?
	`

	sqlDeleteWebhookByUrl = `
	DELETE FROM webhooks
	WHERE url = ?
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

// GetWebhookByName method will search and return webhook by name.
func (h *HeadersDb) GetWebhookByName(ctx context.Context, name string) (*dto.DbWebhook, error) {
	var rWebhook dto.DbWebhook
	if err := h.db.GetContext(ctx, &rWebhook, h.db.Rebind(sqlGetWebhookByName), name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("could not find webhook")
		}
		return nil, errors.Wrapf(err, "failed to get webhook using name %s", name)
	}
	return &rWebhook, nil
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

// DeleteWebhookByName method will delete webhook by name from db.
func (h *HeadersDb) DeleteWebhookByName(ctx context.Context, name string) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := h.db.Prepare(sqlDeleteWebhookByName)
	if err != nil {
		return err
	}
	defer stmt.Close() //nolint:all

	if _, err = stmt.Exec(name); err != nil {
		return errors.Wrap(err, "failed to delete webhook")
	}
	return errors.Wrap(tx.Commit(), "failed to commit tx")
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
