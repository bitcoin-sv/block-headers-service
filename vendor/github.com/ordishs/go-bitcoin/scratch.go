package bitcoin

import (
	"encoding/binary"
	"encoding/hex"

	"bitbucket.org/simon_ordish/cryptolib"
)

type input struct {
	hash            [32]byte // The previous utxo being spent.
	index           uint32   // The previous utxo index being spent.
	unlockingScript []byte   // A script-language script which satisfies the conditions placed in the outpoint’s pubkey script. Should only contain data pushes; see https://bitcoin.org/en/developer-reference#signature_script_modification_warning.
	sequence        uint32   // Sequence number. Default for Bitcoin Core and almost all other programs is 0xffffffff. See https://bitcoin.org/en/glossary/sequence-number
}

func (i *input) toHex() []byte {
	var b []byte

	b = append(b, cryptolib.ReverseBytes(i.hash[:])...)
	b = append(b, cryptolib.GetLittleEndianBytes(i.index, 4)...)
	b = append(b, cryptolib.VarInt(uint64(len(i.unlockingScript)))...)
	b = append(b, i.unlockingScript...)
	b = append(b, cryptolib.GetLittleEndianBytes(i.sequence, 4)...)

	return b
}

func inputFromBytes(b []byte) (*input, int) {
	pos := 0

	var previousOutput [32]byte
	copy(previousOutput[:], cryptolib.ReverseBytes(b[pos:pos+32]))

	pos += 32

	index := binary.LittleEndian.Uint32(b[pos : pos+4])
	pos += 4

	scriptLen, size := cryptolib.DecodeVarInt(b[pos:])
	pos += size

	len := int(scriptLen)

	script := b[pos : pos+len]
	pos += len

	sequence := binary.LittleEndian.Uint32(b[pos : pos+4])
	pos += 4

	return &input{
		hash:            previousOutput,
		index:           index,
		unlockingScript: script,
		sequence:        sequence,
	}, pos
}

type output struct {
	value         uint64 // Number of satoshis to spend. May be zero; the sum of all outputs may not exceed the sum of satoshis previously spent to the outpoints provided in the input section. (Exception: coinbase transactions spend the block subsidy and collected transaction fees.)
	lockingScript []byte // Defines the conditions which must be satisfied to spend this output.
}

func (o *output) toHex() []byte {
	var b []byte

	value := make([]byte, 8)
	binary.LittleEndian.PutUint64(value, o.value)

	b = append(b, value...)
	b = append(b, cryptolib.VarInt(uint64(len(o.lockingScript)))...)
	b = append(b, o.lockingScript...)

	return b
}

func outputFromBytes(b []byte) (*output, int) {
	pos := 0

	value := binary.LittleEndian.Uint64(b[pos : pos+8])
	pos += 8

	scriptLen, size := cryptolib.DecodeVarInt(b[pos:])
	pos += size

	len := int(scriptLen)

	script := b[pos : pos+len]
	pos += len

	return &output{
		value:         value,
		lockingScript: script,
	}, pos

}

type transaction struct {
	Hash     string
	Version  int32    // Transaction version number (note, this is signed); currently version 1 or 2. Programs creating transactions using newer consensus rules may use higher version numbers. Version 2 means that BIP 68 applies.
	Inputs   []input  // Transaction inputs.
	Outputs  []output // Transaction outputs.
	LockTime uint32   // A time (Unix epoch time) or block number. See https://bitcoin.org/en/transactions-guide#locktime_parsing_rules
}

// TransactionFromHex takes a hex string and constructs a Transaction object
func TransactionFromHex(h string) (*transaction, int) {
	s, _ := hex.DecodeString(h)
	return TransactionFromBytes(s)
}

// TransactionFromBytes takes a slice of bytes and constructs a Transaction object
func TransactionFromBytes(b []byte) (*transaction, int) {
	pos := 0

	// extract the version
	version := binary.LittleEndian.Uint32(b[0:4])
	pos += 4

	// Get the number of inputs
	numberOfInputs, size := cryptolib.DecodeVarInt(b[pos:])
	pos += size

	var inputs []input

	for i := uint64(0); i < numberOfInputs; i++ {
		input, size := inputFromBytes(b[pos:])
		pos += size

		inputs = append(inputs, *input)
	}

	// Get the number of outputs
	numberOfOutputs, size := cryptolib.DecodeVarInt(b[pos:])
	pos += size

	var outputs []output

	for i := uint64(0); i < numberOfOutputs; i++ {
		output, size := outputFromBytes(b[pos:])
		pos += size

		outputs = append(outputs, *output)
	}

	locktime := binary.LittleEndian.Uint32(b[pos : pos+4])
	pos += 4

	hash := cryptolib.Sha256d(b[0:pos])

	return &transaction{
		Hash:     hex.EncodeToString(cryptolib.ReverseBytes(hash)),
		Version:  int32(version),
		Inputs:   inputs,
		Outputs:  outputs,
		LockTime: locktime,
	}, pos
}

func (t *transaction) InputCount() int {
	return len(t.Inputs)
}

func (t *transaction) OutputCount() int {
	return len(t.Outputs)
}

func (t *transaction) ToHex() []byte {
	var b []byte

	b = append(b, cryptolib.GetLittleEndianBytes(uint32(t.Version), 4)...)
	b = append(b, cryptolib.VarInt(uint64(t.InputCount()))...)
	for _, input := range t.Inputs {
		b = append(b, input.toHex()...)
	}

	b = append(b, cryptolib.VarInt(uint64(t.OutputCount()))...)
	for _, output := range t.Outputs {
		b = append(b, output.toHex()...)
	}

	b = append(b, cryptolib.GetLittleEndianBytes(t.LockTime, 4)...)

	return b
}
