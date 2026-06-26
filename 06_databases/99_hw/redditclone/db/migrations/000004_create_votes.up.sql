CREATE TABLE votes (
  post_id BIGINT NOT NULL REFERENCES posts(id),
  user_id BIGINT NOT NULL REFERENCES users(id),
  vote SMALLINT NOT NULL,
  PRIMARY KEY (post_id, user_id)
);

CREATE INDEX idx_votes_post ON votes(post_id);