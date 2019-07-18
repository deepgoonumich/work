package process

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContentType(t *testing.T) {
	contentType, ok := ContentTypeMappings["story"]
	assert.True(t, ok)
	assert.Equal(t, Story, contentType)
}
