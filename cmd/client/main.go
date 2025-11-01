package main

import (
	"context"
	"fmt"
	"os"

	gen "github.com/jdk829355/InForest_back/protos/forest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 사용법:
	// go run ./cmd/client get <user_id>
	// go run ./cmd/client create-forest <user_id> <forest_name>
	// go run ./cmd/client create-tree <parent_id> <tree_id> <name> <url>

	if len(os.Args) < 2 {
		fmt.Println("usage: get|create-forest|create-tree ...")
		os.Exit(1)
	}

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}
	addr := fmt.Sprintf("app:%s", port)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to dial gRPC server: %v\n", err)
		os.Exit(2)
	}
	defer conn.Close()

	client := gen.NewForestServiceClient(conn)

	cmd := os.Args[1]
	switch cmd {
	case "get":
		if len(os.Args) < 3 {
			fmt.Println("usage: get <user_id>")
			os.Exit(1)
		}
		userID := os.Args[2]
		resp, err := client.GetForestsByUser(context.Background(), &gen.GetForestsByUserRequest{UserId: userID})
		if err != nil {
			fmt.Fprintf(os.Stderr, "GetForestsByUser error: %v\n", err)
			os.Exit(3)
		}
		fmt.Printf("Got %d forests for user %s:\n", len(resp.GetForests()), userID)
		for i, f := range resp.GetForests() {
			fmt.Printf("%d) id=%s name=%s description=%s user_id=%s\n", i+1, f.GetId(), f.GetName(), f.GetDescription(), f.GetUserId())
		}

	case "create-forest":
		if len(os.Args) < 4 {
			fmt.Println("usage: create-forest <user_id> <forest_name>")
			os.Exit(1)
		}
		userID := os.Args[2]
		name := os.Args[3]
		// create a minimal root tree
		root := &gen.Tree{Id: "root-1", Name: "root", Url: "http://example.com"}
		req := &gen.CreateForestRequest{
			Name:        name,
			Description: "created-by-client",
			UserId:      userID,
			Root:        root,
		}
		created, err := client.CreateForest(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "CreateForest error: %v\n", err)
			os.Exit(3)
		}
		fmt.Printf("Created forest: id=%s name=%s user_id=%s\n", created.GetId(), created.GetName(), created.GetUserId())

	case "create-tree":
		if len(os.Args) < 6 {
			fmt.Println("usage: create-tree <parent_id> <tree_id> <name> <url>")
			os.Exit(1)
		}
		parentID := os.Args[2]
		treeID := os.Args[3]
		name := os.Args[4]
		url := os.Args[5]
		req := &gen.CreateTreeRequest{
			ParentId: parentID,
			Id:       treeID,
			Name:     name,
			Url:      url,
		}
		created, err := client.CreateTree(context.Background(), req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "CreateTree error: %v\n", err)
			os.Exit(3)
		}
		fmt.Printf("Created tree: id=%s name=%s url=%s\n", created.GetId(), created.GetName(), created.GetUrl())

	default:
		fmt.Println("unknown command", cmd)
		os.Exit(1)
	}
}
