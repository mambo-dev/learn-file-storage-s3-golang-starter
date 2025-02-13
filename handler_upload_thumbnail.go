package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	err = r.ParseMultipartForm(maxMemory)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "failed to parse memory", err)
		return
	}

	file, fileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not get file", err)
		return
	}

	mediaTypeHeader := fileHeader.Header.Get("Content-Type")

	mediaType, _, err := mime.ParseMediaType(mediaTypeHeader)

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid content type;", err)
		return
	}

	videoMetadata, err := cfg.db.GetVideo(videoID)

	if err != nil {
		respondWithError(w, http.StatusNotFound, "could not get video metadata", err)
		return
	}

	if videoMetadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Invalid request did not create resource", err)
		return
	}

	extensions, err := mime.ExtensionsByType(mediaType)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	if len(extensions) < 1 {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", errors.New("no extensions found"))
		return
	}

	fileName := fmt.Sprintf("/%v%v", videoMetadata.ID, extensions[0])
	videoFilePath := filepath.Join(cfg.assetsRoot, fileName)

	savedFile, err := os.Create(videoFilePath)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	_, err = io.Copy(savedFile, file)

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	dataUrl := fmt.Sprintf("http://localhost:%v/%v", cfg.port, videoFilePath)

	fmt.Println(dataUrl)

	err = cfg.db.UpdateVideo(database.Video{
		ID:                videoMetadata.ID,
		UpdatedAt:         time.Now(),
		CreatedAt:         videoMetadata.CreatedAt,
		ThumbnailURL:      &dataUrl,
		VideoURL:          videoMetadata.VideoURL,
		CreateVideoParams: videoMetadata.CreateVideoParams,
	})

	if err != nil {
		respondWithError(w, http.StatusBadRequest, "could not update thumbnail url", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoMetadata)
}
