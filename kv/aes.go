package kv

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/hex"
)

func padding(src []byte, blocksize int) []byte {
	padnum := blocksize - len(src)%blocksize
	pad := bytes.Repeat([]byte{byte(padnum)}, padnum)
	return append(src, pad...)
}

func unpadding(src []byte) []byte {
	n := len(src)
	unpadnum := int(src[n-1])
	return src[:n-unpadnum]
}

func encrypt(src string, key string) string {
	bs := []byte(src)
	bk := []byte(key)
	block, _ := aes.NewCipher(bk)

	bs = padding(bs, block.BlockSize())
	blockmode := cipher.NewCBCEncrypter(block, bk)
	blockmode.CryptBlocks(bs, bs)
	return hex.EncodeToString(bs)
}

func decrypt(src string, key string) string {
	bs, _ := hex.DecodeString(src)
	bk := []byte(key)
	block, _ := aes.NewCipher(bk)
	blockmode := cipher.NewCBCDecrypter(block, bk)
	blockmode.CryptBlocks(bs, bs)
	bs = unpadding(bs)
	return string(bs)
}

func toMD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))[0:16]
}
