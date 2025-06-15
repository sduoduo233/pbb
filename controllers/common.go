package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type D map[string]any

func If[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
func randomToken() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
