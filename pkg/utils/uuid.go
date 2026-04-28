package utils

import (
	"github.com/google/uuid"
)

func GenerateV7UID() string {

	id, _ := uuid.NewV7()
	return id.String()
}
