package storage_test

import (
	"errors"
	"io"
	"io/fs"
	"testing"

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

// streamWrite writes data at name through sut.Create, failing the test on any
// error along the way.
func streamWrite(t testing.TB, sut *storage.Storage, name, data string) {
	t.Helper()
	writer, err := sut.Create(name)
	if err != nil {
		t.Fatalf("Create(%q) = %v, want nil", name, err)
	}
	_, err = io.WriteString(writer, data)
	if err != nil {
		t.Fatalf("WriteString(%q) = %v, want nil", name, err)
	}
	err = writer.Close()
	if err != nil {
		t.Fatalf("Close(%q) = %v, want nil", name, err)
	}
}

// statInfo is the subset of [io/fs.FileInfo] compared in these tests.
type statInfo struct {
	Name  string
	Size  int64
	Mode  fs.FileMode
	IsDir bool
}

func infoOf(info fs.FileInfo) statInfo {
	if info == nil {
		return statInfo{}
	}
	return statInfo{
		Name:  info.Name(),
		Size:  info.Size(),
		Mode:  info.Mode(),
		IsDir: info.IsDir(),
	}
}

// dirListing is the subset of [io/fs.DirEntry] compared in these tests.
type dirListing struct {
	Name  string
	IsDir bool
	Type  fs.FileMode
}

// listingsOf projects entries onto the [dirListing] fields compared in these
// tests.
func listingsOf(entries []fs.DirEntry) []dirListing {
	listings := make([]dirListing, 0, len(entries))
	for _, entry := range entries {
		listings = append(listings, dirListing{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Type:  entry.Type(),
		})
	}
	return listings
}

func TestStorage_WriteFile(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		filename string
		data     string
		want     string
		wantErr  error
	}{
		{
			name:     "TopLevelFile",
			filename: "settings.json",
			data:     `{"theme":"dark"}`,
			want:     `{"theme":"dark"}`,
		},
		{
			name:     "NestedFileCreatesParents",
			filename: "profiles/default/settings.json",
			data:     "nested",
			want:     "nested",
		},
		{
			name:     "InvalidPath",
			filename: "../escape",
			data:     "nope",
			wantErr:  fs.ErrInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := storagetest.New("root")

			// Act
			writeErr := sut.WriteFile(tc.filename, []byte(tc.data))
			data, readErr := sut.ReadFile(tc.filename)

			// Assert
			if got, want := writeErr, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("WriteFile(%q) = %v, want %v", tc.filename, got, want)
			}
			if got, want := readErr, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("ReadFile(%q) = %v, want %v", tc.filename, got, want)
			}
			if got, want := string(data), tc.want; got != want {
				t.Errorf("ReadFile(%q) = %q, want %q", tc.filename, got, want)
			}
		})
	}
}

func TestStorage_Create(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := storagetest.New("root")
	streamWrite(t, sut, "nested/dir/file.txt", "streamed")

	// Act
	data, err := sut.ReadFile("nested/dir/file.txt")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ReadFile(...) = %v, want %v", got, want)
	}
	if got, want := string(data), "streamed"; got != want {
		t.Errorf("ReadFile(...) = %q, want %q", got, want)
	}
}

func TestStorage_Open(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		filename string
		wantErr  error
	}{
		{
			name:     "MissingFile",
			filename: "absent.txt",
			wantErr:  fs.ErrNotExist,
		},
		{
			name:     "InvalidEscapingPath",
			filename: "../secret",
			wantErr:  fs.ErrInvalid,
		},
		{
			name:     "InvalidAbsolutePath",
			filename: "/etc/passwd",
			wantErr:  fs.ErrInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := storagetest.New("root")

			// Act
			file, err := sut.Open(tc.filename)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Open(%q) = %v, want %v", tc.filename, got, want)
			}
			if got, want := file, fs.File(nil); got != want {
				t.Errorf("Open(%q) file = %v, want %v", tc.filename, got, want)
			}
		})
	}
}

