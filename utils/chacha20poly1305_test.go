package utils

import (
	"bytes"
	"strconv"
	"testing"
)

func TestSimpleKey(t *testing.T) {
	InitalizeCrypto()

	jwt, err := NewJWTToken().
		SetClaim("email", "vednoc@usw.local").
		GetSignedString(VerifySigningKey)
	if err != nil {
		t.Error(err)
	}

	sealedText := SealText(jwt)
	unSealedText, err := OpenText(B2s(sealedText))
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(S2b(jwt), unSealedText) {
		t.Error("Originial and Unsealed aren't the same string.")
	}
}

func benchamarkChaCha20Poly1305Seal(b *testing.B, buf []byte) {
	b.ReportAllocs()
	b.SetBytes(int64(len(buf)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SealText(B2s(buf[:]))
	}
}

func benchamarkChaCha20Poly1305Open(b *testing.B, buf []byte) {
	b.ReportAllocs()
	b.SetBytes(int64(len(buf)))

	ct := SealText(B2s(buf[:]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = OpenText(B2s(ct[:]))
	}
}

func benchamarkPrepareText(b *testing.B, buf []byte) {
	b.ReportAllocs()
	b.SetBytes(int64(len(buf)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PrepareText(B2s(buf[:]))
	}
}

func benchamarkDecodePreparedText(b *testing.B, buf []byte) {
	b.ReportAllocs()
	b.SetBytes(int64(len(buf)))

	ct := PrepareText(B2s(buf[:]))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecodePreparedText(ct)
	}
}

func BenchmarkPureChaCha20Poly1305(b *testing.B) {
	InitalizeCrypto()
	b.ResetTimer()
	for _, length := range []int{215, 1350, 8 * 1024} {
		b.Run("Open-"+strconv.Itoa(length)+"-X", func(b *testing.B) {
			benchamarkChaCha20Poly1305Open(b, make([]byte, length))
		})
		b.Run("Seal-"+strconv.Itoa(length)+"-X", func(b *testing.B) {
			benchamarkChaCha20Poly1305Seal(b, make([]byte, length))
		})
	}
}

func BenchmarkPrepareText(b *testing.B) {
	InitalizeCrypto()
	b.ResetTimer()
	for _, length := range []int{215, 1350, 8 * 1024} {
		b.Run("Prepare-"+strconv.Itoa(length), func(b *testing.B) {
			benchamarkPrepareText(b, make([]byte, length))
		})
		b.Run("Decode-"+strconv.Itoa(length), func(b *testing.B) {
			benchamarkDecodePreparedText(b, make([]byte, length))
		})
	}
}