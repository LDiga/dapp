package git

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

type ArchiveOptions struct {
	Commit         string
	PathFilter     PathFilter
	WithSubmodules bool
}

func Archive(out io.Writer, repoPath string, opts ArchiveOptions) error {
	clonePath := filepath.Join("/tmp", fmt.Sprintf("git-clone-%s", uuid.NewV4().String()))
	defer os.RemoveAll(clonePath)

	cmd := exec.Command("git", "clone", repoPath, clonePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("temporary clone creation failed: %s\n%s", err, output)
	}

	cmd = exec.Command("git", "-C", clonePath, "checkout", opts.Commit)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("cannot checkout commit `%s`: %s\n%s", err, output)
	}

	if opts.WithSubmodules {
		cmd = exec.Command("git", "-C", clonePath, "submodule", "update", "--init", "--recursive")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("cannot update submodules")
		}
	}

	now := time.Now()
	tw := tar.NewWriter(out)

	trees := []struct{ Path, Commit string }{
		{"/", opts.Commit},
	}

	for len(trees) > 0 {
		tree := trees[0]
		trees = trees[1:]

		cmd = exec.Command(
			"git", "-C", filepath.Join(clonePath, tree.Path),
			"ls-tree", "--long", "--full-tree", "-r", "-z",
			tree.Commit,
		)

		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("cannot list git tree: %s\n%s", err, output)
		}

		lines := strings.Split(string(output), "\000")
		lines = lines[:len(lines)-1]

	HandleEntries:
		for _, line := range lines {
			fields := strings.SplitN(line, " ", 4)

			rawMode := strings.TrimLeft(fields[0], " ")
			objectType := strings.TrimLeft(fields[1], " ")
			objectID := strings.TrimLeft(fields[2], " ")

			fields = strings.SplitN(strings.TrimLeft(fields[3], " "), "\t", 2)

			rawObjectSize := fields[0]
			filePath := filepath.Join(tree.Path, fields[1])
			fullFilePath := filepath.Join(clonePath, filePath)

			mode, err := strconv.ParseInt(rawMode, 8, 64)
			if err != nil {
				return fmt.Errorf("unexpected git ls-tree file mode `%s`: %s", rawMode, err)
			}

			switch objectType {
			case "blob":
				if !opts.PathFilter.IsFilePathValid(filePath) {
					continue HandleEntries
				}
				archiveFilePath := opts.PathFilter.TrimFileBasePath(filePath)

				size, err := strconv.ParseInt(rawObjectSize, 10, 64)
				if err != nil {
					return fmt.Errorf("unexpected git ls-tree file size `%s`: %s", rawObjectSize, err)
				}

				if mode == 0120000 { // symlink
					linkname, err := os.Readlink(fullFilePath)
					if err != nil {
						return fmt.Errorf("cannot read symlink `%s`: %s", fullFilePath, err)
					}

					err = tw.WriteHeader(&tar.Header{
						Format:     tar.FormatGNU,
						Typeflag:   tar.TypeSymlink,
						Name:       archiveFilePath,
						Mode:       mode,
						Linkname:   string(linkname),
						Size:       size,
						ModTime:    now,
						AccessTime: now,
						ChangeTime: now,
					})
					if err != nil {
						return fmt.Errorf("unable to write tar symlink header: %s", err)
					}
				} else {
					err = tw.WriteHeader(&tar.Header{
						Format:     tar.FormatGNU,
						Name:       archiveFilePath,
						Mode:       mode,
						Size:       size,
						ModTime:    now,
						AccessTime: now,
						ChangeTime: now,
					})
					if err != nil {
						return fmt.Errorf("unable to write tar header: %s", err)
					}

					file, err := os.Open(fullFilePath)
					if err != nil {
						return fmt.Errorf("unable to open sss file `%s`: %s", fullFilePath, err)
					}

					_, err = io.Copy(tw, file)
					if err != nil {
						return fmt.Errorf("unable to write data to tar archive: %s", err)
					}
				}

			case "commit":
				if opts.WithSubmodules {
					trees = append(trees, struct{ Path, Commit string }{filePath, objectID})
				}

			default:
				panic(fmt.Sprintf("unexpected object type `%s`", objectType))
			}
		}
	}

	err = tw.Close()
	if err != nil {
		return fmt.Errorf("cannot write tar archive: %s", err)
	}

	return nil
}
