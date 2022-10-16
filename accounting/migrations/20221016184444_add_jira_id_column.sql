-- +goose Up
-- +goose StatementBegin
ALTER TABLE audit
ADD COLUMN jira_id varchar;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE audit
DROP COLUMN jira_id;
-- +goose StatementEnd
