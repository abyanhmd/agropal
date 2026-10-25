package persistence

import (
	"database/sql"
	"errors"
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

// postSelectColumns computes like/dislike/view/comment counts per post on
// read instead of maintaining denormalized counter columns - posts/
// post_comments store no such columns, so these are always derived from
// post_engagements and post_comments directly.
const postSelectColumns = `
	p.*,
	(SELECT COUNT(*) FROM post_engagements WHERE asset_type = 'POST' AND asset_id = p.post_id AND type = 'LIKE') AS like_count,
	(SELECT COUNT(*) FROM post_engagements WHERE asset_type = 'POST' AND asset_id = p.post_id AND type = 'DISLIKE') AS dislike_count,
	(SELECT COUNT(*) FROM post_engagements WHERE asset_type = 'POST' AND asset_id = p.post_id AND type = 'VIEW') AS view_count,
	(SELECT COUNT(*) FROM post_comments WHERE post_id = p.post_id AND parent_id IS NULL) AS comment_count,
	COALESCE((SELECT u.name FROM users u WHERE u.user_id = p.user_id), '') AS author_name`

func (r *postRepo) FindPostByID(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.Post, error) {
	query := `SELECT ` + postSelectColumns + ` FROM posts p WHERE p.post_id = $1;`

	var modelPost PostModel
	if err := q.GetContext(ctx, &modelPost, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dPost.Post{}, dPost.ErrNotFound
		}
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return dPost.Post{}, err
	}

	return modelPost.ToDomain(), nil
}

func (r *postRepo) FindPostBySlug(ctx context.Context, q database.Querier, slug string) (dPost.Post, error) {
	query := `SELECT ` + postSelectColumns + ` FROM posts p WHERE p.slug = $1;`

	var modelPost PostModel
	if err := q.GetContext(ctx, &modelPost, query, slug); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return dPost.Post{}, err
	}

	return modelPost.ToDomain(), nil
}

// orderBy builds an ORDER BY clause from a client's sort. The column is checked against the
// allowed set and the direction against asc/desc; anything else falls back to the default, so
// request input never reaches the SQL text. tiebreak keeps paging stable between equal values.
func orderBy(sort *shared.SortQuery, allowed map[string]struct{}, prefix, tiebreak, fallback string) string {
	if sort == nil || !dPost.ValidSort(sort, allowed) {
		return " ORDER BY " + fallback
	}
	column := sort.Column
	if column == "created_at" || column == "updated_at" || column == "title" {
		column = prefix + column
	}
	direction := "ASC"
	if sort.Dir == "desc" {
		direction = "DESC"
	}
	return " ORDER BY " + column + " " + direction + ", " + tiebreak
}

// escapeLike makes user text match literally inside a LIKE pattern
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// postWhere builds the WHERE clause (with its leading space) and arguments for a post filter
func postWhere(filter dPost.FilterQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if len(filter.UserIDs) > 0 {
		placeholders := make([]string, len(filter.UserIDs))
		for i, id := range filter.UserIDs {
			args = append(args, id)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		conditions = append(conditions, `p.user_id IN (`+strings.Join(placeholders, ", ")+`)`)
	}
	if filter.Slug != nil {
		args = append(args, "%"+escapeLike(strings.ToLower(*filter.Slug))+"%")
		conditions = append(conditions, fmt.Sprintf(`LOWER(p.slug) LIKE $%d ESCAPE '\'`, len(args)))
	}
	if filter.Search != nil {
		args = append(args, "%"+escapeLike(strings.ToLower(*filter.Search))+"%")
		conditions = append(conditions, fmt.Sprintf(`(LOWER(p.title) LIKE $%[1]d ESCAPE '\' OR LOWER(p.content) LIKE $%[1]d ESCAPE '\')`, len(args)))
	}

	if len(conditions) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (r *postRepo) ListPosts(ctx context.Context, q database.Querier, filter dPost.FilterQuery) ([]dPost.Post, error) {
	where, args := postWhere(filter)
	query := `SELECT ` + postSelectColumns + ` FROM posts p` + where +
		orderBy(filter.Sorting, dPost.PostSortColumns, "p.", "p.post_id", "p.created_at DESC, p.post_id")

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

func (r *postRepo) CountPosts(ctx context.Context, q database.Querier, filter dPost.FilterQuery) (int, error) {
	where, args := postWhere(filter)

	var total int
	if err := q.GetContext(ctx, &total, `SELECT COUNT(*) FROM posts p`+where, args...); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return 0, err
	}

	return total, nil
}

func (r *postRepo) ListPostAuthors(ctx context.Context, q database.Querier) ([]dPost.PostAuthor, error) {
	query := `
		SELECT DISTINCT u.user_id, u.name
		FROM posts p
		JOIN users u ON u.user_id = p.user_id
		ORDER BY u.name, u.user_id;`

	var rows []struct {
		UserID uuid.UUID `db:"user_id"`
		Name   string    `db:"name"`
	}
	if err := q.SelectContext(ctx, &rows, query); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return nil, err
	}

	authors := make([]dPost.PostAuthor, len(rows))
	for i, row := range rows {
		authors[i] = dPost.PostAuthor{UserID: row.UserID, Name: row.Name}
	}

	return authors, nil
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

// commentSelectColumns computes like/dislike/view/reply counts per comment
// on read instead of maintaining denormalized counter columns -
// post_comments stores no such columns.
const commentSelectColumns = `
	c.*,
	(SELECT COUNT(*) FROM post_engagements WHERE asset_type = 'COMMENT' AND asset_id = c.comment_id AND type = 'LIKE') AS like_count,
	(SELECT COUNT(*) FROM post_engagements WHERE asset_type = 'COMMENT' AND asset_id = c.comment_id AND type = 'DISLIKE') AS dislike_count,
	(SELECT COUNT(*) FROM post_engagements WHERE asset_type = 'COMMENT' AND asset_id = c.comment_id AND type = 'VIEW') AS view_count,
	(SELECT COUNT(*) FROM post_comments WHERE parent_id = c.comment_id) AS reply_count,
	COALESCE((SELECT u.name FROM users u WHERE u.user_id = c.user_id), '') AS author_name`

func (r *postRepo) FindCommentByID(ctx context.Context, q database.Querier, id uuid.UUID) (dPost.PostComment, error) {
	query := `SELECT ` + commentSelectColumns + ` FROM post_comments c WHERE c.comment_id = $1;`

	var model CommentModel
	if err := q.GetContext(ctx, &model, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dPost.PostComment{}, dPost.ErrCommentNotFound
		}
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return dPost.PostComment{}, err
	}

	return model.ToDomain(), nil
}

func (r *postRepo) ListPostComments(ctx context.Context, q database.Querier, postID uuid.UUID) ([]dPost.PostComment, error) {
	query := `SELECT ` + commentSelectColumns + ` FROM post_comments c WHERE c.post_id = $1 ORDER BY c.created_at, c.comment_id;`

	var models []CommentModel
	if err := q.SelectContext(ctx, &models, query, postID); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return nil, err
	}

	comments := make([]dPost.PostComment, len(models))
	for i, m := range models {
		comments[i] = m.ToDomain()
	}

	return comments, nil
}

