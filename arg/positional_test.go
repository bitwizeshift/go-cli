package arg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/arg"
	"github.com/bitwizeshift/go-cli/arg/argtest"
	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/bitwizeshift/go-cli/internal/completion"
)

// firstPositional returns the [argdef.Positional] registered on cl at position
// zero, so a test can drive its decode closure directly.
func firstPositional(cl *arg.CommandLine) *argdef.Positional {
	return argdef.Positionals((*argdef.CommandLine)(cl))[0]
}

// positionalCompletionOf completes the positional registered on cl at index with
// the partial word toComplete. It fails the test if that positional registered no
// completion.
func positionalCompletionOf(t testing.TB, cl *arg.CommandLine, index int, toComplete string) offered {
	t.Helper()

	fns := argdef.PositionalCompletions((*argdef.CommandLine)(cl))
	fn, ok := fns[index]
	if !ok {
		t.Fatalf("Add(...) registered no completion at index %d, want one", index)
	}
	candidates, directive := fn(toComplete)
	return offered{
		Candidates: candidates,
		Directive:  directive,
	}
}

func TestPositional_Set(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []arg.Option
		value   string
		want    string
		wantErr error
	}{
		{
			name:    "DecodesValue",
			options: nil,
			value:   "alpha",
			want:    "alpha",
			wantErr: nil,
		}, {
			name:    "TransformsValueWithDecoder",
			options: []arg.Option{arg.UnmarshalWith(yell)},
			value:   "abc",
			want:    "ABC",
			wantErr: nil,
		}, {
			name:    "DecoderErrorPropagates",
			options: []arg.Option{arg.UnmarshalWith(failString)},
			value:   "x",
			want:    "",
			wantErr: errDecode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			var dst string
			addPositional(cl, "value", 0, &dst, tc.options...)
			positional := firstPositional(cl)

			// Act
			err := positional.Set(tc.value)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Positional.Set(...) = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Positional.Set(...) value = %q, want %q", got, want)
			}
		})
	}
}

func TestPositional_SetDecodesTypedValue(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst int
	addPositional(cl, "count", 0, &dst)
	positional := firstPositional(cl)

	// Act
	err := positional.Set("42")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Positional.Set(...) = %v, want %v", got, want)
	}
	if got, want := dst, 42; !cmp.Equal(got, want) {
		t.Errorf("Positional.Set(...) value = %d, want %d", got, want)
	}
}

func TestPositional_SetInvokesCallback(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst, seen string
	addPositional(cl, "value", 0, &dst, arg.Callback(func(s string) { seen = s }))
	positional := firstPositional(cl)

	// Act
	err := positional.Set("hello")

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Positional.Set(...) = %v, want %v", got, want)
	}
	if got, want := seen, "hello"; !cmp.Equal(got, want) {
		t.Errorf("Positional.Set(...) callback value = %q, want %q", got, want)
	}
}

func TestPositional_SetCallbackError(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst string
	addPositional(cl, "value", 0, &dst, arg.Callback(func(string) error { return errDecode }))
	positional := firstPositional(cl)

	// Act
	err := positional.Set("x")

	// Assert
	if got, want := err, errDecode; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Positional.Set(...) = %v, want %v", got, want)
	}
}

func TestPositional_Metadata(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var name string
	var count int
	addPositional(cl, "name", 0, &name, arg.Usage("the name"))
	addPositional(cl, "count", 1, &count, arg.Type("number"), arg.Usage("how many"))

	// Act
	positionals := argtest.AllPositionals(cl)

	// Assert
	want := []*argtest.Positional{
		{Index: 0, Name: "name", Type: "string", Usage: "the name"},
		{Index: 1, Name: "count", Type: "number", Usage: "how many"},
	}
	if got, want := positionals, want; !cmp.Equal(got, want) {
		t.Errorf("AllPositionals(...) = %+v, want %+v", got, want)
	}
}

func TestPositional_CompletionOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		option     arg.Option
		toComplete string
		want       offered
	}{
		{
			name:       "CompleteFromMatchesPrefix",
			option:     arg.CompleteFrom("json", "yaml", "jsonl"),
			toComplete: "js",
			want: offered{
				Candidates: []string{"json", "jsonl"},
				Directive:  completion.NoFileComp,
			},
		},
		{
			name:       "CompleterFuncOffersItsCandidates",
			option:     arg.CompleterFunc(suffixCompleter),
			toComplete: "value",
			want: offered{
				Candidates: []string{"value-done"},
				Directive:  completion.NoFileComp,
			},
		},
		{
			name:       "CompleteFilesDefersToShell",
			option:     arg.CompleteFiles(),
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  completion.Default,
			},
		},
		{
			name:       "CompleteFilesMatchingNormalizesExtensions",
			option:     arg.CompleteFilesMatching(".json", "yaml"),
			toComplete: "",
			want: offered{
				Candidates: []string{"json", "yaml"},
				Directive:  completion.FilterFileExt,
			},
		},
		{
			name:       "CompleteDirsFiltersDirectories",
			option:     arg.CompleteDirs(),
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  completion.FilterDirs,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			var dst string
			cl.Add(arg.Positional("value", 0, &dst, tc.option))

			// Act
			offer := positionalCompletionOf(t, cl, 0, tc.toComplete)

			// Assert
			if got, want := offer, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Positional(..., option) completion = %+v, want %+v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}

