-- +goose Up
-- +goose StatementBegin
CREATE TABLE "user" (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid () ,
    username varchar NOT NULL UNIQUE,
    password varchar NOT NULL,
    role varchar NOT NULL,
    email varchar NOT NULL,
    last_modified timestamp NOT NULL
);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE "user";
-- +goose StatementEnd