package crypt4gh

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/ohsu-comp-bio/funnel/logger"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"
)

const privateKeyMagic = "c4gh-v1"
const presumedDirName = ".c4gh"

var base64Decoder *base64.Encoding = base64.StdEncoding.WithPadding(base64.StdPadding)
var log = logger.NewLogger("crypt4gh", logger.DefaultConfig())

type KeyPair struct {
	publicKey []byte
	secretKey []byte
}

// Produces BASE64-encoded public-key where the key is represented just as in
// the public-key file.
func (k *KeyPair) EncodePublicKeyBase64() string {
	header, footer := getKeyFileHeaderFooter("PUBLIC")

	content := bytes.NewBufferString(header)
	content.WriteString(base64Decoder.EncodeToString(k.publicKey))
	content.WriteRune('\n')
	content.WriteString(footer)

	return base64.StdEncoding.EncodeToString(content.Bytes())
}

// Saves the current key-pair to the specified files. If passphrase is not
// empty, the private key will be encrypted using the passphrase
func (k *KeyPair) Save(publicKeyPath, privateKeyPath string, passphrase []byte) error {
	err := saveKeyFile(publicKeyPath, "PUBLIC", k.publicKey)
	if err != nil {
		return err
	}

	encodedKey, err := encodePrivateKey(k.secretKey, passphrase)
	if err != nil {
		return err
	}

	return saveKeyFile(privateKeyPath, "PRIVATE", encodedKey)
}

// Wraps given reader in order to decrypt the Crypt4gh file stream (expecting
// the header part followed by encrypted body).
func (k *KeyPair) Decrypt(r io.Reader) (io.Reader, error) {
	c := Crypt4gh{keyPair: k, stream: r}
	err := c.readHeader()
	return &c, err
}

// Returns a reader providing decrypted data for given Crypt4gh file stream
// (body) and explicit Crypt4gh header information.
func (k *KeyPair) DecryptWithHeader(header []byte, body io.Reader) (io.Reader, error) {
	c := Crypt4gh{keyPair: k, stream: bytes.NewReader(header)}
	err := c.readHeader()
	c.stream = body // After parsing header, switch to the body reader
	return &c, err
}

// Initiates a completely new key-pair, which is stored only in memory.
func NewKeyPair() (*KeyPair, error) {
	edCurve, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		publicKey: edCurve.PublicKey().Bytes(),
		secretKey: edCurve.Bytes(),
	}, nil
}

// Initiates a key-pair from the provided file-paths, where the encrypted secret
// will be accessed using the provided passphrase. Failure to parse the files,
// or decrypt the secret key will result in errors returned by this method.
// Note that the public key file is optional: when it exists, its content will
// be verified to make sure that it pairs with the secret key. Mismatch of keys
// will also result in an error.
// Also note that when the secret key is not encrypted, the passphrase may be
// nil or, if present, its value will be ignored.
func KeyPairFromFiles(publicKeyPath, secretKeyPath string, passphrase []byte) (*KeyPair, error) {
	sec, err := parseSecretKeyFile(secretKeyPath, passphrase)
	if err != nil {
		return nil, err
	}

	secKey, err := ecdh.X25519().NewPrivateKey(sec)
	if err != nil {
		return nil, err
	}

	pub := secKey.PublicKey().Bytes()

	if isFile(publicKeyPath) {
		pubFromFile, err := parsePublicKeyFile(publicKeyPath)

		if err != nil {
			return nil, err
		} else if !bytes.Equal(pub, pubFromFile) {
			return nil, errors.New("The crypt4gh public key from the file " +
				"does not match the private key")
		}
	}

	return &KeyPair{
		publicKey: pub,
		secretKey: sec,
	}, nil
}

