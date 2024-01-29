ALTER TABLE headers RENAME COLUMN previousblock TO previous_block;
ALTER TABLE headers RENAME COLUMN cumulatedWork TO cumulated_work;

ALTER TABLE webhooks RENAME COLUMN tokenHeader TO token_header;
ALTER TABLE webhooks RENAME COLUMN createdAt TO created_at;
ALTER TABLE webhooks RENAME COLUMN lastEmitStatus TO last_emit_status;
ALTER TABLE webhooks RENAME COLUMN lastEmitTimestamp TO last_emit_timestamp;
ALTER TABLE webhooks RENAME COLUMN errorsCount TO errors_count;
ALTER TABLE webhooks RENAME COLUMN active TO is_active;
