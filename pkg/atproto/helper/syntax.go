package helper

import (
	"github.com/bluesky-social/indigo/atproto/syntax"
)

func GenerateTID() string {
	return syntax.NewTIDNow(0).String()
}

func BuildAtURI(uri string) (syntax.ATURI, error) {
	aturi, err := syntax.ParseATURI(uri)
	if err != nil {
		return "", err
	}
	return aturi, nil
}
