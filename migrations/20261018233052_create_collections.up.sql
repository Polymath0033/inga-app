CREATE TABLE collections
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()

);


ALTER TABLE links
    ADD CONSTRAINT fk_links_collection_id
        FOREIGN KEY (collection_id)
            REFERENCES collections(id)
            ON DELETE SET NULL;


ALTER TABLE tags
    DROP CONSTRAINT tags_name_key;