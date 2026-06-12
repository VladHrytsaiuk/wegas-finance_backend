package parsers

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivatBankParser_InvalidInput(t *testing.T) {
	parser := NewPrivatBankParser()
	
	t.Run("Invalid Reader Type", func(t *testing.T) {
		// reader that is NOT ReaderAt
		reader := bytes.NewBuffer([]byte("not a pdf"))
		_, err := parser.Parse(reader, 0, "test.pdf")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must implement io.ReaderAt")
	})

	t.Run("Corrupted PDF", func(t *testing.T) {
		content := []byte("%PDF-1.4\ncorrupted")
		reader := bytes.NewReader(content)
		_, err := parser.Parse(reader, int64(len(content)), "corrupted.pdf")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open pdf")
	})
}
