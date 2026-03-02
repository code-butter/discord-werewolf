-- +goose Up
-- +goose StatementBegin
CREATE TABLE character_actions (
    guild_id VARCHAR(75) NOT NULL,
    user_id VARCHAR(75) NOT NULL,
    target_id VARCHAR(75) NOT NULL,
    action VARCHAR(25) NOT NULL,
    FOREIGN KEY (guild_id) REFERENCES guilds(id) ON DELETE CASCADE,
    FOREIGN KEY (guild_id, user_id) REFERENCES guild_characters(guild_id, id) ON DELETE CASCADE,
    FOREIGN KEY (guild_id, target_id) REFERENCES guild_characters(guild_id, id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
