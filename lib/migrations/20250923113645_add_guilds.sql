-- +goose Up
-- +goose StatementBegin
CREATE TABLE guilds (
    id              VARCHAR(75) PRIMARY KEY,    -- Discord ID
    name            VARCHAR(100),
    channels        TEXT,                       -- json
    game_going      INTEGER DEFAULT 0,          -- bool
    game_mode       INTEGER DEFAULT 0,
    day_night       INTEGER DEFAULT 0,          -- bool
    paused          INTEGER DEFAULT 0,          -- bool
    time_zone       VARCHAR(50) DEFAULT '',
    day_time        VARCHAR(5) DEFAULT '',      -- time-only (UTC)
    night_time      VARCHAR(5) DEFAULT '',      -- time-only (UTC)
    game_settings   TEXT,                       -- json
    last_cycle_ran  TEXT                        -- date-time (UTC)
);

CREATE INDEX guild_playing on guilds (game_going, paused);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX guild_playing;
DROP TABLE guilds;
-- +goose StatementEnd
