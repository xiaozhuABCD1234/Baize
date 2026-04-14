package repository

import (
	"context"
	"testing"

	model "backend/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkMediaRepository_Create(t *testing.T) {
	db := SetupTestDB(t, &model.Work{}, &model.WorkMedia{})
	repo := NewWorkMediaRepository(db)
	ctx := context.Background()

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	media := &model.WorkMedia{
		WorkID:    work.ID,
		MediaType: model.MediaTypeImage,
		URL:       "https://example.com/image.jpg",
	}

	err := repo.Create(ctx, media)
	require.NoError(t, err)
	assert.NotZero(t, media.ID)
}

func TestWorkMediaRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t, &model.Work{}, &model.WorkMedia{})
	repo := NewWorkMediaRepository(db)
	ctx := context.Background()

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	media := &model.WorkMedia{
		WorkID:    work.ID,
		MediaType: model.MediaTypeImage,
		URL:       "https://example.com/image.jpg",
	}
	repo.Create(ctx, media)

	retrieved, err := repo.GetByID(ctx, media.ID)
	require.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, media.URL, retrieved.URL)
}

func TestWorkMediaRepository_ListByWorkID(t *testing.T) {
	db := SetupTestDB(t, &model.Work{}, &model.WorkMedia{})
	repo := NewWorkMediaRepository(db)
	ctx := context.Background()

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	for i := 1; i <= 3; i++ {
		repo.Create(ctx, &model.WorkMedia{
			WorkID:    work.ID,
			MediaType: model.MediaTypeImage,
			URL:       "https://example.com/image" + string(rune('0'+i)) + ".jpg",
		})
	}

	mediaList, err := repo.ListByWorkID(ctx, work.ID)
	require.NoError(t, err)
	assert.Len(t, mediaList, 3)
}

func TestWorkMediaRepository_Update(t *testing.T) {
	db := SetupTestDB(t, &model.Work{}, &model.WorkMedia{})
	repo := NewWorkMediaRepository(db)
	ctx := context.Background()

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	media := &model.WorkMedia{
		WorkID:    work.ID,
		MediaType: model.MediaTypeImage,
		URL:       "https://example.com/old.jpg",
	}
	repo.Create(ctx, media)

	media.URL = "https://example.com/new.jpg"
	err := repo.Update(ctx, media)
	require.NoError(t, err)

	retrieved, _ := repo.GetByID(ctx, media.ID)
	assert.Equal(t, "https://example.com/new.jpg", retrieved.URL)
}

func TestWorkMediaRepository_Delete(t *testing.T) {
	db := SetupTestDB(t, &model.Work{}, &model.WorkMedia{})
	repo := NewWorkMediaRepository(db)
	ctx := context.Background()

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	media := &model.WorkMedia{
		WorkID:    work.ID,
		MediaType: model.MediaTypeImage,
		URL:       "https://example.com/toDelete.jpg",
	}
	repo.Create(ctx, media)

	err := repo.Delete(ctx, media.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, media.ID)
	assert.Error(t, err)
}

func TestWorkMediaRepository_CreateBatch(t *testing.T) {
	db := SetupTestDB(t, &model.Work{}, &model.WorkMedia{})
	repo := NewWorkMediaRepository(db)
	ctx := context.Background()

	work := &model.Work{
		Title:  "TestWork",
		Status: model.WorkStatusPublished,
	}
	db.WithContext(ctx).Create(work)

	mediaList := []model.WorkMedia{
		{WorkID: work.ID, MediaType: model.MediaTypeImage, URL: "https://example.com/1.jpg"},
		{WorkID: work.ID, MediaType: model.MediaTypeImage, URL: "https://example.com/2.jpg"},
	}

	err := repo.CreateBatch(ctx, mediaList)
	require.NoError(t, err)

	retrieved, _ := repo.ListByWorkID(ctx, work.ID)
	assert.Len(t, retrieved, 2)
}
