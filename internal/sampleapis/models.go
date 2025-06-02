package sampleapis

import (
	"encoding/json"
	"fmt"
)

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
