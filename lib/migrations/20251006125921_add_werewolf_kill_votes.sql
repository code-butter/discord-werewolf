-- +goose Up
-- +goose StatementBegin
CREATE TABLE werewolf_kill_votes(
    guild_id VARCHAR(75),
    user_id VARCHAR(75),
    voting_for_id VARCHAR(75),
    FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE,
    FOREIGN KEY (guild_id, user_id) REFERENCES guild_characters(guild_id, id) ON DELETE CASCADE,
    FOREIGN KEY (guild_id, voting_for_id) REFERENCES guild_characters(guild_id, id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE werewolf_kill_votes;
-- +goose StatementEnd
