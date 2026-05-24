package security

import (
	"crypto/sha256"
	"github.com/google/uuid"
)


// nvrNamespace is a custom, static UUIDv4 used as the mathematical root for our NVR.
// DO NOT change this once your system is in production, or all future IDs will shift.
var nvrNamespace = uuid.MustParse("7e297cc6-32fa-4b98-a1bf-ab0a822b7b9b")

func HashUUID(raw string) uuid.UUID {
	return uuid.NewHash(sha256.New(), nvrNamespace, []byte(raw), 5)
}
