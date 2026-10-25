package post

import (
	dPost "agropal/internal/domain/post"
	"agropal/internal/response"
)

func DomainToPostResponse(p dPost.Post) response.PostResponse {
	return response.PostResponse{
		PostID:         p.ID,
		UserID:         p.UserID,
		Title:          p.Title,
		Content:        p.Content,
		Slug:           p.Slug,
		LikeCount:      p.LikeCount,
		DislikeCount:   p.DislikeCount,
		ViewCount:      p.ViewCount,
		CommentCount:   p.CommentCount,
		AuthorName:     p.AuthorName,
		ViewerReaction: string(p.ViewerReaction),
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

func DomainToPostCommentResponse(c dPost.PostComment) response.PostCommentResponse {
	return response.PostCommentResponse{
		CommentID:      c.ID,
		PostID:         c.PostID,
		UserID:         c.UserID,
		ParentID:       c.ParentID,
		Content:        c.Content,
		ReplyCount:     c.ReplyCount,
		LikeCount:      c.LikeCount,
		DislikeCount:   c.DislikeCount,
		ViewCount:      c.ViewCount,
		AuthorName:     c.AuthorName,
		ViewerReaction: string(c.ViewerReaction),
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

func DomainToPostDetailResponse(d dPost.PostDetail) response.PostDetailResponse {
	comments := make([]response.PostCommentResponse, len(d.Comments))
	for i, c := range d.Comments {
		comments[i] = DomainToPostCommentResponse(c)
	}

	return response.PostDetailResponse{
		Post:     DomainToPostResponse(d.Post),
		Comments: comments,
	}
}
