package post

import (
	"agropal/internal/api/middleware"
	dAuth "agropal/internal/domain/auth"
	dNote "agropal/internal/domain/note"
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	transactor database.Transactor
	userRepo   dUser.Repository
	postRepo   dPost.Repository
	noteRepo   dNote.Repository
}

func NewService(
	t database.Transactor,
	ur dUser.Repository,
	pr dPost.Repository,
	nr dNote.Repository,
) *Service {
	return &Service{
		transactor: t,
		userRepo:   ur,
		postRepo:   pr,
		noteRepo:   nr,
	}
}

func (s *Service) CreatePost(ctx context.Context, cmd PostCommand) error {
	user, err := s.verifyAccess(ctx, cmd.UserID)
	if err != nil {
		return err
	}

	postID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	post := cmd.ToDomain(postID, user.ID)

	noteID, err := uuid.NewV7()
	if err != nil {
		return err
	}
	createNote := dNote.Note{
		ID:         noteID,
		CreatedAt:  time.Now().UTC(),
		Type:       dNote.TypeCreate,
		AssetType:  shared.AssetPost,
		AssetID:    postID,
		AuthorID:   user.ID,
		AuthorName: user.Account.Name,
	}

	return s.transactor.WithTx(ctx, func(q database.Querier) error {
		if errPost := s.postRepo.CreatePost(ctx, q, post); errPost != nil {
			return errPost
		}
		if errNote := s.noteRepo.CreateNote(ctx, q, createNote); errNote != nil {
			return errNote
		}
		return nil
	})
}

func (s *Service) GetOnePost(ctx context.Context, postID uuid.UUID) (dPost.PostDetail, error) {
	var postDetail dPost.PostDetail
	err := s.transactor.WithTx(ctx, func(q database.Querier) error {
		post, err := s.postRepo.FindPostByID(ctx, s.transactor.DB(), postID)
		if err != nil {
			return err
		}
		comments, err := s.postRepo.ListCommentReplies(ctx, q, postID, 10, 0)
		if err != nil {
			return err
		}

		postDetail.Post = post
		postDetail.Comments = comments

		return nil
	})
	if err != nil {
		return dPost.PostDetail{}, err
	}

	return postDetail, nil
}

func (s *Service) GetManyPosts(ctx context.Context, filter dPost.FilterQuery) ([]dPost.Post, error) {
	return s.postRepo.ListPosts(ctx, s.transactor.DB(), 10, 0)
}

func (s *Service) CreatePostComment(ctx context.Context, cmd PostCommentCommand) error {
	userID, errParse := middleware.ExtractUserID(ctx)
	if errParse != nil {
		return errParse
	}
	comment := cmd.ToDomain(userID)

	commentID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}
	comment.PostID = commentID

	return s.postRepo.CreatePostComment(ctx, comment)
}

func (s *Service) GetPostComments(ctx context.Context, postIDStr string) ([]dPost.PostComment, error) {
	postID, errParse := uuid.Parse(postIDStr)
	if errParse != nil {
		return nil, errParse
	}

	return s.postRepo.GetPostComments(ctx, postID)
}

func (s *Service) CreatePostEngagement(ctx context.Context, engagementTypeStr, postIDStr string) error {
	userID, errParse := middleware.ExtractUserID(ctx)
	if errParse != nil {
		return errParse
	}

	postID, errParse := uuid.Parse(postIDStr)
	if errParse != nil {
		return errParse
	}
	var engagementType dPost.ActionType
	switch engagementTypeStr {
	case "post_like":
		engagementType = dPost.ActionLike
	case "post_dislike":
		engagementType = dPost.ActionDislike
	case "post_view":
		engagementType = dPost.ActionView
	case "post_comment_like":
		engagementType = dPost.EngagementCommentLike
	case "post_comment_dislike":
		engagementType = dPost.EngagementCommentDislike
	case "post_comment_view":
		engagementType = dPost.EngagementCommentSeen
	default:
		return errors.New("invalid engagement type")
	}

	engagementID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}

	return s.postRepo.CreatePostEngagement(ctx, dPost.PostEngagement{
		ID:     engagementID,
		PostID: postID,
		UserID: userID,
		Type:   engagementType,
	})
}

/* ==================== Helper ====================== */

func (s *Service) verifyAccess(ctx context.Context, targetID uuid.UUID) (dUser.User, error) {
	requesterID, err := shared.ExtractUserID(ctx)
	if err != nil {
		return dUser.User{}, dUser.ErrInvalidUserID
	}

	// Default to self if target is Nil
	if targetID == uuid.Nil {
		targetID = requesterID
	}

	targetUser, err := s.userRepo.FindUserByID(ctx, s.transactor.DB(), targetID)
	if err != nil {
		return dUser.User{}, err
	}

	requesterRole := shared.GetMetadata(ctx, shared.KeyRole)

	isSelf := targetID == requesterID
	isAdmin := requesterRole == string(dAuth.RoleAdmin)

	if !isSelf && !isAdmin {
		return dUser.User{}, shared.ErrForbidden
	}

	return targetUser, nil
}
