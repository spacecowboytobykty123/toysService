package toys

import (
	"context"
	"github.com/spacecowboytobykty123/toysProto/gen/go/toys"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
	subgrpc "toysService/internal/clients/subscriptions/grpc"
	"toysService/internal/contextkeys"
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
	CreateToy(ctx context.Context, inputToy data.Toy) (toys.Status, string, data.Toy)
	DeleteToy(ctx context.Context, toyID int64) (toys.Status, string)
	ChangeToy(ctx context.Context, toy data.Toy) (toys.Status, string)
	GetToy(ctx context.Context, toyID int64) (data.Toy, toys.Status, string)
	GetToysByIds(ctx context.Context, ids []int64) ([]*data.ToySummary, string)
	ListToy(ctx context.Context, to int64, from int64, filters data.Filters, categories []string, skills []string, title string) ([]*data.Toy, toys.Status, string, data.Metadata)
	ListRecommended(ctx context.Context, userID int64) ([]*data.Toy, toys.Status, string, data.Metadata)
}

func New(log *jsonlog.Logger, toysProvider toysProvider, tokenTTL time.Duration, subsClient *subgrpc.Client) *Toys {
	return &Toys{
		log:          log,
		toysProvider: toysProvider,
		tokenTTL:     tokenTTL,
		subsClient:   subsClient,
	}
}

func (t *Toys) CreateToy(ctx context.Context, inputToy data.Toy) (toys.Status, string, data.Toy) {
	t.log.PrintInfo("business logic layer", map[string]string{
		"method": "toys.CreateToy",
	})
	println("До db")
	// TODO: нужен метод проверяющий админ ли или нет
	opStatus, msg, toy := t.toysProvider.CreateToy(ctx, inputToy)
	println("past db")
	println(opStatus.String())

	if opStatus != toys.Status_STATUS_OK {
		t.log.PrintError(status.Error(codes.Internal, "internal error"), map[string]string{
			"method": "toys.CreateToy",
		})

		return toys.Status_STATUS_INTERNAL_ERROR, "internal error", data.Toy{}
	}

	return opStatus, msg, toy
}

func (t *Toys) DeleteToy(ctx context.Context, toyID int64) (toys.Status, string) {
	// TODO: нужен метод проверяющий админ ли или нет
	t.log.PrintInfo("business logic layer", map[string]string{
		"method": "toys.DeleteToy",
	})
	opStatus, msg := t.toysProvider.DeleteToy(ctx, toyID)

	if opStatus != toys.Status_STATUS_OK {
		t.log.PrintError(status.Error(codes.Internal, "internal error"), map[string]string{
			"method": "toys.DeleteToy",
		})

		return toys.Status_STATUS_INTERNAL_ERROR, "internal error"
	}

	return opStatus, msg
}

func (t *Toys) ChangeToy(ctx context.Context, toy data.Toy) (toys.Status, string) {
	// TODO: нужен метод проверяющий админ ли или нет
	t.log.PrintInfo("business logic layer", map[string]string{
		"method": "toys.ChangeToy",
	})
	opStatus, msg := t.toysProvider.ChangeToy(ctx, toy)
	if opStatus != toys.Status_STATUS_OK {
		t.log.PrintError(status.Error(codes.Internal, "internal error"), map[string]string{
			"method": "toys.ChangeToy",
		})

		return toys.Status_STATUS_INTERNAL_ERROR, "internal error"
	}
	return opStatus, msg

}

func (t *Toys) GetToy(ctx context.Context, toyID int64) (data.Toy, toys.Status, string) {
	t.log.PrintInfo("business logic layer", map[string]string{
		"method": "toys.GetToy",
	})
	toy, opStatus, msg := t.toysProvider.GetToy(ctx, toyID)
	if opStatus != toys.Status_STATUS_OK {
		t.log.PrintError(status.Error(codes.Internal, "internal error"), map[string]string{
			"method": "toys.GetToy",
		})

		return data.Toy{}, toys.Status_STATUS_INTERNAL_ERROR, "internal error!"
	}

	return toy, opStatus, msg
}

func (t *Toys) ListToy(ctx context.Context, to int64, from int64, filters data.Filters, categories []string, skills []string, title string) ([]*data.Toy, toys.Status, string, data.Metadata) {
	t.log.PrintInfo("business logic layer", map[string]string{
		"method": "toys.ListToy",
	})
	if from == 0 {
		from = 10
	}
	if to == 0 {
		to = 10000000
	}
	toyList, opStatus, msg, metadata := t.toysProvider.ListToy(ctx, to, from, filters, categories, skills, title)

	if opStatus != toys.Status_STATUS_OK {
		t.log.PrintError(status.Error(codes.Internal, "internal error"), map[string]string{
			"method": "toys.ListToy",
		})

		return []*data.Toy{}, toys.Status_STATUS_INTERNAL_ERROR, "internal error!", data.Metadata{}
	}

	return toyList, opStatus, msg, metadata

}

func (t *Toys) ListRecommended(ctx context.Context) ([]*data.Toy, toys.Status, string, data.Metadata) {
	t.log.PrintInfo("business logic layer", map[string]string{
		"method": "toys.ListRec",
	})
	userId, err := getUserFromContext(ctx)
	if err != nil {
		return []*data.Toy{}, toys.Status_STATUS_INTERNAL_ERROR, "internal error", data.Metadata{}
	}

	toyList, opStatus, msg, metadata := t.toysProvider.ListRecommended(ctx, userId)
	if opStatus != toys.Status_STATUS_OK {
		t.log.PrintError(status.Error(codes.Internal, "internal error"), map[string]string{
			"method": "toys.ListRecommended",
		})

		return []*data.Toy{}, toys.Status_STATUS_INTERNAL_ERROR, "internal error!", data.Metadata{}
	}

	return toyList, opStatus, msg, metadata
}

func (t *Toys) GetToysByIds(ctx context.Context, ids []int64) ([]*data.ToySummary, string) {
	t.log.PrintInfo("business logic part", map[string]string{
		"method": "toys.GetToysByIds",
	})

	toyList, msg := t.toysProvider.GetToysByIds(ctx, ids)
	if len(toyList) < 1 {
		return []*data.ToySummary{}, "could not fetch toys"
	}
	return toyList, msg
}

func getUserFromContext(ctx context.Context) (int64, error) {
	val := ctx.Value(contextkeys.UserIDKey)
	userID, ok := val.(int64)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "user id is missing or invalid in context")
	}

	return userID, nil

}
