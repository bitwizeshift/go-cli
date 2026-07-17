package cli_test

import (
	"context"
	"io"
	"io/fs"
	"testing"

	"github.com/bitwizeshift/go-cli"
	"github.com/bitwizeshift/go-cli/clitest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// configRoot returns an in-memory Config storage root for exercising the
// wrapper's delegation.
func configRoot(t testing.TB) *cli.Storage {
	t.Helper()
	_, app := clitest.WithStorage(context.Background())
	return app.Config
}

// write stores data at name through sut, failing the test if the write fails.
func write(t testing.TB, sut *cli.Storage, name, data string) {
	t.Helper()
	err := sut.WriteFile(name, []byte(data))
	if err != nil {
		t.Fatalf("WriteFile(%q) = %v, want nil", name, err)
	}
}

// streamWrite writes data at name through sut.Create, failing the test on any
// error along the way.
func streamWrite(t testing.TB, sut *cli.Storage, name, data string) {
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

// subStorage opens the dir sub-root of sut, failing the test if it cannot.
func subStorage(t testing.TB, sut *cli.Storage, dir string) *cli.Storage {
	t.Helper()
	sub, err := sut.Sub(dir)
	if err != nil {
		t.Fatalf("Sub(%q) = %v, want nil", dir, err)
	}
	return sub
}

// readAll reads file to completion, failing the test if the read fails.
func readAll(t testing.TB, file fs.File) []byte {
	t.Helper()
	data, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("ReadAll(...) = %v, want nil", err)
	}
	return data
}

// statInfo is the subset of [io/fs.FileInfo] compared in these tests.
type statInfo struct {
	Name string
	Size int64
}

// entryNames returns the filenames of entries, in order.
func entryNames(entries []fs.DirEntry) []string {
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names
}

func TestStorageFrom(t *testing.T) {
	t.Parallel()

	stored, _ := clitest.WithStorage(context.Background())

	testCases := []struct {
		name        string
		ctx         context.Context
		wantPresent bool
	}{
		{
			name:        "StoredStorage",
			ctx:         stored,
			wantPresent: true,
		},
		{
			name:        "NoStorage",
			ctx:         context.Background(),
			wantPresent: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			app := cli.StorageFrom(tc.ctx)

			// Assert
			present := app != nil
			if got, want := present, tc.wantPresent; got != want {
				t.Errorf("StorageFrom(ctx) present = %t, want %t", got, want)
			}
		})
	}
}

func TestStorage_WriteFile(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := configRoot(t)

	// Act
	writeErr := sut.WriteFile("dir/settings.json", []byte("payload"))
	data, readErr := sut.ReadFile("dir/settings.json")

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

func TestStorage_Create(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := configRoot(t)
	streamWrite(t, sut, "stream.txt", "streamed")

	// Act
	data, err := sut.ReadFile("stream.txt")

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

	// Arrange
	sut := configRoot(t)
	write(t, sut, "file.txt", "opened")

	// Act
	file, err := sut.Open("file.txt")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Open(...) = %v, want %v", got, want)
	}
	data := readAll(t, file)
	if got, want := string(data), "opened"; got != want {
		t.Errorf("Open then ReadAll = %q, want %q", got, want)
	}
}

func TestStorage_Stat(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := configRoot(t)
	write(t, sut, "file.txt", "hello")

	// Act
	info, err := sut.Stat("file.txt")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Stat(...) = %v, want %v", got, want)
	}
	meta := statInfo{
		Name: info.Name(),
		Size: info.Size(),
	}
	if got, want := meta, (statInfo{Name: "file.txt", Size: 5}); !cmp.Equal(got, want) {
		t.Errorf("Stat(...) = %+v, want %+v", got, want)
	}
}

func TestStorage_ReadDir(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := configRoot(t)
	write(t, sut, "dir/a.txt", "1")

	// Act
	entries, err := sut.ReadDir("dir")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ReadDir(...) = %v, want %v", got, want)
	}
	names := entryNames(entries)
	if got, want := names, []string{"a.txt"}; !cmp.Equal(got, want) {
		t.Errorf("ReadDir(...) = %v, want %v", got, want)
	}
}

func TestStorage_Remove(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := configRoot(t)
	write(t, sut, "file.txt", "bye")

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

func TestStorage_RemoveAll(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := configRoot(t)
	write(t, sut, "tree/a.txt", "1")

	// Act
	removeErr := sut.RemoveAll("tree")
	_, statErr := sut.Stat("tree")

	// Assert
	if got, want := removeErr, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("RemoveAll(...) = %v, want %v", got, want)
	}
	if got, want := statErr, fs.ErrNotExist; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Errorf("Stat(...) after RemoveAll = %v, want %v", got, want)
	}
}

func TestStorage_Sub(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := configRoot(t)
	sub := subStorage(t, sut, "nested")
	write(t, sub, "file.txt", "scoped")

	// Act
	data, err := sut.ReadFile("nested/file.txt")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("ReadFile(...) through parent = %v, want %v", got, want)
	}
	if got, want := string(data), "scoped"; got != want {
		t.Errorf("ReadFile(...) = %q, want %q", got, want)
	}
}

func TestStorage_Sub_InvalidPath(t *testing.T) {
	t.Parallel()

	// Arrange
	sut := configRoot(t)

	// Act
	sub, err := sut.Sub("../escape")

	// Assert
	if got, want := err, fs.ErrInvalid; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Sub(...) = %v, want %v", got, want)
	}
	if got, want := sub, (*cli.Storage)(nil); got != want {
		t.Errorf("Sub(...) storage = %v, want %v", got, want)
	}
}
