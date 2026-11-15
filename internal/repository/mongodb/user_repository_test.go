package mongodb

import (
	"context"
	"errors"
	"testing"
	"time"

	"user-management/internal/domain"
	"user-management/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func fixtureUser() *domain.User {
	return &domain.User{
		ID:        primitive.NewObjectID(),
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  "hashed",
		CreatedAt: time.Now().Truncate(time.Millisecond),
	}
}

func userDoc(u *domain.User) bson.D {
	return bson.D{
		{Key: "_id", Value: u.ID},
		{Key: "name", Value: u.Name},
		{Key: "email", Value: u.Email},
		{Key: "password", Value: u.Password},
		{Key: "created_at", Value: u.CreatedAt},
	}
}

func assertErrIs(t *testing.T, got, want error) {
	t.Helper()
	if !errors.Is(got, want) { t.Fatalf("error = %v, want %v", got, want) }
}

func cmdErr() bson.D {
	return mtest.CreateCommandErrorResponse(mtest.CommandError{Code: 1, Message: "db error"})
}

func TestEnsureIndexes(t *testing.T) {
	ctx := context.Background()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		if err := NewUserRepository(mt.DB).EnsureIndexes(ctx); err != nil { mt.Fatal(err) }
	})
	mt.Run("error", func(mt *mtest.T) {
		mt.AddMockResponses(cmdErr())
		if err := NewUserRepository(mt.DB).EnsureIndexes(ctx); err == nil { mt.Fatal("expected error") }
	})
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		u := &domain.User{Name: "Test User", Email: "test@example.com"}
		if err := NewUserRepository(mt.DB).Create(ctx, u); err != nil { mt.Fatal(err) }
		if u.ID.IsZero() { mt.Fatal("expected ID to be set") }
	})
	mt.Run("insert error", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{Index: 1, Code: 11000, Message: "dup"}))
		if err := NewUserRepository(mt.DB).Create(ctx, fixtureUser()); err == nil { mt.Fatal("expected error") }
	})
}

func TestGetByIDAndEmail(t *testing.T) {
	ctx := context.Background()
	u := fixtureUser()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("GetByID success", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".users"
		mt.AddMockResponses(mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, userDoc(u)))
		got, err := NewUserRepository(mt.DB).GetByID(ctx, u.ID)
		if err != nil || got.ID != u.ID || got.Email != u.Email { mt.Fatalf("got=%+v err=%v", got, err) }
	})
	mt.Run("GetByID not found", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".users"
		mt.AddMockResponses(mtest.CreateCursorResponse(0, ns, mtest.FirstBatch))
		_, err := NewUserRepository(mt.DB).GetByID(ctx, u.ID)
		assertErrIs(mt.T, err, repository.ErrNotFound)
	})
	mt.Run("GetByID db error", func(mt *mtest.T) {
		mt.AddMockResponses(cmdErr())
		if _, err := NewUserRepository(mt.DB).GetByID(ctx, u.ID); err == nil { mt.Fatal("expected error") }
	})

	mt.Run("GetByEmail success", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".users"
		mt.AddMockResponses(mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, userDoc(u)))
		got, err := NewUserRepository(mt.DB).GetByEmail(ctx, u.Email)
		if err != nil || got.Email != u.Email { mt.Fatalf("got=%+v err=%v", got, err) }
	})
	mt.Run("GetByEmail not found", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".users"
		mt.AddMockResponses(mtest.CreateCursorResponse(0, ns, mtest.FirstBatch))
		_, err := NewUserRepository(mt.DB).GetByEmail(ctx, u.Email)
		assertErrIs(mt.T, err, repository.ErrNotFound)
	})
	mt.Run("GetByEmail db error", func(mt *mtest.T) {
		mt.AddMockResponses(cmdErr())
		if _, err := NewUserRepository(mt.DB).GetByEmail(ctx, u.Email); err == nil { mt.Fatal("expected error") }
	})
}

func TestListAndCount(t *testing.T) {
	ctx := context.Background()
	u := fixtureUser()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("List success", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".users"
		first := mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, userDoc(u))
		kill := mtest.CreateCursorResponse(0, ns, mtest.NextBatch)
		mt.AddMockResponses(first, kill)
		got, err := NewUserRepository(mt.DB).List(ctx)
		if err != nil || len(got) != 1 || got[0].ID != u.ID { mt.Fatalf("List: %+v %v", got, err) }
	})
	mt.Run("List find error", func(mt *mtest.T) {
		mt.AddMockResponses(cmdErr())
		if _, err := NewUserRepository(mt.DB).List(ctx); err == nil { mt.Fatal("expected error") }
	})
	mt.Run("List cursor decode error", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".users"
		mt.AddMockResponses(mtest.CreateCursorResponse(0, ns, mtest.FirstBatch, bson.D{{Key: "_id", Value: "bad_id"}}))
		if _, err := NewUserRepository(mt.DB).List(ctx); err == nil { mt.Fatal("expected error") }
	})

	mt.Run("Count success and error", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".users"
		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, bson.D{{Key: "_id", Value: 1}, {Key: "n", Value: int64(5)}}),
			cmdErr(),
		)
		r := NewUserRepository(mt.DB)
		n, err := r.Count(ctx)
		if err != nil || n != 5 { mt.Fatalf("Count: %d %v", n, err) }
		if _, err := r.Count(ctx); err == nil { mt.Fatal("expected error") }
	})
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	u := fixtureUser()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success with fields", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".users"
		name, email := "Updated", "updated@example.com"
		uUpdated := *u; uUpdated.Name = name; uUpdated.Email = email
		mt.AddMockResponses(
			bson.D{{Key: "ok", Value: 1}, {Key: "n", Value: 1}, {Key: "nModified", Value: 1}},
			mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, userDoc(&uUpdated)),
		)
		got, err := NewUserRepository(mt.DB).Update(ctx, u.ID, domain.UpdateUserInput{Name: &name, Email: &email})
		if err != nil || got.Name != name || got.Email != email { mt.Fatalf("got=%+v err=%v", got, err) }
	})
	mt.Run("empty input skips update and calls GetByID", func(mt *mtest.T) {
		ns := mt.DB.Name() + ".users"
		mt.AddMockResponses(mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, userDoc(u)))
		got, err := NewUserRepository(mt.DB).Update(ctx, u.ID, domain.UpdateUserInput{})
		if err != nil || got.ID != u.ID { mt.Fatalf("got=%+v err=%v", got, err) }
	})
	mt.Run("update error", func(mt *mtest.T) {
		name := "Updated"
		mt.AddMockResponses(cmdErr())
		if _, err := NewUserRepository(mt.DB).Update(ctx, u.ID, domain.UpdateUserInput{Name: &name}); err == nil {
			mt.Fatal("expected error")
		}
	})
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	u := fixtureUser()
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("success, not found, and error", func(mt *mtest.T) {
		mt.AddMockResponses(
			bson.D{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			bson.D{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 0}},
			cmdErr(),
		)
		r := NewUserRepository(mt.DB)
		if err := r.Delete(ctx, u.ID); err != nil { mt.Fatal(err) }
		assertErrIs(mt.T, r.Delete(ctx, u.ID), repository.ErrNotFound)
		if err := r.Delete(ctx, u.ID); err == nil { mt.Fatal("expected error") }
	})
}