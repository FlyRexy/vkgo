CREATE TABLE posts (
  id BIGSERIAL PRIMARY KEY,
  author_id  BIGINT NOT NULL REFERENCES users(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  score BIGINT NOT NULL DEFAULT 0,
  upvote_percentage INT NOT NULL DEFAULT 0,
  title TEXT NOT NULL,
  text TEXT NOT NULL,
  type TEXT NOT NULL CHECK (type IN ('text', 'link')),
  category TEXT NOT NULL,
  url TEXT
);

CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_category ON posts(category);