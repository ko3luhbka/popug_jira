-- +goose Up
-- +goose StatementBegin
CREATE TABLE account (
    id SERIAL PRIMARY KEY,
    assignee_id uuid NOT NULL,
    credit int DEFAULT 0,
    debit int DEFAULT 0,
    created timestamp NOT NULL
);

CREATE TABLE audit (
    id SERIAL PRIMARY KEY,
    event_name varchar NOT NULL,
    assignee_id uuid,
    task_id varchar,
    task_title varchar,
    amount int,
    created timestamp NOT NULL
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE audit;
DROP TABLE account;
-- +goose StatementEnd