-- +goose Up
-- +goose StatementBegin
CREATE TABLE assignee (
    id uuid PRIMARY KEY,
    username varchar NOT NULL UNIQUE
);

CREATE TABLE task (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
    name varchar NOT NULL UNIQUE,
    description text NOT NULL,
    status varchar NOT NULL DEFAULT '',
    assignee_id uuid REFERENCES assignee (id),
    created timestamp NOT NULL
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE task;
DROP TABLE assignee;
-- +goose StatementEnd