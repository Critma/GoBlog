CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email citext UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    likes INTEGER DEFAULT 0,
    published_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE comments (
    id SERIAL PRIMARY KEY,
    article_id INTEGER REFERENCES articles(id),
    user_id INTEGER REFERENCES users(id),
    text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE article_like (
	id SERIAL PRIMARY KEY,
	article_id INTEGER REFERENCES articles(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	UNIQUE(article_id, user_id)
);

CREATE INDEX idx_articles_author ON articles(author_id);
CREATE INDEX idx_comments_article_user ON comments(article_id, user_id);

CREATE VIEW latest_articles AS
    SELECT a.id, a.title, u.username as author_name, a.likes, a.published_at
    FROM articles a JOIN users u ON a.author_id = u.id
    ORDER BY a.published_at DESC LIMIT 10;


CREATE OR REPLACE FUNCTION update_article_likes_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE articles SET likes = likes + 1 WHERE id = NEW.article_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE articles SET likes = likes - 1 WHERE id = OLD.article_id AND likes_count > 0;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_article_likes
    AFTER INSERT OR DELETE ON article_like
    FOR EACH ROW EXECUTE PROCEDURE update_article_likes_count();


CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language plpgsql;

CREATE TRIGGER updated_at_articles
    BEFORE UPDATE ON articles
    FOR EACH ROW EXECUTE PROCEDURE update_modified_column();