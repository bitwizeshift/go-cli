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

// unmatchedOf returns the [argdef.Unmatched] registered on cl, so a test can
// drive its decode closure directly.
func unmatchedOf(cl *arg.CommandLine) *argdef.Unmatched {
	return argdef.GetUnmatched((*argdef.CommandLine)(cl))
}

// unmatchedCompletionOf completes the unmatched binding registered on cl with
// the partial word toComplete. It fails the test if that binding registered no
// completion.
func unmatchedCompletionOf(t testing.TB, cl *arg.CommandLine, toComplete string) offered {
	t.Helper()

	fn := argdef.UnmatchedCompletion((*argdef.CommandLine)(cl))
	if fn == nil {
		t.Fatalf("Add(...) registered no unmatched completion, want one")
	}
	candidates, directive := fn(toComplete)
	return offered{
		Candidates: candidates,
		Directive:  directive,
	}
}

func TestUnmatched_Set(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []arg.Option
		values  []string
		want    []string
		wantErr error
	}{
		{
			name:    "AssignsValues",
			options: nil,
			values:  []string{"a", "b"},
			want:    []string{"a", "b"},
			wantErr: nil,
		}, {
			name:    "EmptySetAssignsNoValues",
			options: nil,
			values:  nil,
			want:    nil,
			wantErr: nil,
		}, {
			name:    "DecodesEachValueIndependently",
			options: []arg.Option{arg.UnmarshalWith(yell)},
			values:  []string{"abc", "def"},
			want:    []string{"ABC", "DEF"},
			wantErr: nil,
		}, {
			name:    "DecoderErrorPropagates",
			options: []arg.Option{arg.UnmarshalWith(failString)},
			values:  []string{"x"},
			want:    nil,
			wantErr: errDecode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			var dst []string
			addUnmatched(cl, &dst, tc.options...)
			unmatched := unmatchedOf(cl)

			// Act
			err := unmatched.Set(tc.values)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Unmatched.Set(...) = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Unmatched.Set(...) values = %v, want %v", got, want)
			}
		})
	}
}

func TestUnmatched_SetDecodesTypedValues(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst []int
	addUnmatched(cl, &dst)
	unmatched := unmatchedOf(cl)

	// Act
	err := unmatched.Set([]string{"1", "2", "3"})

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Unmatched.Set(...) = %v, want %v", got, want)
	}
	if got, want := dst, []int{1, 2, 3}; !cmp.Equal(got, want) {
		t.Errorf("Unmatched.Set(...) values = %v, want %v", got, want)
	}
}

func TestUnmatched_SetLeavesDestinationOnDecodeError(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	dst := []int{7}
	addUnmatched(cl, &dst)
	unmatched := unmatchedOf(cl)

	// Act
	err := unmatched.Set([]string{"1", "not-a-number"})

	// Assert
	if got, want := err, cmpopts.AnyError; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Unmatched.Set(...) = %v, want %v", got, want)
	}
	if got, want := dst, []int{7}; !cmp.Equal(got, want) {
		t.Errorf("Unmatched.Set(...) values = %v, want %v", got, want)
	}
}

func TestUnmatched_SetInvokesCallbackPerValue(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst, seen []string
	addUnmatched(cl, &dst, arg.Callback(func(s string) { seen = append(seen, s) }))
	unmatched := unmatchedOf(cl)

	// Act
	err := unmatched.Set([]string{"a", "b", "c"})

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Unmatched.Set(...) = %v, want %v", got, want)
	}
	if got, want := seen, []string{"a", "b", "c"}; !cmp.Equal(got, want) {
		t.Errorf("Unmatched.Set(...) callback values = %v, want %v", got, want)
	}
}

func TestUnmatched_SetCallbackError(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst []string
	addUnmatched(cl, &dst, arg.Callback(func(string) error { return errDecode }))
	unmatched := unmatchedOf(cl)

	// Act
	err := unmatched.Set([]string{"x"})

	// Assert
	if got, want := err, errDecode; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Unmatched.Set(...) = %v, want %v", got, want)
	}
}

