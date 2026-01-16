-- +goose Up
-- +goose StatementBegin

CREATE TABLE guild_characters (
    id VARCHAR(75),
    guild_id VARCHAR(75),
    character_id varchar(30),
    secondary_character_id varchar(30),
    fake_character_id varchar(30),
    effect_mask INTEGER,
    extra_data TEXT,
    PRIMARY KEY (id, guild_id),
    FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE
);

CREATE INDEX guild_characters_character_id ON guild_characters(character_id);
CREATE INDEX guild_characters_secondary_character_id ON guild_characters(secondary_character_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE guild_characters;
-- +goose StatementEnd
