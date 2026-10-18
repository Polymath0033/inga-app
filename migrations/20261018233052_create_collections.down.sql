ALTER TABLE links
	DROP CONSTRAINT IF EXISTS fk_links_collection_id;

DROP TABLE IF EXISTS collections;