func TestUnmatched_Metadata(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []arg.Option
		want    *argtest.Unmatched
	}{
		{
			name:    "ReportsElementTypeAndUsage",
			options: []arg.Option{arg.Usage("the rest")},
			want: &argtest.Unmatched{
				Type:  "string",
				Usage: "the rest",
			},
		}, {
			name:    "TypeOverridesReportedName",
			options: []arg.Option{arg.Type("path"), arg.Usage("paths to read")},
			want: &argtest.Unmatched{
				Type:  "path",
				Usage: "paths to read",
			},
		}, {
			name:    "NoOptionsReportsBareType",
			options: nil,
			want: &argtest.Unmatched{
				Type:  "string",
				Usage: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			var dst []string
			addUnmatched(cl, &dst, tc.options...)

			// Act
			unmatched := argtest.GetUnmatched(cl)

			// Assert
			if got, want := unmatched, tc.want; !cmp.Equal(got, want) {
				t.Errorf("GetUnmatched(...) = %+v, want %+v", got, want)
			}
		})
	}
}

func TestUnmatched_MetadataReportsTypedElement(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst []int
	addUnmatched(cl, &dst)

	// Act
	unmatched := argtest.GetUnmatched(cl)

	// Assert
	want := &argtest.Unmatched{Type: "int", Usage: ""}
	if got, want := unmatched, want; !cmp.Equal(got, want) {
		t.Errorf("GetUnmatched(...) = %+v, want %+v", got, want)
	}
}

func TestUnmatched_NotBound_ReportsNone(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()

	// Act
	unmatched := argtest.GetUnmatched(cl)

	// Assert
	if got, want := unmatched, (*argtest.Unmatched)(nil); !cmp.Equal(got, want) {
		t.Errorf("GetUnmatched(...) = %+v, want %+v", got, want)
	}
}

func TestUnmatched_BoundTwice_Panics(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var first, second []string
	cl.Add(arg.Unmatched(&first))

	// Act & Assert
	requirePanic(t, func() { cl.Add(arg.Unmatched(&second)) })
}

func TestUnmatched_CompletionOptions(t *testing.T) {
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
		}, {
			name:       "CompleterFuncOffersItsCandidates",
			option:     arg.CompleterFunc(suffixCompleter),
			toComplete: "value",
			want: offered{
				Candidates: []string{"value-done"},
				Directive:  completion.NoFileComp,
			},
		}, {
			name:       "CompleteFilesDefersToShell",
			option:     arg.CompleteFiles(),
			toComplete: "",
			want: offered{
				Candidates: nil,
				Directive:  completion.Default,
			},
		}, {
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
			var dst []string
			addUnmatched(cl, &dst, tc.option)

			// Act
			offer := unmatchedCompletionOf(t, cl, tc.toComplete)

			// Assert
			if got, want := offer, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Unmatched(..., option) completion = %+v, want %+v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}

func TestUnmatched_NoCompletionOption_RegistersNone(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst []string
	addUnmatched(cl, &dst)

	// Act
	fn := argdef.UnmatchedCompletion((*argdef.CommandLine)(cl))

	// Assert
	if got, want := fn == nil, true; !cmp.Equal(got, want) {
		t.Errorf("UnmatchedCompletion(...) = nil %t, want %t", got, want)
	}
}

