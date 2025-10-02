-- +goose Up
-- +goose StatementBegin
CREATE TABLE guild_votes(
    id INT PRIMARY KEY,
    guild_id VARCHAR(75),
    user_id VARCHAR(75),
    voting_for_id VARCHAR(75),
    FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES guild_characters(id) ON DELETE CASCADE,
    FOREIGN KEY  (voting_for_id) REFERENCES  guild_characters(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE guild_votes;
-- +goose StatementEnd
