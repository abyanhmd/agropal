package post

import (
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	"agropal/internal/utils"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RepositoryImpl struct {
	DB     *sqlx.DB
	Logger *utils.Logger
}

func (r *RepositoryImpl) CreatePost(ctx context.Context, post dPost.Post) error {
	insertQuery := `INSERT INTO posts (
		post_id, title, content, user_id, slug, like, dislike, view
	) VALUES ( 
		:post_id, :title, :content, :user_id, :slug, :like, :dislike, :view
	);`

	modelPost := DomainToModelPost(post)
	if _, errInsert := r.DB.NamedExecContext(ctx, insertQuery, modelPost); errInsert != nil {
		r.Logger.Error(ctx, shared.System, errInsert.Error())
		return errInsert
	}

	return nil
}

func (r *RepositoryImpl) GetOnePost(ctx context.Context, postID uuid.UUID) (dPost.PostDetail, error) {
	getPostQuery := `SELECT * FROM posts WHERE post_id = $1;`

	var modelPost ModelPost
	if errGet := r.DB.GetContext(ctx, &modelPost, getPostQuery, postID); errGet != nil {
		r.Logger.Error(ctx, shared.System, errGet.Error())
		return dPost.PostDetail{}, errGet
	}
	post := modelPost.ToDomain()

	getCommentQuery := `SELECT * FROM post_comments WHERE post_id = $1;`

	var modelComments []ModelPostComment
	if errSelect := r.DB.SelectContext(ctx, &modelComments, getCommentQuery, postID); errSelect != nil {
		r.Logger.Error(ctx, shared.System, errSelect.Error())
		return dPost.PostDetail{}, errSelect
	}
	comments := make([]dPost.PostComment, len(modelComments))
	for i, m := range modelComments {
		comments[i] = m.ToDomain()
	}

	return dPost.PostDetail{
		Post:     post,
		Comments: comments,
	}, nil
}

func (r *RepositoryImpl) GetManyPosts(ctx context.Context) ([]dPost.Post, error) {
	getPostQuery := `SELECT * FROM posts`

	var modelPosts []ModelPost
	if errSelect := r.DB.SelectContext(ctx, &modelPosts, getPostQuery); errSelect != nil {
		r.Logger.Error(ctx, shared.System, errSelect.Error())
		return nil, errSelect
	}
	posts := make([]dPost.Post, len(modelPosts))
	for i, m := range modelPosts {
		posts[i] = m.ToDomain()
	}

	return posts, nil
}

func (r *RepositoryImpl) CreatePostComment(ctx context.Context, comment dPost.PostComment) error {
	insertQuery := `INSERT INTO post_comments (
		comment_id, post_id, content, user_id, like, dislike, view
	) VALUES (	          
		:comment_id, :post_id, :content, :user_id, :like, :dislike, :view
	);`

	modelComment := DomainToModelComment(comment)
	if _, errInsert := r.DB.ExecContext(ctx, insertQuery, modelComment); errInsert != nil {
		r.Logger.Error(ctx, shared.System, errInsert.Error())
		return errInsert
	}

	return nil
}

func (r *RepositoryImpl) GetPostComments(ctx context.Context, postID uuid.UUID) ([]dPost.PostComment, error) {
	selectQuery := `SELECT * FROM post_comments WHERE post_id = $1;`

	var modelComments []ModelPostComment
	if errSelect := r.DB.SelectContext(ctx, &modelComments, selectQuery, postID); errSelect != nil {
		r.Logger.Error(ctx, shared.System, errSelect.Error())
		return nil, errSelect
	}
	comments := make([]dPost.PostComment, len(modelComments))
	for i, m := range modelComments {
		comments[i] = m.ToDomain()
	}

	return comments, nil
}

func (r *RepositoryImpl) CreatePostEngagement(ctx context.Context, engagement dPost.PostEngagement) (err error) {
	tx, errTx := r.DB.BeginTxx(ctx, nil)
	if errTx != nil {
		return errTx
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		}
	}()

	insertQuery := `INSERT INTO post_engagements (
		engagement_id, post_id, user_id, type
	) VALUES (
		:engagement_id, :post_id, :user_id, :type
	);`

	var updateEngagementQuery string
	switch engagement.Type {
	case dPost.EngagementPostLike:
		updateEngagementQuery = `UPDATE posts SET "like" = "like" + 1 WHERE post_id = $1`

	case dPost.EngagementPostDislike:
		updateEngagementQuery = `UPDATE posts SET dislike = dislike + 1 WHERE post_id = $1`

	case dPost.EngagementPostView:
		updateEngagementQuery = `UPDATE posts SET view = view + 1 WHERE post_id = $1`

	case dPost.EngagementPostCommentLike:
		updateEngagementQuery = `UPDATE post_comments SET "like" = "like" + 1 WHERE comment_id = $1`

	case dPost.EngagementPostCommentDislike:
		updateEngagementQuery = `UPDATE post_comments SET dislike = dislike + 1 WHERE comment_id = $1`

	case dPost.EngagementPostCommentView:
		updateEngagementQuery = `UPDATE post_comments SET view = view + 1 WHERE comment_id = $1`

	default:
		return errors.New("invalid engagement type")
	}

	modelEngagement := DomainToModelEngagement(engagement)
	if _, errInsert := tx.NamedExecContext(ctx, insertQuery, modelEngagement); errInsert != nil {
		r.Logger.Error(ctx, shared.System, errInsert.Error())
		return errInsert
	}

	if _, errUpdate := tx.ExecContext(ctx, updateEngagementQuery, engagement.PostID); errUpdate != nil {
		r.Logger.Error(ctx, shared.System, errUpdate.Error())
		return errUpdate
	}

	if errCommit := tx.Commit(); errCommit != nil {
		r.Logger.Error(ctx, shared.System, errCommit.Error())
		return errCommit
	}

	return nil
}

func NewRepository(db *sqlx.DB, logger *utils.Logger) dPost.Repository {
	return &RepositoryImpl{
		DB:     db,
		Logger: logger,
	}
}