// The most general-purpose way to load Crypt4gh keys from files, or generate
// and save them when the files cannot be resolved.
//
// First, the public and private key file-paths are resolved from environment
// variables: C4GH_PUBLIC_KEY (optional), C4GH_SECRET_KEY, C4GH_PASSPHRASE
// (optional). If C4GH_SECRET_KEY refers to an unencrypted secret key,
// C4GH_PASSPHRASE may be omitted. If C4GH_PUBLIC_KEY is provided and the file
// exists, it must match with the secret key. Also note that when the files of
// C4GH_PUBLIC_KEY and C4GH_SECRET_KEY do not exist yet, a new key-pair will be
// generated and stored in the specified files (secret key will be encrypted
// with C4GH_SECRET_KEY, if present).
//
// When the variables are declared, the local and home directory files will be
// tried instead: .c4gh/key[.pub] and ~/.c4gh/key[.pub]. If these files
// (especially the secret key) do not exist, a new key-pair will be generated
// and stored in the home-directory file-paths, and, on failure, in the local
// directory file-paths. This method returns an error only when it generates
// new keys but cannot save them to resolved paths.
func ResolveKeyPair() (*KeyPair, error) {
	publicKeyPath := os.Getenv("C4GH_PUBLIC_KEY")
	secretKeyPath := os.Getenv("C4GH_SECRET_KEY")
	passphrase := []byte(os.Getenv("C4GH_PASSPHRASE"))

	defaultKeysDir := resolveKeysDir(secretKeyPath)

	// When existing keys cannot be resolved, default to the in-memory
	// generated key-pair:
	if defaultKeysDir == "" {
		return NewKeyPair()
	}

	// When file-paths are missing, set default values and attempt to use them:
	if secretKeyPath == "" {
		secretKeyPath = path.Join(defaultKeysDir, "key")
	}
	if publicKeyPath == "" {
		publicKeyPath = path.Join(defaultKeysDir, "key.pub")
	}

	// Load existing key:
	if isFile(secretKeyPath) {
		return KeyPairFromFiles(publicKeyPath, secretKeyPath, []byte(passphrase))
	}

	// Generate new key-pair and save it to the files:
	keyPair, err := NewKeyPair()
	if err == nil {
		err = keyPair.Save(publicKeyPath, secretKeyPath, passphrase)
	}
	return keyPair, err
}

// Attempts to resolve the directory of the keys.
// On failure, it returns an empty string.
// Look up order is following:
//
//  1. When the provided file-path is not empty, use its directory (even if it
//     does not exist yet: it will be created).
//  2. Fall back to .c4gh/ directory in the current directory, if it exists.
//  3. When user's home-directory can be resolved, fall back to the ~/.c4gh/
//     directory (creating it, if missing). When the home-directory cannot be
//     resolved, fall back to .c4gh/ directory in the current directory.
//  4. When the directory does not exist and cannot be created, fail by
//     returning "".
//
// To summarise the edge-cases:
//  1. Explicitly provided paths will be always trusted (if the directories
//     don't exist yet, they will be created)
//  2. If no explicit path is provided, keys will be created at
//     ~/.c4gh/key[.pub]
//  3. When the current directory contains the .c4gh directory then that will
//     override the home-directory.
func resolveKeysDir(secretKeyPath string) string {
	var keysDir string

	if secretKeyPath != "" { // explicit path
		keysDir = path.Dir(secretKeyPath)
	} else if isDir(presumedDirName) { // ./.c4gh/
		keysDir = presumedDirName
	} else { // attempting ~/.c4gh/
		var errDir error
		keysDir, errDir = os.UserHomeDir()

		if errDir == nil {
			// Place the keys into a private sub-directory:
			keysDir = path.Join(keysDir, presumedDirName)
		} else {
			keysDir = presumedDirName // Fall-back: ./.c4gh/
		}
	}

	// Check the directory exists or if it can be created:
	directoryExists := isDir(keysDir)

	if !directoryExists {
		if err := os.MkdirAll(keysDir, 0700); err != nil {
			return ""
		}
	}

	return keysDir
}

// Reports whether the path exists and refers to a directory.
func isDir(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && fileInfo != nil && fileInfo.IsDir()
}

// Reports whether the path exists and refers to a regular file.
func isFile(path string) bool {
	fileInfo, err := os.Stat(path)
	return err == nil && fileInfo != nil && fileInfo.Mode().IsRegular()
}

func parsePublicKeyFile(path string) ([]byte, error) {
	b, err := readKeyFile(path, "PUBLIC")

	if err == nil && len(b) != 32 {
		err = fmt.Errorf("The decoded public key has non-expected length: %d (expected 32 bytes)", len(b))
	}

	if err != nil {
		err = errors.Join(fmt.Errorf("Failed to parse the public key from file [%s]", path), err)
	}

	return b, err
}

