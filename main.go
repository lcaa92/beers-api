package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/go-playground/validator/v10"
)

type FormRequest struct {
	Type string `json:"type" validate:"oneof=ale stouts red-ale"`
	Name string `json:"name"`
}

type APIResponseError struct {
	Error   int    `json:"error"`
	Message string `json:"message"`
}

type Rating struct {
	Average float32 `json:"average"`
	Reviews int32   `json:"reviews"`
}

type Beer struct {
	Id     int32  `json:"id"`
	Name   string `json:"name"`
	Price  string `json:"price"`
	Rating Rating `json:"rating"`
	Image  string `json:"image"`
}

func (b *Beer) UnmarshalJSON(data []byte) error {
	var aux struct {
		Id     any    `json:"id"`
		Name   string `json:"name"`
		Rating any    `json:"rating"`
		Price  any    `json:"price"`
		Image  string `json:"image"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.Id.(type) {
	case float64:
		b.Id = int32(v)
	case string:
		var idInt int
		if _, err := fmt.Sscanf(v, "%d", &idInt); err != nil {
			return fmt.Errorf("invalid id string: %v", err)
		}
		b.Id = int32(idInt)
	default:
		return fmt.Errorf("unsupported id type: %T", v)
	}

	switch v := aux.Price.(type) {
	case float64:
		b.Price = fmt.Sprintf("$%.2f", v)
	case string:
		b.Price = v
	default:
		return fmt.Errorf("unsupported price type: %T", v)
	}

	switch r := aux.Rating.(type) {
	case map[string]any:
		if avg, ok := r["average"].(float64); ok {
			b.Rating.Average = float32(avg)
		} else {
			b.Rating.Average = 0.0
		}
		if rev, ok := r["reviews"].(float64); ok {
			b.Rating.Reviews = int32(rev)
		} else {
			b.Rating.Reviews = 0
		}
	case string:
		b.Rating.Average = 0.0
		b.Rating.Reviews = 0
	default:
		return fmt.Errorf("unsupported rating type: %T", r)
	}

	b.Name = aux.Name
	b.Image = aux.Image
	return nil
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/beers", beersHandler)

	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Welcome to the home page!"})
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Hello, world!"})
}

func beersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	formRequest := FormRequest{
		Type: r.FormValue("type"),
		Name: r.FormValue("name"),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(formRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Validation error: %v", err)})
		return
	}

	baseUrl := fmt.Sprintf("https://api.sampleapis.com/beers/%s", formRequest.Type)
	parsedURL, err := url.Parse(baseUrl)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("Error parsing URL:", err)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Error parsing URL: %v", err)})
		return
	}

	queryParams := url.Values{}
	if formRequest.Name != "" {
		queryParams.Add("name", formRequest.Name)
	}
	parsedURL.RawQuery = queryParams.Encode()

	resp, err := http.Get(parsedURL.String())
	if err != nil {
		log.Println("Error fetching beers:", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Error reading response body: %v", err)})
	}
	defer resp.Body.Close()

	// Handle Samples API error response
	if ok := json.Unmarshal(body, &APIResponseError{}); ok == nil {
		var apiError APIResponseError
		log.Println("External API returned an error response: ", string(body))
		if err := json.Unmarshal(body, &apiError); err == nil {
			w.WriteHeader(apiError.Error)
			json.NewEncoder(w).Encode(apiError)
			return
		}
	}

	var beers []Beer
	err = json.Unmarshal(body, &beers)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Validation error: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(beers)
}
