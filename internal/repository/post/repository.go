package post

import (
	"agropal/internal/database"
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type RepositoryImpl struct {
	Logger *utils.Logger
}

func NewRepository(logger *utils.Logger) dPost.Repository {
	return &RepositoryImpl{Logger: logger}
}

func (r *RepositoryImpl) Create(ctx context.Context, q database.Querier, p dPost.Post) error {
	query := `
		INSERT INTO posts (
			post_id, user_id, 
			title, content,
			slug, created_at
		) VALUES (
			:post_id, :user_id, 
			:title, :content,
			:slug, :created_at
		);`

	modelPost := ToPostModel(p)
	if _, err := q.NamedExecContext(ctx, query, modelPost); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) Update(ctx context.Context, q database.Querier, p dPost.Post) error {
	query := `
		UPDATE posts SET
			title = :title,
			content = :content,
			updated_at = :updated_at
		WHERE post_id = :post_id;`

	modelPost := ToPostModel(p)
	if _, err := q.NamedExecContext(ctx, query, modelPost); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) Delete(ctx context.Context, q database.Querier, id uuid.UUID) error {
	query := `DELETE FROM posts WHERE post_id = $1;`

	if _, err := q.ExecContext(ctx, query, id); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) FindByID(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.Post, error) {
	query := `SELECT * FROM posts WHERE post_id = $1;`

	var modelPost PostModel
	if err := q.GetContext(ctx, &modelPost, query, id); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return dPost.Post{}, err
	}

	return modelPost.ToDomain(), nil
}

func (r *RepositoryImpl) FindBySlug(ctx context.Context, q database.Querier, slug string) (dPost.Post, error) {
	query := `SELECT * FROM posts WHERE slug = $1;`

	var modelPost PostModel
	if err := q.GetContext(ctx, &modelPost, query, slug); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return dPost.Post{}, err
	}

	return modelPost.ToDomain(), nil
}

func (r *RepositoryImpl) List(ctx context.Context, q database.Querier, limit, offset int) ([]dPost.Post, error) {
	query := `SELECT * FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2;`

	var models []PostModel
	if err := q.SelectContext(ctx, &models, query, limit, offset); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return nil, err
	}

	posts := make([]dPost.Post, len(models))
	for i, m := range models {
		posts[i] = m.ToDomain()
	}

	return posts, nil
}

func (r *RepositoryImpl) AddComment(ctx context.Context, q database.Querier, c dPost.PostComment) error {
	query := `
		INSERT INTO post_comments (
			comment_id, post_id, 
			user_id, parent_id, 
			content, created_at
		) VALUES (
			:comment_id, :post_id,
			:user_id, :parent_id,
			:content, :created_at
		);`

	modelComment := ToCommentModel(c)
	if _, err := q.NamedExecContext(ctx, query, modelComment); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) ListTopLevel(ctx context.Context, q database.Querier, postID uuid.UUID, limit, offset int) ([]dPost.PostComment, error) {
	query := `
		SELECT * 
		FROM post_comments 
        WHERE post_id = $1 
			AND parent_id IS NULL 
        ORDER BY created_at DESC 
        LIMIT $2 OFFSET $3;`

	var models []CommentModel
	if err := q.SelectContext(ctx, &models, query, postID, limit, offset); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return nil, err
	}

	comments := make([]dPost.PostComment, len(models))
	for i, m := range models {
		comments[i] = m.ToDomain()
	}

	return comments, nil
}

func (r *RepositoryImpl) ListReplies(ctx context.Context, q database.Querier, parentID uuid.UUID, limit, offset int) ([]dPost.PostComment, error) {
	query := `
		SELECT * 
		FROM post_comments 
        WHERE parent_id = $1
        ORDER BY like_count DESC 
		LIMIT $2 OFFSET $3;`

	var models []CommentModel
	if err := q.SelectContext(ctx, &models, query, parentID, limit, offset); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return nil, err
	}

	comments := make([]dPost.PostComment, len(models))
	for i, m := range models {
		comments[i] = m.ToDomain()
	}

	return comments, nil
}

func (r *RepositoryImpl) AddEngagement(ctx context.Context, q database.Querier, e dPost.PostEngagement) error {
	query := `
		INSERT INTO post_engagements (
			engagement_id, asset_id,
			asset_type, user_id,
			type, created_at
		) VALUES (
			:engagement_id, :asset_id,
			:asset_type, :user_id,
			:type, :created_at
		);`

	modelEngagement := ToEngagementModel(e)
	if _, err := q.NamedExecContext(ctx, query, modelEngagement); err != nil {
		r.Logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *RepositoryImpl) SyncCounts(ctx context.Context, q database.Querier, assetType dPost.PostType, assetID uuid.UUID) error {
	var tableName, idColumn, extraCountColumn, extraCountQuery string
	var assetTypeStr shared.AssetType

	switch assetType {
	case dPost.TypePost:
		assetTypeStr = shared.AssetPost
		tableName = "posts"
		idColumn = "post_id"
		extraCountColumn = "comment_count"
		extraCountQuery = "(SELECT COUNT(*) FROM post_comments WHERE post_id = $1 AND parent_id IS NULL)"
	case dPost.TypeComment:
		assetTypeStr = shared.AssetComment
		tableName = "post_comments"
		idColumn = "comment_id"
		extraCountColumn = "reply_count"
		extraCountQuery = "(SELECT COUNT(*) FROM post_comments WHERE parent_id = $1)"
	default:
		return dPost.ErrInvalidAssetType
	}

	query := fmt.Sprintf(`
        WITH engagement_stats AS (
            SELECT 
                COUNT(*) FILTER (WHERE type = 'LIKE') as likes,
                COUNT(*) FILTER (WHERE type = 'DISLIKE') as dislikes,
                COUNT(*) FILTER (WHERE type = 'VIEW') as views
            FROM post_engagements 
            WHERE asset_type = $2 AND asset_id = $1
        )
        UPDATE %s SET 
            like_count = stats.likes,
            dislike_count = stats.dislikes,
            view_count = stats.views,
            %s = %s
        FROM engagement_stats stats
        WHERE %s = $1`, tableName, extraCountColumn, extraCountQuery, idColumn)

	if _, err := q.ExecContext(ctx, query, assetID, assetType); err != nil {
		r.Logger.Error(ctx, assetTypeStr, err.Error())
		return err
	}

	return nil
}
