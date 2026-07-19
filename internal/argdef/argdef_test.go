package argdef_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/bitwizeshift/go-cli/internal/argdef"
	"github.com/bitwizeshift/go-cli/internal/completion"
)

func TestBind(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		indices       []int
		unmatched     bool
		args          []string
		wantBound     map[int]string
		wantUnmatched []string
	}{
		{
			name:      "BindsPositionalByIndex",
			indices:   []int{0, 1},
			args:      []string{"alpha", "beta"},
			wantBound: map[int]string{0: "alpha", 1: "beta"},
		}, {
			name:      "SkipsOutOfRangeIndex",
			indices:   []int{0, 2},
			args:      []string{"alpha"},
			wantBound: map[int]string{0: "alpha"},
		}, {
			name:          "UnmatchedCollectsAllWhenNoPositionals",
			unmatched:     true,
			args:          []string{"a", "b", "c"},
			wantUnmatched: []string{"a", "b", "c"},
		}, {
			name:          "UnmatchedExcludesClaimedPreservingOrder",
			indices:       []int{0, 2},
			unmatched:     true,
			args:          []string{"a", "b", "c", "d"},
			wantBound:     map[int]string{0: "a", 2: "c"},
			wantUnmatched: []string{"b", "d"},
		}, {
			name:          "UnmatchedEmptyWhenAllClaimed",
			indices:       []int{0},
			unmatched:     true,
			args:          []string{"a"},
			wantBound:     map[int]string{0: "a"},
			wantUnmatched: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			cl := argdef.New()
			bound := map[int]string{}
			for _, index := range tc.indices {
				argdef.AddPositional(cl, &argdef.Positional{
					Index: index,
					Set:   func(value string) error { bound[index] = value; return nil },
				})
			}
			var rest []string
			if tc.unmatched {
				argdef.SetUnmatched(cl, &argdef.Unmatched{
					Set: func(values []string) error { rest = values; return nil },
				})
			}
			ctx := context.Background()

			// Act
			err := argdef.Bind(ctx, cl, tc.args)

			// Assert
			if got, want := err, error(nil); !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Bind(...) = %v, want %v", got, want)
			}
			if got, want := bound, tc.wantBound; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Bind(...) bound = %v, want %v", got, want)
			}
			if got, want := rest, tc.wantUnmatched; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("Bind(...) unmatched = %v, want %v", got, want)
			}
		})
	}
}

func TestBind_PositionalError(t *testing.T) {
	t.Parallel()

	// Arrange
	testErr := errors.New("boom")
	cl := argdef.New()
	argdef.AddPositional(cl, &argdef.Positional{
		Index: 0,
		Set:   func(string) error { return testErr },
	})
	ctx := context.Background()

	// Act
	err := argdef.Bind(ctx, cl, []string{"x"})

	// Assert
	if got, want := err, testErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
}

func TestBind_UnmatchedError(t *testing.T) {
	t.Parallel()

	// Arrange
	testErr := errors.New("boom")
	cl := argdef.New()
	argdef.SetUnmatched(cl, &argdef.Unmatched{
		Set: func([]string) error { return testErr },
	})
	ctx := context.Background()

	// Act
	err := argdef.Bind(ctx, cl, []string{"x"})

	// Assert
	if got, want := err, testErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
}

// constantFallback returns a fallback function yielding value unconditionally.
func constantFallback(value string) argdef.FallbackFunc {
	return func(context.Context) (string, error) { return value, nil }
}

// failingFallback returns a fallback function that always fails with err.
func failingFallback(err error) argdef.FallbackFunc {
	return func(context.Context) (string, error) { return "", err }
}

