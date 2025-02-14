package crypt4gh

import (
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
)

// Handles Crypt4gh decryption context per source stream.
type Crypt4gh struct {
	keyPair               *KeyPair
	stream                io.Reader
	headerPacketCount     uint32
	headerPacketProcessed uint32
	dataKeys              []cipher.AEAD
	dataBlock             []byte
	dataBlockPos          int
	dataBlockCount        int
	editListLengths       []uint64
	editListSkip          bool
}

// Reads the magic-number and the version number at the beginning of the file
// to check if the file might be considered to be a supported Crypt4gh file.
func IsCrypt4ghFile(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}

	defer file.Close()

	buffer := make([]byte, 12)
	_, err = file.Read(buffer)

	return err == nil &&
		string(buffer[0:8]) == "crypt4gh" &&
		readInt32(buffer[8:12]) == 1
}

func (c *Crypt4gh) Read(buffer []byte) (n int, err error) {
	addedCount := 0

	for n < len(buffer) && err == nil {
		addedCount, err = c.copyTo(buffer[n:])
		n += addedCount
	}

	if n > 0 {
		return n, nil
	}

	return n, err
}

func (c *Crypt4gh) copyTo(buffer []byte) (int, error) {
	err := c.loadDataBlock()
	if err != nil {
		return 0, err
	}

	start, end, amount := c.getAvailableRange(len(buffer))

	copy(buffer, c.dataBlock[start:end])

	c.dataBlockPos = end

	return amount, nil
}

func (c *Crypt4gh) loadDataBlock() error {
	c.applyEditListSkip()

	// Load next unprocessed block, and skip bytes defined in the edit-list:
	for c.dataBlockPos >= len(c.dataBlock) {
		err := c.decryptDataBlock()
		if err != nil {
			return err
		}
		c.applyEditListSkip()
	}
	return nil
}

func (c *Crypt4gh) readHeader() error {
	err := c.checkMagicNumber()

	if err == nil {
		err = c.checkVersion()
	}

	if err == nil {
		err = c.storeHeaderPacketCount()
	}

	if err == nil {
		for i := uint32(0); i < c.headerPacketCount; i++ {
			err = c.readHeaderPacket()
			if err != nil {
				break
			}
		}
	}

	if err == nil {
		err = c.checkAtLeastOneDataKey()
	}

	return err
}

func (c *Crypt4gh) checkMagicNumber() error {
	magicNumber, errRead := c.readBytes(8)

	if string(magicNumber) == "crypt4gh" {
		return nil
	}

	errMagic := errors.New("Not a Crypt4gh file (missing/wrong magic number)")
	return errors.Join(errMagic, errRead)
}

func (c *Crypt4gh) checkVersion() error {
	version, errRead := c.readInt32()

	if version == 1 {
		return nil
	}

	errVersion := fmt.Errorf("Crypt4gh file version (%d) not supported", version)
	return errors.Join(errVersion, errRead)
}

func (c *Crypt4gh) checkAtLeastOneDataKey() error {
	if len(c.dataKeys) == 0 {
		return fmt.Errorf("The Crypt4gh file is not shared to the provided "+
			"key-pair (scanned %d header packets)", c.headerPacketProcessed)
	}
	return nil
}

func (c *Crypt4gh) storeHeaderPacketCount() error {
	var errRead error
	c.headerPacketCount, errRead = c.readInt32()
	return errRead
}

func (c *Crypt4gh) readHeaderPacket() error {
	c.headerPacketProcessed += 1

	packetLength, errRead := c.readInt32()
	if errRead != nil {
		return errRead
	}

	headerEncryptionMethod, errRead := c.readInt32()
	if errRead != nil {
		return errRead
	}

	if headerEncryptionMethod != 0 {
		log.Warn(fmt.Sprintf("Unrecognized header packet encryption method "+
			"value (%d). Only 0 (X25519_chacha20_ietf_poly1305) is supported. "+
			"Skipping the header packet.", headerEncryptionMethod))
		return nil
	}

	writerPublicKey, errRead := c.readBytes(32)
	if errRead != nil {
		return errRead
	}

	nonce, errRead := c.readBytes(12)
	if errRead != nil {
		return errRead
	}

	// Subtracting the length of previously read items to get the length:
	remainingLength := uint(packetLength) - 52
	encryptedPayloadWithMac, errRead := c.readBytes(remainingLength)

	payload := c.decryptPacketPayload(encryptedPayloadWithMac, writerPublicKey, nonce)
	if len(payload) > 0 {
		c.parseHeaderPayload(payload)
	}

	return errRead
}

