package utils

import (
	"math"
)

// PagedResponse contains the fields for pagination in the JSON response and all model items
type PagedResponse struct {
	Total       uint        `json:"total"`
	Limit       uint        `json:"limit"`
	PerPage     uint        `json:"per_page"`
	Pages       uint        `json:"pages"`
	CurrentPage uint        `json:"current_page"`
	Items       interface{} `json:"items"`
}

// Get query limit and total pages
func (p *PagedResponse) Get() {
	p.Limit = ((p.CurrentPage - 1) * p.PerPage)
	p.Pages = uint(math.Ceil(float64(p.Total) / float64(p.PerPage)))

	if p.Pages == 0 {
		p.Pages = 1
	}

}
