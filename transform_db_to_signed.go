package main

import (
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	videoURL := *video.VideoURL
	splitURL := strings.Split(videoURL, ",")

	if len(splitURL) != 2 {
		return video, nil
	}

	bucket := splitURL[0]
	key := splitURL[1]

	presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Hour)

	if err != nil {
		return database.Video{}, err
	}

	err = cfg.db.UpdateVideo(database.Video{
		VideoURL:          &presignedURL,
		ID:                video.ID,
		CreatedAt:         video.CreatedAt,
		UpdatedAt:         time.Now(),
		ThumbnailURL:      video.ThumbnailURL,
		CreateVideoParams: video.CreateVideoParams,
	})

	if err != nil {
		return database.Video{}, err
	}

	return video, nil

}
