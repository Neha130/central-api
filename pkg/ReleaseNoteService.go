package pkg

import (
	"context"
	"encoding/json"
	util "github.com/devtron-labs/central-api/client"
	"github.com/devtron-labs/central-api/common"
	"github.com/google/go-github/github"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"time"
)

type ReleaseNoteService interface {
	GetReleases() ([]*common.Release, error)
	UpdateReleases(requestBodyBytes []byte) (bool, error)
}

type ReleaseNoteServiceImpl struct {
	logger       *zap.SugaredLogger
	client       *util.GitHubClient
	releaseCache *util.ReleaseCache
}

func NewReleaseNoteServiceImpl(logger *zap.SugaredLogger, client *util.GitHubClient, releaseCache *util.ReleaseCache) *ReleaseNoteServiceImpl {
	return &ReleaseNoteServiceImpl{
		logger:       logger,
		client:       client,
		releaseCache: releaseCache,
	}
}

func (impl *ReleaseNoteServiceImpl) UpdateReleases(requestBodyBytes []byte) (bool, error) {
	data := make(map[string]interface{})
	err := json.Unmarshal(requestBodyBytes, &data)
	if err != nil {
		impl.logger.Errorw("unmarshal error", "err", err)
		return false, err
	}
	action := data["action"].(string)
	if action != "published" {
		return false, nil
	}
	releaseData := data["release"].(map[string]interface{})
	releaseName := releaseData["name"].(string)
	tagName := releaseData["tag_name"].(string)
	createdAtString := releaseData["created_at"].(string)
	createdAt, error := time.Parse("2006-01-02T15:04", createdAtString)
	if error != nil {
		impl.logger.Error(error)
		//return false, nil
	}
	body := releaseData["body"].(string)
	releaseInfo := &common.Release{
		TagName:     tagName,
		ReleaseName: releaseName,
		Body:        body,
		CreatedAt:   createdAt,
	}

	//updating cache, fetch existing object and append new item
	var releaseList []*common.Release
	cachedReleases := impl.releaseCache.GetReleaseCache()
	if cachedReleases != nil {
		itemMap, ok := cachedReleases.(map[string]cache.Item)
		if !ok {
			// Can't assert, handle error.
		}
		impl.logger.Info(itemMap)
		if itemMap != nil {
			items := itemMap["releases"]
			if items.Object != nil {
				releases := items.Object.([]*common.Release)
				releaseList = append(releaseList, releases...)
			}
		}
	}
	releaseList = append(releaseList, releaseInfo)
	impl.releaseCache.UpdateReleaseCache(releaseList)
	return true, nil
}

func (impl *ReleaseNoteServiceImpl) GetReleases() ([]*common.Release, error) {
	var releaseList []*common.Release
	cachedReleases := impl.releaseCache.GetReleaseCache()
	if cachedReleases != nil {
		itemMap, ok := cachedReleases.(map[string]cache.Item)
		if !ok {
			// Can't assert, handle error.
		}
		impl.logger.Info(itemMap)
		if itemMap != nil {
			items := itemMap["releases"]
			if items.Object != nil {
				releases := items.Object.([]*common.Release)
				releaseList = append(releaseList, releases...)
			}
		}
	}
	//todo - get it from cache

	if releaseList == nil {
		releases, _, err := impl.client.GitHubClient.Repositories.ListReleases(context.Background(), impl.client.GitHubConfig.GitHubOrg, "release-notes", &github.ListOptions{})
		if err != nil {
			responseErr, ok := err.(*github.ErrorResponse)
			if !ok || responseErr.Response.StatusCode != 404 {
				impl.logger.Errorw("error in fetching releases from github", "err", err, "config", "config")
				return nil, err
			} else {
				//newFile = true
			}
		}
		result := &common.ReleaseList{}
		var releasesDto []*common.Release
		for _, item := range releases {
			dto := &common.Release{
				TagName:     *item.TagName,
				ReleaseName: *item.Name,
				CreatedAt:   item.CreatedAt.Time,
				PublishedAt: item.PublishedAt.Time,
				Body:        *item.Body,
			}
			releasesDto = append(releasesDto, dto)
		}
		//todo pagination
		result.Releases = releasesDto
		releaseList = releasesDto
		impl.releaseCache.UpdateReleaseCache(releaseList)
	}

	return releaseList, nil
}