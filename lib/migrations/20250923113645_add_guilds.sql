-- +goose Up
-- +goose StatementBegin
CREATE TABLE guilds (
    id varchar(75) PRIMARY KEY, -- Discord ID
    name VARCHAR(100),
    channels TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE guilds;
-- +goose StatementEnd
