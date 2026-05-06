package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

func TestShareProofValidate(t *testing.T) {
	type testCase struct {
		name    string
		sp      ShareProof
		root    []byte
		wantErr bool
	}

	testCases := []testCase{
		{
			name:    "empty share proof returns error",
			sp:      ShareProof{},
			root:    shareProofRoot,
			wantErr: true,
		},
		{
			name:    "valid share proof returns no error",
			sp:      validShareProof(),
			root:    shareProofRoot,
			wantErr: false,
		},
		{
			name:    "share proof with empty share proofs returns error",
			sp:      emptyShareProofs(),
			root:    shareProofRoot,
			wantErr: true,
		},
		{
			name:    "share proof with empty proof bytes returns error",
			sp:      emptyKZGProofBytes(),
			root:    shareProofRoot,
			wantErr: true,
		},
		{
			name:    "share proof with nil commitment proof returns error",
			sp:      nilCommitmentProof(),
			root:    shareProofRoot,
			wantErr: true,
		},
		{
			name:    "share proof with empty column proofs returns error",
			sp:      emptyColumnProofs(),
			root:    shareProofRoot,
			wantErr: true,
		},
		{
			name:    "share proof with mismatched column indices returns error",
			sp:      mismatchedColumnProofs(),
			root:    shareProofRoot,
			wantErr: true,
		},
		{
			name:    "valid share proof with incorrect root returns error",
			sp:      validShareProof(),
			root:    shareProofWrongRoot,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.sp.Validate(tc.root)
			if tc.wantErr {
				assert.Error(t, got)
				return
			}
			assert.NoError(t, got)
		})
	}
}

const (
	validRootSize = 32
)

var (
	shareProofRoot      = bytesOfSize(validRootSize, 0x11)
	shareProofWrongRoot = bytesOfSize(validRootSize, 0x22)
)

func bytesOfSize(size int, value byte) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = value
	}
	return data
}

func emptyShareProofs() ShareProof {
	sp := validShareProof()
	sp.ShareProofs = nil
	return sp
}

func emptyKZGProofBytes() ShareProof {
	sp := validShareProof()
	sp.ShareProofs[0].Proof = nil
	return sp
}

func nilCommitmentProof() ShareProof {
	sp := validShareProof()
	sp.CommitmentProof = nil
	return sp
}

func emptyColumnProofs() ShareProof {
	sp := validShareProof()
	sp.CommitmentProof.ColumnProofs = nil
	return sp
}

func mismatchedColumnProofs() ShareProof {
	sp := validShareProof()
	sp.CommitmentProof.ColumnIndices = []uint32{}
	return sp
}

func validShareProof() ShareProof {
	return ShareProof{
		Data: [][]byte{{0x01, 0x02}},
		ShareProofs: []*tmproto.KZGMultiProof{
			{Proof: []byte{0x0a}},
		},
		NamespaceID: []byte{0x01, 0x02},
		CommitmentProof: &tmproto.CommitmentProof{
			ColumnProofs: []*tmproto.KZGMultiProof{
				{Proof: []byte{0x0b}},
			},
			ColumnIndices:  []uint32{0},
			RootCommitment: shareProofRoot,
		},
		NamespaceVersion: 0,
	}
}
