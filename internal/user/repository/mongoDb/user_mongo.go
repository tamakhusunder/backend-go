package repository

import (
	"backend-go/constants"
	model "backend-go/models"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository interface {
	Create(ctx context.Context, creds model.User) (interface{}, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateByID(ctx context.Context, id string, updatedData bson.M) (*model.User, error)
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
	_, existErr := r.FindByEmail(ctx, creds.Email)
	if existErr == nil {
		return nil, fmt.Errorf("user already exists with email %s", creds.Email)
	}
	if existErr != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("error checking existing user: %v", existErr)
	}

	result, err := r.collection.InsertOne(ctx, creds)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *userRepositoryImpl) FindByID(ctx context.Context, id string) (*model.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}

	var user model.User
	collErr := r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if collErr != nil {
		return nil, collErr
	}
	log.Println("repo: user repo find by id", user)

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

func (r *userRepositoryImpl) UpdateByID(ctx context.Context, id string, updatedData bson.M) (*model.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %v", err)
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": updatedData}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated model.User
	collErr := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updated)
	if collErr != nil {
		return nil, collErr
	}

	return &updated, nil
}
func (r *userRepositoryImpl) DeleteByID(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %v", err)
	}
	filter := bson.M{"_id": objectID}
	_, collErr := r.collection.DeleteOne(ctx, filter)
	if collErr != nil {
		log.Printf("Failed to update user: %v", id)
		return collErr
	}
	return nil
}
