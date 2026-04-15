package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	apperrs "backend/internal/errors"
	model "backend/internal/models"
	"backend/internal/repository"
	"backend/pkg/response"
)

var (
	ErrCommentNotFound  = apperrs.ErrCommentNotFound
	ErrCannotReplyChild = apperrs.ErrCannotReplyChild
	ErrInvalidStatus    = apperrs.ErrInvalidStatus
)

type CommentService interface {
	Create(ctx context.Context, req *model.CreateCommentRequest, userID uint) (*model.CommentResponse, error)
	Update(ctx context.Context, id uint, content string) (*model.CommentResponse, error)
	Delete(ctx context.Context, id uint) error
	GetByID(ctx context.Context, id uint) (*model.CommentResponse, error)
	GetByIDWithUser(ctx context.Context, id uint) (*model.CommentResponse, error)
	ListByWorkID(ctx context.Context, workID uint, page, pageSize int) ([]model.CommentResponse, int64, error)
	ListRootByWorkID(ctx context.Context, workID uint) ([]model.CommentResponse, error)
	ListByUserID(ctx context.Context, userID uint) ([]model.CommentResponse, error)
	UpdateStatus(ctx context.Context, id uint, status model.CommentStatus) error
	IncrementLikeCount(ctx context.Context, id uint, delta int) error
}

type commentService struct {
	commentRepo repository.CommentRepository
	workRepo    repository.WorkRepository
	userRepo    repository.UserRepository
	logger      *slog.Logger
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	workRepo repository.WorkRepository,
	userRepo repository.UserRepository,
	logger *slog.Logger,
) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		workRepo:    workRepo,
		userRepo:    userRepo,
		logger:      logger,
	}
}

