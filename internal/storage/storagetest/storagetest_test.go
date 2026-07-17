package storagetest_test

import (
	"io/fs"
	"testing"
	"time"

	"github.com/bitwizeshift/go-cli/internal/storage"
	"github.com/bitwizeshift/go-cli/internal/storage/storagetest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// write stores data at name through sut, failing the test if the write fails.
func write(t testing.TB, sut *storage.Storage, name, data string) {
	t.Helper()
	err := sut.WriteFile(name, []byte(data))
	if err != nil {
		t.Fatalf("WriteFile(%q) = %v, want nil", name, err)
	}
}

// emptyDir leaves an empty directory at name by writing a file beneath it and
// removing it, failing the test on any error.
func emptyDir(t testing.TB, sut *storage.Storage, name string) {
	t.Helper()
	placeholder := name + "/placeholder"
	write(t, sut, placeholder, "")
	err := sut.Remove(placeholder)
	if err != nil {
		t.Fatalf("Remove(%q) = %v, want nil", placeholder, err)
	}
}

// statMeta is the subset of [io/fs.FileInfo] compared in these tests.
type statMeta struct {
	Name  string
	Size  int64
	IsDir bool
	Mode  fs.FileMode
}

// entryMeta is the subset of [io/fs.DirEntry] compared in these tests.
type entryMeta struct {
	Name  string
	IsDir bool
	Type  fs.FileMode
}

// entriesOf projects entries onto the [entryMeta] fields compared in these
// tests.
func entriesOf(entries []fs.DirEntry) []entryMeta {
	metas := make([]entryMeta, 0, len(entries))
	for _, entry := range entries {
		metas = append(metas, entryMeta{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Type:  entry.Type(),
		})
	}
	return metas
}

func TestNew_RoundTrips(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := storagetest.New("root")

	// Act
	writeErr := sut.WriteFile("dir/file.txt", []byte("payload"))
	data, readErr := sut.ReadFile("dir/file.txt")

	// Assert
	if got, want := writeErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("WriteFile(...) = %v, want %v", got, want)
	}
	if got, want := readErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ReadFile(...) = %v, want %v", got, want)
	}
	if got, want := string(data), "payload"; got != want {
		t.Errorf("ReadFile(...) = %q, want %q", got, want)
	}
}

func TestNew_Stat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		target  string
		want    statMeta
		wantErr error
	}{
		{
			name:   "File",
			target: "dir/file.txt",
			want: statMeta{
				Name:  "file.txt",
				Size:  5,
				IsDir: false,
				Mode:  0o644,
			},
		},
		{
			name:   "Directory",
			target: "dir",
			want: statMeta{
				Name:  "dir",
				Size:  0,
				IsDir: true,
				Mode:  fs.ModeDir | 0o755,
			},
		},
		{
			name:    "Missing",
			target:  "absent",
			wantErr: fs.ErrNotExist,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := storagetest.New("root")
			write(t, sut, "dir/file.txt", "hello")

			// Act
			info, err := sut.Stat(tc.target)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Stat(%q) = %v, want %v", tc.target, got, want)
			}
			if err != nil {
				return
			}
			meta := statMeta{
				Name:  info.Name(),
				Size:  info.Size(),
				IsDir: info.IsDir(),
				Mode:  info.Mode(),
			}
			if got, want := meta, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Stat(%q) = %+v, want %+v", tc.target, got, want)
			}
		})
	}
}

func TestNew_ReadDir(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		dir     string
		want    []entryMeta
		wantErr error
	}{
		{
			name: "ListsImmediateChildren",
			dir:  "top",
			want: []entryMeta{
				{
					Name:  "a.txt",
					IsDir: false,
					Type:  0,
				},
				{
					Name:  "sub",
					IsDir: true,
					Type:  fs.ModeDir,
				},
			},
		},
		{
			name:    "MissingDirectory",
			dir:     "absent",
			wantErr: fs.ErrNotExist,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := storagetest.New("root")
			write(t, sut, "top/a.txt", "1")
			write(t, sut, "top/sub/b.txt", "2")

			// Act
			entries, err := sut.ReadDir(tc.dir)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("ReadDir(%q) = %v, want %v", tc.dir, got, want)
			}
			metas := entriesOf(entries)
			if got, want := metas, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("ReadDir(%q) = %+v, want %+v", tc.dir, got, want)
			}
		})
	}
}

