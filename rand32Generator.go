package main

import (
	"crypto/rand"
)

func random32Generator() ([]byte, error) {
	number := make([]byte, 32)

	_, err := rand.Read(number)

	if err != nil {

		return make([]byte, 0), err
	}

	return number, nil
}
