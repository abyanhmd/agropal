-- +goose Up
CREATE TABLE IF NOT EXISTS posts (
	post_id UUID PRIMARY KEY,
	user_id UUID NOT NULL REFERENCES users (user_id),
	title VARCHAR(255) NOT NULL,
	content TEXT NOT NULL,
	slug VARCHAR(255) NOT NULL UNIQUE,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS post_comments (
	comment_id UUID PRIMARY KEY,
	post_id UUID REFERENCES posts (post_id),
	user_id UUID REFERENCES users (user_id),
	parent_id UUID REFERENCES post_comments (comment_id),
	content TEXT NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS post_engagements(
    engagement_id UUID PRIMARY KEY,
    asset_id UUID NOT NULL,
    asset_type VARCHAR(255) NOT NULL,
	user_id UUID REFERENCES users (user_id),
	type VARCHAR(255) NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS post_engagements;

DROP TABLE IF EXISTS post_comments;

DROP TABLE IF EXISTS posts;
