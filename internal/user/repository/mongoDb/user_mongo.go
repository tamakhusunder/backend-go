package repository

import (
	"backend-go/constants"
	model "backend-go/models"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	Create(ctx context.Context, creds model.User) (interface{}, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateByID(ctx context.Context, id string, updatedData model.User) (*model.User, error)
	DeleteByID(ctx context.Context, id string) error
}

type userRepositoryImpl struct {
	collection *mongo.Collection
}

func NewUserRepository(database *mongo.Database) UserRepository {
	return &userRepositoryImpl{
		collection: database.Collection(constants.USER_COLLECTION),
	}
}

func (r *userRepositoryImpl) Create(ctx context.Context, creds model.User) (interface{}, error) {
	result, err := r.collection.InsertOne(ctx, creds)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepositoryImpl) UpdateByID(ctx context.Context, id string, updatedData model.User) (*model.User, error) {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": updatedData}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated model.User
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}
func (r *userRepositoryImpl) DeleteByID(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Printf("Failed to update user: %v", id)
		return err
	}
	return nil
}
