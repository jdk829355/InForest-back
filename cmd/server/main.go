package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jdk829355/InForest_back/config"
	app "github.com/jdk829355/InForest_back/internal/grpc/forestservice"
	"github.com/jdk829355/InForest_back/internal/store"
	gen "github.com/jdk829355/InForest_back/protos/forest"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

func main() {
	// 1. Zap 로거 초기화 (프로덕션 환경용)
	// 개발 중에는 zap.NewDevelopment()를 사용하면 더 읽기 편한 로그가 출력됩니다.
	logger, err := zap.NewProduction()
	if err != nil {
		// 로거 초기화 실패 시 표준 log로 최후의 메시지를 남기고 패닉
		// (이 시점에서는 logger.Fatal을 사용할 수 없음)
		fmt.Printf("failed to initialize zap logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() // 애플리케이션 종료 전 버퍼 로그 플러시

	// 2. 설정 로딩
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}
	logger.Info("Configuration loaded successfully")

	// 3. 데이터베이스 연결
	driver, err := store.InitNeo4jStore(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to Neo4j", zap.Error(err))
	}
	store := store.NewStore(*driver)
	ctx := context.Background()

	// defer를 사용하여 main 함수 종료 시 DB 연결을 닫도록 설정
	defer func() {
		logger.Info("Closing database connection...")
		if err := store.Neo4j.Close(ctx); err != nil {
			logger.Error("Failed to close database connection", zap.Error(err))
		}
	}()
	logger.Info("Database connection established")

	// 4. gRPC 서비스 로직 초기화
	forestService := app.NewForestService(store)

	// 5. gRPC 리스너 설정
	listenAddr := fmt.Sprintf(":%s", cfg.GRPC_PORT)
	l, e := net.Listen("tcp", listenAddr)
	if e != nil {
		logger.Fatal("Failed to listen on address",
			zap.String("address", listenAddr),
			zap.Error(e),
		)
	}

	// 6. Zap 로깅 및 Recovery 인터셉터 설정
	// (이전 대화에서 설명한 내용)
	loggingOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ms", duration.Milliseconds())
		}),
	}

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			// 1. Recovery: 패닉이 발생하면 복구하고 Internal 에러를 반환
			grpc_recovery.UnaryServerInterceptor(),
			// 2. Logging: Zap을 사용해 요청/응답을 로깅
			grpc_zap.UnaryServerInterceptor(logger, loggingOpts...),
			// (추가 인터셉터가 있다면 여기에...)
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			// 1. Recovery
			grpc_recovery.StreamServerInterceptor(),
			// 2. Logging
			grpc_zap.StreamServerInterceptor(logger, loggingOpts...),
		)),
	}

	// 7. gRPC 서버 설정 및 등록
	// 기존: s := grpc.NewServer()
	s := grpc.NewServer(serverOptions...) // 인터셉터 옵션 적용

	gen.RegisterForestServiceServer(s, forestService)

	// 8. gRPC 서버 시작 (고루틴)
	go func() {
		logger.Info("Starting gRPC server", zap.String("address", listenAddr))
		if err := s.Serve(l); err != nil {
			// Serve가 정상 종료(GracefulStop) 외의 이유로 중단되면 Fatal 로깅
			logger.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// 9. 우아한 종료(Graceful Shutdown) 설정
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 시그널이 수신될 때까지 대기
	sig := <-quit
	logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

	// gRPC 서버의 우아한 종료 (진행 중인 요청 완료 대기)
	s.GracefulStop()
	logger.Info("gRPC server stopped gracefully.")
}
