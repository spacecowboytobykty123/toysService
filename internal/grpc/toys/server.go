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
	"toysService/internal/jsonlog"
	"toysService/internal/validator"
	"toysService/storage/postgres"
)

type serverAPI struct {
	toys.UnimplementedToysServer
	toys Toys
	log  *jsonlog.Logger
}

type Toys interface {
	CreateToy(ctx context.Context, toy data.Toy) (toys.Status, string, data.Toy)
	DeleteToy(ctx context.Context, toyID int64) (toys.Status, string)
	ChangeToy(ctx context.Context, toy data.Toy) (toys.Status, string)
	GetToy(ctx context.Context, toyID int64) (data.Toy, toys.Status, string)
	ListToy(ctx context.Context, to int64, from int64, filters data.Filters, categories []string, skills []string, title string) ([]*data.Toy, toys.Status, string, data.Metadata)
	ListRecommended(ctx context.Context) ([]*data.Toy, toys.Status, string, data.Metadata)
	GetToysByIds(ctx context.Context, ids []int64) ([]*data.ToySummary, string)
}

func Register(gRPC *grpc.Server, toy Toys, log *jsonlog.Logger) {
	toys.RegisterToysServer(gRPC, &serverAPI{toys: toy, log: log})
}

func (s *serverAPI) GetToysByIds(ctx context.Context, r *toys.GetToysByIdsRequest) (*toys.GetToysByIdsResponse, error) {
	s.log.PrintInfo("server part", map[string]string{
		"method": "server.GetToysByIds",
	})
	toyIds := r.GetId()
	if len(toyIds) < 1 {
		s.log.PrintError(fmt.Errorf("no toy ids were provided!"), map[string]string{
			"method": "server.GetToysByIds",
		})
		return nil, status.Error(codes.NotFound, "toy ids not provided")
	}

	toysList, msg := s.toys.GetToysByIds(ctx, toyIds)

	if len(toysList) < 1 {
		return nil, status.Error(codes.NotFound, "failed to fetch toys!")
	}

	return &toys.GetToysByIdsResponse{
		Toy: mapToySummary(toysList),
		Msg: msg,
	}, nil
}

func (s *serverAPI) CreateToy(ctx context.Context, r *toys.CreateToyRequest) (*toys.CreateToyResponse, error) {
	s.log.PrintInfo("server part", map[string]string{
		"method": "server.CreateToy",
	})
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

	inputToy := data.Toy{
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

	if postgres.ValidateToy(v, &inputToy); !v.Valid() {
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
	//s.log.PrintInfo("server part", map[string]string{
	//	"method": "server.DeleteToy",
	//})
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
	//s.log.PrintInfo("server part", map[string]string{
	//	"method": "server.ChangeToy",
	//})
	v := validator.New()
	toyProto := r.GetToy()

	existingToy, opStatus, msg := s.toys.GetToy(ctx, toyProto.Id)
	if opStatus != toys.Status_STATUS_OK {
		return nil, status.Error(codes.Internal, "internal error")
	}

	if toyProto.Title != nil {
		existingToy.Title = *toyProto.Title
	}
	if toyProto.Desc != nil {
		existingToy.Desc = *toyProto.Desc
	}
	if toyProto.Value != nil {
		existingToy.Value = *toyProto.Value
	}
	if toyProto.Images != nil {
		existingToy.Images = toyProto.Images
	}
	if toyProto.Categories != nil {
		existingToy.Categories = toyProto.Categories
	}
	if toyProto.Skills != nil {
		existingToy.Skills = toyProto.Skills
	}
	if toyProto.Manufacturer != nil {
		existingToy.Manufacturer = *toyProto.Manufacturer
	}
	if toyProto.RecommendedAge != nil {
		existingToy.RecAge = *toyProto.RecommendedAge
	}
	if toyProto.IsAvailable != nil {
		existingToy.IsAvailable = *toyProto.IsAvailable
	}

	if postgres.ValidateToy(v, &existingToy); !v.Valid() {
		return nil, collectErrors(v)
	}

	opStatus, msg = s.toys.ChangeToy(ctx, existingToy)
	if opStatus != toys.Status_STATUS_OK {
		return nil, status.Error(codes.Internal, "internal error!")
	}

	return &toys.ChangeToyResponse{
		Status:   opStatus,
		ErrorMsg: msg,
	}, nil
}

func (s *serverAPI) GetToy(ctx context.Context, r *toys.GetToyRequest) (*toys.GetToyResponse, error) {
	//s.log.PrintInfo("server part", map[string]string{
	//	"method": "server.GetToy",
	//})
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
	//s.log.PrintInfo("server part", map[string]string{
	//	"method": "server.ListToy",
	//})
	from := r.GetFrom()
	categories := r.GetCategories()
	skills := r.GetSkills()
	to := r.GetTo()
	title := r.GetTitle()
	v := validator.New()

	filters := &data.Filters{
		Page:         r.GetPage(),
		PageSize:     r.GetPageSize(),
		Sort:         r.GetSort(),
		SortSafelist: []string{"id", "title", "skills", "categories", "recAge", "value", "from", "to", "-id", "-title", "-skills", "-categories", "-recAge", "-value"},
	}
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}
	if filters.Sort == "" {
		filters.Sort = "id"
	}
	//if from == 0 {
	//	return nil, status.Error(codes.InvalidArgument, "Invalid filter(from)")
	//}
	//if to == 0 {
	//	return nil, status.Error(codes.InvalidArgument, "Invalid filter(to)")
	//}

	if data.ValidateFilters(v, *filters); !v.Valid() {
		return nil, collectErrors(v)
	}

	toyList, opStatus, msg, metadata := s.toys.ListToy(ctx, to, from, *filters, categories, skills, title)

	return &toys.ListToyResponse{
		Toys:     mapDataListToGrpc(toyList),
		Status:   opStatus,
		ErrorMsg: msg,
		Metadata: mapDataMetToGrpc(metadata),
	}, nil

}

func (s *serverAPI) ListRecommended(ctx context.Context, r *toys.ListRecommendedRequest) (*toys.ListRecommendedResponse, error) {
	//s.log.PrintInfo("server part", map[string]string{
	//	"method": "server.ListRec",
	//})
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
	println(b.String())
	return status.Error(codes.InvalidArgument, b.String())
}

func mapDataToGRPCToy(toy data.Toy) *toys.Toy {
	return &toys.Toy{
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
	}
}

func mapDataListToGrpc(toyList []*data.Toy) []*toys.Toy {
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

func mapToySummary(toyList []*data.ToySummary) []*toys.ToySummary {
	items := make([]*toys.ToySummary, 0, len(toyList))
	for _, toy := range toyList {
		items = append(items, &toys.ToySummary{
			Id:       toy.ID,
			Title:    toy.Title,
			Value:    toy.Value,
			ImageUrl: toy.URL,
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
