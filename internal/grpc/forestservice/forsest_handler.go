package forestservice

import (
	"context"
	"errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jdk829355/InForest-back/models"
	"github.com/jdk829355/InForest-back/protos/forest"
	"go.uber.org/zap"
)

func (s *ForestService) GetForestsByUser(ctx context.Context, req *forest.GetForestsByUserRequest) (*forest.GetForestsByUserResponse, error) {
	user_id := ctx.Value("user_id")
	if user_id == "" {
		return nil, errors.New("invalid user_id")
	}
	forests, err := s.Store.Neo4j.GetForestByUser(ctx, user_id.(string), req.GetIncludeChildren())
	if err != nil {
		return nil, err
	}
	forestsProto := make([]*forest.Forest, len(forests))
	for i, f := range forests {
		forestsProto[i] = f.ToProto()
	}
	return &forest.GetForestsByUserResponse{
		Forests: forestsProto,
	}, nil
}

func (s *ForestService) CreateForest(ctx context.Context, req *forest.CreateForestRequest) (*forest.Forest, error) {
	user_id := ctx.Value("user_id")
	if user_id == "" {
		return nil, errors.New("invalid user_id")
	}

	root := &models.Tree{
		Id:   req.GetRoot().GetId(),
		Name: req.GetRoot().GetName(),
		Url:  req.GetRoot().GetUrl(),
	}
	forestModel := &models.Forest{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		UserId:      user_id.(string),
		Root:        root,
	}
	if err := s.Store.Neo4j.CreateForest(ctx, forestModel, root); err != nil {
		return nil, err
	}
	s.Store.Supabase.CreateMemo(user_id.(string), root.Id, nil)
	return forestModel.ToProto(), nil
}

func (s *ForestService) GetForest(ctx context.Context, req *forest.GetForestRequest) (*forest.GetForestResponse, error) {
	forestModel, err := s.Store.Neo4j.GetForest(ctx, req.GetForestId(), req.GetIncludeChildren())
	if err != nil {
		return nil, err
	}
	return &forest.GetForestResponse{
		Forest: forestModel.ToProto(),
	}, nil
}

func (s *ForestService) UpdateForest(ctx context.Context, req *forest.UpdateForestRequest) (*forest.Forest, error) {
	inputForestModel := &models.Forest{
		Id:          req.GetForestId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
	}
	forestModel, err := s.Store.Neo4j.UpdateForest(ctx, inputForestModel)
	if err != nil {
		return nil, err
	}
	return forestModel.ToProto(), nil
}

func (s *ForestService) DeleteForest(ctx context.Context, req *forest.DeleteForestRequest) (*forest.DeleteForestResponse, error) {
	idsToDelete, err := s.Store.Neo4j.DeleteForest(ctx, req.GetForestId())
	ctxzap.Extract(ctx).Info("Deleted forest", zap.String("forest_id", req.GetForestId()), zap.Strings("idsToDelete", idsToDelete))
	if err != nil {
		return &forest.DeleteForestResponse{
			Success: false,
		}, err
	}
	user_id := ctx.Value("user_id")
	if user_id == "" {
		return nil, errors.New("invalid user_id")
	}
	for _, treeID := range idsToDelete {
		_, err := s.Store.Supabase.DeleteMemo(user_id.(string), treeID)
		if err != nil {
			ctxzap.Extract(ctx).Error("Failed to delete memo", zap.String("tree_id", treeID), zap.Error(err))
		}
	}
	return &forest.DeleteForestResponse{
		Success: true,
	}, nil
}
