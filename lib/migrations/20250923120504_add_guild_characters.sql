-- +goose Up
-- +goose StatementBegin
CREATE TABLE guild_characters (
    id VARCHAR(75),
    guild_id VARCHAR(75),
    character_id INTEGER,
    secondary_character_id INTEGER,
    effect_mask INTEGER,
    extra_data TEXT,
    PRIMARY KEY (id, guild_id),
    FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE guild_characters;
-- +goose StatementEnd
