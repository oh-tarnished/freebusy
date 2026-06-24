// Package code mints the human-entered code string for a promo code created
// with CodeGeneration AUTO — a random alphanumeric suffix (e.g. "K7QM4P2X"). It
// is a pure, side-effect-free helper so the rules can be unit-tested without a
// server or database; the random source is injectable for deterministic tests.
package code

import (
	"crypto/rand"
	"io"
)

const (
	// defaultLength is the number of random characters in a generated code.
	defaultLength = 8
	// alphanumeric is the alphabet auto-generated codes draw from. It omits the
	// visually ambiguous characters (O/0, I/1) so codes read back cleanly.
	alphanumeric = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
)

// Generate returns a fresh random promo code using a cryptographically-secure
// random source. Used when CreatePromoCode is called with CODE_GENERATION_AUTO.
func Generate() string {
	return generate(rand.Reader)
}

// generate is Generate with an injectable random source, for tests.
func generate(r io.Reader) string {
	return randomString(r, alphanumeric, defaultLength)
}

// randomString returns n characters drawn from alphabet using r as the entropy
// source. crypto/rand.Reader never fails in practice; on a read error the bytes
// stay zero so a full-length code is still produced.
func randomString(r io.Reader, alphabet string, n int) string {
	buf := make([]byte, n)
	_, _ = io.ReadFull(r, buf)
	out := make([]byte, n)
	for i, b := range buf {
		out[i] = alphabet[int(b)%len(alphabet)]
	}
	return string(out)
}
