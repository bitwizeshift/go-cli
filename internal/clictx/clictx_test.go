package clictx_test

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/bitwizeshift/go-cli/internal/clictx"
	"github.com/bitwizeshift/go-cli/richtext"
)

func TestWriters(t *testing.T) {
	t.Parallel()

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	testCases := []struct {
		name        string
		ctx         context.Context
		wantOutType reflect.Type
		wantErrType reflect.Type
	}{
		{
			name:        "StoredWriters",
			ctx:         clictx.WithWriters(context.Background(), stdout, stderr),
			wantOutType: reflect.TypeFor[*bytes.Buffer](),
			wantErrType: reflect.TypeFor[*bytes.Buffer](),
		},
		{
			name:        "NoWriters",
			ctx:         context.Background(),
			wantOutType: reflect.TypeFor[*richtext.Writer](),
			wantErrType: reflect.TypeFor[*richtext.Writer](),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			outWriter, errWriter := clictx.Writers(tc.ctx)

			// Assert
			if got, want := reflect.TypeOf(outWriter), tc.wantOutType; got != want {
				t.Errorf("Writers(ctx) stdout = %v, want %v", got, want)
			}
			if got, want := reflect.TypeOf(errWriter), tc.wantErrType; got != want {
				t.Errorf("Writers(ctx) stderr = %v, want %v", got, want)
			}
		})
	}
}
