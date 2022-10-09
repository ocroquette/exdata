package exdata

import (
	"os"
	"path/filepath"
)

type Repository struct {
	BaseDir string
}

func MakeRepository(baseDir string) Repository {
	p := new(Repository)
	p.BaseDir = baseDir
	return *p
}

func (*Repository) SubPathForChecksum(cs string) string {
	if len(cs) < 8 {
		panic("Checksum too short: " + cs)
	}
	part1 := cs[0:4]
	part2 := cs[4:8]
	return filepath.Join(part1, part2)
}

func (r *Repository) SubPathForTemporaryFiles() string {
	return "tmp"
}

func MkdirOrPanic(path string) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		panic("Creating \"" + path + "\" : " + err.Error())
	}
}

func (r *Repository) DirectoryForChecksum(cs string) string {
	d := filepath.Join(r.BaseDir, r.SubPathForChecksum(cs))
	MkdirOrPanic(d)
	return d
}

func (r *Repository) FilePathForChecksum(cs string) string {
	return filepath.Join(r.BaseDir, r.SubPathForChecksum(cs), cs)
}

func (r *Repository) DirectoryForTemporaryFiles() string {
	d := filepath.Join(r.BaseDir, r.SubPathForTemporaryFiles())
	MkdirOrPanic(d)
	return d
}
