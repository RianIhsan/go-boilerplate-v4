package constants

// ContextKey is a distinct type for context.WithValue keys, avoiding
// collisions with keys set by other packages using plain strings.
type ContextKey string

const (
	ContextKeyUserID ContextKey = "user_id"
	ContextKeyEmail  ContextKey = "email"
)

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100

	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	CacheKeyUserPrefix = "user:"
	CacheTTLUser       = 300 // seconds

	// MaxRequestBodyBytes caps the size of a decoded JSON request body.
	// Applied to endpoints reachable without authentication (e.g. register,
	// login) so an oversized payload can't be used for memory-exhaustion DoS.
	MaxRequestBodyBytes = 1 << 20 // 1 MiB
)
