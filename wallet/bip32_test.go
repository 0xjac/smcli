package wallet

import (
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/hex"
	"github.com/spacemeshos/smkeys/bip32"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdg-go/pbkdf2"
)

var goodSeed = []byte("abandon abandon abandon abandon abandon abandon abandon abandon ")

func generateTestMasterKeyPair() (*EDKeyPair, error) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	password := "Winning together!"
	seed := pbkdf2.Key([]byte(mnemonic), []byte("mnemonic"+password), 2048, 64, sha512.New)
	return NewMasterKeyPair(seed)
}

func TestNewMasterBIP32EDKeyPair(t *testing.T) {
	// first iteration
	keyPair1, err := generateTestMasterKeyPair()
	require.NoError(t, err)
	require.NotEmpty(t, keyPair1)

	// second iteration
	keyPair2, err := generateTestMasterKeyPair()
	require.NoError(t, err)
	require.NotEmpty(t, keyPair2)

	msg := []byte("master test")

	// Testing the private key signature generated from the first iteration and verifying with public key from the second iteration
	sig1 := ed25519.Sign(ed25519.PrivateKey(keyPair1.Private), msg)
	valid1 := ed25519.Verify(ed25519.PublicKey(keyPair2.Public), msg, sig1)
	require.True(t, valid1)

	// Same test with swapped private and public key
	sig2 := ed25519.Sign(ed25519.PrivateKey(keyPair2.Private), msg)
	valid2 := ed25519.Verify(ed25519.PublicKey(keyPair1.Public), msg, sig2)
	require.True(t, valid2)
}

func TestNonHardenedPath(t *testing.T) {
	path1Str := "m/44'/540'/0'/0'/0'"
	path1Hd, err := StringToHDPath(path1Str)
	require.NoError(t, err)
	require.True(t, IsPathCompletelyHardened(path1Hd))

	path2Str := "m/44'/540'/0'/0'/0"
	path2Hd, err := StringToHDPath(path2Str)
	require.NoError(t, err)
	require.False(t, IsPathCompletelyHardened(path2Hd))
}

// Test that path string produces expected path and vice-versa
func TestPath(t *testing.T) {
	path1Str := "m/44'/540'/0'/0'/0'"
	path1Hd, err := StringToHDPath(path1Str)
	require.NoError(t, err)
	path2Hd := DefaultPath().Extend(BIP44HardenedAccountIndex(0))
	path2Str := HDPathToString(path2Hd)
	require.Equal(t, path1Str, path2Str)
	require.Equal(t, path1Hd, path2Hd)
}

// Test deriving a child keypair
func TestChildKeyPair(t *testing.T) {
	path := DefaultPath().Extend(BIP44HardenedAccountIndex(0))

	// generate first keypair
	masterKeyPair, err := NewMasterKeyPair(goodSeed)
	require.NoError(t, err)
	childKeyPair1, err := masterKeyPair.NewChildKeyPair(goodSeed, 0)
	require.Equal(t, path, childKeyPair1.Path)
	require.Equal(t, ed25519.PrivateKeySize, len(childKeyPair1.Private))
	require.Equal(t, ed25519.PublicKeySize, len(childKeyPair1.Public))
	require.NoError(t, err)
	require.NotEmpty(t, childKeyPair1)

	// test signing
	msg := []byte("child test")
	sig := ed25519.Sign(ed25519.PrivateKey(childKeyPair1.Private), msg)
	valid := ed25519.Verify(ed25519.PublicKey(childKeyPair1.Public), msg, sig)
	require.True(t, valid)

	// generate second keypair and check lengths
	childKeyPair2, err := bip32.Derive(HDPathToString(path), goodSeed)
	require.NoError(t, err)
	require.Equal(t, ed25519.PrivateKeySize, len(childKeyPair2))
	privkey2 := PrivateKey(childKeyPair2[:])
	require.Equal(t, ed25519.PrivateKeySize, len(privkey2))
	edpubkey2 := ed25519.PrivateKey(privkey2).Public().(ed25519.PublicKey)
	require.Equal(t, ed25519.PublicKeySize, len(edpubkey2))
	pubkey2 := PublicKey(edpubkey2)
	require.Equal(t, ed25519.PublicKeySize, len(pubkey2))

	// make sure they agree
	require.Equal(t, "feae6977b42bf3441d04314d09c72c5d6f2d1cb4bf94834680785b819f8738dd", hex.EncodeToString(pubkey2))
	require.Equal(t, hex.EncodeToString(childKeyPair1.Public), hex.EncodeToString(pubkey2))
	require.Equal(t, "05fe9affa5562ca833faf3803ce5f6f7615d3c37c4a27903492027f6853e486dfeae6977b42bf3441d04314d09c72c5d6f2d1cb4bf94834680785b819f8738dd", hex.EncodeToString(privkey2))
	require.Equal(t, hex.EncodeToString(childKeyPair1.Private), hex.EncodeToString(privkey2))
}
