package errors

type ABFError string

func (e ABFError) Error() string {
	return string(e)
}

var (
	ErrorRedisConnect    = ABFError("unable to connect redis")
	ErrorInvalidReqBody  = ABFError("invalid request body: some field is empty")
	ErrorContextTimeout  = ABFError("context deadline exceeded")
	ErrorInvalidIP       = ABFError("invalid IP address")
	ErrorInvalidIPNet    = ABFError("invalid IP/net address")
	ErrorInvalidListName = ABFError("invalid list name")
	ErrorInternalDB      = ABFError("internal db error")
	ErrorDBNotFound      = ABFError("key not found")
)
