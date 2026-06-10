package persistence

import (
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	"agropal/internal/infra/database"
	"agropal/pkg"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type postRepo struct {
	logger *pkg.Logger
}

func NewPostRepository(l *pkg.Logger) dPost.Repository {
	return &postRepo{logger: l}
}

func (r *postRepo) CreatePost(ctx context.Context, q database.Querier, p dPost.Post) error {
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
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *postRepo) UpdatePost(ctx context.Context, q database.Querier, p dPost.Post) error {
	query := `
		UPDATE posts SET
			title = :title,
			content = :content,
			updated_at = :updated_at
		WHERE post_id = :post_id;`

	modelPost := ToPostModel(p)
	if _, err := q.NamedExecContext(ctx, query, modelPost); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *postRepo) DeletePost(ctx context.Context, q database.Querier, id uuid.UUID) error {
	query := `DELETE FROM posts WHERE post_id = $1;`

	if _, err := q.ExecContext(ctx, query, id); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *postRepo) FindPostByID(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.Post, error) {
	query := `SELECT * FROM posts WHERE post_id = $1;`

	var modelPost PostModel
	if err := q.GetContext(ctx, &modelPost, query, id); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return dPost.Post{}, err
	}

	return modelPost.ToDomain(), nil
}

func (r *postRepo) FindPostBySlug(ctx context.Context, q database.Querier, slug string) (dPost.Post, error) {
	query := `SELECT * FROM posts WHERE slug = $1;`

	var modelPost PostModel
	if err := q.GetContext(ctx, &modelPost, query, slug); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return dPost.Post{}, err
	}

	return modelPost.ToDomain(), nil
}

func (r *postRepo) ListPosts(ctx context.Context, q database.Querier, filter dPost.FilterQuery) ([]dPost.Post, error) {
	query := `SELECT * FROM posts`
	args := []interface{}{}

	var conditions []string
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf(`user_id = $%d`, len(args)+1))
		args = append(args, *filter.UserID)
	}
	if filter.Slug != nil {
		conditions = append(conditions, fmt.Sprintf(`LOWER(slug) LIKE $%d`, len(args)+1))
		lowerSearch := strings.ToLower(*filter.Slug)
		args = append(args, fmt.Sprintf("%%%s%%", lowerSearch))
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if filter.Sorting != nil {
		query += fmt.Sprintf(` ORDER BY %s %s`, filter.Sorting.Column, filter.Sorting.Dir)
	} else {
		query += ` ORDER BY created_at DESC`
	}

	if filter.Pagination != nil {
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, filter.Pagination.Limit, filter.Pagination.Offset)
	}

	var models []PostModel
	if err := q.SelectContext(ctx, &models, query, args...); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return nil, err
	}

	posts := make([]dPost.Post, len(models))
	for i, m := range models {
		posts[i] = m.ToDomain()
	}

	return posts, nil
}

func (r *postRepo) AddComment(ctx context.Context, q database.Querier, c dPost.PostComment) error {
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
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *postRepo) ListTopLevelComments(ctx context.Context, q database.Querier, postID uuid.UUID, filter dPost.FilterQuery) ([]dPost.PostComment, error) {
	query := `SELECT * FROM post_comments`

	args := []interface{}{postID}
	condition := []string{
		`post_id = $1`,
		`parent_id IS NULL`,
	}

	if filter.UserID != nil {
		condition = append(condition, fmt.Sprintf(` user_id = $%d`, len(args)+1))
		args = append(args, *filter.UserID)
	}
	if filter.Slug != nil {
		condition = append(condition, fmt.Sprintf(` LOWER(slug) LIKE $%d`, len(args)+1))
		lowerSearch := strings.ToLower(*filter.Slug)
		args = append(args, fmt.Sprintf("%%%s%%", lowerSearch))
	}
	if len(condition) > 0 {
		query += " WHERE " + strings.Join(condition, " AND ")
	}

	if filter.Sorting != nil {
		query += fmt.Sprintf(` ORDER BY %s %s`, filter.Sorting.Column, filter.Sorting.Dir)
	} else {
		query += ` ORDER BY created_at DESC`
	}
	if filter.Pagination != nil {
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, filter.Pagination.Limit, filter.Pagination.Offset)
	}

	var models []CommentModel
	if err := q.SelectContext(ctx, &models, query, args...); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return nil, err
	}

	comments := make([]dPost.PostComment, len(models))
	for i, m := range models {
		comments[i] = m.ToDomain()
	}

	return comments, nil
}

func (r *postRepo) ListCommentReplies(ctx context.Context, q database.Querier, parentID uuid.UUID, filter dPost.FilterQuery) ([]dPost.PostComment, error) {
	query := `SELECT * FROM post_comments`

	args := []interface{}{parentID}
	condition := []string{
		`parent_id = $1`,
	}

	if filter.UserID != nil {
		condition = append(condition, fmt.Sprintf(` user_id = $%d`, len(args)+1))
		args = append(args, *filter.UserID)
	}
	if filter.Slug != nil {
		condition = append(condition, fmt.Sprintf(` LOWER(slug) LIKE $%d`, len(args)+1))
		lowerSearch := strings.ToLower(*filter.Slug)
		args = append(args, fmt.Sprintf("%%%s%%", lowerSearch))
	}
	if len(condition) > 0 {
		query += " WHERE " + strings.Join(condition, " AND ")
	}

	if filter.Sorting != nil {
		query += fmt.Sprintf(` ORDER BY %s %s`, filter.Sorting.Column, filter.Sorting.Dir)
	} else {
		query += ` ORDER BY like_count DESC`
	}
	if filter.Pagination != nil {
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, filter.Pagination.Limit, filter.Pagination.Offset)
	}

	var models []CommentModel
	if err := q.SelectContext(ctx, &models, query, args...); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return nil, err
	}

	comments := make([]dPost.PostComment, len(models))
	for i, m := range models {
		comments[i] = m.ToDomain()
	}

	return comments, nil
}

func (r *postRepo) AddEngagement(ctx context.Context, q database.Querier, e dPost.PostEngagement) error {
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
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *postRepo) SyncPostEngagements(ctx context.Context, q database.Querier, assetType dPost.PostType, assetID uuid.UUID) error {
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
		r.logger.Error(ctx, assetTypeStr, err.Error())
		return err
	}

	return nil
}
