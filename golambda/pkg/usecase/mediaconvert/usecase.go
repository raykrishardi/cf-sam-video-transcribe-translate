package mediaconvert

import (
	"cf-sam-video-transcription-translate/config"
	"context"

	"cf-sam-video-transcription-translate/pkg/entity"
	"cf-sam-video-transcription-translate/pkg/repository"
)

type MediaConvertUseCase struct {
	AppConfig        *config.AppConfig
	MediaConvertRepo repository.MediaConvertRepo
}

func NewMediaConvertUseCase(appConfig *config.AppConfig, mediaConvertRepo repository.MediaConvertRepo) *MediaConvertUseCase {
	return &MediaConvertUseCase{
		AppConfig:        appConfig,
		MediaConvertRepo: mediaConvertRepo,
	}
}

func (uc *MediaConvertUseCase) ConvertMP4ToMP3(ctx context.Context, params entity.ConvertMP4ToMP3Input) error {
	return uc.MediaConvertRepo.ConvertMP4ToMP3(ctx, params)
}
