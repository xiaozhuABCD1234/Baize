package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegionRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	region := &model.Region{
		Name:  "北京市",
		Code:  "110000",
		Level: 1,
	}

	err := repo.Create(ctx, region)
	require.NoError(t, err)
	assert.NotZero(t, region.ID)
}

func TestRegionRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	region := &model.Region{
		Name:  "北京市",
		Code:  "110000",
		Level: 1,
	}
	repo.Create(ctx, region)

	retrieved, err := repo.GetByID(ctx, region.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, region.Name, retrieved.Name)
}

func TestRegionRepository_GetByCode(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	region := &model.Region{
		Name:  "北京市",
		Code:  "110000",
		Level: 1,
	}
	repo.Create(ctx, region)

	retrieved, err := repo.GetByCode(ctx, "110000")
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, region.Name, retrieved.Name)
}

func TestRegionRepository_List(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		repo.Create(ctx, &model.Region{
			Name:  "Region" + string(rune('0'+i)),
			Code:  "code" + string(rune('0'+i)),
			Level: 1,
		})
	}

	regions, err := repo.List(ctx, "")
	require.NoError(t, err)
	assert.Len(t, regions, 3)
}

func TestRegionRepository_ListRoot(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.Region{
		Name:     "Root",
		Code:     "root",
		Level:    1,
		ParentID: 0,
	})

	roots, err := repo.ListRoot(ctx)
	require.NoError(t, err)
	assert.Len(t, roots, 1)
}

func TestRegionRepository_ListByLevel(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.Region{
		Name:  "Province",
		Code:  "province",
		Level: 1,
	})

	regions, err := repo.ListByLevel(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, regions, 1)
}

func TestRegionRepository_ListHeritageCenters(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &model.Region{
		Name:             "Heritage Center",
		Code:             "hc1",
		IsHeritageCenter: true,
	})

	centers, err := repo.ListHeritageCenters(ctx)
	require.NoError(t, err)
	assert.Len(t, centers, 1)
}

func TestRegionRepository_Update(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	region := &model.Region{
		Name:  "OldName",
		Code:  "110000",
		Level: 1,
	}
	repo.Create(ctx, region)

	region.Name = "NewName"
	err := repo.Update(ctx, region)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, region.ID)
	assert.Equal(t, "NewName", retrieved.Name)
}

func TestRegionRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	region := &model.Region{
		Name:  "ToDelete",
		Code:  "110000",
		Level: 1,
	}
	repo.Create(ctx, region)

	err := repo.Delete(ctx, region.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, region.ID)
	assert.Error(t, err)
}

func TestRegionRepository_UpsertBatch(t *testing.T) {
	db := SetupTestDB(t, &model.Region{})
	repo := NewRegionRepository(db)
	ctx := context.Background()

	regions := []model.Region{
		{Name: "Region1", Code: "r1", Level: 1},
		{Name: "Region2", Code: "r2", Level: 1},
	}

	err := repo.UpsertBatch(ctx, regions)
	require.NoError(t, err)

	retrieved, _ := repo.List(ctx, "")
	assert.Len(t, retrieved, 2)
}
