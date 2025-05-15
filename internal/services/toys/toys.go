package toys

import (
	"context"
	"github.com/spacecowboytobykty123/toysProto/gen/go/toys"
	"time"
	subgrpc "toysService/internal/clients/subscriptions/grpc"
	"toysService/internal/data"
	"toysService/internal/jsonlog"
)

type Toys struct {
	log          *jsonlog.Logger
	toysProvider toysProvider
	tokenTTL     time.Duration
	subsClient   *subgrpc.Client
}

type toysProvider interface {
	CreateToy(ctx context.Context, title string, desc string, value int64, images []string, skills []string, categories []string, recAge string, manufacturer string, isAvailable bool) (toys.Status, string, data.Toy)
	DeleteToy(ctx context.Context, toyID int64) (toys.Status, string)
	ChangeToy(ctx context.Context, toy data.Toy) (toys.Status, string)
	GetToy(ctx context.Context, toyID int64) (data.Toy, toys.Status, string)
	ListToy(ctx context.Context, to int64, from int64, filters data.Filters) ([]data.Toy, toys.Status, string, data.Metadata)
	ListRecommended(ctx context.Context) ([]data.Toy, toys.Status, string, data.Metadata)
}

func New(log *jsonlog.Logger, toyProvider toysProvider, tokenTTL time.Duration, subsClient *subgrpc.Client) *Toys {
	return &Toys{
		log:          log,
		toysProvider: toyProvider,
		tokenTTL:     tokenTTL,
		subsClient:   subsClient,
	}
}

func (t Toys) CreateToy(ctx context.Context, title string, desc string, value int64, images []string, skills []string, categories []string, recAge string, manufacturer string, isAvailable bool) (toys.Status, string, data.Toy) {
	//TODO implement me
	panic("implement me")
}

func (t Toys) DeleteToy(ctx context.Context, toyID int64) (toys.Status, string) {
	//TODO implement me
	panic("implement me")
}

func (t Toys) ChangeToy(ctx context.Context, toy data.Toy) (toys.Status, string) {
	//TODO implement me
	panic("implement me")
}

func (t Toys) GetToy(ctx context.Context, toyID int64) (data.Toy, toys.Status, string) {
	//TODO implement me
	panic("implement me")
}

func (t Toys) ListToy(ctx context.Context, to int64, from int64, filters data.Filters) ([]data.Toy, toys.Status, string, data.Metadata) {
	//TODO implement me
	panic("implement me")
}

func (t Toys) ListRecommended(ctx context.Context) ([]data.Toy, toys.Status, string, data.Metadata) {
	//TODO implement me
	panic("implement me")
}
