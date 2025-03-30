-- Write your migrate up statements here
CREATE TABLE users_password_history (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) NOT NULL,
    pass_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_password_history_user_id ON users_password_history(user_id);

---- create above / drop below ----

DROP INDEX idx_users_password_history_user_id;
DROP TABLE users_password_history;

-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
