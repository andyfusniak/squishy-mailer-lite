package secrets_test

import (
	"testing"

	"github.com/andyfusniak/squishy-mailer-lite/internal/secrets"
	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("abcdefghijklmnop")
	mgr, err := secrets.New(secrets.AESGCMWithRandomNonce, key)
	assert.NoError(t, err)

	tests := []struct {
		ptPassword string
	}{
		{ptPassword: "secret1"},
		{ptPassword: "secret2"},
		{ptPassword: "secret3"},
		{ptPassword: "secret4"},
		{ptPassword: "secret5"},
		{ptPassword: "secret6"},
		{ptPassword: "secret7"},
		{ptPassword: "secret8"},
		{ptPassword: "secret9"},
		{ptPassword: "secret10"},
	}

	for _, tt := range tests {
		nonce, ciphertext, err := mgr.Encrypt([]byte(tt.ptPassword))
		assert.NoError(t, err)

		t.Logf("nonce:\t\t%x", nonce)
		t.Logf("ciphertext:\t%x", ciphertext)

		plaintext, err := mgr.Decrypt(nonce, ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, []byte(tt.ptPassword), plaintext)

		t.Logf("plaintext:\t%s", plaintext)
	}
}

func TestEncryptHexEncode(t *testing.T) {
	key := []byte("abcdefghijklmnop")
	mgr, err := secrets.New(secrets.AESGCMWithRandomNonce, key)
	assert.NoError(t, err)

	tests := []struct {
		ptPassword string
	}{
		{ptPassword: "secret1"},
		{ptPassword: "secret2"},
		{ptPassword: "secret3"},
		{ptPassword: "secret4"},
		{ptPassword: "secret5"},
		{ptPassword: "secret6"},
		{ptPassword: "secret7"},
		{ptPassword: "secret8"},
		{ptPassword: "secret9"},
		{ptPassword: "secret10"},
	}

	for _, tt := range tests {
		nonce, ciphertext, err := mgr.EncryptHexEncode(tt.ptPassword)
		assert.NoError(t, err)

		t.Logf("nonce:\t\t%s", nonce)
		t.Logf("ciphertext:\t%s", ciphertext)

		plaintext, err := mgr.HexDecodeDecrypt(nonce, ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, tt.ptPassword, plaintext)

		t.Logf("plaintext:\t%s", plaintext)
	}
}