func parseSecretKeyFile(path string, passphrase []byte) ([]byte, error) {
	b, err := readKeyFile(path, "PRIVATE")

	if err != nil {
		return nil, errors.Join(fmt.Errorf("Failed to parse the secret key from file [%s]", path), err)
	}

	return parsePrivateKey(b, passphrase)
}

func readKeyFile(path, keyType string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	expectHeader, expectFooter := getKeyFileHeaderFooter(keyType)

	r := bufio.NewReader(file)

	err = checkLine(r, expectHeader)
	if err != nil {
		return nil, err
	}

	line, _, err := r.ReadLine()
	if err != nil {
		return nil, err
	}

	b, err := base64Decoder.DecodeString(string(line))
	if err != nil {
		return nil, errors.Join(errors.New("Failed to decode the Base64 string in the Crypt4gh key file"), err)
	}

	log.Info(fmt.Sprint("Reading Crypt4gh", keyType, "key from file", path))

	return b, checkLine(r, expectFooter)
}

func saveKeyFile(path, keyType string, data []byte) error {
	log.Info(fmt.Sprint("Saving Crypt4gh", keyType, "key to file", path))
	header, footer := getKeyFileHeaderFooter(keyType)

	content := bytes.NewBufferString(header)
	content.WriteString(base64Decoder.EncodeToString(data))
	content.WriteRune('\n')
	content.WriteString(footer)

	return os.WriteFile(path, content.Bytes(), 0400)
}

func getKeyFileHeaderFooter(keyType string) (string, string) {
	// keyType should be "PUBLIC" or "PRIVATE"
	return "-----BEGIN CRYPT4GH " + keyType + " KEY-----\n",
		"-----END CRYPT4GH " + keyType + " KEY-----\n"
}

// Checks that the next len(line) bytes () match the given "line" string.
// On success, the method returns nil.
func checkLine(r io.Reader, line string) error {
	b := make([]byte, len(line))

	n, err := r.Read(b)
	if err != nil {
		return err
	}

	value := string(b[:n])

	if value != line {
		return fmt.Errorf("Mismatch of expected line: expected [%s] but got [%s]",
			strings.TrimRight(line, "\n"), strings.TrimRight(value, "\n"))
	}

	return nil
}

// Extract a number of bytes from given list. The length of the bytes to be
// returned is specified by the first two bytes (big-endian) at the starting
// position. The second returned int indicates the position after the extracted
// bytes.
func readBytes(bytes []byte, startPos int) ([]byte, int) {
	length := int(bytes[startPos])<<8 | int(bytes[startPos+1])
	start := startPos + 2
	end := start + length
	return bytes[start:end], end
}

func readString(bytes []byte, startPos int) (string, int) {
	b, end := readBytes(bytes, startPos)
	return string(b), end
}

// Returns a two-byte list holding the provided int in big-endian encoding.
func getLengthBytes(l int) []byte {
	b := [2]byte{byte(l >> 8), byte(l)}
	return b[:]
}

func parsePrivateKey(payload, passphrase []byte) ([]byte, error) {
	pos := len(privateKeyMagic)
	magic := string(payload[0:pos])

	if magic != privateKeyMagic {
		return nil, fmt.Errorf("Unexpected magic [%s] (expected: [%s])",
			magic, privateKeyMagic)
	}

	kdfname, pos := readString(payload, pos)

	if kdfname != "none" &&
		kdfname != "bcrypt" &&
		kdfname != "scrypt" &&
		kdfname != "pbkdf2_hmac_sha256" {
		return nil, fmt.Errorf("Unsupported Key Derivation Function [%s] "+
			"(expected [none], [bcrypt], [scrypt], or [pbkdf2_hmac_sha256])", kdfname)
	}

	var roundsSalt []byte
	if kdfname != "none" {
		roundsSalt, pos = readBytes(payload, pos)
	}

	ciphername, pos := readString(payload, pos)

	if ciphername != "none" && ciphername != "chacha20_poly1305" {
		return nil, fmt.Errorf("Unsupported cipher alorithm for the key "+
			"protection [%s] (expected [none] or [chacha20_poly1305])",
			ciphername)
	}

	encryptedKey, _ := readBytes(payload, pos)

	return decryptPrivateKey(encryptedKey, passphrase, kdfname, roundsSalt, ciphername)
}

