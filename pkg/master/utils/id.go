package utils

import "github.com/google/uuid"

func GenerateMasterId() string {
	return uuid.New().String()
}