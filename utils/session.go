package utils

import "github.com/gorilla/sessions"

var (
	key   = []byte("secret")
	store = sessions.NewCookieStore(key)
)