// commentWhere builds the conditions shared by the comment listings: the scope (bound to $1)
// plus an optional author filter
func commentWhere(scope string, id uuid.UUID, filter dPost.FilterQuery) (string, []interface{}) {
	conditions := []string{scope}
	args := []interface{}{id}

	if len(filter.UserIDs) > 0 {
		placeholders := make([]string, len(filter.UserIDs))
		for i, userID := range filter.UserIDs {
			args = append(args, userID)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		conditions = append(conditions, `c.user_id IN (`+strings.Join(placeholders, ", ")+`)`)
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (r *postRepo) ListTopLevelComments(ctx context.Context, q database.Querier, postID uuid.UUID, filter dPost.FilterQuery) ([]dPost.PostComment, error) {
	where, args := commentWhere(`c.post_id = $1 AND c.parent_id IS NULL`, postID, filter)
	query := `SELECT ` + commentSelectColumns + ` FROM post_comments c` + where +
		orderBy(filter.Sorting, dPost.CommentSortColumns, "c.", "c.comment_id", "c.created_at DESC, c.comment_id")

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
	where, args := commentWhere(`c.parent_id = $1`, parentID, filter)
	query := `SELECT ` + commentSelectColumns + ` FROM post_comments c` + where +
		orderBy(filter.Sorting, dPost.CommentSortColumns, "c.", "c.comment_id", "like_count DESC, c.comment_id")

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
		)`
	// One like or dislike per user and asset (see idx_post_engagements_one_reaction): reacting again
	// swaps the earlier one. Views are not covered by the index, so each one is its own row.
	if e.Type == dPost.ActionLike || e.Type == dPost.ActionDislike {
		query += `
		ON CONFLICT (asset_type, asset_id, user_id) WHERE type IN ('LIKE', 'DISLIKE')
		DO UPDATE SET type = EXCLUDED.type, created_at = EXCLUDED.created_at`
	}
	query += ";"

	modelEngagement := ToEngagementModel(e)
	if _, err := q.NamedExecContext(ctx, query, modelEngagement); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *postRepo) RemoveReaction(ctx context.Context, q database.Querier, assetType dPost.PostType, assetID, userID uuid.UUID) error {
	query := `DELETE FROM post_engagements WHERE asset_type = $1 AND asset_id = $2 AND user_id = $3 AND type IN ('LIKE', 'DISLIKE');`

	if _, err := q.ExecContext(ctx, query, string(assetType), assetID, userID); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return err
	}

	return nil
}

func (r *postRepo) ListReactions(ctx context.Context, q database.Querier, userID, postID uuid.UUID) (map[uuid.UUID]dPost.ActionType, error) {
	query := `
		SELECT asset_id, type
		FROM post_engagements
		WHERE user_id = $1
		  AND type IN ('LIKE', 'DISLIKE')
		  AND (
		    (asset_type = 'POST' AND asset_id = $2)
		    OR (asset_type = 'COMMENT' AND asset_id IN (SELECT comment_id FROM post_comments WHERE post_id = $2))
		  );`

	var rows []struct {
		AssetID uuid.UUID `db:"asset_id"`
		Type    string    `db:"type"`
	}
	if err := q.SelectContext(ctx, &rows, query, userID, postID); err != nil {
		r.logger.Error(ctx, shared.AssetPost, err.Error())
		return nil, err
	}

	reactions := make(map[uuid.UUID]dPost.ActionType, len(rows))
	for _, row := range rows {
		reactions[row.AssetID] = dPost.ActionType(row.Type)
	}

	return reactions, nil
}
