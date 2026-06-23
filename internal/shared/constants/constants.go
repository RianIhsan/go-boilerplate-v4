package constants

const (
	ContextKeyUserID = "user_id"
	ContextKeyEmail  = "email"

	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100

	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"

	CacheKeyUserPrefix = "user:"
	CacheTTLUser       = 300 // seconds
)
