package bbs

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const postFileVersion = 1

// PostFile persists posts per board in JSON files.
// File format:
//
//	{
//	  "version": 1,
//	  "board": "general",
//	  "posts": [ { ... Post fields ... } ]
//	}
type PostFile struct {
	Dir           string
	EncryptionKey []byte // 32 bytes for AES-256
}

func (f PostFile) path(board string) string {
	return filepath.Join(f.Dir, board+".json")
}

func (f PostFile) Load(board string) ([]Post, error) {
	if f.Dir == "" {
		return nil, nil
	}
	path := f.path(board)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read posts file: %w", err)
	}

	// Try to decrypt if key is present
	if len(f.EncryptionKey) > 0 {
		decrypted, err := f.decrypt(data)
		if err == nil {
			data = decrypted
		}
		// If decryption fails, we assume it might be plain JSON (migration scenario)
		// or it's just broken. We try to unmarshal whatever we have.
	}

	var wrapper struct {
		Version int    `json:"version"`
		Board   string `json:"board"`
		Posts   []Post `json:"posts"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("parse posts file: %w", err)
	}
	return wrapper.Posts, nil
}

func (f PostFile) Save(board string, posts []Post) error {
	if f.Dir == "" {
		return nil
	}
	if err := os.MkdirAll(f.Dir, 0o755); err != nil {
		return fmt.Errorf("make posts dir: %w", err)
	}
	path := f.path(board)
	tmp := path + ".tmp"
	payload := struct {
		Version int    `json:"version"`
		Board   string `json:"board"`
		Posts   []Post `json:"posts"`
	}{
		Version: postFileVersion,
		Board:   board,
		Posts:   posts,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal posts: %w", err)
	}

	// Encrypt if key is present
	if len(f.EncryptionKey) > 0 {
		encrypted, err := f.encrypt(data)
		if err != nil {
			return fmt.Errorf("encrypt posts: %w", err)
		}
		data = encrypted
	}

	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write posts temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename posts file: %w", err)
	}
	return nil
}

func (f PostFile) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(f.EncryptionKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func (f PostFile) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(f.EncryptionKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
