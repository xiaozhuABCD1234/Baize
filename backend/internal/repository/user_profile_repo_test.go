package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserProfileRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserProfileRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	profile := &model.UserProfile{
		UserID:   user.ID,
		Nickname: "TestNick",
		Bio:      "Test bio",
		IsMaster: true,
	}

	err := repo.Create(ctx, profile)
	require.NoError(t, err)
	assert.NotZero(t, profile.ID)
}

func TestUserProfileRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserProfileRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	profile := &model.UserProfile{
		UserID:   user.ID,
		Nickname: "TestNick",
		Bio:      "Test bio",
	}
	repo.Create(ctx, profile)

	retrieved, err := repo.GetByID(ctx, profile.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, profile.Nickname, retrieved.Nickname)
}

func TestUserProfileRepository_GetByUserID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserProfileRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	profile := &model.UserProfile{
		UserID:   user.ID,
		Nickname: "TestNick",
	}
	repo.Create(ctx, profile)

	retrieved, err := repo.GetByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, profile.ID, retrieved.ID)
}

func TestUserProfileRepository_Update(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserProfileRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	profile := &model.UserProfile{
		UserID:   user.ID,
		Nickname: "OldNick",
	}
	repo.Create(ctx, profile)

	profile.Nickname = "NewNick"
	err := repo.Update(ctx, profile)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, profile.ID)
	assert.Equal(t, "NewNick", retrieved.Nickname)
}

func TestUserProfileRepository_UpdateByUserID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserProfileRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	profile := &model.UserProfile{
		UserID:   user.ID,
		Nickname: "OldNick",
	}
	repo.Create(ctx, profile)

	err := repo.UpdateByUserID(ctx, user.ID, map[string]interface{}{"nickname": "NewNick"})
	require.NoError(t, err)

	retrieved, _ := repo.GetByUserID(ctx, user.ID)
	assert.Equal(t, "NewNick", retrieved.Nickname)
}

func TestUserProfileRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserProfileRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	profile := &model.UserProfile{
		UserID:   user.ID,
		Nickname: "TestNick",
	}
	repo.Create(ctx, profile)

	err := repo.Delete(ctx, profile.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, profile.ID)
	assert.Error(t, err)
}

func TestUserProfileRepository_Upsert(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserProfileRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	db.WithContext(ctx).Create(user)

	profile := &model.UserProfile{
		UserID:   user.ID,
		Nickname: "Original",
	}
	repo.Create(ctx, profile)

	profile.Nickname = "Updated"
	err := repo.Upsert(ctx, profile)
	require.NoError(t, err)

	retrieved, _ := repo.GetByUserID(ctx, user.ID)
	assert.Equal(t, "Updated", retrieved.Nickname)
}