func TestNew_Remove(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := storagetest.New("root")
	write(t, sut, "file.txt", "content")

	// Act
	removeErr := sut.Remove("file.txt")
	_, statErr := sut.Stat("file.txt")

	// Assert
	if got, want := removeErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Remove(...) = %v, want %v", got, want)
	}
	if got, want := statErr, fs.ErrNotExist; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Errorf("Stat(...) after Remove = %v, want %v", got, want)
	}
}

func TestNew_Remove_EmptyDirectory(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := storagetest.New("root")
	emptyDir(t, sut, "dir")

	// Act
	removeErr := sut.Remove("dir")
	_, statErr := sut.Stat("dir")

	// Assert
	if got, want := removeErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Remove(dir) = %v, want %v", got, want)
	}
	if got, want := statErr, fs.ErrNotExist; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Errorf("Stat(dir) after Remove = %v, want %v", got, want)
	}
}

func TestNew_Remove_Errors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		target  string
		wantErr error
	}{
		{
			name:    "MissingEntry",
			target:  "absent.txt",
			wantErr: fs.ErrNotExist,
		},
		{
			name:    "NonEmptyDirectory",
			target:  "dir",
			wantErr: fs.ErrInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := storagetest.New("root")
			write(t, sut, "dir/file.txt", "content")

			// Act
			err := sut.Remove(tc.target)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Remove(%q) = %v, want %v", tc.target, got, want)
			}
		})
	}
}

func TestNew_RemoveAll(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		target string
	}{
		{
			name:   "RemovesSubtree",
			target: "tree",
		},
		{
			name:   "AbsentTargetIsNoError",
			target: "gone",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := storagetest.New("root")
			write(t, sut, "tree/a.txt", "1")
			write(t, sut, "tree/sub/b.txt", "2")

			// Act
			removeErr := sut.RemoveAll(tc.target)
			_, statErr := sut.Stat(tc.target)

			// Assert
			if got, want := removeErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("RemoveAll(%q) = %v, want %v", tc.target, got, want)
			}
			if got, want := statErr, fs.ErrNotExist; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Errorf("Stat(%q) after RemoveAll = %v, want %v", tc.target, got, want)
			}
		})
	}
}

func TestNewAppStorage_RootsAreIsolated(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := storagetest.NewAppStorage()

	// Act
	writeErr := sut.Config.WriteFile("settings.json", []byte("config"))
	_, cacheErr := sut.Cache.ReadFile("settings.json")

	// Assert
	if got, want := writeErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Config.WriteFile(...) = %v, want %v", got, want)
	}
	if got, want := cacheErr, fs.ErrNotExist; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Errorf("Cache.ReadFile(...) = %v, want %v", got, want)
	}
}

// fileMeta is the metadata compared for a stored file's [io/fs.FileInfo].
type fileMeta struct {
	ModTime time.Time
	Sys     any
}

func TestNewAppStorage_FileInfoMetadata(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := storagetest.NewAppStorage()
	write(t, sut.Data, "state.bin", "x")

	// Act
	info, err := sut.Data.Stat("state.bin")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Data.Stat(...) = %v, want %v", got, want)
	}
	meta := fileMeta{
		ModTime: info.ModTime(),
		Sys:     info.Sys(),
	}
	if got, want := meta, (fileMeta{ModTime: time.Time{}, Sys: nil}); !cmp.Equal(got, want) {
		t.Errorf("Stat(...) metadata = %+v, want %+v", got, want)
	}
}
