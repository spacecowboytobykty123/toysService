package data

import (
	"math"
	"strings"
	"toysService/internal/validator"
)

type Filters struct {
	Page         int32
	PageSize     int32
	Sort         string
	SortSafelist []string
}

type Metadata struct {
	CurrentPage  int32
	PageSize     int32
	FirstPage    int32
	LastPage     int32
	TotalRecords int32
}

func (f Filters) CalculateMetadata(totalRecords int, page, pageSize int32) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int32(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: int32(totalRecords),
	}
}

func (f Filters) Limit() int32 {
	return f.PageSize
}

func (f Filters) Offset() int32 {
	return (f.Page - 1) * f.PageSize
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "maximum is 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "maximum is 100")

	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort type")
}

func (f Filters) SortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) SortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}
