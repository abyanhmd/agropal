-- +goose Up
CREATE TABLE IF NOT EXISTS posts (
	post_id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users (user_id) ON DELETE CASCADE,
	title VARCHAR(255) NOT NULL,
	content TEXT NOT NULL,
	slug VARCHAR(255) NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS post_comments (
	comment_id UUID PRIMARY KEY,
	post_id UUID REFERENCES posts (post_id) ON DELETE CASCADE,
	user_id UUID REFERENCES users (user_id) ON DELETE CASCADE,
	parent_id UUID REFERENCES post_comments (comment_id) ON DELETE CASCADE,
	content TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS post_engagements(
    engagement_id UUID PRIMARY KEY,
    asset_id UUID NOT NULL,
    asset_type VARCHAR(255) NOT NULL,
	user_id UUID REFERENCES users (user_id) ON DELETE CASCADE,
	type VARCHAR(255) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts (user_id);
CREATE INDEX IF NOT EXISTS idx_post_comments_post_id ON post_comments (post_id);
CREATE INDEX IF NOT EXISTS idx_post_comments_parent_id ON post_comments (parent_id);
CREATE INDEX IF NOT EXISTS idx_post_engagements_asset ON post_engagements (asset_type, asset_id);

-- +goose Down
DROP INDEX IF EXISTS idx_post_engagements_asset;

DROP INDEX IF EXISTS idx_post_comments_parent_id;

DROP INDEX IF EXISTS idx_post_comments_post_id;

DROP INDEX IF EXISTS idx_posts_user_id;

DROP TABLE IF EXISTS post_engagements;

DROP TABLE IF EXISTS post_comments;

DROP TABLE IF EXISTS posts;
