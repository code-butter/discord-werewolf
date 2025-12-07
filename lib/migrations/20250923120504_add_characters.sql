-- +goose Up
-- +goose StatementBegin

CREATE TABLE characters(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label VARCHAR(20) UNIQUE
);

CREATE INDEX characters_label ON characters(label);

CREATE TABLE secondary_characters(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    label VARCHAR(20) UNIQUE
);

CREATE INDEX secondary_characters_label ON secondary_characters(label);

CREATE TABLE guild_characters (
    id VARCHAR(75),
    guild_id VARCHAR(75),
    character_id INTEGER REFERENCES characters(id),
    secondary_character_id INTEGER REFERENCES secondary_characters(id),
    effect_mask INTEGER,
    extra_data TEXT,
    PRIMARY KEY (id, guild_id),
    FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE guild_characters;
DROP TABLE characters;
-- +goose StatementEnd
