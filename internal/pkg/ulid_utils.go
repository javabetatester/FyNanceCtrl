package pkg

import (
	"errors"
	"strconv"
	"time"

	"github.com/oklog/ulid/v2"
)

func GenerateULID() string {
	entropy := ulid.DefaultEntropy()
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func GenerateULIDObject() ulid.ULID {
	entropy := ulid.DefaultEntropy()
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
}

func ParseULID(ulidStr string) (ulid.ULID, error) {
	if ulidStr == "" {
		return ulid.ULID{}, errors.New("ULID string cannot be empty")
	}

	parsedULID, err := ulid.Parse(ulidStr)
	if err != nil {
		return ulid.ULID{}, errors.New("invalid ULID format")
	}

	return parsedULID, nil
}

func ULIDToString(id ulid.ULID) string {
	return id.String()
}
func IsValidULID(ulidStr string) bool {
	_, err := ulid.Parse(ulidStr)
	return err == nil
}

func IsEmptyULID(id ulid.ULID) bool {
	return id == ulid.ULID{}
}

func SetTimestamps() time.Time {
	return time.Now()
}

func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
