// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package main

import (
	"github.com/devtron-labs/central-api/api"
	"github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/internal/logger"
	"github.com/devtron-labs/central-api/pkg"
	"github.com/devtron-labs/central-api/pkg/releaseNote"
	"github.com/devtron-labs/central-api/pkg/sql"
)

// Injectors from Wire.go:

func InitializeApp() (*App, error) {
	sugaredLogger := logger.NewSugardLogger()
	gitHubClient, err := util.NewGitHubClient(sugaredLogger)
	if err != nil {
		return nil, err
	}
	moduleConfig, err := util.NewModuleConfig(sugaredLogger)
	if err != nil {
		return nil, err
	}
	config, err := sql.ParseConfig()
	if err != nil {
		return nil, err
	}
	db, err := sql.NewDbConnection(config, sugaredLogger)
	if err != nil {
		return nil, err
	}
	releaseNoteRepositoryImpl := releaseNote.NewReleaseNoteRepositoryImpl(db)
	releaseNoteServiceImpl, err := pkg.NewReleaseNoteServiceImpl(sugaredLogger, gitHubClient, moduleConfig, releaseNoteRepositoryImpl)
	if err != nil {
		return nil, err
	}
	webhookSecretValidatorImpl := pkg.NewWebhookSecretValidatorImpl(sugaredLogger, gitHubClient)
	ciBuildMetadataServiceImpl := pkg.NewCiBuildMetadataServiceImpl(sugaredLogger)
	restHandlerImpl := api.NewRestHandlerImpl(sugaredLogger, releaseNoteServiceImpl, webhookSecretValidatorImpl, gitHubClient, ciBuildMetadataServiceImpl)
	muxRouter := api.NewMuxRouter(sugaredLogger, restHandlerImpl)
	app := NewApp(muxRouter, sugaredLogger)
	return app, nil
}
