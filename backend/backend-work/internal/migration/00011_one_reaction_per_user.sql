-- +goose Up
-- Posts, comments and engagements used to be stored without a timestamp, which left created_at at
-- year 1. Their ids are UUIDv7, which start with the creation time in milliseconds, so restore it
-- from there.
UPDATE posts SET created_at = to_timestamp(('x' || substr(replace(post_id::text, '-', ''), 1, 12))::bit(48)::bigint / 1000.0)
WHERE created_at < '2000-01-01';
UPDATE post_comments SET created_at = to_timestamp(('x' || substr(replace(comment_id::text, '-', ''), 1, 12))::bit(48)::bigint / 1000.0)
WHERE created_at < '2000-01-01';
UPDATE post_engagements SET created_at = to_timestamp(('x' || substr(replace(engagement_id::text, '-', ''), 1, 12))::bit(48)::bigint / 1000.0)
WHERE created_at < '2000-01-01';

-- Posts also never had a slug set, which is required and unique: give any such post one from its id.
UPDATE posts SET slug = 'post-' || right(replace(post_id::text, '-', ''), 8) WHERE slug = '';

-- A like or dislike is one user's reaction to one post or comment, so there can only be one of the two
-- per user. Views stay as they are: every view is its own row.
-- Earlier rows could pile up (the same user liking twice, or liking then disliking); keep each user's latest.
DELETE FROM post_engagements e
USING post_engagements n
WHERE e.type IN ('LIKE', 'DISLIKE')
  AND n.type IN ('LIKE', 'DISLIKE')
  AND e.asset_type = n.asset_type
  AND e.asset_id = n.asset_id
  AND e.user_id = n.user_id
  AND (e.created_at, e.engagement_id) < (n.created_at, n.engagement_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_post_engagements_one_reaction
    ON post_engagements (asset_type, asset_id, user_id)
    WHERE type IN ('LIKE', 'DISLIKE');

-- +goose Down
-- The timestamps repaired and the rows removed above can't be brought back.
DROP INDEX IF EXISTS idx_post_engagements_one_reaction;
