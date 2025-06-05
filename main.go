package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/go-playground/validator/v10"
	formrequest "github.com/lcaa92/beers-api/internal/form_request"
	"github.com/lcaa92/beers-api/internal/sampleapis"
)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/beers", beersHandler)

	http.ListenAndServe(":8080", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"version": "0.1.0", "message": "Welcome to the Beers API!"})
}

func beersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	formRequest := formrequest.FormRequest{
		Type:    r.FormValue("type"),
		Name:    r.FormValue("name"),
		OrderBy: r.FormValue("order_by"),
		Order:   r.FormValue("order"),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("one_off_or_empty", formrequest.ValidateOneOfOrEmpty)

	err := validate.Struct(formRequest)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Validation error: %v", err)})
		return
	}

	baseUrl := fmt.Sprintf("https://api.sampleapis.com/beers/%s", formRequest.Type)
	resp, err := http.Get(baseUrl)
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
	if ok := json.Unmarshal(body, &sampleapis.APIResponseError{}); ok == nil {
		var apiError sampleapis.APIResponseError
		log.Println("External API returned an error response: ", string(body))
		if err := json.Unmarshal(body, &apiError); err == nil {
			w.WriteHeader(apiError.Error)
			json.NewEncoder(w).Encode(apiError)
			return
		}
	}

	var beers []sampleapis.Beer
	err = json.Unmarshal(body, &beers)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Validation error: %v", err)})
		return
	}

	if formRequest.Name != "" {
		beers = filterBeersByName(beers, formRequest.Name)
	}

	if formRequest.OrderBy != "" {
		beers = orderBy(beers, formRequest.OrderBy, formRequest.Order)
	}

	json.NewEncoder(w).Encode(beers)
}

func filterBeersByName(beers []sampleapis.Beer, name string) []sampleapis.Beer {
	if name == "" {
		return beers
	}

	var filtered []sampleapis.Beer
	for _, beer := range beers {
		if strings.Contains(strings.ToLower(beer.Name), strings.ToLower(name)) {
			filtered = append(filtered, beer)
		}
	}
	return filtered
}

func orderBy(beers []sampleapis.Beer, orderBy string, order string) []sampleapis.Beer {
	if orderBy == "" {
		return beers // No ordering specified
	}

	if order == "" {
		order = "asc" // Default order
	}

	// Sort beers based on the orderBy field
	switch orderBy {
	case "id":
		sort.Slice(beers, func(i, j int) bool {
			if order == "asc" {
				return beers[i].Id < beers[j].Id
			}
			return beers[i].Id > beers[j].Id
		})
	case "name":
		sort.Slice(beers, func(i, j int) bool {
			if order == "asc" {
				return beers[i].Name < beers[j].Name
			}
			return beers[i].Name > beers[j].Name
		})
	case "price":
		sort.Slice(beers, func(i, j int) bool {
			if order == "asc" {
				return beers[i].Price < beers[j].Price
			}
			return beers[i].Price > beers[j].Price
		})
	case "average":
		sort.Slice(beers, func(i, j int) bool {
			if order == "asc" {
				return beers[i].Rating.Average < beers[j].Rating.Average
			}
			return beers[i].Rating.Average > beers[j].Rating.Average
		})
	case "reviews":
		sort.Slice(beers, func(i, j int) bool {
			if order == "asc" {
				return beers[i].Rating.Reviews < beers[j].Rating.Reviews
			}
			return beers[i].Rating.Reviews > beers[j].Rating.Reviews
		})
	default:
		log.Println("Invalid order_by value:", orderBy)
	}

	return beers
}
