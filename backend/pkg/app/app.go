// Package app arma el router HTTP y los handlers de la API.
package app

import (
	"chiro/pkg/auth"
	"chiro/pkg/store"
)

// App agrupa las dependencias compartidas por los handlers.
type App struct {
	Store *store.Store
	Auth  *auth.Manager
}

func New(st *store.Store, authMgr *auth.Manager) *App {
	return &App{Store: st, Auth: authMgr}
}
