package handler

type Handlers struct {
	Auth    *AuthHandler
	OAuth   *OAuthHandler
	Private *PrivateHandler
}

func New(auth *AuthHandler, oauth *OAuthHandler, private *PrivateHandler) Handlers {
	return Handlers{Auth: auth, OAuth: oauth, Private: private}
}
