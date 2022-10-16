-- +goose Up
-- +goose StatementBegin
ALTER TABLE task
ADD COLUMN jira_id varchar;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE task
DROP COLUMN jira_id;
-- +goose StatementEnd
