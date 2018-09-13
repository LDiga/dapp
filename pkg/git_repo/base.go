package git_repo

import (
	"fmt"
	"io"
	"os"
	"strings"

	git_util "github.com/flant/dapp/pkg/git"
	go_git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Base struct {
	Name   string
	TmpDir string
}

func (repo *Base) getHeadCommitForRepo(repoPath string) (string, error) {
	ref, err := repo.getReferenceForRepo(repoPath)
	if err != nil {
		return "", fmt.Errorf("cannot get repo head: %s", err)
	}

	return fmt.Sprintf("%s", ref.Hash()), nil
}

func (repo *Base) getHeadBranchNameForRepo(repoPath string) (string, error) {
	ref, err := repo.getReferenceForRepo(repoPath)
	if err != nil {
		return "", fmt.Errorf("cannot get repo head: %s", err)
	}

	if ref.Name().IsBranch() {
		branchRef := ref.Name()
		return strings.Split(string(branchRef), "refs/heads/")[1], nil
	} else {
		return "", fmt.Errorf("cannot get branch name: HEAD refers to a specific revision that is not associated with a branch name")
	}
}

func (repo *Base) getReferenceForRepo(repoPath string) (*plumbing.Reference, error) {
	var err error

	repository, err := go_git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	return repository.Head()
}

func (repo *Base) String() string {
	return repo.Name
}

func (repo *Base) HeadCommit() (string, error) {
	panic("not implemented")
}

func (repo *Base) HeadBranchName() (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (repo *Base) LatestBranchCommit(branch string) (string, error) {
	panic("not implemented")
}

func (repo *Base) LatestTagCommit(branch string) (string, error) {
	panic("not implemented")
}

func (repo *Base) ArchiveType(ArchiveOptions) (ArchiveType, error) {
	panic("not implemented")
}

func (repo *Base) CreatePatch(io.Writer, PatchOptions) error {
	panic("not implemented")
}

func (repo *Base) IsAnyChanges(PatchOptions) (bool, error) {
	panic("not implemented")
}

func (repo *Base) IsAnyEntries(ArchiveOptions) (bool, error) {
	panic("not implemented")
}

func (repo *Base) CreateArchiveTar(io.Writer, ArchiveOptions) error {
	panic("not implemented")
}

func (repo *Base) ArchiveChecksum(ArchiveOptions) (string, error) {
	panic("not implemented")
}

func (repo *Base) createPatch(repoPath string, opts PatchOptions) (Patch, error) {
	repository, err := go_git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	fromHash := plumbing.NewHash(opts.FromCommit)
	_, err = repository.CommitObject(fromHash)
	if err != nil {
		return nil, fmt.Errorf("bad `from` commit `%s`: %s", opts.FromCommit, err)
	}

	toHash := plumbing.NewHash(opts.ToCommit)
	_, err = repository.CommitObject(toHash)
	if err != nil {
		return nil, fmt.Errorf("bad `to` commit `%s`: %s", opts.ToCommit, err)
	}

	patch := NewTmpPatchFile()

	fileFandler, err := os.OpenFile(patch.GetFilePath(), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open patch file: %s", err)
	}

	err = git_util.Diff(fileFandler, repoPath, git_util.DiffOptions{
		FromCommit: opts.FromCommit,
		ToCommit:   opts.ToCommit,
		PathFilter: git_util.PathFilter{
			BasePath:     opts.BasePath,
			IncludePaths: opts.IncludePaths,
			ExcludePaths: opts.ExcludePaths,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating diff between `%s` and `%s` commits: %s", opts.FromCommit, opts.ToCommit, err)
	}

	err = fileFandler.Close()
	if err != nil {
		return nil, fmt.Errorf("error creating diff file: %s", err)
	}

	return patch, nil
}

func (repo *Base) createArchive(repoPath string, opts ArchiveOptions) (Archive, error) {
	repository, err := go_git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot open repo: %s", err)
	}

	commitHash := plumbing.NewHash(opts.Commit)
	_, err = repository.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("bad commit `%s`: %s", opts.Commit, err)
	}

	archive := NewTmpArchiveFile()

	fileFandler, err := os.OpenFile(archive.GetFilePath(), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open archive file: %s", err)
	}

	err = git_util.Archive(fileFandler, repoPath, git_util.ArchiveOptions{
		Commit: opts.Commit,
		PathFilter: git_util.PathFilter{
			BasePath:     opts.BasePath,
			IncludePaths: opts.IncludePaths,
			ExcludePaths: opts.ExcludePaths,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error creating archive for commit `%s`: %s", opts.Commit, err)
	}

	return archive, nil
}
