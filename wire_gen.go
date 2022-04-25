// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package main

import (
	"github.com/devtron-labs/central-api/api"
	"github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/internal/logger"
	"github.com/devtron-labs/central-api/pkg"
)

// Injectors from Wire.go:

func InitializeApp() (*App, error) {
	sugaredLogger := logger.NewSugardLogger()
	gitHubClient, err := util.NewGitHubClient(sugaredLogger)
	if err != nil {
		return nil, err
	}
	releaseCache := util.NewReleaseCache(sugaredLogger)
	releaseNoteServiceImpl := pkg.NewReleaseNoteServiceImpl(sugaredLogger, gitHubClient, releaseCache)
	webhookSecretValidatorImpl := pkg.NewWebhookSecretValidatorImpl(sugaredLogger, gitHubClient)
	restHandlerImpl := api.NewRestHandlerImpl(sugaredLogger, releaseNoteServiceImpl, webhookSecretValidatorImpl, gitHubClient)
	muxRouter := api.NewMuxRouter(sugaredLogger, restHandlerImpl)
	app := NewApp(muxRouter, sugaredLogger)
	return app, nil
}