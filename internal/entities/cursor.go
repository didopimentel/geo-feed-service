package entities

import (
	"encoding/base64"
	"encoding/json"
	"errors"
)

func EncodeCursor(c *Cursor) (string, error) {
	if c == nil {
		return "", nil
	}

	raw, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}

	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("invalid cursor encoding")
	}

	var cursor Cursor
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return nil, errors.New("invalid cursor payload")
	}

	if len(cursor.ID) != 16 {
		return nil, errors.New("invalid cursor id")
	}

	return &cursor, nil
}