// constantDefault returns an [arg.DefaultFunc] yielding value unconditionally.
func constantDefault(value string) arg.DefaultFunc {
	return func(context.Context) (string, error) { return value, nil }
}

// failingDefault returns an [arg.DefaultFunc] that always fails with err.
func failingDefault(err error) arg.DefaultFunc {
	return func(context.Context) (string, error) { return "", err }
}

func TestPositional_Fallbacks(t *testing.T) {
	errCompute := errors.New("compute failed")

	testCases := []struct {
		name    string
		env     map[string]string
		options []arg.Option
		want    string
		wantErr error
	}{
		{
			name:    "DefaultFromEnvAppliesWhenArgMissing",
			env:     map[string]string{"POS_ONE": "from-env"},
			options: []arg.Option{arg.DefaultFromEnv("POS_ONE")},
			want:    "from-env",
			wantErr: nil,
		}, {
			name:    "DefaultFromFuncAppliesWhenArgMissing",
			options: []arg.Option{arg.DefaultFromFunc(constantDefault("from-func"))},
			want:    "from-func",
			wantErr: nil,
		}, {
			name: "DefaultFromEnvPrecedesDefaultFromFunc",
			env:  map[string]string{"POS_ONE": "from-env"},
			options: []arg.Option{
				arg.DefaultFromEnv("POS_ONE"),
				arg.DefaultFromFunc(constantDefault("from-func")),
			},
			want:    "from-env",
			wantErr: nil,
		}, {
			name: "FirstNonEmptyDefaultFromFuncWins",
			options: []arg.Option{
				arg.DefaultFromFunc(constantDefault("")),
				arg.DefaultFromFunc(constantDefault("second")),
				arg.DefaultFromFunc(constantDefault("third")),
			},
			want:    "second",
			wantErr: nil,
		}, {
			name:    "DefaultFromFuncErrorPropagates",
			options: []arg.Option{arg.DefaultFromFunc(failingDefault(errCompute))},
			want:    "",
			wantErr: argdef.ErrComputingFuncFlag,
		}, {
			name:    "DefaultFromEnvIgnoresUnsetKey",
			options: []arg.Option{arg.DefaultFromEnv("POS_UNSET")},
			want:    "",
			wantErr: nil,
		}, {
			name:    "NoFallbackLeavesZeroValue",
			options: nil,
			want:    "",
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			cl := argtest.NewCommandLine()
			var dst string
			cl.Add(arg.Positional("value", 0, &dst, tc.options...))
			ctx := context.Background()

			// Act
			err := argdef.Bind(ctx, (*argdef.CommandLine)(cl), nil)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Bind(...) = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Bind(...) value = %q, want %q", got, want)
			}
		})
	}
}

func TestPositional_FallbackDecodesTypedValue(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst int
	cl.Add(arg.Positional("count", 0, &dst, arg.DefaultFromFunc(constantDefault("42"))))
	ctx := context.Background()

	// Act
	err := argdef.Bind(ctx, (*argdef.CommandLine)(cl), nil)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
	if got, want := dst, 42; !cmp.Equal(got, want) {
		t.Errorf("Bind(...) value = %d, want %d", got, want)
	}
}

func TestPositional_FallbackDecodeErrorIsWrapped(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst string
	cl.Add(arg.Positional("value", 0, &dst,
		arg.UnmarshalWith(failString),
		arg.DefaultFromFunc(constantDefault("x")),
	))
	ctx := context.Background()

	// Act
	err := argdef.Bind(ctx, (*argdef.CommandLine)(cl), nil)

	// Assert
	if got, want := err, argdef.ErrSettingFuncFlag; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
	if got, want := dst, ""; !cmp.Equal(got, want) {
		t.Errorf("Bind(...) value = %q, want %q", got, want)
	}
}

func TestPositional_FallbackInvokesCallback(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst, seen string
	cl.Add(arg.Positional("value", 0, &dst,
		arg.DefaultFromFunc(constantDefault("from-func")),
		arg.Callback(func(s string) { seen = s }),
	))
	ctx := context.Background()

	// Act
	err := argdef.Bind(ctx, (*argdef.CommandLine)(cl), nil)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
	if got, want := seen, "from-func"; !cmp.Equal(got, want) {
		t.Errorf("Bind(...) callback value = %q, want %q", got, want)
	}
}

func TestPositional_NoCompletionOption_RegistersNone(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst string
	cl.Add(arg.Positional("value", 0, &dst))

	// Act
	fns := argdef.PositionalCompletions((*argdef.CommandLine)(cl))

	// Assert
	if got, want := len(fns), 0; !cmp.Equal(got, want) {
		t.Errorf("PositionalCompletions(...) = %d completions, want %d", got, want)
	}
}
