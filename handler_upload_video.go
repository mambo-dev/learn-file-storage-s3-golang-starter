package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	maxLimit := 1 >> 30
	http.MaxBytesReader(w, r.Body, int64(maxLimit))

	videoID := r.PathValue("videoID")

	videoUUID, err := uuid.Parse(videoID)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	userID, msg, err := authenticateUser(cfg.jwtSecret, r.Header, w)

	if err != nil {
		respondWithError(w, http.StatusUnauthorized, msg, err)
		return
	}

	videoMetadata, err := cfg.db.GetVideo(videoUUID)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video id", err)
		return
	}

	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Cannot upload video", err)
		return
	}

	videoFile, fileHeader, err := r.FormFile("video")

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't upload video", err)
		return
	}

	defer videoFile.Close()
	mediaTypeHeader := fileHeader.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(mediaTypeHeader)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid content type.", err)
		return
	}

	if mediaTypeHeader != "video/mp4" {
		respondWithError(w, http.StatusForbidden, "Only .png and .jpeg files allowed.", err)
		return
	}

	file, err := os.CreateTemp("", "tubely-mp4")

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	defer os.Remove(file.Name())
	defer file.Close()

	io.Copy(file, videoFile)
	file.Seek(0, io.SeekStart)

	random32, err := random32Generator()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	aspectRatio, err := getVideoAspectRatio(file.Name())

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	processedVideo, err := processVideoForFastStart(file.Name())

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	processedFile, err := os.Open(processedVideo)
	defer os.Remove(processedFile.Name())
	defer processedFile.Close()

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	extensions, err := mime.ExtensionsByType(mediaType)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	display := ""

	if aspectRatio == "16:9" {
		display = "landscape"
	} else if aspectRatio == "9:16" {
		display = "portrait"
	} else {
		display = "other"
	}

	s3Key := fmt.Sprintf("%v/%v%v", display, hex.EncodeToString(random32), extensions[3])

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &s3Key,
		Body:        processedFile,
		ContentType: &mediaType,
	})

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to upload video", err)
		return
	}

	videoURL := fmt.Sprintf("https://%v.s3.%v.amazonaws.com/%v", cfg.s3Bucket, cfg.s3Region, s3Key)

	err = cfg.db.UpdateVideo(database.Video{
		VideoURL:          &videoURL,
		ID:                videoMetadata.ID,
		CreatedAt:         videoMetadata.CreatedAt,
		UpdatedAt:         time.Now(),
		ThumbnailURL:      videoMetadata.ThumbnailURL,
		CreateVideoParams: videoMetadata.CreateVideoParams,
	})

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}
