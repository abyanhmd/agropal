package post

import (
	"agropal/config"
	dNote "agropal/internal/domain/note"
	dPost "agropal/internal/domain/post"
	"agropal/internal/domain/shared"
	dUser "agropal/internal/domain/user"
	"agropal/internal/infra/database"
	"context"
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

func (s *Service) GetOnePost(ctx context.Context, postID uuid.UUID, filter dPost.FilterQuery) (dPost.PostDetail, error) {
	var postDetail dPost.PostDetail
	err := s.transactor.WithTx(ctx, func(q database.Querier) error {
		post, err := s.postRepo.FindPostByID(ctx, q, postID)
		if err != nil {
			return err
		}
		comments, err := s.postRepo.ListCommentReplies(ctx, q, postID, filter)
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
	return s.postRepo.ListPosts(ctx, s.transactor.DB(), filter)
}

func (s *Service) CreatePostComment(ctx context.Context, cmd PostCommentCommand) error {
	userID, errParse := shared.ExtractUserID(ctx)
	if errParse != nil {
		return errParse
	}
	comment := cmd.ToDomain(userID)

	commentID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}
	comment.PostID = commentID

	return s.postRepo.AddComment(ctx, s.transactor.DB(), comment)
}

func (s *Service) CreatePostEngagement(ctx context.Context, postType dPost.PostType, actionType dPost.ActionType, postID uuid.UUID) error {
	userID, errParse := shared.ExtractUserID(ctx)
	if errParse != nil {
		return errParse
	}

	engagementID, errGen := uuid.NewV7()
	if errGen != nil {
		return errGen
	}

	return s.postRepo.AddEngagement(ctx, s.transactor.DB(), dPost.PostEngagement{
		ID:        engagementID,
		AssetID:   postID,
		AssetType: postType,
		UserID:    userID,
		Type:      actionType,
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
	isAdmin := requesterRole == string(config.RoleAdmin)

	if !isSelf && !isAdmin {
		return dUser.User{}, shared.ErrForbidden
	}

	return targetUser, nil
}
