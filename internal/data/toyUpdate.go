package data

type UpdatedToy struct {
	ID           int64
	Title        *string
	Desc         *string
	Value        *int64
	Images       []string
	Skills       []string
	Categories   []string
	RecAge       *string
	Manufacturer *string
	IsAvailable  *bool
}
