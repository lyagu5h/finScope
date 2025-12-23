package client

import (
	ledgerv1 "github.com/lyagu5h/finScope/gateway/internal/delivery/protos/ledger/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client ledgerv1.LedgerServiceClient
}

func New(addr string) (*Client, error) {

	c, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: ledgerv1.NewLedgerServiceClient(c),
	}, nil
}

func (c *Client) Ledger() ledgerv1.LedgerServiceClient {
	return c.client
}