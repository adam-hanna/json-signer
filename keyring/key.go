package keyring

import (
	"fmt"
	"math/big"

	"github.com/tbruyelle/legacykey/codec"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	ledger "github.com/cosmos/ledger-cosmos-go"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmoskeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type Key struct {
	Name string
	// Record is not nil if the key is proto-encoded
	Record *cosmoskeyring.Record
	// Info is not nil if the key is amino-encoded
	Info cosmoskeyring.LegacyInfo
}

func discoverLedger() (*ledger.LedgerCosmos, error) {
	return ledger.FindLedgerCosmosUserApp()
}

func (k Key) Sign(bz []byte) ([]byte, cryptotypes.PubKey, error) {
	if k.IsAmino() {
		switch k.Info.GetType() {
		case cosmoskeyring.TypeLocal:
			privKey, err := privKeyFromInfo(k.Info)
			if err != nil {
				return nil, nil, err
			}
			signature, err := privKey.Sign(bz)
			if err != nil {
				return nil, nil, err
			}
			return signature, privKey.PubKey(), nil

		case cosmoskeyring.TypeLedger:
			device, err := discoverLedger()
			if err != nil {
				return nil, nil, err
			}
			path, err := k.Info.GetPath()
			if err != nil {
				return nil, nil, err
			}
			pubKey, err := getLedgerPubKey(device, path.DerivationPath())
			if err != nil {
				return nil, nil, err
			}
			signature, err := device.SignSECP256K1(path.DerivationPath(), bz, 0)
			if err != nil {
				return nil, nil, err
			}
			signature, err = convertDERtoBER(signature)
			if err != nil {
				return nil, nil, err
			}
			return signature, pubKey, nil
		}
		return nil, nil, fmt.Errorf("unhandled key type %q", k.Info.GetType())
	}
	switch k.Record.GetType() {
	case cosmoskeyring.TypeLocal:
		privKey, err := privKeyFromRecord(k.Record)
		if err != nil {
			return nil, nil, err
		}
		signature, err := privKey.Sign(bz)
		if err != nil {
			return nil, nil, err
		}
		return signature, privKey.PubKey(), nil
	}
	return nil, nil, fmt.Errorf("unhandled key type %q", k.Record.GetType())
}

func getLedgerPubKey(device *ledger.LedgerCosmos, bip32Path []uint32) (cryptotypes.PubKey, error) {
	pubKey, err := device.GetPublicKeySECP256K1(bip32Path)
	if err != nil {
		return nil, err
	}
	// re-serialize in the 33-byte compressed format
	cmp, err := btcec.ParsePubKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}

	compressedPublicKey := make([]byte, secp256k1.PubKeySize)
	copy(compressedPublicKey, cmp.SerializeCompressed())

	return &secp256k1.PubKey{Key: compressedPublicKey}, nil
}

func convertDERtoBER(signatureDER []byte) ([]byte, error) {
	sigDER, err := ecdsa.ParseDERSignature(signatureDER)
	if err != nil {
		return nil, err
	}

	sigStr := sigDER.Serialize()
	// The format of a DER encoded signature is as follows:
	// 0x30 <total length> 0x02 <length of R> <R> 0x02 <length of S> <S>
	r, s := new(big.Int), new(big.Int)
	r.SetBytes(sigStr[4 : 4+sigStr[3]])
	s.SetBytes(sigStr[4+sigStr[3]+2:])

	sModNScalar := new(btcec.ModNScalar)
	sModNScalar.SetByteSlice(s.Bytes())
	// based on https://github.com/tendermint/btcd/blob/ec996c5/btcec/signature.go#L33-L50
	if sModNScalar.IsOverHalfOrder() {
		s = new(big.Int).Sub(btcec.S256().N, s)
	}

	sigBytes := make([]byte, 64)
	// 0 pad the byte arrays from the left if they aren't big enough.
	copy(sigBytes[32-len(r.Bytes()):32], r.Bytes())
	copy(sigBytes[64-len(s.Bytes()):64], s.Bytes())

	return sigBytes, nil
}

func (k Key) IsAmino() bool {
	return k.Info != nil
}

func (k Key) RecordToInfo() (cosmoskeyring.LegacyInfo, error) {
	return legacyInfoFromRecord(k.Record)
}

func extractPrivKeyFromLocal(rl *cosmoskeyring.Record_Local) (cryptotypes.PrivKey, error) {
	if rl.PrivKey == nil {
		return nil, cosmoskeyring.ErrPrivKeyNotAvailable
	}

	priv, ok := rl.PrivKey.GetCachedValue().(cryptotypes.PrivKey)
	if !ok {
		return nil, cosmoskeyring.ErrCastAny
	}

	return priv, nil
}

func privKeyFromRecord(record *cosmoskeyring.Record) (cryptotypes.PrivKey, error) {
	switch record.GetType() {
	case cosmoskeyring.TypeLocal:
		return extractPrivKeyFromLocal(record.GetLocal())
	}
	return nil, fmt.Errorf("unhandled Record type %q", record.GetType())
}

func privKeyFromInfo(info cosmoskeyring.LegacyInfo) (privKey cryptotypes.PrivKey, err error) {
	switch info.GetType() {
	case cosmoskeyring.TypeLocal:
		err = codec.Amino.Unmarshal([]byte(info.(legacyLocalInfo).GetPrivKeyArmor()), &privKey)
		return
	}
	return nil, fmt.Errorf("unhandled Info type %q", info.GetType())
}

// legacyInfoFromLegacyInfo turns a Record into a LegacyInfo.
func legacyInfoFromRecord(record *cosmoskeyring.Record) (cosmoskeyring.LegacyInfo, error) {
	switch record.GetType() {
	case cosmoskeyring.TypeLocal:
		pk, err := record.GetPubKey()
		if err != nil {
			return nil, err
		}
		privKey, err := extractPrivKeyFromLocal(record.GetLocal())
		if err != nil {
			return nil, err
		}
		privBz, err := codec.Amino.Marshal(privKey)
		if err != nil {
			return nil, err
		}
		return legacyLocalInfo{
			Name:         record.Name,
			PubKey:       pk,
			Algo:         hd.PubKeyType(pk.Type()),
			PrivKeyArmor: string(privBz),
		}, nil

	case cosmoskeyring.TypeLedger:
		pk, err := record.GetPubKey()
		if err != nil {
			return nil, err
		}
		return legacyLedgerInfo{
			Name:   record.Name,
			PubKey: pk,
			Algo:   hd.PubKeyType(pk.Type()),
			Path:   *record.GetLedger().Path,
		}, nil

	case cosmoskeyring.TypeMulti:
		panic("record type TypeMulti unhandled")

	case cosmoskeyring.TypeOffline:
		panic("record type TypeOffline unhandled")
	}
	panic(fmt.Sprintf("record type %s unhandled", record.GetType()))
}
