package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
)

type Transaction struct {
	ID      []byte     // represents the id of the transaction
	Inputs  []TxInput  // represents the inputs of the transaction
	Outputs []TxOutput // represents the outputs of the transaction
}

type TxOutput struct {
	Value  int    // represents the value in tokens
	PubKey string // represents the public key
}

type TxInput struct {
	ID  []byte // represents the transaction that the output is
	Out int    // represents the index where the output appears
	Sig string // represents the data wich is use in the output pubkey
}

// setID will generate a hashed id for the transaction
func (tx *Transaction) setID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	CheckError(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// CoinbasTx will generate a new transaction instance with the given data
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txin := TxInput{ID: []byte{}, Out: -1, Sig: data}
	txout := TxOutput{Value: 100, PubKey: to}

	tx := Transaction{
		ID:      nil,
		Inputs:  []TxInput{txin},
		Outputs: []TxOutput{txout},
	}

	tx.setID()
	return &tx
}

// IsCoinbase will determine if the current transaction is a coinbase
func (tx *Transaction) IsCoinBase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// CanUnlock check if the given data is equal to the
// input sig wich is the pub key of the transaction
// if are equal means that the data of the input can be unlocked
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

// CanBeUnlocked check if the given data is equal to the
// output pub key if are equal means that the data of the output can be unlocked
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
