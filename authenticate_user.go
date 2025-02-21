package main

import (
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func authenticateUser(jwtSecret string, headers http.Header, w http.ResponseWriter) (uuid.UUID, string, error) {

	token, err := auth.GetBearerToken(headers)
	if err != nil {
		return uuid.New(), "Couldn't find JWT", err
	}

	userID, err := auth.ValidateJWT(token, jwtSecret)
	if err != nil {
		return uuid.New(), "Couldn't validate JWT", err
	}

	return userID, "", nil

}
