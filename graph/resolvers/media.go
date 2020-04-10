package resolvers

import (
	"context"
	"errors"

	"github.com/factly/dega-api/graph/logger"
	"github.com/factly/dega-api/graph/models"
	"github.com/factly/dega-api/graph/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func (r *queryResolver) Media(ctx context.Context) ([]*models.Medium, error) {
	client := ctx.Value("client").(string)

	if client == "" {
		return nil, errors.New("client id missing")
	}

	query := bson.M{
		"client_id": client,
	}

	cursor, err := mongo.Core.Collection("media").Find(ctx, query)

	if err != nil {
		logger.Error(err)
		return nil, nil
	}

	var results []*models.Medium

	for cursor.Next(ctx) {
		var each *models.Medium
		err := cursor.Decode(&each)
		if err != nil {
			logger.Error(err)
			return nil, nil
		}
		results = append(results, each)
	}

	return results, nil
}
