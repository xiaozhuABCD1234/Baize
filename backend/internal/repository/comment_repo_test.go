package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Comment{})
	repo := NewCommentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	comment := &model.Comment{
		WorkID:  work.ID,
		UserID:  user.ID,
		Content: "Test comment",
		Status:  model.CommentStatusActive,
	}

	err := repo.Create(ctx, comment)
	require.NoError(t, err)
	assert.NotZero(t, comment.ID)
}

func TestCommentRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Comment{})
	repo := NewCommentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	comment := &model.Comment{
		WorkID:  work.ID,
		UserID:  user.ID,
		Content: "Test comment",
		Status:  model.CommentStatusActive,
	}
	repo.Create(ctx, comment)

	retrieved, err := repo.GetByID(ctx, comment.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, comment.Content, retrieved.Content)
}

func TestCommentRepository_ListByWorkID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Comment{})
	repo := NewCommentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	for i := 1; i <= 3; i++ {
		repo.Create(ctx, &model.Comment{
			WorkID:  work.ID,
			UserID:  user.ID,
			Content: "Comment" + string(rune('0'+i)),
			Status:  model.CommentStatusActive,
		})
	}

	comments, err := repo.ListByWorkID(ctx, work.ID)
	require.NoError(t, err)
	assert.Len(t, comments, 3)
}

func TestCommentRepository_ListRootByWorkID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Comment{})
	repo := NewCommentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	repo.Create(ctx, &model.Comment{
		WorkID:   work.ID,
		UserID:   user.ID,
		Content:  "Root comment",
		ParentID: 0,
		Status:   model.CommentStatusActive,
	})

	comments, err := repo.ListRootByWorkID(ctx, work.ID)
	require.NoError(t, err)
	assert.Len(t, comments, 1)
}

func TestCommentRepository_Update(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Comment{})
	repo := NewCommentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	comment := &model.Comment{
		WorkID:  work.ID,
		UserID:  user.ID,
		Content: "OldContent",
		Status:  model.CommentStatusActive,
	}
	repo.Create(ctx, comment)

	comment.Content = "NewContent"
	err := repo.Update(ctx, comment)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, comment.ID)
	assert.Equal(t, "NewContent", retrieved.Content)
}

func TestCommentRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Comment{})
	repo := NewCommentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	comment := &model.Comment{
		WorkID:  work.ID,
		UserID:  user.ID,
		Content: "ToDelete",
		Status:  model.CommentStatusActive,
	}
	repo.Create(ctx, comment)

	err := repo.Delete(ctx, comment.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, comment.ID)
	assert.Error(t, err)
}

func TestCommentRepository_WithTransaction(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Comment{})
	repo := NewCommentRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	txRepo := repo.WithTransaction(db)
	comment := &model.Comment{
		WorkID:  work.ID,
		UserID:  user.ID,
		Content: "TransactionComment",
		Status:  model.CommentStatusActive,
	}
	err := txRepo.Create(ctx, comment)
	require.NoError(t, err)

	retrieved, _ := txRepo.GetByID(ctx, comment.ID)
	assert.NotNil(t, retrieved)
}