func TestUnmatched_Fallbacks(t *testing.T) {
	errCompute := errors.New("compute failed")

	testCases := []struct {
		name    string
		env     map[string]string
		options []arg.Option
		args    []string
		want    []string
		wantErr error
	}{
		{
			name:    "DefaultFromEnvAppliesWhenSetEmpty",
			env:     map[string]string{"UNMATCHED_ONE": "from-env"},
			options: []arg.Option{arg.DefaultFromEnv("UNMATCHED_ONE")},
			args:    nil,
			want:    []string{"from-env"},
			wantErr: nil,
		}, {
			name:    "DefaultFromEnvSplitsOnComma",
			env:     map[string]string{"UNMATCHED_ONE": "alpha,beta,gamma"},
			options: []arg.Option{arg.DefaultFromEnv("UNMATCHED_ONE")},
			args:    nil,
			want:    []string{"alpha", "beta", "gamma"},
			wantErr: nil,
		}, {
			name:    "DefaultFromEnvHonoursCSVQuoting",
			env:     map[string]string{"UNMATCHED_ONE": `"alpha,beta",gamma`},
			options: []arg.Option{arg.DefaultFromEnv("UNMATCHED_ONE")},
			args:    nil,
			want:    []string{"alpha,beta", "gamma"},
			wantErr: nil,
		}, {
			name:    "DefaultFromFuncAppliesWhenSetEmpty",
			options: []arg.Option{arg.DefaultFromFunc(constantDefault("one,two"))},
			args:    nil,
			want:    []string{"one", "two"},
			wantErr: nil,
		}, {
			name: "DefaultFromEnvPrecedesDefaultFromFunc",
			env:  map[string]string{"UNMATCHED_ONE": "from-env"},
			options: []arg.Option{
				arg.DefaultFromEnv("UNMATCHED_ONE"),
				arg.DefaultFromFunc(constantDefault("from-func")),
			},
			args:    nil,
			want:    []string{"from-env"},
			wantErr: nil,
		}, {
			name:    "SuppressedWhenAnyArgUnclaimed",
			env:     map[string]string{"UNMATCHED_ONE": "from-env"},
			options: []arg.Option{arg.DefaultFromEnv("UNMATCHED_ONE")},
			args:    []string{"typed"},
			want:    []string{"typed"},
			wantErr: nil,
		}, {
			name:    "DefaultFromFuncErrorPropagates",
			options: []arg.Option{arg.DefaultFromFunc(failingDefault(errCompute))},
			args:    nil,
			want:    nil,
			wantErr: argdef.ErrComputingFuncFlag,
		}, {
			name:    "DefaultFromEnvIgnoresUnsetKey",
			options: []arg.Option{arg.DefaultFromEnv("UNMATCHED_UNSET")},
			args:    nil,
			want:    nil,
			wantErr: nil,
		}, {
			name:    "NoFallbackLeavesEmptySet",
			options: nil,
			args:    nil,
			want:    nil,
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
			var dst []string
			cl.Add(arg.Unmatched(&dst, tc.options...))
			ctx := context.Background()

			// Act
			err := argdef.Bind(ctx, (*argdef.CommandLine)(cl), tc.args)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Bind(...) = %v, want %v", got, want)
			}
			if got, want := dst, tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Bind(...) values = %v, want %v", got, want)
			}
		})
	}
}

func TestUnmatched_FallbackDecodesTypedValues(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst []int
	cl.Add(arg.Unmatched(&dst, arg.DefaultFromFunc(constantDefault("1,2,3"))))
	ctx := context.Background()

	// Act
	err := argdef.Bind(ctx, (*argdef.CommandLine)(cl), nil)

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
	if got, want := dst, []int{1, 2, 3}; !cmp.Equal(got, want) {
		t.Errorf("Bind(...) values = %v, want %v", got, want)
	}
}

func TestUnmatched_FallbackDecodeErrorIsWrapped(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var dst []string
	cl.Add(arg.Unmatched(&dst,
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
	if got, want := dst, []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Bind(...) values = %v, want %v", got, want)
	}
}

func TestUnmatched_FallbackEnvDecodeErrorIsWrapped(t *testing.T) {
	// Arrange
	t.Setenv("UNMATCHED_BAD", "x")
	cl := argtest.NewCommandLine()
	var dst []string
	cl.Add(arg.Unmatched(&dst,
		arg.UnmarshalWith(failString),
		arg.DefaultFromEnv("UNMATCHED_BAD"),
	))
	ctx := context.Background()

	// Act
	err := argdef.Bind(ctx, (*argdef.CommandLine)(cl), nil)

	// Assert
	if got, want := err, argdef.ErrSettingEnvFlag; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
	if got, want := dst, []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Bind(...) values = %v, want %v", got, want)
	}
}

func TestUnmatched_ClaimsOnlyArgsPositionalsLeave(t *testing.T) {
	t.Parallel()

	// Arrange
	cl := argtest.NewCommandLine()
	var first string
	var rest []string
	cl.Add(
		arg.Positional("first", 0, &first),
		arg.Unmatched(&rest),
	)
	ctx := context.Background()

	// Act
	err := argdef.Bind(ctx, (*argdef.CommandLine)(cl), []string{"a", "b", "c"})

	// Assert
	if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
	if got, want := first, "a"; !cmp.Equal(got, want) {
		t.Errorf("Bind(...) positional = %q, want %q", got, want)
	}
	if got, want := rest, []string{"b", "c"}; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Bind(...) values = %v, want %v", got, want)
	}
}

func TestUnmatched_Required(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		options []arg.Option
		want    bool
	}{
		{
			name:    "OptionalByDefault",
			options: nil,
			want:    false,
		}, {
			name:    "MarkedRequired",
			options: []arg.Option{arg.Required()},
			want:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argtest.NewCommandLine()
			var rest []string
			cl.Add(arg.Unmatched(&rest, tc.options...))

			// Act
			unmatched := unmatchedOf(cl)

			// Assert
			if got, want := unmatched.Required, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Unmatched(...) required = %t, want %t", got, want)
			}
		})
	}
}