func (c *Crypt4gh) parseHeaderPayload(payload []byte) {
	packetType := readInt32(payload[0:4])
	dataEncryptionParameters := packetType == 0
	dataEditList := packetType == 1

	if dataEncryptionParameters {
		if len(payload) != 40 {
			c.warnPacket("payload with data encryption parameters has a "+
				"non-expected length [%d] (expected: 40).", len(payload))
		}

		dataEncryptionMethod := readInt32(payload[4:8])

		if dataEncryptionMethod != 0 {
			c.warnPacket("specifies an unsupported data encryption method "+
				"[%d] while the only supported method is "+
				"[chacha20_ietf_poly1305 = 0].", dataEncryptionMethod)
			return
		}

		if dataKey, errKey := chacha20poly1305.New(payload[8:40]); errKey != nil {
			c.warnPacket("ChaCha20-IETF-Poly1305 data-key error: %v", errKey)
		} else {
			c.dataKeys = append(c.dataKeys, dataKey)
			log.Info("Successfully received data encryption keys from the " +
				"file header")
		}

	} else if dataEditList {
		numberLengths := readInt32(payload[4:8])
		expectedLength := 8 + 8*int(numberLengths)

		if len(payload) != expectedLength {
			c.warnPacket("payload with data edit list has a non expected "+
				"length [%d] (expected: [%d]).", len(payload), expectedLength)
		}

		if len(c.editListLengths) > 0 {
			c.warnPacket("supplies another edit-list (only one permitted)")
			return
		}

		// Read and store the lengths of the edit list
		for startPos := 8; numberLengths > 0; numberLengths-- {
			c.editListLengths = append(c.editListLengths,
				readInt64(payload[startPos:startPos+8]))
			startPos += 8
		}

		log.Info(fmt.Sprintf("Header defines an edit-list: %v", c.editListLengths))

		// The first length is about skipping a number of bytes
		c.editListSkip = true

	} else {
		c.warnPacket("specifies an unsupported packet type [%d] while only "+
			"[data_encryption_parameters = 0] and [data_edit_list = 1] are "+
			"supported.", packetType)
	}
}

func (c *Crypt4gh) decryptPacketPayload(encryptedPayloadWithMac, writerPublicKey, nonce []byte) []byte {
	curve, errCurve := curve25519.X25519(c.keyPair.secretKey, writerPublicKey)
	if errCurve != nil {
		c.warnPacket("curve25519 error: %v", errCurve)
		return nil
	}

	length1 := len(curve)
	length2 := length1 + len(c.keyPair.publicKey)
	length3 := length2 + len(writerPublicKey)

	keys := make([]byte, length3)

	copy(keys, curve)
	copy(keys[length1:], c.keyPair.publicKey)
	copy(keys[length2:], writerPublicKey)

	sharedKey := blake2b.Sum512(keys)

	aead, errKey := chacha20poly1305.New(sharedKey[:32])
	if errKey != nil {
		c.warnPacket("ChaCha20-IETF-Poly1305 shared-key error: %v", errKey)
		return nil
	}

	plaintext, errOpen := aead.Open(nil, nonce, encryptedPayloadWithMac, nil)

	if errOpen != nil {
		c.warnPacket("ChaCha20-IETF-Poly1305 deciphering error : %v", errOpen)
		return nil // This error and the packet payload must be ignored
	}

	return plaintext
}

