package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
