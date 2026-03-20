-- +goose Up

ALTER TABLE document_versions ADD COLUMN is_signed BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE document_versions ADD COLUMN is_final BOOLEAN NOT NULL DEFAULT false;

CREATE TABLE document_version_tags (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  version_id UUID NOT NULL REFERENCES document_versions(id) ON DELETE CASCADE,
  tag VARCHAR(50) NOT NULL,
  UNIQUE(version_id, tag)
);

-- +goose Down

DROP TABLE IF EXISTS document_version_tags;
ALTER TABLE document_versions DROP COLUMN IF EXISTS is_final;
ALTER TABLE document_versions DROP COLUMN IF EXISTS is_signed;
