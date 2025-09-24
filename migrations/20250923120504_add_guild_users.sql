-- +goose Up
-- +goose StatementBegin
CREATE TABLE guild_users (
    id INTEGER,
    guild_id INTEGER,
    character_id INTEGER,
    secondary_character_mask INTEGER,
    status INTEGER,
    PRIMARY KEY (id, guild_id),
    FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE guild_users
-- +goose StatementEnd
