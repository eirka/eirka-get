package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagination(t *testing.T) {

	paged := PagedResponse{
		CurrentPage: 1,
		PerPage:     50,
		Total:       132,
	}

	paged.Get()

	assert.Equal(t, paged.Limit, uint(0), "Limit should match")
	assert.Equal(t, paged.Pages, uint(3), "Pages should match")

	onepage := PagedResponse{
		CurrentPage: 1,
		PerPage:     50,
		Total:       33,
	}

	onepage.Get()

	assert.Equal(t, onepage.Limit, uint(0), "Limit should match")
	assert.Equal(t, onepage.Pages, uint(1), "Pages should match")

	manypages := PagedResponse{
		CurrentPage: 3,
		PerPage:     50,
		Total:       344,
	}

	manypages.Get()

	assert.Equal(t, manypages.Limit, uint(100), "Limit should match")
	assert.Equal(t, manypages.Pages, uint(7), "Pages should match")

	zeropage := PagedResponse{
		CurrentPage: 0,
		PerPage:     900,
		Total:       900,
		Limit:       0,
	}

	zeropage.Get()

	assert.Equal(t, zeropage.Pages, uint(1), "Pages should match")

}