func (s *commentService) Create(ctx context.Context, req *model.CreateCommentRequest, userID uint) (*model.CommentResponse, error) {
	start := time.Now()
	s.logger.Info("CommentService.Create", "work_id", req.WorkID, "user_id", userID)

	if err := s.validateWorkExists(ctx, req.WorkID); err != nil {
		return nil, err
	}

	if req.ParentID != 0 {
		parentComment, err := s.commentRepo.GetByID(ctx, req.ParentID)
		if err != nil {
			if errors.Is(err, apperrs.ErrCommentNotFound) {
				return nil, fmt.Errorf("%s: %w", response.BadRequest, apperrs.ErrCommentNotFound)
			}
			return nil, fmt.Errorf("%s: %w", response.InternalError, err)
		}

		if parentComment.RootID != 0 {
			return nil, fmt.Errorf("%s: %w", response.BadRequest, apperrs.ErrCannotReplyChild)
		}

		if req.RootID == 0 {
			req.RootID = req.ParentID
		}
	}

	comment := &model.Comment{
		WorkID:   req.WorkID,
		UserID:   userID,
		ParentID: req.ParentID,
		RootID:   req.RootID,
		Content:  req.Content,
		MediaURL: req.MediaURL,
		Status:   model.CommentStatusActive,
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		s.logger.Error("Failed to create comment", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.workRepo.IncrementCount(ctx, req.WorkID, "comment_count", 1); err != nil {
		s.logger.Warn("Failed to increment work comment count", "error", err, "work_id", req.WorkID)
	}

	resp, err := s.GetByIDWithUser(ctx, comment.ID)
	if err != nil {
		return nil, err
	}

	s.logger.Info("CommentService.Create success", "comment_id", comment.ID, "duration", time.Since(start))
	return resp, nil
}

func (s *commentService) Update(ctx context.Context, id uint, content string) (*model.CommentResponse, error) {
	start := time.Now()
	s.logger.Info("CommentService.Update", "comment_id", id)

	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCommentNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCommentNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if comment.Status == model.CommentStatusDeleted {
		return nil, fmt.Errorf("%s: 评论已被删除", response.BadRequest)
	}

	comment.Content = content

	if err := s.commentRepo.Update(ctx, comment); err != nil {
		s.logger.Error("Failed to update comment", "error", err, "comment_id", id, "duration", time.Since(start))
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp, err := s.GetByIDWithUser(ctx, id)
	if err != nil {
		return nil, err
	}

	s.logger.Info("CommentService.Update success", "comment_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *commentService) Delete(ctx context.Context, id uint) error {
	start := time.Now()
	s.logger.Info("CommentService.Delete", "comment_id", id)

	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCommentNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCommentNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.commentRepo.Delete(ctx, id); err != nil {
		s.logger.Error("Failed to delete comment", "error", err, "comment_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.workRepo.IncrementCount(ctx, comment.WorkID, "comment_count", -1); err != nil {
		s.logger.Warn("Failed to decrement work comment count", "error", err, "work_id", comment.WorkID)
	}

	s.logger.Info("CommentService.Delete success", "comment_id", id, "duration", time.Since(start))
	return nil
}

func (s *commentService) GetByID(ctx context.Context, id uint) (*model.CommentResponse, error) {
	start := time.Now()

	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCommentNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCommentNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toCommentResponse(comment)

	s.logger.Info("CommentService.GetByID success", "comment_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *commentService) GetByIDWithUser(ctx context.Context, id uint) (*model.CommentResponse, error) {
	start := time.Now()

	comment, err := s.commentRepo.GetByIDWithUser(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCommentNotFound) {
			return nil, fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCommentNotFound)
		}
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	resp := s.toCommentResponse(comment)
	if comment.User.ID != 0 {
		resp.User = s.toUserResponse(&comment.User)
	}

	s.logger.Info("CommentService.GetByIDWithUser success", "comment_id", id, "duration", time.Since(start))
	return resp, nil
}

func (s *commentService) ListByWorkID(ctx context.Context, workID uint, page, pageSize int) ([]model.CommentResponse, int64, error) {
	start := time.Now()
	s.logger.Info("CommentService.ListByWorkID", "work_id", workID, "page", page, "pageSize", pageSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	comments, total, err := s.commentRepo.ListWithPagination(ctx, workID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.CommentResponse
	for _, comment := range comments {
		resp := s.toCommentResponse(&comment)
		if comment.User.ID != 0 {
			resp.User = s.toUserResponse(&comment.User)
		}
		responses = append(responses, *resp)
	}

	s.logger.Info("CommentService.ListByWorkID success", "total", total, "count", len(responses), "duration", time.Since(start))
	return responses, total, nil
}

func (s *commentService) ListRootByWorkID(ctx context.Context, workID uint) ([]model.CommentResponse, error) {
	start := time.Now()
	s.logger.Info("CommentService.ListRootByWorkID", "work_id", workID)

	comments, err := s.commentRepo.ListRootByWorkID(ctx, workID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.CommentResponse
	for _, comment := range comments {
		resp := s.toCommentResponse(&comment)
		if comment.User.ID != 0 {
			resp.User = s.toUserResponse(&comment.User)
		}

		replies, err := s.commentRepo.ListByRootID(ctx, comment.ID)
		if err != nil {
			s.logger.Warn("Failed to get replies", "error", err, "root_id", comment.ID)
		}
		for _, reply := range replies {
			replyResp := s.toCommentResponse(&reply)
			if reply.User.ID != 0 {
				replyResp.User = s.toUserResponse(&reply.User)
			}
			resp.Replies = append(resp.Replies, *replyResp)
		}

		responses = append(responses, *resp)
	}

	s.logger.Info("CommentService.ListRootByWorkID success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *commentService) ListByUserID(ctx context.Context, userID uint) ([]model.CommentResponse, error) {
	start := time.Now()
	s.logger.Info("CommentService.ListByUserID", "user_id", userID)

	comments, err := s.commentRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", response.InternalError, err)
	}

	var responses []model.CommentResponse
	for _, comment := range comments {
		responses = append(responses, *s.toCommentResponse(&comment))
	}

	s.logger.Info("CommentService.ListByUserID success", "count", len(responses), "duration", time.Since(start))
	return responses, nil
}

func (s *commentService) UpdateStatus(ctx context.Context, id uint, status model.CommentStatus) error {
	start := time.Now()
	s.logger.Info("CommentService.UpdateStatus", "comment_id", id, "status", status)

	if !s.isValidStatus(status) {
		return fmt.Errorf("%s: %w", response.BadRequest, apperrs.ErrInvalidStatus)
	}

	_, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCommentNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCommentNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.commentRepo.UpdateStatus(ctx, id, status); err != nil {
		s.logger.Error("Failed to update comment status", "error", err, "comment_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("CommentService.UpdateStatus success", "comment_id", id, "duration", time.Since(start))
	return nil
}

func (s *commentService) IncrementLikeCount(ctx context.Context, id uint, delta int) error {
	start := time.Now()
	s.logger.Info("CommentService.IncrementLikeCount", "comment_id", id, "delta", delta)

	_, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrs.ErrCommentNotFound) {
			return fmt.Errorf("%s: %w", response.UserNotFound, apperrs.ErrCommentNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	if err := s.commentRepo.IncrementLikeCount(ctx, id, delta); err != nil {
		s.logger.Error("Failed to increment like count", "error", err, "comment_id", id, "duration", time.Since(start))
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}

	s.logger.Info("CommentService.IncrementLikeCount success", "comment_id", id, "duration", time.Since(start))
	return nil
}

func (s *commentService) validateWorkExists(ctx context.Context, workID uint) error {
	_, err := s.workRepo.GetByID(ctx, workID)
	if err != nil {
		if errors.Is(err, apperrs.ErrWorkNotFound) {
			return fmt.Errorf("%s: %w", response.BadRequest, apperrs.ErrWorkNotFound)
		}
		return fmt.Errorf("%s: %w", response.InternalError, err)
	}
	return nil
}

func (s *commentService) isValidStatus(status model.CommentStatus) bool {
	switch status {
	case model.CommentStatusDeleted, model.CommentStatusActive, model.CommentStatusReviewing:
		return true
	default:
		return false
	}
}

func (s *commentService) toCommentResponse(comment *model.Comment) *model.CommentResponse {
	if comment == nil {
		return nil
	}

	return &model.CommentResponse{
		ID:        comment.ID,
		CreatedAt: comment.CreatedAt,
		WorkID:    comment.WorkID,
		UserID:    comment.UserID,
		ParentID:  comment.ParentID,
		RootID:    comment.RootID,
		Content:   comment.Content,
		MediaURL:  comment.MediaURL,
		LikeCount: comment.LikeCount,
		Status:    int8(comment.Status),
	}
}

func (s *commentService) toUserResponse(user *model.User) *model.UserResponse {
	if user == nil {
		return nil
	}

	return &model.UserResponse{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		UserType:  string(user.UserType),
		Status:    int8(user.Status),
	}
}

var _ CommentService = (*commentService)(nil)
