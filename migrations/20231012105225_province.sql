-- +goose Up
-- +goose StatementBegin
CREATE TABLE province (
	psgc_code TEXT PRIMARY KEY,
	reg_code TEXT,
	name TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE province
-- +goose StatementEnd