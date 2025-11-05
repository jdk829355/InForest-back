package forestservice

import (
	"context"
	"errors"

	"github.com/jdk829355/InForest_back/models"
	"github.com/jdk829355/InForest_back/protos/forest"
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
	return forestModel.ToProto(), nil
}
func (s *ForestService) CreateTree(ctx context.Context, req *forest.CreateTreeRequest) (*forest.Tree, error) {
	treeModel := &models.Tree{
		Id:   req.GetId(),
		Name: req.GetName(),
		Url:  req.GetUrl(),
	}
	if err := s.Store.Neo4j.CreateTree(ctx, treeModel, req.GetParentId()); err != nil {
		return nil, err
	}
	return treeModel.ToProto(), nil
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
	err := s.Store.Neo4j.DeleteForest(ctx, req.GetForestId())
	if err != nil {
		return &forest.DeleteForestResponse{
			Success: false,
		}, err
	}
	return &forest.DeleteForestResponse{
		Success: true,
	}, nil
}

func (s *ForestService) GetTree(ctx context.Context, req *forest.GetTreeRequest) (*forest.Tree, error) {
	tree, err := s.Store.Neo4j.GetTreeByID(ctx, req.GetTreeId(), req.GetIncludeChildren())
	if err != nil {
		return nil, err
	}
	return tree.ToProto(), nil
}

func (s *ForestService) UpdateTree(ctx context.Context, req *forest.UpdateTreeRequest) (*forest.Tree, error) {
	inputTreeModel := &models.Tree{
		Id:   req.GetTreeId(),
		Name: req.GetName(),
		Url:  req.GetUrl(),
	}
	treeModel, err := s.Store.Neo4j.UpdateTree(ctx, inputTreeModel)
	if err != nil {
		return nil, err
	}
	return treeModel.ToProto(), nil
}

func (s *ForestService) DeleteTree(ctx context.Context, req *forest.DeleteTreeRequest) (*forest.DeleteTreeResponse, error) {
	err := s.Store.Neo4j.DeleteTree(ctx, req.GetTreeId(), req.GetCascade())
	if err != nil {
		return &forest.DeleteTreeResponse{
			Success: false,
		}, err
	}
	return &forest.DeleteTreeResponse{
		Success: true,
	}, nil
}
