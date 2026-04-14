package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"backend/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func createTestUser(t *testing.T, db *gorm.DB, user *models.User) *models.User {
	if user.ID == 0 {
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}
	return user
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "hashedpassword",
			UserType: models.UserTypeUser,
			Status:   models.UserStatusActive,
		}

		err := repo.Create(ctx, user)
		if err != nil {
			t.Errorf("Create() error = %v, want nil", err)
		}
		if user.ID == 0 {
			t.Error("Create() user.ID should not be zero")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		user1 := &models.User{Username: "user1", Email: "same@example.com", Password: "hash", UserType: models.UserTypeUser}
		db.Create(user1)

		user2 := &models.User{Username: "user2", Email: "same@example.com", Password: "hash", UserType: models.UserTypeUser}
		err := repo.Create(ctx, user2)
		if err == nil {
			t.Error("Create() expected error for duplicate email, got nil")
		}
	})

	t.Run("duplicate username", func(t *testing.T) {
		user1 := &models.User{Username: "sameuser", Email: "a@example.com", Password: "hash", UserType: models.UserTypeUser}
		db.Create(user1)

		user2 := &models.User{Username: "sameuser", Email: "b@example.com", Password: "hash", UserType: models.UserTypeUser}
		err := repo.Create(ctx, user2)
		if err == nil {
			t.Error("Create() expected error for duplicate username, got nil")
		}
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &models.User{Username: "getbyid", Email: "getbyid@example.com", Password: "hash", UserType: models.UserTypeUser}
		db.Create(user)

		found, err := repo.GetByID(ctx, user.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v, want nil", err)
		}
		if found == nil {
			t.Fatal("GetByID() returned nil user")
		}
		if found.Email != user.Email {
			t.Errorf("GetByID() email = %v, want %v", found.Email, user.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		found, err := repo.GetByID(ctx, 9999)
		if err != nil {
			t.Errorf("GetByID() error = %v, want nil", err)
		}
		if found != nil {
			t.Error("GetByID() should return nil for non-existent user")
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &models.User{Username: "getbyemail", Email: "getbyemail@example.com", Password: "hash", UserType: models.UserTypeUser}
		db.Create(user)

		found, err := repo.GetByEmail(ctx, user.Email)
		if err != nil {
			t.Errorf("GetByEmail() error = %v, want nil", err)
		}
		if found == nil {
			t.Fatal("GetByEmail() returned nil user")
		}
		if found.Username != user.Username {
			t.Errorf("GetByEmail() username = %v, want %v", found.Username, user.Username)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		found, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		if err != nil {
			t.Errorf("GetByEmail() error = %v, want nil", err)
		}
		if found != nil {
			t.Error("GetByEmail() should return nil for non-existent email")
		}
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &models.User{Username: "getbyusername", Email: "getbyusername@example.com", Password: "hash", UserType: models.UserTypeUser}
		db.Create(user)

		found, err := repo.GetByUsername(ctx, user.Username)
		if err != nil {
			t.Errorf("GetByUsername() error = %v, want nil", err)
		}
		if found == nil {
			t.Fatal("GetByUsername() returned nil user")
		}
		if found.Email != user.Email {
			t.Errorf("GetByUsername() email = %v, want %v", found.Email, user.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		found, err := repo.GetByUsername(ctx, "nonexistent")
		if err != nil {
			t.Errorf("GetByUsername() error = %v, want nil", err)
		}
		if found != nil {
			t.Error("GetByUsername() should return nil for non-existent username")
		}
	})
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success with users", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			user := &models.User{
				Username: "listuser",
				Email:    "listuser@example.com",
				Password: "hash",
				UserType: models.UserTypeUser,
			}
			db.Create(user)
		}

		users, err := repo.List(ctx)
		if err != nil {
			t.Errorf("List() error = %v, want nil", err)
		}
		if len(users) < 1 {
			t.Errorf("List() count = %d, want >= 1", len(users))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		db.Exec("DELETE FROM users")
		users, err := repo.List(ctx)
		if err != nil {
			t.Errorf("List() error = %v, want nil", err)
		}
		if len(users) != 0 {
			t.Errorf("List() count = %d, want 0", len(users))
		}
	})
}

func TestUserRepository_ListWithPagination(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	for i := 0; i < 15; i++ {
		user := &models.User{
			Username: fmt.Sprintf("pageuser%d", i),
			Email:    fmt.Sprintf("pageuser%d@example.com", i),
			Password: "hash",
			UserType: models.UserTypeUser,
			Phone:    fmt.Sprintf("1380000000%d", i),
		}
		if err := db.Create(user).Error; err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
	}

	t.Run("first page", func(t *testing.T) {
		users, total, err := repo.ListWithPagination(ctx, 1, 10)
		if err != nil {
			t.Errorf("ListWithPagination() error = %v, want nil", err)
		}
		if len(users) != 10 {
			t.Errorf("ListWithPagination() users count = %d, want 10", len(users))
		}
		if total != 15 {
			t.Errorf("ListWithPagination() total = %d, want 15", total)
		}
	})

	t.Run("second page", func(t *testing.T) {
		users, total, err := repo.ListWithPagination(ctx, 2, 10)
		if err != nil {
			t.Errorf("ListWithPagination() error = %v, want nil", err)
		}
		if len(users) != 5 {
			t.Errorf("ListWithPagination() users count = %d, want 5", len(users))
		}
		if total != 15 {
			t.Errorf("ListWithPagination() total = %d, want 15", total)
		}
	})

	t.Run("page beyond data", func(t *testing.T) {
		users, total, err := repo.ListWithPagination(ctx, 10, 10)
		if err != nil {
			t.Errorf("ListWithPagination() error = %v, want nil", err)
		}
		if len(users) != 0 {
			t.Errorf("ListWithPagination() users count = %d, want 0", len(users))
		}
		if total != 15 {
			t.Errorf("ListWithPagination() total = %d, want 15", total)
		}
	})

	t.Run("page size 0 returns no limit", func(t *testing.T) {
		users, _, err := repo.ListWithPagination(ctx, 1, 0)
		if err != nil {
			t.Errorf("ListWithPagination() error = %v, want nil", err)
		}
		if len(users) != 0 {
			t.Errorf("ListWithPagination() users count = %d, want 0", len(users))
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &models.User{Username: "updateuser", Email: "updateuser@example.com", Password: "hash", UserType: models.UserTypeUser}
		db.Create(user)

		user.Email = "newemail@example.com"
		err := repo.Update(ctx, user)
		if err != nil {
			t.Errorf("Update() error = %v, want nil", err)
		}

		var updated models.User
		db.First(&updated, user.ID)
		if updated.Email != "newemail@example.com" {
			t.Errorf("Update() email = %v, want newemail@example.com", updated.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		user := &models.User{Username: "nonexistent", Email: "nonexistent@example.com"}
		user.ID = 9999
		err := repo.Update(ctx, user)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("Update() error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &models.User{Username: "passuser", Email: "passuser@example.com", Password: "oldhash", UserType: models.UserTypeUser}
		db.Create(user)

		err := repo.UpdatePassword(ctx, user.ID, "newhash")
		if err != nil {
			t.Errorf("UpdatePassword() error = %v, want nil", err)
		}

		var updated models.User
		db.First(&updated, user.ID)
		if updated.Password != "newhash" {
			t.Error("UpdatePassword() did not update password")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		err := repo.UpdatePassword(ctx, 9999, "newhash")
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("UpdatePassword() error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestUserRepository_UpdateEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &models.User{Username: "emailuser", Email: "oldemail@example.com", Password: "hash", UserType: models.UserTypeUser}
		db.Create(user)

		err := repo.UpdateEmail(ctx, user.ID, "newemail@example.com")
		if err != nil {
			t.Errorf("UpdateEmail() error = %v, want nil", err)
		}

		var updated models.User
		db.First(&updated, user.ID)
		if updated.Email != "newemail@example.com" {
			t.Errorf("UpdateEmail() email = %v, want newemail@example.com", updated.Email)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		err := repo.UpdateEmail(ctx, 9999, "newemail@example.com")
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("UpdateEmail() error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("soft delete success", func(t *testing.T) {
		user := &models.User{Username: "deleteuser", Email: "deleteuser@example.com", Password: "hash", UserType: models.UserTypeUser}
		db.Create(user)

		err := repo.Delete(ctx, user.ID)
		if err != nil {
			t.Errorf("Delete() error = %v, want nil", err)
		}

		var deleted models.User
		err = db.Unscoped().First(&deleted, user.ID).Error
		if err != nil {
			t.Errorf("Delete() user still exists: %v", err)
		}

		var softDeleted models.User
		err = db.First(&softDeleted, user.ID).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Error("Delete() should soft delete user")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		err := repo.Delete(ctx, 9999)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("Delete() error = %v, want %v", err, ErrUserNotFound)
		}
	})

	t.Run("idempotent delete", func(t *testing.T) {
		user := &models.User{Username: "idempotent", Email: "idempotent@example.com", Password: "hash", UserType: models.UserTypeUser, Phone: "13900000000"}
		db.Create(user)

		err := repo.Delete(ctx, user.ID)
		if err != nil {
			t.Errorf("Delete() first call error = %v, want nil", err)
		}

		err = repo.Delete(ctx, user.ID)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("Delete() second call error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestUserRepository_ForceDelete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	t.Run("force delete success", func(t *testing.T) {
		user := &models.User{Username: "forceuser", Email: "forceuser@example.com", Password: "hash", UserType: models.UserTypeUser}
		db.Create(user)

		err := repo.ForceDelete(ctx, user.ID)
		if err != nil {
			t.Errorf("ForceDelete() error = %v, want nil", err)
		}

		var deleted models.User
		err = db.Unscoped().First(&deleted, user.ID).Error
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Error("ForceDelete() should permanently delete user")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		err := repo.ForceDelete(ctx, 9999)
		if !errors.Is(err, ErrUserNotFound) {
			t.Errorf("ForceDelete() error = %v, want %v", err, ErrUserNotFound)
		}
	})
}

func TestUserRepository_DBError(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	t.Run("GetByID db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		_, err := repo.GetByID(ctx, 1)
		if err == nil {
			t.Error("GetByID() expected error with closed db, got nil")
		}
	})

	t.Run("GetByEmail db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		_, err := repo.GetByEmail(ctx, "test@example.com")
		if err == nil {
			t.Error("GetByEmail() expected error with closed db, got nil")
		}
	})

	t.Run("GetByUsername db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		_, err := repo.GetByUsername(ctx, "testuser")
		if err == nil {
			t.Error("GetByUsername() expected error with closed db, got nil")
		}
	})

	t.Run("List db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		_, err := repo.List(ctx)
		if err == nil {
			t.Error("List() expected error with closed db, got nil")
		}
	})

	t.Run("ListWithPagination db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		_, _, err := repo.ListWithPagination(ctx, 1, 10)
		if err == nil {
			t.Error("ListWithPagination() expected error with closed db, got nil")
		}
	})

	t.Run("Create db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		user := &models.User{Username: "dbuser", Email: "db@example.com", Password: "hash"}
		err := repo.Create(ctx, user)
		if err == nil {
			t.Error("Create() expected error with closed db, got nil")
		}
	})

	t.Run("Update db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		user := &models.User{Username: "dbuser", Email: "db@example.com"}
		user.ID = 1
		err := repo.Update(ctx, user)
		if err == nil {
			t.Error("Update() expected error with closed db, got nil")
		}
	})

	t.Run("UpdatePassword db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		err := repo.UpdatePassword(ctx, 1, "newhash")
		if err == nil {
			t.Error("UpdatePassword() expected error with closed db, got nil")
		}
	})

	t.Run("UpdateEmail db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		err := repo.UpdateEmail(ctx, 1, "new@example.com")
		if err == nil {
			t.Error("UpdateEmail() expected error with closed db, got nil")
		}
	})

	t.Run("Delete db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		err := repo.Delete(ctx, 1)
		if err == nil {
			t.Error("Delete() expected error with closed db, got nil")
		}
	})

	t.Run("ForceDelete db error", func(t *testing.T) {
		repo := NewUserRepository(db)
		sqlDB, _ := db.DB()
		sqlDB.Close()

		err := repo.ForceDelete(ctx, 1)
		if err == nil {
			t.Error("ForceDelete() expected error with closed db, got nil")
		}
	})
}
