package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestICHCategoryRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{})
	repo := NewICHCategoryRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{
		Name:  "传统技艺",
		Level: 1,
	}

	err := repo.Create(ctx, category)
	require.NoError(t, err)
	assert.NotZero(t, category.ID)
}

func TestICHCategoryRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{})
	repo := NewICHCategoryRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{
		Name:  "传统技艺",
		Level: 1,
	}
	repo.Create(ctx, category)

	retrieved, err := repo.GetByID(ctx, category.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, category.Name, retrieved.Name)
}

func TestICHCategoryRepository_GetByName(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{})
	repo := NewICHCategoryRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{
		Name:  "传统技艺",
		Level: 1,
	}
	repo.Create(ctx, category)

	retrieved, err := repo.GetByName(ctx, "传统技艺")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, category.ID, retrieved.ID)
}

func TestICHCategoryRepository_List(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{})
	repo := NewICHCategoryRepository(db)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		repo.Create(ctx, &model.ICHCategory{
			Name:  "Category" + string(rune('0'+i)),
			Level: 1,
		})
	}

	categories, err := repo.List(ctx, "")
	require.NoError(t, err)
	assert.Len(t, categories, 3)
}

func TestICHCategoryRepository_ListRoot(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{})
	repo := NewICHCategoryRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.ICHCategory{
		Name:     "Root",
		Level:    1,
		ParentID: 0,
	})

	roots, err := repo.ListRoot(ctx)
	require.NoError(t, err)
	assert.Len(t, roots, 1)
}

func TestICHCategoryRepository_ListActive(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{})
	repo := NewICHCategoryRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.ICHCategory{
		Name:   "Active",
		Level:  1,
		Status: 1,
	})

	active, err := repo.ListActive(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 1)
}

func TestICHCategoryRepository_Update(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{})
	repo := NewICHCategoryRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{
		Name:  "OldName",
		Level: 1,
	}
	repo.Create(ctx, category)

	category.Name = "NewName"
	err := repo.Update(ctx, category)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, category.ID)
	assert.Equal(t, "NewName", retrieved.Name)
}

func TestICHCategoryRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{})
	repo := NewICHCategoryRepository(db)
	ctx := context.Background()

	category := &model.ICHCategory{
		Name:  "ToDelete",
		Level: 1,
	}
	repo.Create(ctx, category)

	err := repo.Delete(ctx, category.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, category.ID)
	assert.Error(t, err)
}

func TestICHCategoryRepository_UpsertBatch(t *testing.T) {
	db := SetupTestDB(t, &model.ICHCategory{})
	repo := NewICHCategoryRepository(db)
	ctx := context.Background()

	categories := []model.ICHCategory{
		{Name: "Cat1", Level: 1},
		{Name: "Cat2", Level: 1},
	}

	err := repo.UpsertBatch(ctx, categories)
	require.NoError(t, err)

	retrieved, _ := repo.List(ctx, "")
	assert.Len(t, retrieved, 2)
}
