// Package processor holds the deterministic parts of video conversion.
package processor

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"
)

func ForcedFailure(filename, prefix string) error {
	if prefix != "" && strings.HasPrefix(filename, prefix) {
		return errors.New("forced processing failure for local notification test")
	}
	return nil
}

func OutputObjectKey(id uuid.UUID) string { return "results/" + id.String() + "/frames.zip" }

// ArchiveFrames creates an archive with stable, basename-only entries.
func ArchiveFrames(directory string) (string, error) {
	frames, _ := filepath.Glob(filepath.Join(directory, "frame-*.jpg"))
	if len(frames) == 0 {
		return "", os.ErrNotExist
	}
	sort.Strings(frames)
	archivePath := filepath.Join(directory, "frames.zip")
	file, err := os.Create(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	archive := zip.NewWriter(file)
	for _, frame := range frames {
		contents, readErr := os.ReadFile(frame)
		if readErr != nil {
			return "", readErr
		}
		entry, _ := archive.Create(filepath.Base(frame))
		_, _ = entry.Write(contents)
	}
	if err = archive.Close(); err != nil {
		return "", err
	}
	return archivePath, nil
}
