package formrequest

type FormRequest struct {
	Type string `json:"type" validate:"oneof=ale stouts red-ale"`
	Name string `json:"name"`
}
