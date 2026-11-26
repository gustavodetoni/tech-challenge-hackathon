package processor

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestForcedFailure(t *testing.T) {
	if err := ForcedFailure("force-failure-demo.mov", "force-failure-"); err == nil {
		t.Fatal("expected forced failure")
	}
	if err := ForcedFailure("movie.mov", "force-failure-"); err != nil {
		t.Fatal(err)
	}
	if err := ForcedFailure("force-failure-demo.mov", ""); err != nil {
		t.Fatal(err)
	}
}

func TestOutputObjectKey(t *testing.T) {
	id := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	if got, want := OutputObjectKey(id), "results/11111111-1111-1111-1111-111111111111/frames.zip"; got != want {
		t.Fatalf("got %q", got)
	}
}

func TestArchiveFramesCreatesSortedArchive(t *testing.T) {
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "frame-000002.jpg"), []byte("two"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "frame-000001.jpg"), []byte("one"), 0600); err != nil {
		t.Fatal(err)
	}
	path, err := ArchiveFrames(directory)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	if len(archive.File) != 2 || archive.File[0].Name != "frame-000001.jpg" || archive.File[1].Name != "frame-000002.jpg" {
		t.Fatalf("unexpected archive: %#v", archive.File)
	}
}

func TestArchiveFramesRequiresFrames(t *testing.T) {
	if _, err := ArchiveFrames(t.TempDir()); !os.IsNotExist(err) {
		t.Fatalf("expected missing frames error, got %v", err)
	}
}

func TestArchiveFramesReturnsReadError(t *testing.T) {
	directory := t.TempDir()
	if err := os.Symlink(filepath.Join(directory, "missing.jpg"), filepath.Join(directory, "frame-000001.jpg")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := ArchiveFrames(directory); err == nil {
		t.Fatal("expected read error")
	}
}
