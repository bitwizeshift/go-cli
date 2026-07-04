package strcase_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/bitwizeshift/go-cli/internal/strcase"
)

func TestToKebab(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "single word",
			input: "foo",
			want:  "foo",
		},
		{
			name:  "camel boundary",
			input: "fooBar",
			want:  "foo-bar",
		},
		{
			name:  "acronym prefix",
			input: "HTTPServer",
			want:  "http-server",
		},
		{
			name:  "acronym two letter",
			input: "IOReader",
			want:  "io-reader",
		},
		{
			name:  "acronym suffix",
			input: "myID",
			want:  "my-id",
		},
		{
			name:  "digit before uppercase",
			input: "Base64Binary",
			want:  "base64-binary",
		},
		{
			name:  "camel digit before uppercase",
			input: "base64Binary",
			want:  "base64-binary",
		},
		{
			name:  "acronym before digits",
			input: "AES512",
			want:  "aes512",
		},
		{
			name:  "digit before lowercase",
			input: "base64binary",
			want:  "base64binary",
		},
		{
			name:  "underscore separator",
			input: "foo_bar",
			want:  "foo-bar",
		},
		{
			name:  "hyphen separator",
			input: "foo-bar",
			want:  "foo-bar",
		},
		{
			name:  "space separator",
			input: "foo bar",
			want:  "foo-bar",
		},
		{
			name:  "surrounding and repeated separators",
			input: "__foo--bar__",
			want:  "foo-bar",
		},
		{
			name:  "caseless letter",
			input: "中Foo",
			want:  "中-foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			input := tc.input

			// Act
			kebab := strcase.ToKebab(input)

			// Assert
			if got, want := kebab, tc.want; !cmp.Equal(got, want) {
				t.Errorf("strcase.ToKebab(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func TestToSnake(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "camel boundary",
			input: "fooBar",
			want:  "foo_bar",
		},
		{
			name:  "acronym prefix",
			input: "HTTPServer",
			want:  "http_server",
		},
		{
			name:  "digit before uppercase",
			input: "Base64Binary",
			want:  "base64_binary",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			input := tc.input

			// Act
			snake := strcase.ToSnake(input)

			// Assert
			if got, want := snake, tc.want; !cmp.Equal(got, want) {
				t.Errorf("strcase.ToSnake(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func TestToScreamingSnake(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "camel boundary",
			input: "fooBar",
			want:  "FOO_BAR",
		},
		{
			name:  "acronym prefix",
			input: "HTTPServer",
			want:  "HTTP_SERVER",
		},
		{
			name:  "digit before uppercase",
			input: "Base64Binary",
			want:  "BASE64_BINARY",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			input := tc.input

			// Act
			screaming := strcase.ToScreamingSnake(input)

			// Assert
			if got, want := screaming, tc.want; !cmp.Equal(got, want) {
				t.Errorf("strcase.ToScreamingSnake(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func TestToPascal(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "snake input",
			input: "foo_bar",
			want:  "FooBar",
		},
		{
			name:  "camel boundary",
			input: "fooBar",
			want:  "FooBar",
		},
		{
			name:  "acronym normalized",
			input: "HTTPServer",
			want:  "HttpServer",
		},
		{
			name:  "digit before uppercase",
			input: "base64Binary",
			want:  "Base64Binary",
		},
		{
			name:  "caseless first rune",
			input: "中Foo",
			want:  "中Foo",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			input := tc.input

			// Act
			pascal := strcase.ToPascal(input)

			// Assert
			if got, want := pascal, tc.want; !cmp.Equal(got, want) {
				t.Errorf("strcase.ToPascal(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func TestToCamel(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty",
			input: "",
			want:  "",
		},
		{
			name:  "single word lowercased",
			input: "Foo",
			want:  "foo",
		},
		{
			name:  "snake input",
			input: "foo_bar",
			want:  "fooBar",
		},
		{
			name:  "acronym normalized",
			input: "HTTPServer",
			want:  "httpServer",
		},
		{
			name:  "digit before uppercase",
			input: "Base64Binary",
			want:  "base64Binary",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			input := tc.input

			// Act
			camel := strcase.ToCamel(input)

			// Assert
			if got, want := camel, tc.want; !cmp.Equal(got, want) {
				t.Errorf("strcase.ToCamel(%q) = %q, want %q", input, got, want)
			}
		})
	}
}
