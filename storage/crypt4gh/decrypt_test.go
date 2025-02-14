package crypt4gh

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"io"
	"os"
	"path"
	"testing"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

// Test plain-text content for the test-cases
const content = `First line
Second line
Third line
`

// Holds the key-pair for encrypting/decrypting the content
var c4gh, _ = KeyPairFromFiles("testdata/key.pub", "testdata/key.plain.sec", nil)

// Verifies the IsCrypt4ghFile() function.
func TestIsCrypt4ghFile(t *testing.T) {
	tempDir := t.TempDir()
	correctFile := path.Join(tempDir, "correct.c4gh")

	if err := os.WriteFile(correctFile, encryptContent(0, -1).Bytes(), 0600); err != nil {
		t.Error("Failed to create a test-file")
	}

	if !IsCrypt4ghFile(correctFile) {
		t.Error("Failed to recognise the correct Crypt4gh file as valid")
	}

	if IsCrypt4ghFile(path.Join(tempDir, "unknown.c4gh")) {
		t.Error("Failed to recognise a non-existent Crypt4gh file as invalid")
	}
}

// Encrypts test-data in memory (as it would be in a Crypt4gh file) and then decrypts the content.
// No edit-list used.
func TestDecryptFullText(t *testing.T) {
	decrypted, err := encryptAndDecryptContent(0, -1)

	if err != nil {
		t.Error("Failed to parse the encrypted content", err)
	} else if decrypted != content {
		t.Errorf("Decrypted content does not match the original one (got: %v)", decrypted)
	}
}

// Encrypts test-data in memory together with edit-list (as it would be in a
// Crypt4gh file) for requesting only the second line to be rendered, and then
// decrypts the content (expecting only the second line).
func TestDecryptWithDataEditList(t *testing.T) {
	decrypted, err := encryptAndDecryptContent(11, 11)

	if err != nil {
		t.Error("Failed to parse the encrypted content", err)
	} else if decrypted != "Second line" {
		t.Errorf("Decrypted content does not match the second line (got: %v)", decrypted)
	}
}

func encryptAndDecryptContent(rangeStart, rangeLength int) (string, error) {
	reader, err := c4gh.Decrypt(encryptContent(rangeStart, rangeLength))
	if err != nil {
		return "", err
	}

	buffer := new(bytes.Buffer)
	_, _ = io.Copy(buffer, reader)
	return buffer.String(), nil
}

// Very simplified approach for encrypting some content. Good enough for testing.
// When range is provided, an edit-list packet is added to the header so that
// receiver would look for the part of content defined by the start position
// and length. Specify start=0 and length=-1 to avoid the edit-list.
// Returns the Cryp4gh formatted encrypted data with header in the buffer.
func encryptContent(rangeStart, rangeLength int) *bytes.Buffer {
	sharedKey := generateSharedKey()
	aead, _ := chacha20poly1305.New(sharedKey)

	buffer := new(bytes.Buffer)      // Stores Crypt4gh v1 formatted encrypted content
	buffer.WriteString("crypt4gh")   // "magic" text
	buffer.Write([]byte{1, 0, 0, 0}) // Version "1"

	if rangeStart == 0 && rangeLength < 1 {
		buffer.Write([]byte{1, 0, 0, 0}) // Packet count == 1
	} else {
		buffer.Write([]byte{2, 0, 0, 0}) // Packet count == 2
	}

	// Writes a header packet for storing the key
	encryptionPacket := [40]byte{}
	copy(encryptionPacket[8:40], sharedKey) // key
	writeHeaderPacket(&aead, buffer, encryptionPacket[:])

	// Writes a header packet for storing a data-edit list
	if rangeStart >= 0 && rangeStart < len(content) && rangeLength > 0 {
		afterRange := len(content) - rangeStart - rangeLength
		dataEditListPacket := make([]byte, 32)
		dataEditListPacket[0] = 1                                                     // specifies that packetType = data-edit list
		dataEditListPacket[4] = 3                                                     // specifies that number of following lengths
		binary.LittleEndian.PutUint64(dataEditListPacket[8:16], uint64(rangeStart))   // Skip this number of bytes
		binary.LittleEndian.PutUint64(dataEditListPacket[16:24], uint64(rangeLength)) // Keep this number of bytes
		binary.LittleEndian.PutUint64(dataEditListPacket[24:32], uint64(afterRange))  // Skip this number of bytes

		writeHeaderPacket(&aead, buffer, dataEditListPacket)
	}

	// Writes encrypted content
	nonce := nonce()
	buffer.Write(nonce)                                       // nonce (12 bytes) of the encrypted payload
	buffer.Write(aead.Seal(nil, nonce, []byte(content), nil)) // encrypted payload
	return buffer
}

func writeHeaderPacket(aead *cipher.AEAD, buffer *bytes.Buffer, payload []byte) {
	nonce := nonce()
	encrypted := (*aead).Seal(nil, nonce, payload, nil)

	buffer.Write(binary.LittleEndian.AppendUint32(nil, uint32(52+len(encrypted)))) // packet length
	buffer.Write([]byte{0, 0, 0, 0})                                               // packet encryption method (0)
	buffer.Write(c4gh.publicKey[:32])                                              // writer's public key (32 bytes)
	buffer.Write(nonce)                                                            // nonce (12 bytes) of the encrypted payload
	buffer.Write(encrypted)                                                        // encrypted payload
}

func generateSharedKey() []byte {
	diffieHellmanKey, _ := curve25519.X25519(c4gh.secretKey, c4gh.publicKey)
	diffieHellmanKey = append(diffieHellmanKey, c4gh.publicKey...) // reader's
	diffieHellmanKey = append(diffieHellmanKey, c4gh.publicKey...) // writer's
	hash := blake2b.Sum512(diffieHellmanKey)
	return hash[:chacha20poly1305.KeySize]
}

func nonce() []byte {
	nonce := make([]byte, 12)
	_, _ = rand.Read(nonce)
	return nonce
}
