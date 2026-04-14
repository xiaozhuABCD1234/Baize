package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	repo.Create(ctx, user)

	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, user.Username, retrieved.Username)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	repo.Create(ctx, user)

	retrieved, err := repo.GetByEmail(ctx, "test@example.com")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, user.Username, retrieved.Username)
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	repo.Create(ctx, user)

	retrieved, err := repo.GetByUsername(ctx, "testuser")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, user.Email, retrieved.Email)
}

func TestUserRepository_Update(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	repo.Create(ctx, user)

	user.Username = "updateduser"
	err := repo.Update(ctx, user)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, user.ID)
	assert.Equal(t, "updateduser", retrieved.Username)
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "oldpassword",
		Status:   model.UserStatusActive,
	}
	repo.Create(ctx, user)

	err := repo.UpdatePassword(ctx, user.ID, "newpassword")
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, user.ID)
	assert.Equal(t, "newpassword", retrieved.Password)
}

func TestUserRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	repo.Create(ctx, user)

	err := repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, user.ID)
	assert.Error(t, err)
}

func TestUserRepository_WithTransaction(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		profile := &model.UserProfile{
			UserID:   user.ID,
			Nickname: "TestNick",
		}
		return tx.Create(profile).Error
	})
	require.NoError(t, err)

	txRepo := repo.WithTransaction(db)
	retrieved, err := txRepo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
}

func TestUserRepository_Upsert(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	repo.Create(ctx, user)

	user.Password = "updatedpassword"
	err := repo.Upsert(ctx, user)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, user.ID)
	assert.Equal(t, "updatedpassword", retrieved.Password)
}

func TestUserRepository_GetByPhone(t *testing.T) {
	db := SetupTestDB(t, &model.User{}, &model.UserProfile{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Phone:    "1234567890",
		Password: "hashedpassword",
		Status:   model.UserStatusActive,
	}
	repo.Create(ctx, user)

	retrieved, err := repo.GetByPhone(ctx, "1234567890")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, user.Username, retrieved.Username)
}
