package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/argon2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <password>\n", os.Args[0])
		os.Exit(1)
	}

	password := os.Args[1]

	// Generate random 16-byte salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating salt: %v\n", err)
		os.Exit(1)
	}

	// Parameters that work with version 1.2.2
	// These match the working hash format: $argon2i$v=19$m=4096,t=3,p=1$salt$hash
	memory := uint32(4096) // 4 MB in KB (matches working version)
	time := uint32(3)      // 3 iterations
	threads := uint8(1)    // 1 thread (matches working version)
	keyLen := uint32(32)   // 32 bytes output

	// Generate Argon2i hash (matches working version)
	hash := argon2.Key([]byte(password), salt, time, memory, threads, keyLen)

	// Format: $argon2i$v=19$m=4096,t=3,p=1$salt$hash
	encodedHash := fmt.Sprintf("$argon2i$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		memory,
		time,
		threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	fmt.Println(encodedHash)
}
