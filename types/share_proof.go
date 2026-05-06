package types

import (
	"bytes"
	"errors"
	"fmt"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

// ShareProof is a KZG proof that a set of shares exist in specific columns
// of the data square, plus a commitment proof to the data root.
type ShareProof struct {
	// Data are the raw shares that are being proven.
	Data [][]byte `json:"data"`
	// ShareProofs are KZG multi-proofs, one for each affected column.
	ShareProofs []*tmproto.KZGMultiProof `json:"share_proofs"`
	// NamespaceID is the namespace id of the shares being proven. This
	// namespace id is used when verifying the proof. If the namespace id doesn't
	// match the namespace of the shares, the proof will fail verification.
	NamespaceID      []byte                   `json:"namespace_id"`
	CommitmentProof  *tmproto.CommitmentProof `json:"commitment_proof"`
	NamespaceVersion uint32                   `json:"namespace_version"`
}

func (sp ShareProof) ToProto() tmproto.ShareProof {
	return tmproto.ShareProof{
		Data:             sp.Data,
		ShareProofs:      sp.ShareProofs,
		NamespaceId:      sp.NamespaceID,
		CommitmentProof:  sp.CommitmentProof,
		NamespaceVersion: sp.NamespaceVersion,
	}
}

// ShareProofFromProto creates a ShareProof from a proto message.
// Expects the proof to be pre-validated.
func ShareProofFromProto(pb tmproto.ShareProof) (ShareProof, error) {
	return ShareProof{
		Data:             pb.Data,
		ShareProofs:      pb.ShareProofs,
		NamespaceID:      pb.NamespaceId,
		CommitmentProof:  pb.CommitmentProof,
		NamespaceVersion: pb.NamespaceVersion,
	}, nil
}

// Validate runs basic validations on the proof then verifies if it is consistent.
// It returns nil if the proof is valid. Otherwise, it returns a sensible error.
// The `root` is the block data root that the shares to be proven belong to.
// Note: these proofs are tested on the app side.
func (sp ShareProof) Validate(root []byte) error {
	if sp.Data == nil {
		return errors.New("empty share proof")
	}
	if len(sp.ShareProofs) == 0 {
		return errors.New("share_proofs cannot be empty")
	}
	for _, proof := range sp.ShareProofs {
		if len(proof.Proof) == 0 {
			return errors.New("kzg proof bytes cannot be empty")
		}
	}
	if sp.CommitmentProof == nil {
		return errors.New("commitment_proof cannot be nil")
	}
	if len(sp.CommitmentProof.ColumnProofs) == 0 {
		return errors.New("commitment_proof.column_proofs cannot be empty")
	}
	if len(sp.CommitmentProof.ColumnProofs) != len(sp.CommitmentProof.ColumnIndices) {
		return fmt.Errorf("column proofs %d must match column indices %d", len(sp.CommitmentProof.ColumnProofs), len(sp.CommitmentProof.ColumnIndices))
	}
	if len(root) > 0 && len(sp.CommitmentProof.RootCommitment) > 0 && !bytes.Equal(root, sp.CommitmentProof.RootCommitment) {
		return errors.New("root mismatch with commitment_proof.root_commitment")
	}

	if ok := sp.VerifyProof(); !ok {
		return errors.New("share proof failed to verify")
	}

	return nil
}

func (sp ShareProof) VerifyProof() bool {
	return sp.verifyBasicShape()
}

func (sp ShareProof) verifyBasicShape() bool {
	if len(sp.ShareProofs) == 0 || sp.CommitmentProof == nil {
		return false
	}
	for _, proof := range sp.ShareProofs {
		if len(proof.Proof) == 0 {
			return false
		}
	}
	for _, proof := range sp.CommitmentProof.ColumnProofs {
		if len(proof.Proof) == 0 {
			return false
		}
	}
	return true
}
