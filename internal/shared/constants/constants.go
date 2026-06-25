package constants

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
	CacheTTLUser       = 300

	MaxRequestBodyBytes = 1 << 20

	UploadTempDir = "tmp/uploads"

	MaxUploadFileBytes = 500 << 20
)
