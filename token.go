/**
  Copyright (c) 2022 Zander Schwid & Co. LLC. All rights reserved.
*/

package main

import (
	"crypto/rand"
	"crypto/sha256"
	"github.com/schwid/base62"
	"io"
)

func generateToken() (string, error) {
	nonce := make([]byte, 256 / 8)
	if _, err := io.ReadFull(rand.Reader, nonce); err == nil {
		return base62.StdEncoding.EncodeToString(nonce), nil
	} else {
		return "", err
	}
}

func hashToken(token string) (string, error) {

	bin, err := base62.StdEncoding.DecodeString(token)
	if err != nil {
		return "", err
	}

	h := sha256.New()
	h.Write(bin)
	s := h.Sum(nil)

	return base62.StdEncoding.EncodeToString(s), nil
}
