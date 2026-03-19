-- +goose Up

ALTER TABLE estimation_items DROP COLUMN hours;
ALTER TABLE estimation_items ADD COLUMN min_hours DECIMAL(8,2) NOT NULL DEFAULT 0;
ALTER TABLE estimation_items ADD COLUMN likely_hours DECIMAL(8,2) NOT NULL DEFAULT 0;
ALTER TABLE estimation_items ADD COLUMN max_hours DECIMAL(8,2) NOT NULL DEFAULT 0;
ALTER TABLE estimation_items ADD COLUMN note TEXT NOT NULL DEFAULT '';

-- +goose Down

ALTER TABLE estimation_items DROP COLUMN note;
ALTER TABLE estimation_items DROP COLUMN max_hours;
ALTER TABLE estimation_items DROP COLUMN likely_hours;
ALTER TABLE estimation_items DROP COLUMN min_hours;
ALTER TABLE estimation_items ADD COLUMN hours DECIMAL(8,2) NOT NULL DEFAULT 0;
