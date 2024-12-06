CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_comments_content ON comments USING gin (content gin_trgm_ops);
CREATE INDEX idx_posts_title ON posts USING gin (title gin_trgm_ops);
CREATE INDEX idx_posts_tags ON posts USING gin (tags gin_trgm_ops);
CREATE INDEX idx_users_username ON users USING gin (username gin_trgm_ops);

-- Use btree or avoid indexing these columns if incompatible with GIN
CREATE INDEX idx_posts_user_id ON posts USING btree (user_id);
CREATE INDEX idx_comments_post_id ON comments USING btree (post_id);
