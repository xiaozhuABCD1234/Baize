package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFavoriteRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Favorite{})
	repo := NewFavoriteRepository(db)
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

	favorite := &model.Favorite{
		UserID: user.ID,
		WorkID: work.ID,
	}

	err := repo.Create(ctx, favorite)
	require.NoError(t, err)
	assert.NotZero(t, favorite.ID)
}

func TestFavoriteRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Favorite{})
	repo := NewFavoriteRepository(db)
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

	favorite := &model.Favorite{
		UserID: user.ID,
		WorkID: work.ID,
	}
	repo.Create(ctx, favorite)

	retrieved, err := repo.GetByID(ctx, favorite.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestFavoriteRepository_GetByUserAndWork(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Favorite{})
	repo := NewFavoriteRepository(db)
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

	favorite := &model.Favorite{
		UserID: user.ID,
		WorkID: work.ID,
	}
	repo.Create(ctx, favorite)

	retrieved, err := repo.GetByUserAndWork(ctx, user.ID, work.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestFavoriteRepository_ListByUserID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Favorite{})
	repo := NewFavoriteRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	work1 := &model.Work{Title: "Work1", Status: model.WorkStatusPublished}
	work2 := &model.Work{Title: "Work2", Status: model.WorkStatusPublished}
	db.WithContext(ctx).Create(work1)
	db.WithContext(ctx).Create(work2)

	repo.Create(ctx, &model.Favorite{UserID: user.ID, WorkID: work1.ID})
	repo.Create(ctx, &model.Favorite{UserID: user.ID, WorkID: work2.ID})

	favorites, err := repo.ListByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Len(t, favorites, 2)
}

func TestFavoriteRepository_Exists(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Favorite{})
	repo := NewFavoriteRepository(db)
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

	favorite := &model.Favorite{
		UserID: user.ID,
		WorkID: work.ID,
	}
	repo.Create(ctx, favorite)

	exists, err := repo.Exists(ctx, user.ID, work.ID)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFavoriteRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Favorite{})
	repo := NewFavoriteRepository(db)
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

	favorite := &model.Favorite{
		UserID: user.ID,
		WorkID: work.ID,
	}
	repo.Create(ctx, favorite)

	err := repo.Delete(ctx, favorite.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, favorite.ID)
	assert.Error(t, err)
}

func TestFavoriteRepository_DeleteByUserAndWork(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.Work{}, &model.Favorite{})
	repo := NewFavoriteRepository(db)
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

	favorite := &model.Favorite{
		UserID: user.ID,
		WorkID: work.ID,
	}
	repo.Create(ctx, favorite)

	err := repo.DeleteByUserAndWork(ctx, user.ID, work.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, favorite.ID)
	assert.Error(t, err)
}
