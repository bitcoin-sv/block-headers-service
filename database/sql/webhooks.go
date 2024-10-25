package sql

import (
	"context"
	"time"

	"github.com/bitcoin-sv/block-headers-service/bhserrors"
	"github.com/bitcoin-sv/block-headers-service/repository/dto"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	sqlInsertWebhook = `
	INSERT INTO webhooks(url, token_header, token, created_at)
	VALUES(:url, :token_header, :token, :created_at)
	`

	sqlGetWebhookByURL = ` 
	SELECT url, token_header, token, created_at, last_emit_status, last_emit_timestamp, errors_count, is_active
	FROM webhooks
	WHERE url = ?
	`

	sqlGetAllWebhooks = `
	SELECT url, token_header, token, created_at, last_emit_status, last_emit_timestamp, errors_count, is_active
	FROM webhooks
	`

	sqlDeleteWebhookByURL = `
	DELETE FROM webhooks
	WHERE url = :url
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
		return bhserrors.ErrCreateWebhook.Wrap(err)
	}

	if err = tx.Commit(); err != nil {
		return bhserrors.ErrCreateWebhook.Wrap(err)
	}

	return nil
}

// GetWebhookByURL method will search and return webhook by url.
func (h *HeadersDb) GetWebhookByURL(ctx context.Context, url string) (*dto.DbWebhook, error) {
	var rWebhook dto.DbWebhook
	if err := h.db.GetContext(ctx, &rWebhook, h.db.Rebind(sqlGetWebhookByURL), url); err != nil {
		return nil, bhserrors.ErrWebhookNotFound.Wrap(err)
	}

	return &rWebhook, nil
}

// GetAllWebhooks method will return all webhooks from db.
func (h *HeadersDb) GetAllWebhooks(ctx context.Context) ([]*dto.DbWebhook, error) {
	var rWebhooks []*dto.DbWebhook
	if err := h.db.SelectContext(ctx, &rWebhooks, h.db.Rebind(sqlGetAllWebhooks)); err != nil {
		return nil, bhserrors.ErrGetAllWebhooks.Wrap(err)
	}

	return rWebhooks, nil
}

// DeleteWebhookByURL method will delete webhook by url from db.
func (h *HeadersDb) DeleteWebhookByURL(ctx context.Context, url string) error {
	tx, err := h.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	params := map[string]interface{}{"url": url}

	if _, err = tx.NamedExecContext(ctx, h.db.Rebind(sqlDeleteWebhookByURL), params); err != nil {
		return bhserrors.ErrDeleteWebhook.Wrap(err)
	}

	if err = tx.Commit(); err != nil {
		return bhserrors.ErrDeleteWebhook.Wrap(err)
	}

	return nil
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
