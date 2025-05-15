package toys

import (
	"context"
	"fmt"
	"github.com/spacecowboytobykty123/toysProto/gen/go/toys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"toysService/internal/data"
	"toysService/internal/validator"
	"toysService/storage/postgres"
)

type serverAPI struct {
	toys.UnimplementedToysServer
	toys Toys
}

type Toys interface {
	CreateToy(ctx context.Context, toy *data.Toy) (toys.Status, string, data.Toy)
	DeleteToy(ctx context.Context, toyID int64) (toys.Status, string)
	ChangeToy(ctx context.Context, toy data.Toy) (toys.Status, string)
	GetToy(ctx context.Context, toyID int64) (data.Toy, toys.Status, string)
	ListToy(ctx context.Context, to int64, from int64, filters data.Filters) ([]data.Toy, toys.Status, string, data.Metadata)
	ListRecommended(ctx context.Context) ([]data.Toy, toys.Status, string, data.Metadata)
}

func Register(gRPC *grpc.Server, toy Toys) {
	toys.RegisterToysServer(gRPC, &serverAPI{toys: toy})
}

func (s *serverAPI) CreateToy(ctx context.Context, r *toys.CreateToyRequest) (*toys.CreateToyResponse, error) {
	v := validator.New()
	toyTitle := r.GetTitle()
	toyDesc := r.GetDesc()
	toyValue := r.GetValue()
	toyImages := r.GetImages()
	toySkills := r.GetSkills()
	toyCategories := r.GetCategories()
	toyRecAge := r.GetRecommendedAge()
	toyManufacturer := r.GetManufacturer()
	isToyAvailable := r.GetIsAvailable()

	inputToy := &data.Toy{
		Title:        toyTitle,
		Desc:         toyDesc,
		Value:        toyValue,
		Images:       toyImages,
		Skills:       toySkills,
		Categories:   toyCategories,
		RecAge:       toyRecAge,
		Manufacturer: toyManufacturer,
		IsAvailable:  isToyAvailable,
	}

	if postgres.ValidateToy(v, inputToy); !v.Valid() {
		return nil, collectErrors(v)
	}

	opStatus, msg, toy := s.toys.CreateToy(ctx, inputToy)

	if opStatus != toys.Status_STATUS_OK {
		return nil, status.Error(codes.Internal, "internal error!")
	}

	return &toys.CreateToyResponse{
		Status:   opStatus,
		ErrorMsg: msg,
		Toy:      mapDataToGRPCToy(toy),
	}, nil

}

func (s *serverAPI) DeleteToy(ctx context.Context, r *toys.DeleteToyRequest) (*toys.DeleteToyResponse, error) {
	toyId := r.GetToyId()
	if toyId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid ToyId")
	}

	opStatus, msg := s.toys.DeleteToy(ctx, toyId)
	if opStatus != toys.Status_STATUS_OK {
		return nil, status.Error(codes.Internal, "internal error!")
	}

	return &toys.DeleteToyResponse{
		Status:   opStatus,
		ErrorMsg: msg,
	}, nil
}

func (s *serverAPI) ChangeToy(ctx context.Context, r *toys.ChangeToyRequest) (*toys.ChangeToyResponse, error) {
	v := validator.New()
	toyProto := r.GetToy()

	inputToy := &data.Toy{
		ID:           toyProto.Id,
		Title:        toyProto.Title,
		Desc:         toyProto.Desc,
		Value:        toyProto.Value,
		Images:       toyProto.Images,
		Skills:       toyProto.Skills,
		Categories:   toyProto.Categories,
		RecAge:       toyProto.RecommendedAge,
		Manufacturer: toyProto.Manufacturer,
		IsAvailable:  toyProto.IsAvailable,
	}

	if postgres.ValidateToy(v, inputToy); !v.Valid() {
		return nil, collectErrors(v)
	}

	opStatus, msg := s.toys.ChangeToy(ctx, *inputToy)
	if opStatus != toys.Status_STATUS_OK {
		return nil, status.Error(codes.Internal, "internal error!")
	}

	return &toys.ChangeToyResponse{
		Status:   opStatus,
		ErrorMsg: msg,
	}, nil
}

