package repository

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"user-management/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]domain.User, error)
	Update(ctx context.Context, id primitive.ObjectID, input domain.UpdateUserInput) (*domain.User, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
}
