package filesystem

import (
	"testing"

	"github.com/casdoor/oss/tests"
)

func TestAll(t *testing.T) {
	fileSystem := New("/tmp")
	tests.TestAll(fileSystem, t)
}
