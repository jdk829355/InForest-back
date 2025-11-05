package auth

import "context"

type service struct {
	secret []byte
}

func NewAuthService(secret string) (*service, error) {
	return &service{
		secret: []byte(secret),
	}, nil
}

func (s *service) ValidateToken(_ context.Context, token string) (string, error) {
	// TODO 실제 검증 로직 작성하기
	return "jdk829355", nil
}
