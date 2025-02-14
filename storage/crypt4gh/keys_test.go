package crypt4gh

import (
	"os"
	"testing"
)

const encodedPublicKey = "LS0tLS1CRUdJTiBDUllQVDRHSCBQVUJMSUMgS0VZLS0tLS0KQ3N4Mk5KNVhicTJuM3Q4dWdSeTJabGRLYXRoWDRLa0haZ1dzcGhuTTlSbz0KLS0tLS1FTkQgQ1JZUFQ0R0ggUFVCTElDIEtFWS0tLS0tCg=="

func TestLoadingEncryptedKeys(t *testing.T) {
	os.Setenv("C4GH_PUBLIC_KEY", "testdata/key.pub")
	os.Setenv("C4GH_SECRET_KEY", "testdata/key.encrypted.sec")
	os.Setenv("C4GH_PASSPHRASE", "abcDEFghi")

	c4gh, err := ResolveKeyPair()

	if err != nil {
		t.Error("Could not load encrypted Crypt4gh key-pair", err)
	} else if len(c4gh.publicKey) == 0 {
		t.Error("Loaded Crypt4gh public key is empty", err)
	} else if len(c4gh.secretKey) == 0 {
		t.Error("Loaded Crypt4gh secret key is empty", err)
	} else if c4gh.EncodePublicKeyBase64() != encodedPublicKey {
		t.Error("Unexpected BASE64-encoded public key:", c4gh.EncodePublicKeyBase64())
	}
}

func TestLoadingNonEncryptedKeys(t *testing.T) {
	os.Setenv("C4GH_PUBLIC_KEY", "testdata/key.pub")
	os.Setenv("C4GH_SECRET_KEY", "testdata/key.plain.sec")
	os.Setenv("C4GH_PASSPHRASE", "to-be-ignored")

	c4gh, err := ResolveKeyPair()

	if err != nil {
		t.Error("Could not load plain-text Crypt4gh key-pair", err)
	} else if len(c4gh.publicKey) == 0 {
		t.Error("Loaded Crypt4gh public key is empty", err)
	} else if len(c4gh.secretKey) == 0 {
		t.Error("Loaded Crypt4gh secret key is empty", err)
	} else if c4gh.EncodePublicKeyBase64() != encodedPublicKey {
		t.Error("Unexpected BASE64-encoded public key:", c4gh.EncodePublicKeyBase64())
	}
}

func TestGeneratingAndSavingNewKeys(t *testing.T) {
	c4gh, err := NewKeyPair()

	if err != nil {
		t.Error("Could not generate a Crypt4gh key-pair", err)
	} else if len(c4gh.publicKey) == 0 {
		t.Error("Generated Crypt4gh public key is empty", err)
	} else if len(c4gh.secretKey) == 0 {
		t.Error("Generated Crypt4gh secret key is empty", err)
	}

	pubPath := "tmp_key.pub"
	secPath := "tmp_key.sec"

	defer os.Remove(pubPath)
	defer os.Remove(secPath)

	err = c4gh.Save(pubPath, secPath, nil)
	if err != nil {
		t.Error("Could not save a Crypt4gh key-pair to file", err)
	}

	os.Remove(pubPath)
	os.Remove(secPath)

	err = c4gh.Save(pubPath, secPath, []byte("abcDEFghi"))
	if err != nil {
		t.Error("Could not save a Crypt4gh key-pair to file", err)
	}

	_, err = KeyPairFromFiles(pubPath, secPath, []byte("abcDEFghi"))
	if err != nil {
		t.Error("Could not reload the saved Crypt4gh key-pair", err)
	}
}
