package toys

import (
	"context"
	"github.com/spacecowboytobykty123/toysProto/gen/go/toys"
	"google.golang.org/grpc"
	"toysService/internal/data"
)

type serverAPI struct {
	toys.UnimplementedToysServer
	toys Toys
}

type Toys interface {
	CreateToy(ctx context.Context, title string, desc string, value int64, images []string, skills []string, categories []string, recAge string, manufacturer string, isAvailable bool) (toys.Status, string, data.Toy)
	DeleteToy(ctx context.Context, toyID int64) (toys.Status, string)
	ChangeToy(ctx context.Context, toy data.Toy) (toys.Status, string)
	GetToy(ctx context.Context, toyID int64) (data.Toy, toys.Status, string)
	ListToy(ctx context.Context, to int64, from int64, filters data.Filters) ([]data.Toy, toys.Status, string, toys.Metadata)
	ListRecommended(ctx context.Context) ([]data.Toy, toys.Status, string, toys.Metadata)
}

func Register(gRPC *grpc.Server, toy Toys) {
	toys.RegisterToysServer(gRPC, &serverAPI{toys: toy})
}
