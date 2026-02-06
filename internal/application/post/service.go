package post

import (
	dPost "agropal/internal/domain/post"
	"agropal/internal/utils"
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	PostRepository dPost.Repository
}

func NewService(postRepository dPost.Repository) Service {
	return Service{
		PostRepository: postRepository,
	}
}

func (s *Service) CreatePost(ctx context.Context, cmd PostCommand) error {
	userID, errParse := utils.GetUserIDFromContext(ctx)
	if errParse != nil {
		return errParse
	}
	post := cmd.ToDomain(userID)

	postID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}
	post.PostID = postID
	post.Like = 0
	post.Dislike = 0
	post.View = 0

	// TODO: Validate post input

	return s.PostRepository.CreatePost(ctx, post)
}

func (s *Service) GetOnePost(ctx context.Context, postIDStr string) (dPost.PostDetail, error) {
	postID, errParse := uuid.Parse(postIDStr)
	if errParse != nil {
		return dPost.PostDetail{}, errParse
	}

	return s.PostRepository.GetOnePost(ctx, postID)
}

func (s *Service) GetManyPosts(ctx context.Context) ([]dPost.Post, error) {
	return s.PostRepository.GetManyPosts(ctx)
}

func (s *Service) CreatePostComment(ctx context.Context, cmd PostCommentCommand) error {
	userID, errParse := utils.GetUserIDFromContext(ctx)
	if errParse != nil {
		return errParse
	}
	comment := cmd.ToDomain(userID)

	commentID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}
	comment.PostID = commentID

	return s.PostRepository.CreatePostComment(ctx, comment)
}

func (s *Service) GetPostComments(ctx context.Context, postIDStr string) ([]dPost.PostComment, error) {
	postID, errParse := uuid.Parse(postIDStr)
	if errParse != nil {
		return nil, errParse
	}

	return s.PostRepository.GetPostComments(ctx, postID)
}

func (s *Service) CreatePostEngagement(ctx context.Context, engagementTypeStr, postIDStr string) error {
	userID, errParse := utils.GetUserIDFromContext(ctx)
	if errParse != nil {
		return errParse
	}

	postID, errParse := uuid.Parse(postIDStr)
	if errParse != nil {
		return errParse
	}
	var engagementType dPost.EngagementType
	switch engagementTypeStr {
	case "post_like":
		engagementType = dPost.EngagementPostLike
	case "post_dislike":
		engagementType = dPost.EngagementPostDislike
	case "post_view":
		engagementType = dPost.EngagementPostView
	case "post_comment_like":
		engagementType = dPost.EngagementPostCommentLike
	case "post_comment_dislike":
		engagementType = dPost.EngagementPostCommentDislike
	case "post_comment_view":
		engagementType = dPost.EngagementPostCommentView
	default:
		return errors.New("invalid engagement type")
	}

	engagementID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}

	return s.PostRepository.CreatePostEngagement(ctx, dPost.PostEngagement{
		EngagementID: engagementID,
		PostID:       postID,
		UserID:       userID,
		Type:         engagementType,
	})
}