func (c *Crypt4gh) decryptDataBlock() error {
	c.dataBlockCount++

	block, err := c.readBytesMax(65564)
	if err != nil {
		return err
	}

	for i := range c.dataKeys {
		key := c.dataKeys[i]

		nonce := block[0:key.NonceSize()]
		encryptedDataWithMac := block[key.NonceSize():]

		plaintext, errOpen := key.Open(nil, nonce, encryptedDataWithMac, nil)

		if errOpen == nil {
			c.dataBlock = plaintext
			c.dataBlockPos = 0
			log.Debug("Successfully decrypted a data block",
				"data_block_number", c.dataBlockCount,
			)
			return nil
		} else {
			log.Warn("Failed to decrypt a data block with a key",
				"data_block_number", c.dataBlockCount,
				"tried_key_number", i+1,
				"keys_count", len(c.dataKeys),
			)
		}
	}

	log.Warn("Failed to decrypt a data block (tried all keys)",
		"data_block_number", c.dataBlockCount,
		"keys_count", len(c.dataKeys),
	)

	return nil
}

func (c *Crypt4gh) applyEditListSkip() {
	if !c.editListSkip || len(c.editListLengths) == 0 {
		return
	}

	remainingAmount := uint64(len(c.dataBlock) - c.dataBlockPos)
	skipAmount := &c.editListLengths[0]

	if remainingAmount == 0 {
		return
	}

	if *skipAmount <= remainingAmount {
		c.dataBlockPos += int(*skipAmount)
		*skipAmount = 0
	} else {
		c.dataBlockPos += int(remainingAmount)
		*skipAmount -= remainingAmount
	}

	if *skipAmount == 0 {
		c.editListLengths = c.editListLengths[1:]
		c.editListSkip = false
	}
}

func (c *Crypt4gh) getAvailableRange(amount int) (start, end, providedAmount int) {
	providedAmount = amount

	if c.dataBlockPos+providedAmount > len(c.dataBlock) {
		providedAmount = len(c.dataBlock) - c.dataBlockPos
	}

	// apply Edit-List:
	if len(c.editListLengths) > 0 && !c.editListSkip {
		keepAmount := &c.editListLengths[0]

		// Reduce the available amount of bytes to read, if necessary
		if *keepAmount < uint64(providedAmount) {
			providedAmount = int(*keepAmount)
		}

		// Reduce the amount of bytes to keep (in the edit list):
		*keepAmount -= uint64(providedAmount)

		// Switch to edit-list skip-mode, once keep-mode is exhausted:
		if *keepAmount == 0 {
			c.editListLengths = c.editListLengths[1:]
			c.editListSkip = true
		}
	}

	return c.dataBlockPos, c.dataBlockPos + providedAmount, providedAmount
}

func (c *Crypt4gh) warnPacket(msg string, args ...any) {
	log.Warn(fmt.Sprintf("Header packet [%d/%d] %s\n",
		c.headerPacketProcessed,
		c.headerPacketCount,
		fmt.Sprintf(msg, args...)))
}

func (c *Crypt4gh) readBytes(count uint) ([]byte, error) {
	b, err := c.readBytesMax(count)

	if err == nil && count != uint(len(b)) {
		err = fmt.Errorf("Could not read the entire value (got %d out of %d bytes)", len(b), count)
	}

	return b, err
}

func (c *Crypt4gh) readBytesMax(count uint) ([]byte, error) {
	b := make([]byte, count)
	actualCount, err := c.stream.Read(b)
	return b[:actualCount], err
}

func (c *Crypt4gh) readInt32() (uint32, error) {
	bytes, err := c.readBytes(4)
	var num uint32

	if err == nil {
		num = readInt32(bytes)
	}

	return num, err
}

// Reads 4 bytes in the Little-Endian order to compute an unsigned integer.
func readInt32(bytes []byte) uint32 {
	var num uint32
	for i := range bytes[0:4] {
		num = uint32(bytes[i])<<(8*i) | num
	}
	return num
}

// Reads 8 bytes in the Little-Endian order to compute an unsigned integer.
func readInt64(bytes []byte) uint64 {
	var num uint64
	for i := range bytes[0:8] {
		num = uint64(bytes[i])<<(8*i) | num
	}
	return num
}