func TestStorage_Stat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		filename string
		want     statInfo
		wantErr  error
	}{
		{
			name:     "File",
			filename: "dir/file.txt",
			want: statInfo{
				Name:  "file.txt",
				Size:  5,
				Mode:  0o644,
				IsDir: false,
			},
		},
		{
			name:     "Directory",
			filename: "dir",
			want: statInfo{
				Name:  "dir",
				Size:  0,
				Mode:  fs.ModeDir | 0o755,
				IsDir: true,
			},
		},
		{
			name:     "Missing",
			filename: "absent",
			wantErr:  fs.ErrNotExist,
		},
		{
			name:     "InvalidPath",
			filename: "../escape",
			wantErr:  fs.ErrInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := storagetest.New("root")
			write(t, sut, "dir/file.txt", "hello")

			// Act
			info, err := sut.Stat(tc.filename)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Stat(%q) = %v, want %v", tc.filename, got, want)
			}
			if got, want := infoOf(info), tc.want; !cmp.Equal(got, want) {
				t.Errorf("Stat(%q) = %+v, want %+v", tc.filename, got, want)
			}
		})
	}
}

func TestStorage_ReadDir(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		dir     string
		want    []dirListing
		wantErr error
	}{
		{
			name: "ListsImmediateChildren",
			dir:  "top",
			want: []dirListing{
				{
					Name:  "a.txt",
					IsDir: false,
					Type:  0,
				},
				{
					Name:  "b.txt",
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
		{
			name:    "InvalidPath",
			dir:     "../escape",
			wantErr: fs.ErrInvalid,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := storagetest.New("root")
			write(t, sut, "top/a.txt", "1")
			write(t, sut, "top/b.txt", "2")
			write(t, sut, "top/sub/c.txt", "3")

			// Act
			entries, err := sut.ReadDir(tc.dir)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("ReadDir(%q) = %v, want %v", tc.dir, got, want)
			}
			listings := listingsOf(entries)
			if got, want := listings, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("ReadDir(%q) = %+v, want %+v", tc.dir, got, want)
			}
		})
	}
}

func TestStorage_Remove(t *testing.T) {
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

func TestStorage_Remove_Errors(t *testing.T) {
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
		{
			name:    "InvalidPath",
			target:  "../escape",
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

func TestStorage_RemoveAll(t *testing.T) {
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

func TestStorage_RemoveAll_InvalidPath(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := storagetest.New("root")

	// Act
	err := sut.RemoveAll("../escape")

	// Assert
	if got, want := err, fs.ErrInvalid; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("RemoveAll(...) = %v, want %v", got, want)
	}
}

func TestStorage_ErrorBackend(t *testing.T) {
	t.Parallel()

	// Arrange
	unavailable := errors.New("home unavailable")
	sut := storage.New("root", storage.ErrFS(unavailable))

	// Act
	_, createErr := sut.Create("file.txt")
	_, openErr := sut.Open("file.txt")

	// Assert
	errs := []error{createErr, openErr}
	want := []error{unavailable, unavailable}
	if got, want := errs, want; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Errorf("operation errors = %v, want all %v", got, unavailable)
	}
}

func TestStorage_Sub(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := storagetest.New("root")
	sub, subErr := sut.Sub("nested")

	// Act
	writeErr := sub.WriteFile("file.txt", []byte("scoped"))
	data, readErr := sut.ReadFile("nested/file.txt")

	// Assert
	if got, want := subErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Sub(...) = %v, want %v", got, want)
	}
	if got, want := writeErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("WriteFile(...) through sub = %v, want %v", got, want)
	}
	if got, want := readErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ReadFile(...) through parent = %v, want %v", got, want)
	}
	if got, want := string(data), "scoped"; got != want {
		t.Errorf("ReadFile(...) = %q, want %q", got, want)
	}
}

func TestStorage_Sub_InvalidPath(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := storagetest.New("root")

	// Act
	sub, err := sut.Sub("../escape")

	// Assert
	if got, want := err, fs.ErrInvalid; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Sub(...) = %v, want %v", got, want)
	}
	if got, want := sub, (*storage.Storage)(nil); got != want {
		t.Errorf("Sub(...) storage = %v, want %v", got, want)
	}
}
