package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{})
	repo := NewWorkRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		UserID:  user.ID,
		Title:   "Test Work",
		Content: "Test content",
		Status:  model.WorkStatusPublished,
	}

	err := repo.Create(ctx, work)
	require.NoError(t, err)
	assert.NotZero(t, work.ID)
}

func TestWorkRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{})
	repo := NewWorkRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		UserID:  user.ID,
		Title:   "Test Work",
		Content: "Test content",
		Status:  model.WorkStatusPublished,
	}
	repo.Create(ctx, work)

	retrieved, err := repo.GetByID(ctx, work.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, work.Title, retrieved.Title)
}

func TestWorkRepository_List(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{})
	repo := NewWorkRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	for i := 1; i <= 3; i++ {
		repo.Create(ctx, &model.Work{
			UserID:  user.ID,
			Title:   "Work" + string(rune('0'+i)),
			Content: "Content",
			Status:  model.WorkStatusPublished,
		})
	}

	works, err := repo.List(ctx, "")
	require.NoError(t, err)
	assert.Len(t, works, 3)
}

func TestWorkRepository_ListByStatus(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{})
	repo := NewWorkRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	repo.Create(ctx, &model.Work{
		UserID: user.ID,
		Title:  "Published",
		Status: model.WorkStatusPublished,
	})

	works, err := repo.ListByStatus(ctx, model.WorkStatusPublished)
	require.NoError(t, err)
	assert.Len(t, works, 1)
}

func TestWorkRepository_Update(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{})
	repo := NewWorkRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		UserID: user.ID,
		Title:  "OldTitle",
		Status: model.WorkStatusPublished,
	}
	repo.Create(ctx, work)

	work.Title = "NewTitle"
	err := repo.Update(ctx, work)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, work.ID)
	assert.Equal(t, "NewTitle", retrieved.Title)
}

func TestWorkRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{})
	repo := NewWorkRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work := &model.Work{
		UserID: user.ID,
		Title:  "ToDelete",
		Status: model.WorkStatusPublished,
	}
	repo.Create(ctx, work)

	err := repo.Delete(ctx, work.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, work.ID)
	assert.Error(t, err)
}

func TestWorkRepository_WithTransaction(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{})
	repo := NewWorkRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	txRepo := repo.WithTransaction(db)
	work := &model.Work{
		UserID: user.ID,
		Title:  "TransactionWork",
		Status: model.WorkStatusPublished,
	}
	err := txRepo.Create(ctx, work)
	require.NoError(t, err)

	retrieved, _ := txRepo.GetByID(ctx, work.ID)
	assert.NotNil(t, retrieved)
}