func decryptPrivateKey(
	encryptedKey, passphrase []byte,
	kdfname string,
	roundsSalt []byte,
	ciphername string,
) ([]byte, error) {

	// With these parameters, the key is actually not encrypted:
	if kdfname == "none" && ciphername == "none" {
		return encryptedKey, nil
	}

	if kdfname == "none" || ciphername == "none" {
		return nil, fmt.Errorf("Invalid key encryption information: "+
			"kdfname=%s, ciphername=%s", kdfname, ciphername)
	}

	if len(passphrase) == 0 {
		return nil, errors.New("The secret key is encrypted but no passphrase was provided")
	}

	rounds := int(roundsSalt[0])<<24 | int(roundsSalt[1])<<16 | int(roundsSalt[2])<<8 | int(roundsSalt[3])
	salt := roundsSalt[4:]
	keySize := chacha20poly1305.KeySize

	var err error

	// Deriving a key from passphrase.
	// Parameters inspired by https://github.com/EGA-archive/crypt4gh/blob/master/crypt4gh/keys/kdf.py
	switch kdfname {
	case "bcrypt":
		passphrase, err = bcrypt.GenerateFromPassword(passphrase, rounds)
	case "scrypt":
		passphrase, err = scrypt.Key(passphrase, salt, 1<<14, 8, 1, keySize)
	case "pbkdf2_hmac_sha256":
		passphrase = pbkdf2.Key(passphrase, salt, rounds, keySize, crypto.SHA256.New)
	default:
		err = errors.New("Given KDF is not supported")
	}

	if err != nil {
		return nil, errors.Join(fmt.Errorf("Failed to derive a key using the "+
			"provided passphrase and function [%s]", kdfname), err)
	}

	// Starting to decrypt the key.
	dataKey, err := chacha20poly1305.New(passphrase)
	if err != nil {
		return nil, err
	}

	nonce := encryptedKey[0:chacha20poly1305.NonceSize]
	ciphertext := encryptedKey[chacha20poly1305.NonceSize:]

	decryptedKey, err := dataKey.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		err = errors.Join(errors.New("Failed to decrypt the secret key (wrong passphrase)"), err)
	}

	return decryptedKey, err
}

func encodePrivateKey(key, passphrase []byte) ([]byte, error) {
	key, kdfname, roundsSalt, ciphername, err := encryptPrivateKey(key, passphrase)
	if err != nil {
		return nil, err
	}

	content := bytes.NewBufferString(privateKeyMagic)

	content.Write(getLengthBytes(len(kdfname)))
	content.Write([]byte(kdfname))

	if kdfname != "none" {
		content.Write(getLengthBytes(len(roundsSalt)))
		content.Write(roundsSalt)
	}

	content.Write(getLengthBytes(len(ciphername)))
	content.Write([]byte(ciphername))

	content.Write(getLengthBytes(len(key)))
	content.Write(key)

	return content.Bytes(), nil
}

func encryptPrivateKey(key, passphrase []byte) (
	encryptedKey []byte,
	kdfname string,
	roundsSalt []byte,
	ciphername string,
	err error,
) {
	if len(passphrase) == 0 {
		return key, "none", nil, "none", nil
	}

	salt := make([]byte, 16)
	if _, err = rand.Reader.Read(salt); err != nil {
		return nil, "", nil, "", err
	}

	// Derive a key from the passphrase:
	derivedKey, err := scrypt.Key(passphrase, salt, 1<<14, 8, 1, chacha20poly1305.KeySize)
	if err != nil {
		return nil, "", nil, "", err
	}

	// Initialise chacha20poly1305 from derived key for symmetric encryption:
	aead, err := chacha20poly1305.New(derivedKey)
	if err != nil {
		return nil, "", nil, "", err
	}

	// Initialise encrypted key with a random nonce, and leave capacity for the ciphertext.
	encryptedKey = make([]byte, aead.NonceSize(), aead.NonceSize()+len(key)+aead.Overhead())
	if _, err := rand.Read(encryptedKey); err != nil {
		return nil, "", nil, "", err
	}

	// Encrypt the key:
	encryptedKey = aead.Seal(encryptedKey, encryptedKey, key, nil)

	// Prepare rounds+salt byte-array for the private key file:
	roundsSalt = make([]byte, 4+len(salt))
	copy(roundsSalt[4:], salt) // First 4 bytes ("rounds") remain all zeros.

	return encryptedKey, "scrypt", roundsSalt, "chacha20_poly1305", nil
}
