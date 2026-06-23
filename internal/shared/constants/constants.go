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
)
