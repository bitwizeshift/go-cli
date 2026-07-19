package update

import (
	"errors"

	"github.com/bitwizeshift/go-cli/internal/updatecheck"
)

var (
	// ErrInvalidVersion indicates a provider received a version that is not valid
	// semver.
	ErrInvalidVersion = updatecheck.ErrInvalidVersion

	// ErrUnexpectedStatus indicates a provider received a non-2xx HTTP response
	// while looking up the latest version.
	ErrUnexpectedStatus = errors.New("update: unexpected response status")

	// ErrDecodeResponse indicates a provider could not decode the response body
	// returned by its channel.
	ErrDecodeResponse = errors.New("update: decoding response")
)
