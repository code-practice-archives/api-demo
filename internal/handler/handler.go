package handler

type Handlers struct {
	Auth    *AuthHandler
	Private *PrivateHandler
}

func New(auth *AuthHandler, private *PrivateHandler) Handlers {
	return Handlers{Auth: auth, Private: private}
}
