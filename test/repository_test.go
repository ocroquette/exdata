package repository_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ocroquette/exdata/internal/exdata"
	"github.com/stretchr/testify/assert"
)

func TestSubPathForChecksum(t *testing.T) {
	r := exdata.MakeRepository("/the/path")

	assert.Equal(t,
		"c5a8/9f74",
		r.SubPathForChecksum("c5a89f7473a049daa5e476fb96d59b9113a8aaab900b2923ba8dda6c5800c86e"))
}

func TestDirectoryForChecksum(t *testing.T) {
	tmpDir := t.TempDir()

	r := exdata.MakeRepository(tmpDir)
	dir := r.DirectoryForChecksum("c5a89f7473a049daa5e476fb96d59b9113a8aaab900b2923ba8dda6c5800c86e")
	assert.Equal(t, filepath.Join(tmpDir, "c5a8/9f74"), dir)

	stat, _ := os.Stat(dir)
	assert.True(t, stat.IsDir())
}
