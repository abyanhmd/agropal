package post

import (
	"agropal/config"
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

// CreatePost returns the new post, whose slug is its address
func (s *Service) CreatePost(ctx context.Context, cmd PostCommand) (dPost.Post, error) {
	user, err := s.verifyAccess(ctx, cmd.UserID)
	if err != nil {
		return dPost.Post{}, err
	}

	postID, err := uuid.NewV7()
	if err != nil {
		return dPost.Post{}, err
	}
	nowUTC := time.Now().UTC()
	post := cmd.ToDomain(postID, user.ID)
	post.Slug = slugFor(post.Title, postID)
	post.CreatedAt = nowUTC

	noteID, err := uuid.NewV7()
	if err != nil {
		return dPost.Post{}, err
	}
	createNote := dNote.Note{
		ID:         noteID,
		CreatedAt:  nowUTC,
		Type:       dNote.TypeCreate,
		AssetType:  shared.AssetPost,
		AssetID:    postID,
		AuthorID:   user.ID,
		AuthorName: user.Account.Name,
	}

	err = s.transactor.WithTx(ctx, func(q database.Querier) error {
		if errPost := s.postRepo.CreatePost(ctx, q, post); errPost != nil {
			return errPost
		}
		if errNote := s.noteRepo.CreateNote(ctx, q, createNote); errNote != nil {
			return errNote
		}
		return nil
	})
	if err != nil {
		return dPost.Post{}, err
	}

	return post, nil
}

func (s *Service) GetOnePost(ctx context.Context, postID uuid.UUID) (dPost.PostDetail, error) {
	var postDetail dPost.PostDetail
	err := s.transactor.WithTx(ctx, func(q database.Querier) error {
		post, err := s.postRepo.FindPostByID(ctx, q, postID)
		if err != nil {
			return err
		}
		comments, err := s.postRepo.ListPostComments(ctx, q, postID)
		if err != nil {
			return err
		}

		// Reactions are the requester's own; without a signed-in user there are none to show
		if userID, errUser := shared.ExtractUserID(ctx); errUser == nil {
			reactions, errReactions := s.postRepo.ListReactions(ctx, q, userID, postID)
			if errReactions != nil {
				return errReactions
			}
			post.ViewerReaction = reactions[post.ID]
			for i := range comments {
				comments[i].ViewerReaction = reactions[comments[i].ID]
			}
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

// CountPosts counts the posts a filter matches, ignoring its pagination, so a client can page through them.
func (s *Service) CountPosts(ctx context.Context, filter dPost.FilterQuery) (int, error) {
	return s.postRepo.CountPosts(ctx, s.transactor.DB(), filter)
}

func (s *Service) GetPostAuthors(ctx context.Context) ([]dPost.PostAuthor, error) {
	return s.postRepo.ListPostAuthors(ctx, s.transactor.DB())
}

func (s *Service) CreatePostComment(ctx context.Context, cmd PostCommentCommand) (uuid.UUID, error) {
	userID, errParse := shared.ExtractUserID(ctx)
	if errParse != nil {
		return uuid.Nil, errParse
	}

	// A reply must answer a comment on the same post
	if cmd.ParentID != nil {
		parent, err := s.postRepo.FindCommentByID(ctx, s.transactor.DB(), *cmd.ParentID)
		if err != nil {
			if errors.Is(err, dPost.ErrCommentNotFound) {
				return uuid.Nil, dPost.ErrInvalidParentID
			}
			return uuid.Nil, err
		}
		if parent.PostID != cmd.PostID {
			return uuid.Nil, dPost.ErrInvalidParentID
		}
	}
	comment := cmd.ToDomain(userID)

	commentID, errGen := uuid.NewV7()
	if errGen != nil {
		return uuid.Nil, errGen
	}
	comment.ID = commentID
	comment.CreatedAt = time.Now().UTC()

	if err := s.postRepo.AddComment(ctx, s.transactor.DB(), comment); err != nil {
		return uuid.Nil, err
	}

	return commentID, nil
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
		CreatedAt: time.Now().UTC(),
	})
}

// RemovePostEngagement takes back the requester's like or dislike on a post or comment.
func (s *Service) RemovePostEngagement(ctx context.Context, postType dPost.PostType, assetID uuid.UUID) error {
	userID, errParse := shared.ExtractUserID(ctx)
	if errParse != nil {
		return errParse
	}

	return s.postRepo.RemoveReaction(ctx, s.transactor.DB(), postType, assetID, userID)
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