func TestBind_PositionalFallbacks(t *testing.T) {
	errSet := errors.New("set failed")
	errCompute := errors.New("compute failed")

	testCases := []struct {
		name          string
		env           map[string]string
		envFallbacks  []string
		funcFallbacks []argdef.FallbackFunc
		setErr        error
		args          []string
		want          string
		wantErr       error
	}{
		{
			name:         "EnvAppliesWhenArgMissing",
			env:          map[string]string{"ARG_ONE": "from-env"},
			envFallbacks: []string{"ARG_ONE"},
			args:         nil,
			want:         "from-env",
			wantErr:      nil,
		}, {
			name:          "ArgumentTakesPrecedenceOverFallbacks",
			env:           map[string]string{"ARG_ONE": "from-env"},
			envFallbacks:  []string{"ARG_ONE"},
			funcFallbacks: []argdef.FallbackFunc{constantFallback("from-func")},
			args:          []string{"given"},
			want:          "given",
			wantErr:       nil,
		}, {
			name:         "FirstSetEnvKeyWins",
			env:          map[string]string{"ARG_SECOND": "beta"},
			envFallbacks: []string{"ARG_FIRST", "ARG_SECOND"},
			args:         nil,
			want:         "beta",
			wantErr:      nil,
		}, {
			name: "EmptyEnvValueIsSkipped",
			env: map[string]string{
				"ARG_EMPTY": "",
				"ARG_NEXT":  "next",
			},
			envFallbacks: []string{"ARG_EMPTY", "ARG_NEXT"},
			args:         nil,
			want:         "next",
			wantErr:      nil,
		}, {
			name:         "EnvAssignmentErrorIsWrapped",
			env:          map[string]string{"ARG_ONE": "from-env"},
			envFallbacks: []string{"ARG_ONE"},
			setErr:       errSet,
			args:         nil,
			want:         "from-env",
			wantErr:      argdef.ErrSettingEnvFlag,
		}, {
			name:          "FuncAppliesWhenNoEnvMatches",
			envFallbacks:  []string{"ARG_UNSET"},
			funcFallbacks: []argdef.FallbackFunc{constantFallback("from-func")},
			args:          nil,
			want:          "from-func",
			wantErr:       nil,
		}, {
			name:          "EnvTakesPrecedenceOverFunc",
			env:           map[string]string{"ARG_ONE": "from-env"},
			envFallbacks:  []string{"ARG_ONE"},
			funcFallbacks: []argdef.FallbackFunc{constantFallback("from-func")},
			args:          nil,
			want:          "from-env",
			wantErr:       nil,
		}, {
			name: "EmptyFuncResultIsSkipped",
			funcFallbacks: []argdef.FallbackFunc{
				constantFallback(""),
				constantFallback("second"),
			},
			args:    nil,
			want:    "second",
			wantErr: nil,
		}, {
			name: "FuncErrorIsWrapped",
			funcFallbacks: []argdef.FallbackFunc{
				failingFallback(errCompute),
				constantFallback("unreached"),
			},
			args:    nil,
			want:    "",
			wantErr: argdef.ErrComputingFuncFlag,
		}, {
			name:          "FuncAssignmentErrorIsWrapped",
			funcFallbacks: []argdef.FallbackFunc{constantFallback("from-func")},
			setErr:        errSet,
			args:          nil,
			want:          "from-func",
			wantErr:       argdef.ErrSettingFuncFlag,
		}, {
			name:    "NoFallbacksLeavesValueUnassigned",
			args:    nil,
			want:    "",
			wantErr: nil,
		}, {
			name:          "AllFuncsEmptyLeavesValueUnassigned",
			funcFallbacks: []argdef.FallbackFunc{constantFallback(""), constantFallback("")},
			args:          nil,
			want:          "",
			wantErr:       nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			for key, value := range tc.env {
				t.Setenv(key, value)
			}
			cl := argdef.New()
			var bound string
			argdef.AddPositional(cl, &argdef.Positional{
				Index:         0,
				EnvFallbacks:  tc.envFallbacks,
				FuncFallbacks: tc.funcFallbacks,
				Set:           func(value string) error { bound = value; return tc.setErr },
			})
			ctx := context.Background()

			// Act
			err := argdef.Bind(ctx, cl, tc.args)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("Bind(...) = %v, want %v", got, want)
			}
			if got, want := bound, tc.want; !cmp.Equal(got, want) {
				t.Errorf("Bind(...) bound = %q, want %q", got, want)
			}
		})
	}
}

func TestBind_FallbackErrorPrecedesUnmatched(t *testing.T) {
	t.Parallel()

	// Arrange
	testErr := errors.New("compute failed")
	cl := argdef.New()
	argdef.AddPositional(cl, &argdef.Positional{
		Index:         5,
		FuncFallbacks: []argdef.FallbackFunc{failingFallback(testErr)},
		Set:           func(string) error { return nil },
	})
	var rest []string
	argdef.SetUnmatched(cl, &argdef.Unmatched{
		Set: func(values []string) error { rest = values; return nil },
	})
	ctx := context.Background()

	// Act
	err := argdef.Bind(ctx, cl, []string{"a"})

	// Assert
	if got, want := err, argdef.ErrComputingFuncFlag; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
		t.Fatalf("Bind(...) = %v, want %v", got, want)
	}
	if got, want := rest, []string(nil); !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
		t.Errorf("Bind(...) unmatched = %v, want %v", got, want)
	}
}

// completerOf returns a completion function offering candidate alone, so that a
// collected completion can be identified by the value it returns.
func completerOf(candidate string) completion.Func {
	return func(string) ([]string, completion.Directive) {
		return []string{candidate}, completion.NoFileComp
	}
}

// candidatesOf invokes each collected completion function and returns the single
// candidate it offers, keyed by the index it completes. Functions are not
// comparable, so this reduces them to comparable values.
func candidatesOf(fns map[int]completion.Func) map[int]string {
	result := map[int]string{}
	for index, fn := range fns {
		candidates, _ := fn("")
		result[index] = candidates[0]
	}
	return result
}

func TestPositionalCompletions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		positionals []*argdef.Positional
		want        map[int]string
	}{
		{
			name: "CollectsCompletersByIndex",
			positionals: []*argdef.Positional{
				{Index: 0, Complete: completerOf("first")},
				{Index: 1, Complete: completerOf("second")},
			},
			want: map[int]string{
				0: "first",
				1: "second",
			},
		},
		{
			name: "SkipsPositionalsWithoutCompleter",
			positionals: []*argdef.Positional{
				{Index: 0, Complete: completerOf("first")},
				{Index: 1, Complete: nil},
			},
			want: map[int]string{
				0: "first",
			},
		},
		{
			name:        "NoPositionals",
			positionals: nil,
			want:        map[int]string{},
		},
		{
			name: "NoCompleters",
			positionals: []*argdef.Positional{
				{Index: 0, Complete: nil},
				{Index: 1, Complete: nil},
			},
			want: map[int]string{},
		},
		{
			name: "LaterRegistrationWins",
			positionals: []*argdef.Positional{
				{Index: 0, Complete: completerOf("first")},
				{Index: 0, Complete: completerOf("second")},
			},
			want: map[int]string{
				0: "second",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := argdef.New()
			for _, p := range tc.positionals {
				argdef.AddPositional(sut, p)
			}

			// Act
			fns := argdef.PositionalCompletions(sut)

			// Assert
			if got, want := candidatesOf(fns), tc.want; !cmp.Equal(got, want, cmpopts.EquateEmpty()) {
				t.Errorf("PositionalCompletions(...) = %v, want %v\n%s", got, want, cmp.Diff(want, got, cmpopts.EquateEmpty()))
			}
		})
	}
}
