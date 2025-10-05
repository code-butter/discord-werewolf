-- +goose Up
-- +goose StatementBegin
CREATE TABLE guild_votes(
    guild_id VARCHAR(75),
    user_id VARCHAR(75),
    voting_for_id VARCHAR(75),
    PRIMARY KEY (guild_id, user_id),
    FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE,
    FOREIGN KEY (guild_id, user_id) REFERENCES guild_characters(guild_id, id) ON DELETE CASCADE,
    FOREIGN KEY (guild_id, voting_for_id) REFERENCES guild_characters(guild_id, id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE guild_votes;
-- +goose StatementEnd
