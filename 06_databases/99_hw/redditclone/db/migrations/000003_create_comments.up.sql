CREATE TABLE comments (
  id BIGSERIAL PRIMARY KEY,
  body TEXT NOT NULL,
  author_id BIGINT NOT NULL REFERENCES users(id),
  post_id BIGINT NOT NULL REFERENCES posts(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_post ON comments(post_id);
