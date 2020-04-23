package zkflowexample

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/iden3/go-circom-prover-verifier/parsers"
	"github.com/iden3/go-circom-prover-verifier/prover"
	"github.com/iden3/go-circom-prover-verifier/verifier"
	witnesscalc "github.com/iden3/go-circom-witnesscalc"
	"github.com/stretchr/testify/require"
)

func TestFullFlow(t *testing.T) {
	// compile circuits & compute trusted setup:
	// using compile-and-trustedsetup.sh

	// build the testing environment: claims, identities, merkletrees, etc
	printT("Generate testing environment: claims, identities, merkletrees, etc")
	printT("- Generate inputs")
	inputsJson, err := IdStateInputs()
	require.Nil(t, err)
	err = ioutil.WriteFile("testdata/inputs.json", []byte(inputsJson), 0644)
	require.Nil(t, err)

	// calculate witness
	printT("- Parse witness files")
	wasmFilename := "testdata/circuit.wasm"
	inputsFilename := "testdata/inputs.json"

	wasmBytes, err := ioutil.ReadFile(wasmFilename)
	require.Nil(t, err)
	inputsBytes, err := ioutil.ReadFile(inputsFilename)
	require.Nil(t, err)
	inputs, err := witnesscalc.ParseInputs(inputsBytes)
	require.Nil(t, err)

	printT("- Calculate witness")
	wit, err := ComputeWitness(wasmBytes, inputs)
	require.Nil(t, err)

	printT("- Write witness output files")
	wJSON, err := json.Marshal(witnesscalc.WitnessJSON(wit))
	require.Nil(t, err)
	err = ioutil.WriteFile("testdata/witness.json", []byte(wJSON), 0644)
	require.Nil(t, err)

	// generate zk proof
	// read ProvingKey & Witness files
	printT("- Read proving_key.json & witness.json files")
	provingKeyJson, err := ioutil.ReadFile("testdata/proving_key.json")
	require.Nil(t, err)
	witnessJson, err := ioutil.ReadFile("testdata/witness.json")
	require.Nil(t, err)

	// parse Proving Key
	pk, err := parsers.ParsePk(provingKeyJson)
	require.Nil(t, err)
	// parse Witness
	w, err := parsers.ParseWitness(witnessJson)
	require.Nil(t, err)

	// generate the proof
	printT("- Generate zkProof")
	proof, pubSignals, err := prover.GenerateProof(pk, w)
	require.Nil(t, err)

	// print proof & publicSignals
	proofStr, err := parsers.ProofToJson(proof)
	require.Nil(t, err)
	publicStr, err := json.Marshal(parsers.ArrayBigIntToString(pubSignals))
	require.Nil(t, err)

	err = ioutil.WriteFile("testdata/proof.json", proofStr, 0644)
	require.Nil(t, err)
	err = ioutil.WriteFile("testdata/public.json", publicStr, 0644)
	require.Nil(t, err)

	// verify zk proof
	// read proof & verificationKey & publicSignals
	proofJson, err := ioutil.ReadFile("testdata/proof.json")
	require.Nil(t, err)
	printT("- Read verification_key.json & public.json files")
	vkJson, err := ioutil.ReadFile("testdata/verification_key.json")
	require.Nil(t, err)
	publicJson, err := ioutil.ReadFile("testdata/public.json")
	require.Nil(t, err)

	// parse proof & verificationKey & publicSignals
	public, err := parsers.ParsePublicSignals(publicJson)
	require.Nil(t, err)
	proofParsed, err := parsers.ParseProof(proofJson)
	require.Nil(t, err)
	vk, err := parsers.ParseVk(vkJson)
	require.Nil(t, err)

	// verify the proof with the given verificationKey & publicSignals
	printT("- Verify zkProof")
	v := verifier.Verify(vk, proof, pubSignals)
	fmt.Println("verifier.Verify", v)

	// verify but using the stored files
	vkJson, err = ioutil.ReadFile("testdata/verification_key.json")
	require.Nil(t, err)
	vk, err = parsers.ParseVk(vkJson)
	require.Nil(t, err)
	v2 := verifier.Verify(vk, proofParsed, public)
	fmt.Println("verifier.Verify", v2)
}
