package forestservice

import (
	"context"
	"errors"

	"github.com/jdk829355/InForest_back/protos/forest"
)

func (s *ForestService) GetMemo(ctx context.Context, req *forest.GetMemoRequest) (*forest.Memo, error) {
	user_id := ctx.Value("user_id")
	if user_id == "" {
		return nil, errors.New("invalid user_id")
	}
	memo, err := s.Store.Supabase.GetMemo(user_id.(string), req.GetTreeId())
	if err != nil {
		return nil, err
	}
	return memo.ToProto(), nil
}
