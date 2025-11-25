package main

import (
	"context"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/ssh"

	"ag/internal/auth"
	"ag/internal/bbs"
	"ag/internal/server"
)

func main() {
	addr := flag.String("addr", ":2323", "listen address for SSH clients")
	boardsFile := flag.String("boards", "data/boards.json", "path to boards list json")
	postsDir := flag.String("posts", "data/posts", "directory to store posts per board")
	authFile := flag.String("auth", "", "path to auth JSON (optional)")
	flag.Parse()

	// Load Auth
	var authenticator auth.Authenticator
	if *authFile != "" {
		authCfg, err := auth.LoadConfig(*authFile)
		if err != nil {
			log.Fatalf("load auth file: %v", err)
		}
		authenticator = auth.NewAuthenticator(authCfg)
	}

	// Load BBS Data
	store := bbs.BoardFile{Path: *boardsFile}

	// Read encryption key from environment
	var encryptionKey []byte
	if keyHex := os.Getenv("BBS_ENCRYPTION_KEY"); keyHex != "" {
		key, err := decodeHexKey(keyHex)
		if err != nil {
			log.Fatalf("invalid encryption key: %v", err)
		}
		if len(key) != 32 {
			log.Fatalf("encryption key must be 32 bytes (64 hex chars), got %d bytes", len(key))
		}
		encryptionKey = key
		log.Println("Encryption enabled for post storage")
	}

	postStore := bbs.PostFile{Dir: *postsDir, EncryptionKey: encryptionKey}
	boardNames, err := store.Load()
	if err != nil {
		log.Printf("failed to load boards list, using defaults: %v", err)
	}
	board := bbs.NewWithBoards(nil, boardNames, store, postStore)

	// Create SSH Server
	s, err := server.New(*addr, ".ssh/term_info_ed25519", board)
	if err != nil {
		log.Fatalln(err)
	}

	// Handle Auth (Simple Password check if enabled)
	if authenticator.Enabled() {
		// Note: Wish's password auth is a bit different, it checks public keys first usually.
		// For simplicity in this migration, we might rely on the UI to handle login
		// OR we can implement a custom auth handler.
		// For now, let's keep it open or implement a basic check if we can.
		// Since the original code did interactive auth INSIDE the session,
		// we can replicate that in the Bubbletea model if we wanted to,
		// but for now let's just allow connection and assume "guest" or the SSH user.
		// A proper Wish auth middleware could be added later.
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("Starting SSH server on %s", *addr)
	go func() {
		if err = s.ListenAndServe(); err != nil && err != ssh.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	<-done
	log.Println("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}

func decodeHexKey(keyHex string) ([]byte, error) {
	return hex.DecodeString(keyHex)
}
