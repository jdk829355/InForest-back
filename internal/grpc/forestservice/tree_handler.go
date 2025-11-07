package forestservice

import (
	"context"
	"errors"

	"github.com/jdk829355/InForest_back/models"
	"github.com/jdk829355/InForest_back/protos/forest"
)

// 메모 적용 완료
func (s *ForestService) CreateTree(ctx context.Context, req *forest.CreateTreeRequest) (*forest.CreateTreeResponse, error) {
	user_id := ctx.Value("user_id")
	if user_id == "" {
		return nil, errors.New("invalid user_id")
	}
	treeModel := &models.Tree{
		Id:   req.GetId(),
		Name: req.GetName(),
		Url:  req.GetUrl(),
	}
	id, err := s.Store.Neo4j.CreateTree(ctx, treeModel, req.GetParentId())
	if id == "" || err != nil {
		return nil, err
	}
	// 트리 생성 후 해당 메모 생성
	memo, err := s.Store.Supabase.CreateMemo(user_id.(string), id, nil)
	if err != nil {
		// 메모 생성 실패 시 트리도 삭제
		// TODO 다른 RPC 핸들러 호출해야겠음
		_, _ = s.Store.Neo4j.DeleteTree(ctx, id, true)
		return nil, err
	}
	return &forest.CreateTreeResponse{
		Tree: treeModel.ToProto(),
		Memo: memo.ToProto(),
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

// 트리 삭제 시 메모도 같이 삭제
func (s *ForestService) DeleteTree(ctx context.Context, req *forest.DeleteTreeRequest) (*forest.DeleteTreeResponse, error) {
	user_id := ctx.Value("user_id")
	if user_id == "" {
		return nil, errors.New("invalid user_id")
	}
	deletedIds, err := s.Store.Neo4j.DeleteTree(ctx, req.GetTreeId(), req.GetCascade())
	if err != nil {
		return &forest.DeleteTreeResponse{
			Success: false,
		}, err
	}
	deletedMemos := map[string]models.Memo{}
	for _, treeID := range deletedIds {
		memo, err := s.Store.Supabase.DeleteMemo(user_id.(string), treeID)
		if err != nil {
			for _, m := range deletedMemos {
				// 롤백: 삭제된 메모 복구
				_, _ = s.Store.Supabase.CreateMemo(m.UserID, m.TreeID, map[string]interface{}{
					"content": m.Content,
					"version": m.Version,
				})
			}
			return nil, err
		} else {
			deletedMemos[memo.TreeID] = *memo
		}
	}
	return &forest.DeleteTreeResponse{
		Success: true,
	}, nil
}
