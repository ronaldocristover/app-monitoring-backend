package pagination

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMeta_FirstPage(t *testing.T) {
	meta := NewMeta(1, 20, 100)

	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 20, meta.PageSize)
	assert.Equal(t, int64(100), meta.TotalItems)
	assert.Equal(t, 5, meta.TotalPages)
}

func TestNewMeta_ExactDivision(t *testing.T) {
	meta := NewMeta(1, 10, 50)

	assert.Equal(t, 5, meta.TotalPages)
}

func TestNewMeta_Remainder(t *testing.T) {
	meta := NewMeta(1, 10, 55)

	assert.Equal(t, 6, meta.TotalPages)
}

func TestNewMeta_SingleItem(t *testing.T) {
	meta := NewMeta(1, 20, 1)

	assert.Equal(t, 1, meta.TotalPages)
}

func TestNewMeta_ZeroItems(t *testing.T) {
	meta := NewMeta(1, 20, 0)

	assert.Equal(t, 0, meta.TotalPages)
	assert.Equal(t, int64(0), meta.TotalItems)
}

func TestNewMeta_LargeTotal(t *testing.T) {
	meta := NewMeta(3, 25, 1000)

	assert.Equal(t, 3, meta.Page)
	assert.Equal(t, 25, meta.PageSize)
	assert.Equal(t, int64(1000), meta.TotalItems)
	assert.Equal(t, 40, meta.TotalPages)
}
