package utils

import "github.com/google/uuid"

func GenerateNodeId() string {
	return uuid.New().String()
}

func GenerateChunkId() string {
	return uuid.New().String()
}