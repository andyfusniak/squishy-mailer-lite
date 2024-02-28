package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// Manager secret manager.
type Manager struct {
	mode Mode
	key  []byte // private key
}

// Mode type.
type Mode int

const (
	// AESGCMWithRandomNonce encryption and decryption scheme.
	AESGCMWithRandomNonce = iota
)

// New creates a new secret manger.
func New(m Mode, key []byte) (*Manager, error) {
	if m != AESGCMWithRandomNonce {
		return nil, fmt.Errorf(
			"AESGCMWithRandomNonce is currently the only supported mode of operation")
	}
	if len(key) != 16 {
		return nil, fmt.Errorf("secret manager key must be 16 bytes in length")
	}
	return &Manager{
		mode: m,
		key:  key,
	}, nil
}

// Encrypt accepts the plaintext password and returns a random IV with
// the encrypted ciphertext. The IV should be stored alongside the
func (m *Manager) Encrypt(plaintext []byte) (nonce, ciphertext []byte, err error) {
	// TODO: find out if it is safe to move the NewCipher and NewGCM
	// to the Manager.
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return nil, nil, err
	}

	// nonce (96 bits) (32 bits reserved for the counter)
	nonce = make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}

	// GCM Mode (not constant-time)
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	// encypt
	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)

	return nonce, ciphertext, nil
}

// EncryptHexEncode performs a Encrypt and hex encodes the resulting nonce and ciphertext.
func (m *Manager) EncryptHexEncode(plaintext string) (nonce, ciphertext string, err error) {
	n, c, err := m.Encrypt([]byte(plaintext))
	if err != nil {
		return "", "", err
	}

	ndst := make([]byte, hex.EncodedLen(len(n)))
	_ = hex.Encode(ndst, n)

	cdst := make([]byte, hex.EncodedLen(len(c)))
	_ = hex.Encode(cdst, c)

	return string(ndst), string(cdst), nil
}

// Decrypt accepts a nonce and ciphertext pair and returns the unencrypted plaintext.
func (m *Manager) Decrypt(nonce, ciphertext []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return nil, err
	}

	// GCM Mode (not constant-time)
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// decrypt
	plaintext, err = aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// HexDecodeDecrypt first decodes the nonceHex and ciphertextHex to []byte before calling Decrypt.
func (m *Manager) HexDecodeDecrypt(nonceHex, ciphertextHex string) (plaintext string, err error) {
	nonce, err := hex.DecodeString(nonceHex)
	if err != nil {
		return "", err
	}

	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", err
	}

	plaintextbs, err := m.Decrypt(nonce, ciphertext)
	if err != nil {
		return "", err
	}

	return string(plaintextbs), nil
}
