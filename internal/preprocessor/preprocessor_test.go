package preprocessor_test

import (
	"testing"

	"github.com/grqphical/webc/internal/preprocessor"
	"github.com/stretchr/testify/assert"
)

func TestDefinitions(t *testing.T) {
	input := `#define FOOBAR 1
	#define BARFOO Hello World`

	pp := preprocessor.New()
	err := pp.Parse(input)
	assert.NoError(t, err)

	assert.Equal(t, 4, len(pp.Definitions)) // 2 new ones plus 2 default ones

	assert.Equal(t, "1", pp.Definitions["FOOBAR"])
	assert.Equal(t, "Hello World", pp.Definitions["BARFOO"])
}
