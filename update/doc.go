// Package update reports whether a newer release of the running application is
// available from the channel it was distributed through.
//
// An application supplies its [BuildInfo] (the running version and the source it
// was installed from) and a [ProviderRegistry] mapping source names to the
// [Provider] that knows how to look up the latest version for that channel. A
// [Checker] ties the two together and decides whether an update exists.
//
// Providers reach the network on every call. Wrap one in a [CacheProvider] to
// memoize the result on disk so repeated invocations do not re-query the
// channel. Lookups are best-effort: callers are expected to treat any error as
// "no information available" rather than surfacing it to the user.
package update
