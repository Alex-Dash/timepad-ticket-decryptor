package main

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"syscall/js"
)

func main() {

}

func decryptorWrapper() js.Func {
	decrFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 2 {
			return fmt.Sprintf("Incorrect amount of args passed, expected 2 got %d", len(args))
		}
		pubkey := args[0].String()
		ticket := args[1].String()

		if len(pubkey) == 0 || len(ticket) == 0 {
			fmt.Println("Cannot use empty strings as input")
			return errors.New("cannot use empty strings as input").Error()
		}

		fmt.Printf("pubkey %s\n", pubkey)
		fmt.Printf("ticket %s\n", ticket)

		data, err := decryptTicket("", "")
		if err != nil {
			fmt.Printf("Error encountered: %s\n", err)
			return err.Error()
		}
		return data
	})
	return decrFunc
}

func decryptTicket(pubKeyPem string, ticket string) (string, error) {

	// Import public key
	pubKey, err3 := ImportSPKIPublicKeyPEM(pubKeyPem)
	if err3 != nil {
		return "", errors.New("failed to decode b64 contents of a ticket")
	}
	// Base64 decode the ticket contents
	ciphertextBytes, err4 := base64.StdEncoding.DecodeString(ticket)
	if err4 != nil {
		return "", errors.New("failed to decode b64 contents of a ticket")
	}

	// Split ciphertext into signature chunks a 32 bytes and decrypt each chunk
	reader := bytes.NewReader(ciphertextBytes)
	var writer bytes.Buffer
	ciphertextBytesChunk := make([]byte, 32)
	for {
		n, err := io.ReadFull(reader, ciphertextBytesChunk)
		if err != nil {
			return "", errors.New("failed to read ciphertext bytes chunk")
		}
		if n == 0 {
			break
		}
		decryptChunk(ciphertextBytesChunk, &writer, pubKey)
	}
	// Concatenate decrypted signature chunks
	decryptedData := writer.String()
	return decryptedData, nil
}

func ImportSPKIPublicKeyPEM(spkiPEM string) (*rsa.PublicKey, error) {
	body, _ := pem.Decode([]byte(spkiPEM))
	if body == nil || body.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}
	publicKey, e1 := x509.ParsePKIXPublicKey(body.Bytes)
	if e1 != nil {
		return nil, errors.New("failed to parse X509 cert from key")
	}
	if publicKey, ok := publicKey.(*rsa.PublicKey); ok {
		return publicKey, nil
	} else {
		return nil, errors.New("public key import failed or key is empty")
	}
}

func decryptChunk(ciphertextBytesChunk []byte, writer *bytes.Buffer, pubKey *rsa.PublicKey) {
	// Decrypt each signature chunk
	ciphertextInt := new(big.Int)
	ciphertextInt.SetBytes(ciphertextBytesChunk)
	decryptedPaddedInt := decrypt(new(big.Int), pubKey, ciphertextInt)
	// Remove padding
	decryptedPaddedBytes := make([]byte, pubKey.Size())
	decryptedPaddedInt.FillBytes(decryptedPaddedBytes)
	start := bytes.Index(decryptedPaddedBytes[1:], []byte{0}) + 1 // // 0001FF...FF00<data>: Find index after 2nd 0x00
	decryptedBytes := decryptedPaddedBytes[start:]
	// Write decrypted signature chunk
	writer.Write(decryptedBytes)
}

func decrypt(c *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	// Textbook RSA
	e := big.NewInt(int64(pub.E))
	c.Exp(m, e, pub.N)
	return c
}