func (s *serverAPI) GetToy(ctx context.Context, r *toys.GetToyRequest) (*toys.GetToyResponse, error) {
	toyId := r.GetToyId()
	if toyId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid ToyId")
	}

	toy, opStatus, msg := s.toys.GetToy(ctx, toyId)

	return &toys.GetToyResponse{
		Toy:    mapDataToGRPCToy(toy),
		Status: opStatus,
		Msg:    msg,
	}, nil
}

func (s *serverAPI) ListToy(ctx context.Context, r *toys.ListToyRequest) (*toys.ListToyResponse, error) {
	from := r.GetFrom()
	to := r.GetTo()
	v := validator.New()

	filters := &data.Filters{
		Page:         r.GetPage(),
		PageSize:     r.GetPageSize(),
		Sort:         r.GetSort(),
		SortSafelist: r.GetSortSafeList(),
	}
	if from == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid filter(from)")
	}
	if to == 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid filter(to)")
	}

	if data.ValidateFilters(v, *filters); !v.Valid() {
		return nil, collectErrors(v)
	}

	toyList, opStatus, msg, metadata := s.toys.ListToy(ctx, to, from, *filters)

	return &toys.ListToyResponse{
		Toys:     mapDataListToGrpc(toyList),
		Status:   opStatus,
		ErrorMsg: msg,
		Metadata: mapDataMetToGrpc(metadata),
	}, nil

}

func (s *serverAPI) ListRecommended(ctx context.Context, r *toys.ListRecommendedRequest) (*toys.ListRecommendedResponse, error) {
	toyList, opStatus, msg, metadata := s.toys.ListRecommended(ctx)

	return &toys.ListRecommendedResponse{
		Toys:     mapDataListToGrpc(toyList),
		Status:   opStatus,
		ErrorMsg: msg,
		Metadata: mapDataMetToGrpc(metadata),
	}, nil
}

func collectErrors(v *validator.Validator) error {
	var b strings.Builder
	for field, msg := range v.Errors {
		fmt.Fprintf(&b, "%s:%s; ", field, msg)
	}
	return status.Error(codes.InvalidArgument, b.String())
}

func mapDataToGRPCToy(toy data.Toy) *toys.Toy {
	return &toys.Toy{
		Title:          toy.Title,
		Desc:           toy.Desc,
		Value:          toy.Value,
		Images:         toy.Images,
		Skills:         toy.Skills,
		Categories:     toy.Categories,
		RecommendedAge: toy.RecAge,
		Manufacturer:   toy.Manufacturer,
		IsAvailable:    toy.IsAvailable,
	}
}

func mapDataListToGrpc(toyList []data.Toy) []*toys.Toy {
	items := make([]*toys.Toy, 0, len(toyList))
	for _, toy := range toyList {
		items = append(items, &toys.Toy{
			Id:             toy.ID,
			Title:          toy.Title,
			Desc:           toy.Desc,
			Value:          toy.Value,
			Images:         toy.Images,
			Skills:         toy.Skills,
			Categories:     toy.Categories,
			RecommendedAge: toy.RecAge,
			Manufacturer:   toy.Manufacturer,
			IsAvailable:    toy.IsAvailable,
		})
	}
	return items
}

func mapDataMetToGrpc(metadata data.Metadata) *toys.Metadata {
	return &toys.Metadata{
		CurrentPage:  metadata.CurrentPage,
		PageSize:     metadata.PageSize,
		FirstPage:    metadata.FirstPage,
		LastPage:     metadata.LastPage,
		TotalRecords: metadata.TotalRecords,
	}
}
