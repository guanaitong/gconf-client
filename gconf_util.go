package gconf

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"math/big"
)

func Decrypt(encryptedPassword string) string {
	if encryptedPassword == "" {
		return ""
	}
	encryptedDecodeBytes, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return ""
	}
	publicKey := GetGlobalConfigCollection().GetValue("publicKey").Raw()
	key, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return ""
	}
	pubKey, err := x509.ParsePKIXPublicKey(key)
	if err != nil {
		return ""
	}
	pub := pubKey.(*rsa.PublicKey)
	return string(rsaPublicDecrypt(pub, encryptedDecodeBytes))
}

func rsaPublicDecrypt(pubKey *rsa.PublicKey, data []byte) []byte {
	c := new(big.Int)
	m := new(big.Int)
	m.SetBytes(data)
	e := big.NewInt(int64(pubKey.E))
	c.Exp(m, e, pubKey.N)
	out := c.Bytes()
	skip := 0
	for i := 2; i < len(out); i++ {
		if i+1 >= len(out) {
			break
		}
		if out[i] == 0xff && out[i+1] == 0 {
			skip = i + 2
			break
		}
	}
	return out[skip:]
}
