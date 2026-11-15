package mongodb

import (
	"context"
	"errors"
	"user-management/internal/domain"
	"user-management/internal/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repository.UserRepository = (*UserRepository)(nil)

type UserRepository struct{ collection *mongo.Collection }

func NewUserRepository(db *mongo.Database) *UserRepository {
	return &UserRepository{collection: db.Collection("users")}
}
func (r *UserRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)})
	return err
}
func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	res, err := r.collection.InsertOne(ctx, u)
	if err != nil {
		return err
	}
	u.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}
func (r *UserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repository.ErrNotFound
	}
	return &user, err
}
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repository.ErrNotFound
	}
	return &user, err
}
func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	cur, err := r.collection.Find(ctx, bson.D{}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	users := []domain.User{}
	if err = cur.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}
func (r *UserRepository) Update(ctx context.Context, id primitive.ObjectID, input domain.UpdateUserInput) (*domain.User, error) {
	set := bson.M{}
	if input.Name != nil {
		set["name"] = *input.Name
	}
	if input.Email != nil {
		set["email"] = *input.Email
	}
	if len(set) == 0 {
		return r.GetByID(ctx, id)
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}
func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.D{})
}
