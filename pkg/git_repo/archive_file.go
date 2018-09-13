package git_repo

import (
	"fmt"
	"path/filepath"

	uuid "github.com/satori/go.uuid"
)

type ArchiveFile struct {
	FilePath string
}

func NewTmpArchiveFile() *ArchiveFile {
	path := filepath.Join("/tmp", fmt.Sprintf("dapp-%s.archive.tar", uuid.NewV4().String()))
	return &ArchiveFile{FilePath: path}
}

func (a *ArchiveFile) GetFilePath() string {
	return a.FilePath
}

func (a *ArchiveFile) GetType() (ArchiveType, error) {
	// TODO
	return FileArchive, nil
}

func (a *ArchiveFile) IsEmpty() (bool, error) {
	// TODO
	return false, nil
}
