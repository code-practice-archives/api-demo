package handler

type Handlers struct {
	Auth *AuthHandler
}

func New(auth *AuthHandler) Handlers {
	return Handlers{Auth: auth}
}
