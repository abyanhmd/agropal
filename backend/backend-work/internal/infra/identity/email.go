package identity

import (
	cryptoRand "crypto/rand"
	"math/big"
)

func (g *Generator) GenerateOTP() string {
	lengthOTP := 6
	digits := []byte("0123456789")

	otp := make([]byte, lengthOTP)

	for i := range lengthOTP {
		idxBig, _ := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(digits))))
		otp[i] = digits[idxBig.Int64()]
	}

	return string(otp)
}

func (g *Generator) GenerateReferral() string {
	lengthReferral := 8
	charsets := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	referral := make([]byte, lengthReferral)

	for i := range referral {
		idxBig, _ := cryptoRand.Int(cryptoRand.Reader, big.NewInt(int64(len(charsets))))
		referral[i] = charsets[idxBig.Int64()]
	}

	return string(referral)
}
