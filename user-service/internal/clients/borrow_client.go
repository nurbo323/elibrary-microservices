package clients

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"elibrary/gen/borrowpb"
)

type BorrowClient struct {
	conn *grpc.ClientConn
	cli  borrowpb.BorrowServiceClient
}

func NewBorrowClient(addr string) (*BorrowClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial borrow: %w", err)
	}
	return &BorrowClient{conn: conn, cli: borrowpb.NewBorrowServiceClient(conn)}, nil
}

func (c *BorrowClient) GetActiveByUser(ctx context.Context, userID string) (*borrowpb.ListBorrowsResponse, error) {
	return c.cli.GetActiveBorrows(ctx, &borrowpb.GetActiveBorrowsRequest{
		Limit:  10,
		Offset: 0,
	})
}

func (c *BorrowClient) GetHistoryByUser(ctx context.Context, userID string) (*borrowpb.ListBorrowsResponse, error) {
	return c.cli.GetUserBorrowHistory(ctx, &borrowpb.GetUserBorrowHistoryRequest{UserId: userID})
}

func (c *BorrowClient) Close() error { return c.conn.Close() }
