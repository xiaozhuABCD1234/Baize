package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFollowRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	follow := &model.Follow{
		FollowerID:  1,
		FollowingID: 2,
	}

	err := repo.Create(ctx, follow)
	require.NoError(t, err)
}

func TestFollowRepository_Exists(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	follow := &model.Follow{
		FollowerID:  1,
		FollowingID: 2,
	}
	repo.Create(ctx, follow)

	exists, err := repo.Exists(ctx, 1, 2)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFollowRepository_NotExists(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	exists, err := repo.Exists(ctx, 999, 888)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFollowRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	follow := &model.Follow{
		FollowerID:  1,
		FollowingID: 2,
	}
	repo.Create(ctx, follow)

	err := repo.Delete(ctx, 1, 2)
	require.NoError(t, err)

	exists, _ := repo.Exists(ctx, 1, 2)
	assert.False(t, exists)
}

func TestFollowRepository_GetFollowingList(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 3})

	follows, err := repo.GetFollowingList(ctx, 1, "")
	require.NoError(t, err)
	assert.Len(t, follows, 2)
}

func TestFollowRepository_GetFollowerList(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	repo.Create(ctx, &model.Follow{FollowerID: 3, FollowingID: 2})

	follows, err := repo.GetFollowerList(ctx, 2, "")
	require.NoError(t, err)
	assert.Len(t, follows, 2)
}

func TestFollowRepository_CountFollowing(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 3})

	count, err := repo.CountFollowing(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestFollowRepository_CountFollowers(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})
	repo.Create(ctx, &model.Follow{FollowerID: 3, FollowingID: 2})

	count, err := repo.CountFollowers(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestFollowRepository_IsFollowing(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.Follow{FollowerID: 1, FollowingID: 2})

	isFollowing, err := repo.IsFollowing(ctx, 1, 2)
	require.NoError(t, err)
	assert.True(t, isFollowing)
}

func TestFollowRepository_WithTransaction(t *testing.T) {
	db := SetupTestDB(t, &model.Follow{})
	repo := NewFollowRepository(db)
	ctx := context.Background()

	txRepo := repo.WithTransaction(db)
	follow := &model.Follow{
		FollowerID:  10,
		FollowingID: 20,
	}
	err := txRepo.Create(ctx, follow)
	require.NoError(t, err)

	exists, _ := txRepo.Exists(ctx, 10, 20)
	assert.True(t, exists)
}
