package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCraftRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	craft := &model.Craft{
		CategoryID: category.ID,
		Name:       "苏绣",
		Difficulty: 5,
	}

	err := repo.Create(ctx, craft)
	require.NoError(t, err)
	assert.NotZero(t, craft.ID)
}

func TestCraftRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	craft := &model.Craft{
		CategoryID: category.ID,
		Name:       "苏绣",
		Difficulty: 5,
	}
	repo.Create(ctx, craft)

	retrieved, err := repo.GetByID(ctx, craft.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, craft.Name, retrieved.Name)
}

func TestCraftRepository_GetByIDWithCategory(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	craft := &model.Craft{
		CategoryID: category.ID,
		Name:       "苏绣",
	}
	repo.Create(ctx, craft)

	retrieved, err := repo.GetByIDWithCategory(ctx, craft.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved.Category)
	assert.Equal(t, category.Name, retrieved.Category.Name)
}

func TestCraftRepository_GetByName(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	craft := &model.Craft{
		CategoryID: category.ID,
		Name:       "苏绣",
	}
	repo.Create(ctx, craft)

	retrieved, err := repo.GetByName(ctx, "苏绣")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, craft.ID, retrieved.ID)
}

func TestCraftRepository_List(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	for i := 1; i <= 3; i++ {
		repo.Create(ctx, &model.Craft{
			CategoryID: category.ID,
			Name:       "Craft" + string(rune('0'+i)),
		})
	}

	crafts, err := repo.List(ctx, "")
	require.NoError(t, err)
	assert.Len(t, crafts, 3)
}

func TestCraftRepository_ListByCategoryID(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	repo.Create(ctx, &model.Craft{
		CategoryID: category.ID,
		Name:       "苏绣",
	})

	crafts, err := repo.ListByCategoryID(ctx, category.ID)
	require.NoError(t, err)
	assert.Len(t, crafts, 1)
}

func TestCraftRepository_ListByDifficulty(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	repo.Create(ctx, &model.Craft{
		CategoryID: category.ID,
		Name:       "苏绣",
		Difficulty: 5,
	})

	crafts, err := repo.ListByDifficulty(ctx, 5)
	require.NoError(t, err)
	assert.Len(t, crafts, 1)
}

func TestCraftRepository_Update(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	craft := &model.Craft{
		CategoryID: category.ID,
		Name:       "OldName",
	}
	repo.Create(ctx, craft)

	craft.Name = "NewName"
	err := repo.Update(ctx, craft)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, craft.ID)
	assert.Equal(t, "NewName", retrieved.Name)
}

func TestCraftRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	craft := &model.Craft{
		CategoryID: category.ID,
		Name:       "ToDelete",
	}
	repo.Create(ctx, craft)

	err := repo.Delete(ctx, craft.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, craft.ID)
	assert.Error(t, err)
}

func TestCraftRepository_UpsertBatch(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{}, &model.Craft{})
	repo := NewCraftRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{Name: "传统技艺", Level: 1}
	db.WithContext(ctx).Create(category)

	crafts := []model.Craft{
		{CategoryID: category.ID, Name: "Craft1"},
		{CategoryID: category.ID, Name: "Craft2"},
	}

	err := repo.UpsertBatch(ctx, crafts)
	require.NoError(t, err)

	retrieved, _ := repo.List(ctx, "")
	assert.Len(t, retrieved, 2)
}
