package resolvers

import (
	"context"
	"errors"

	"github.com/factly/dega-api/graph/generated"
	"github.com/factly/dega-api/graph/loaders"
	"github.com/factly/dega-api/graph/logger"
	"github.com/factly/dega-api/graph/models"
	"github.com/factly/dega-api/graph/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *claimResolver) Rating(ctx context.Context, obj *models.Claim) (*models.Rating, error) {
	return loaders.GetRatingLoader(ctx).Load(obj.Rating.ID)
}

func (r *claimResolver) Claimant(ctx context.Context, obj *models.Claim) (*models.Claimant, error) {
	return loaders.GetClaimantLoader(ctx).Load(obj.Claimant.ID)
}

func (r *queryResolver) Claims(ctx context.Context, ratings []string, claimants []string, page *int, limit *int, sortBy *string, sortOrder *string) (*models.ClaimsPaging, error) {

	client := ctx.Value("client").(string)

	if client == "" {
		return nil, errors.New("client id missing")
	}

	query := bson.M{
		"client_id": client,
	}

	if len(ratings) > 0 {
		keys := []primitive.ObjectID{}

		for _, id := range ratings {
			rid, err := primitive.ObjectIDFromHex(id)

			if err == nil {
				keys = append(keys, rid)
			}
		}

		query["rating.$id"] = bson.M{"$in": keys}
	}

	if len(claimants) > 0 {
		keys := []primitive.ObjectID{}

		for _, id := range claimants {
			cid, err := primitive.ObjectIDFromHex(id)

			if err == nil {
				keys = append(keys, cid)
			}
		}

		query["claimant.$id"] = bson.M{"$in": keys}
	}

	pageLimit := 10
	pageNo := 1
	pageSortBy := "created_date"
	pageSortOrder := -1

	if limit != nil {
		pageLimit = *limit
	}
	if page != nil {
		pageNo = *page
	}

	if sortBy != nil {
		pageSortBy = *sortBy
	}
	if sortOrder != nil && *sortOrder == "ASC" {
		pageSortOrder = 1
	}

	opts := options.Find().SetSort(bson.D{{pageSortBy, pageSortOrder}}).SetSkip(int64((pageNo - 1) * pageLimit)).SetLimit(int64(pageLimit))
	cursor, err := mongo.Factcheck.Collection("claim").Find(ctx, query, opts)

	if err != nil {
		logger.Error(err)
		return nil, nil
	}

	count, err := mongo.Factcheck.Collection("claim").CountDocuments(ctx, query)

	if err != nil {
		logger.Error(err)
		return nil, nil
	}

	var nodes []*models.Claim

	for cursor.Next(ctx) {
		var each *models.Claim
		err := cursor.Decode(&each)
		if err != nil {
			logger.Error(err)
			return nil, nil
		}
		nodes = append(nodes, each)
	}

	var result *models.ClaimsPaging = new(models.ClaimsPaging)

	result.Nodes = nodes
	result.Total = int(count)

	return result, nil
}

// Claim model resolver
func (r *Resolver) Claim() generated.ClaimResolver { return &claimResolver{r} }

type claimResolver struct{ *Resolver }
