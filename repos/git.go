package repos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	w "github.com/0xdevar/waraqah"
)

type Git struct {
	owner  string
	repo   string
	branch string
	dest   string
}

func (s *Git) String() string {
	return fmt.Sprintf(
		"owner=%s,repo=%s,branch=%s,dest=%s",
		s.owner,
		s.repo,
		s.branch,
		s.dest,
	)
}

type GitError struct {
	Message string
	Git     *Git
}

func (e *GitError) Error() string {
	return fmt.Sprintf("git operation error: %s (%s)", e.Message, e.Git.String())
}

type MetaError struct {
	Message  string
	Filename string
}

func (e *MetaError) Error() string {
	return fmt.Sprintf("meta operation error: %s (%s)", e.Message, e.Filename)
}

type metaImage struct {
	Name       string   `json:"name"`
	Size       int      `json:"size"`
	Tags       []string `json:"tags"`
	Resolution [2]int   `json:"resolution"`
}

type metaType struct {
	Name   string      `json:"name"`
	Images []metaImage `json:"images"`
}

func runGitCommand(dir *string, arg string) (string, error) {
	args := strings.Fields(arg)
	cmd := exec.Command("git", args...)

	if dir != nil {
		cmd.Dir = *dir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return stdout.String(), fmt.Errorf("%v: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func getMeta(filename string) (metaType, error) {
	buffer, err := os.ReadFile(filename)

	if err != nil {
		return metaType{}, &MetaError{
			Message:  err.Error(),
			Filename: filename,
		}
	}

	var meta metaType

	err = json.Unmarshal(buffer, &meta)

	return meta, err
}

func toImage(imagePath string, m metaImage) w.Wallpaper {
	return w.Wallpaper{
		Path: imagePath,
		WallpaperMeta: w.WallpaperMeta{
			Resolution: m.Resolution,
			Size:       m.Size,
		},
	}

}

func NewGitRepo(owner, repo, branch, dest string) *Git {
	return &Git{owner, repo, branch, dest}
}

func (s *Git) isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

func (s *Git) GetWallpapers() (out []w.WallpaperCollection, err error) {
	if !s.isGitRepo(s.dest) {
		gitCmd := fmt.Sprintf("clone --depth=1 --filter=blob:none --sparse --no-checkout "+
			"https://github.com/%s/%s.git --branch %s %s",
			s.owner, s.repo, s.branch, s.dest)

		// TODO: change directory to something else
		if _, err := runGitCommand(nil, gitCmd); err != nil {
			return nil, &GitError{
				Message: err.Error(),
				Git:     s,
			}
		}
	}

	{
		if _, err := runGitCommand(&s.dest, "sparse-checkout init --no-cone"); err != nil {
			return nil, &GitError{
				Message: err.Error(),
				Git:     s,
			}
		}

		if _, err := runGitCommand(&s.dest, "sparse-checkout set ''"); err != nil {
			return nil, &GitError{
				Message: err.Error(),
				Git:     s,
			}
		}

		if _, err := runGitCommand(&s.dest, "checkout"); err != nil {
			return nil, &GitError{
				Message: err.Error(),
				Git:     s,
			}
		}
	}

	{
		filesAsString, err := runGitCommand(&s.dest, "ls-tree main -d --name-only")

		if err != nil {
			return nil, &GitError{
				Message: err.Error(),
				Git:     s,
			}
		}

		fileNames := strings.SplitSeq(strings.TrimSpace(filesAsString), "\n")

		if _, err = runGitCommand(&s.dest, "sparse-checkout add **/*.json **/thumbnail.jpg"); err != nil {
			return nil, &GitError{
				Message: err.Error(),
				Git:     s,
			}
		}

		for fileName := range fileNames {
			// TODO: what to do here, fail for all?
			meta, _ := getMeta(fmt.Sprintf("%s/%s/meta.json", s.dest, fileName))

			images := func() []w.Wallpaper {
				var out []w.Wallpaper

				for _, image := range meta.Images {
					path := fmt.Sprintf("%s/%s/images/%s", s.dest, fileName, image.Name)
					out = append(out, toImage(path, image))
				}

				return out
			}()

			out = append(out, w.WallpaperCollection{
				Name:     fileName,
				Thumnail: fmt.Sprintf("%s/%s/%s", s.dest, fileName, "thumbnail.jpg"),
				Images:   images,
			})
		}
	}

	return
}

func (s *Git) DownloadWallpaper(wallpaper w.WallpaperCollection) error {
	folder := wallpaper.Name

	if _, err := runGitCommand(&s.dest, fmt.Sprintf("sparse-checkout add %s", folder)); err != nil {
		return &GitError{
			Message: err.Error(),
			Git:     s,
		}
	}

	return nil
}
