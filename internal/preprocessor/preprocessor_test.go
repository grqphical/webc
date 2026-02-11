package preprocessor_test

import (
	"embed"
	"testing"

	"github.com/grqphical/webc/internal/preprocessor"
	"github.com/stretchr/testify/assert"
)

func TestDefinitions(t *testing.T) {
	input := `#define FOOBAR 1
	#define BARFOO Hello World`

	pp := preprocessor.New(embed.FS{})
	_, err := pp.Parse(input)
	assert.NoError(t, err)

	assert.Equal(t, 8, len(pp.Definitions)) // 2 new ones plus 6 default ones

	assert.Equal(t, "1", pp.Definitions["FOOBAR"])
	assert.Equal(t, "Hello World", pp.Definitions["BARFOO"])
}

func TestIfStatements(t *testing.T) {
	input := `#define FOOBAR 1
	#ifdef FOOBAR
	blah blah blah
	#endif
	#ifndef FOOBAR
	boo boo boo
	#else
	blah blah blah
	#endif
	boo
`

	pp := preprocessor.New(embed.FS{})
	source, err := pp.Parse(input)
	assert.NoError(t, err)

	assert.Equal(t, 7, len(pp.Definitions)) // 1 new one plus 6 default ones

	assert.Equal(t, "1", pp.Definitions["FOOBAR"])
	assert.Equal(t, "blah blah blah\nblah blah blah\nboo\n", source)
}
