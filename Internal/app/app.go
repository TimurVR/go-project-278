package app

import (
	"context"
	"go-project-278/Internal/handler"
	"go-project-278/Internal/repository"
	"github.com/gin-gonic/gin"
	"database/sql"
)

type App struct {
	Ctx       context.Context
	Repo      *repository.Repository
	Handler   *handler.App
}

func NewApp(ctx context.Context, dbpool *sql.DB) *App {
	repo := repository.NewLinkRepository(dbpool)
	handlerApp := &handler.App{
		Ctx:  ctx,
		Repo: repo,
	}
	return &App{
		Ctx:       ctx,
		Repo:      repo,
		Handler:   handlerApp,
	}
}

func (a *App) Routes(r *gin.Engine) {
	a.Handler.Routes(r)
}

