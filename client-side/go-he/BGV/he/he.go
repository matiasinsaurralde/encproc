package he

import (
	"fmt"

	"github.com/tuneinsight/lattigo/v6/core/rlwe"
	"github.com/tuneinsight/lattigo/v6/schemes/bgv"
)

type HE struct {
	Sk     *rlwe.SecretKey
	Pk     *rlwe.PublicKey
	Params bgv.Parameters
}

func (init *HE) GenerateKeypair() {
	init.Params = SetupParams()

	// Key Generator
	kgen := rlwe.NewKeyGenerator(init.Params)

	// Secret and Public Key
	init.Sk = kgen.GenSecretKeyNew()
	init.Pk = kgen.GenPublicKeyNew(init.Sk)
}

func (he *HE) Decrypt_result(sk *rlwe.SecretKey, ct *rlwe.Ciphertext) ([]uint64, error) {
	var err error
	// Encoder
	ecd := bgv.NewEncoder(he.Params)
	// Decryptor
	dec := bgv.NewDecryptor(he.Params, sk)
	// Decrypts the vector of plaintext values

	pt := rlwe.NewPlaintext(he.Params, he.Params.MaxLevel())
	dec.Decrypt(ct, pt)
	// Vector of plaintext values
	values := make([]uint64, he.Params.MaxSlots())
	// Encoder
	if err = ecd.Decode(pt, values); err != nil {
		panic(err)
	}
	fmt.Println(values)
	return values, nil
}

// Encrypt input
func (he *HE) EncryptInput(inputAr []uint64, pkBin []byte) ([]byte, error) {
	var err error
	fmt.Println(inputAr)
	// Initialize Public Key
	pk := rlwe.NewPublicKey(he.Params)
	err = pk.UnmarshalBinary(pkBin)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public key: %v", err)
	}

	// Initialize Plaintext
	pt := bgv.NewPlaintext(he.Params, he.Params.MaxLevel())

	// Encoder
	ecd := bgv.NewEncoder(he.Params)

	// Encryptor
	encryptor := bgv.NewEncryptor(he.Params, pk)

	// Encode and encrypt the input
	err = ecd.Encode(inputAr, pt)
	if err != nil {
		return nil, fmt.Errorf("failed to encode input: %v", err)
	}

	ct := bgv.NewCiphertext(he.Params, 1, he.Params.MaxLevel())
	err = encryptor.Encrypt(pt, ct)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt input: %v", err)
	}

	// Serialize the ciphertext
	ctByte, err := ct.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize ciphertext: %v", err)
	}

	return ctByte, nil
}

func (he *HE) ExportBytes() ([]byte, []byte, error) {
	pkBytes, err := he.Pk.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	skBytes, err := he.Sk.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	return pkBytes, skBytes, nil
}
