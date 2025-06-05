package formrequest

type FormRequest struct {
	Type    string `json:"type" validate:"oneof=ale stouts red-ale"`
	Name    string `json:"name"`
	OrderBy string `json:"order_by" validate:"one_off_or_empty=id name price average reviews" default:"id"` // Custom validation tag
	Order   string `json:"order" validate:"one_off_or_empty=asc desc" default:"asc"`
}
