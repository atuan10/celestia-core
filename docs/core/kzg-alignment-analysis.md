# KZG-first Alignment Analysis: celestia-app vs celestia-core

## 1. Proto Message Structures

### 1.1 ShareProof - MAJOR CHANGE

**celestia-app (KZG-first):**
```proto
message ShareProof {
  repeated bytes data = 1;
  repeated KZGMultiProof share_proofs = 2;
  bytes namespace_id = 3;
  CommitmentProof commitment_proof = 4;
  uint32 namespace_version = 5;
}
```

**celestia-core (CURRENT - NMT-based):**
```proto
message ShareProof {
  repeated bytes data = 1;
  repeated NMTProof share_proofs = 2;
  bytes namespace_id = 3;
  RowProof row_proof = 4;
  uint32 namespace_version = 5;
}
```

**Key Changes:**
- Field 2: `NMTProof[]` → `KZGMultiProof[]`
- Field 4: `RowProof` → `CommitmentProof`

### 1.2 New Proto Messages (NOT in celestia-core yet)

**KZGMultiProof:**
```proto
message KZGMultiProof {
  bytes proof = 1;  // G1 point commitment (48 bytes for BN254)
}
```

**CommitmentProof:**
```proto
message CommitmentProof {
  repeated KZGMultiProof column_proofs = 1;
  repeated uint32 column_indices = 2;
  bytes root_commitment = 3;
}
```

### 1.3 DataAvailabilityHeader - EXTENDED (in Data message)

**celestia-app (CURRENT):**
```proto
message DataAvailabilityHeader {
  repeated bytes row_roots = 1 [deprecated = true];
  repeated bytes column_roots = 2 [deprecated = true];
  repeated bytes piece_commitments = 3;
  repeated bytes column_commitments = 4;
  repeated NamespaceRangeEntry namespace_index = 5;
}
```

**celestia-core (CURRENT - in Block.Data):**
- Contains only basic fields (no KZG commitments)
- Deprecated row_roots and column_roots are still active

**NamespaceRangeEntry:**
```proto
message NamespaceRangeEntry {
  bytes namespace_id = 1;
  uint32 start = 2;
  uint32 end = 3;
}
```

---

## 2. Go Type Structures & Functions

### 2.1 celestia-app/pkg/proof/ Key Types

#### ShareProof
- **Fields:**
  - `Data [][]byte` - Raw shares
  - `ShareProofs []*KZGMultiProof` - KZG proofs per column
  - `NamespaceId []byte`
  - `NamespaceVersion uint32`
  - `CommitmentProof *CommitmentProof` - Column commitment inclusion
  
- **Methods:**
  - `Validate(root []byte) error` - Basic validation
  - `VerifyProof() bool` - Basic shape verification
  - `VerifyProofWithKZGRange(rangeProof, dah, codec, provider) bool` - Full cryptographic verification
  - `verifyBasicShape() bool` - Internal shape checks

#### KZGMultiProof
- **Fields:**
  - `Proof []byte` - Commitment to quotient polynomial (G1 point)

#### CommitmentProof
- **Fields:**
  - `ColumnProofs []*KZGMultiProof` - KZG proofs for column commitments
  - `ColumnIndices []uint32` - Which columns are proven
  - `RootCommitment []byte` - Data root as commitment point (optional)

#### KZGRangeProof (INTERNAL - used for proof generation)
- **Fields:**
  - `NamespaceID []byte`
  - `NamespaceVersion uint32`
  - `StartCell uint32`
  - `EndCell uint32` (end-exclusive)
  - `DataRoot []byte`
  - `Width uint32`
  - `MaxChunks uint32`
  - `Seed int`
  - `ColumnProofs []KZGColumnProof`
  - `CellProofs []KZGCellProof`
  - `RowProofs []KZGRowProof`
  - `RowBatch *KZGRowBatch`

#### KZGColumnProof
- **Fields:**
  - `Column uint32`
  - `Commitment []byte`
  - `Proof *Proof`

#### KZGCellProof
- **Fields:**
  - `Row uint32`
  - `Column uint32`
  - `ShareData []byte`
  - `PieceOpenProofs [][]byte`

#### KZGRowProof
- **Fields:**
  - `Row uint32`
  - `Columns []uint32`
  - `Coeffs []byte`
  - `CellCombinedProofs [][]byte`
  - `CombinedProof []byte`
  - `CombinedCommitment []byte`

#### KZGRowBatch
- **Fields:**
  - `Transcript []byte`
  - `Coeffs []byte`
  - `CombinedProof []byte`
  - `CombinedCommitment []byte`

### 2.2 celestia-app/pkg/proof/ Key Functions

#### ShareProof Generation
- `NewTxInclusionProof(txs [][]byte, txIndex, _ uint64) (ShareProof, error)`
- `NewShareInclusionProof(dataSquare, namespace, shareRange) (ShareProof, error)`
- `NewShareInclusionProofFromEDS(eds, namespace, shareRange) (ShareProof, error)`

#### Internal Conversion
- `shareProofFromKZGRange(rangeProof *KZGRangeProof, root []byte) (ShareProof, error)` - Converts KZGRangeProof to wire format

#### KZGRangeProof Generation
- `NewKZGRangeProofFromEDS(eds, dah, namespace, codec, provider, seed) (*KZGRangeProof, error)`
- `NewKZGRangeProofForRangeFromEDS(eds, dah, namespace, shareRange, codec, provider, seed) (*KZGRangeProof, error)`
- `newKZGRangeProofFromEDSWithBounds(...) (*KZGRangeProof, error)`

#### Legacy Functions (to deprecate in core)
- `CreateShareToRowRootProofs(...)` - Legacy NMT-based proof creation
- `errorsNewLegacyUnsupported() error`

### 2.3 celestia-app/pkg/da/ Key Types & Functions

#### DataAvailabilityHeader
- **Fields:**
  - `PieceComm [][]byte` - N*k Kate commitments
  - `ColumnComm [][]byte` - N combined column commitments
  - `NamespaceIndex map[string]Range` - Namespace → share range mapping
  - `hash []byte` - Memoized root hash

#### Range (from go-square)
- **Fields:**
  - `Start int`
  - `End int` (end-exclusive)

#### Key Functions
- `NewDataAvailabilityHeader(eds) (*DataAvailabilityHeader, error)`
- `Hash() []byte` - Computes root from all commitments
- `NamespaceRange(namespaceID) (Range, bool)` - Looks up namespace range

---

## 3. Validation Contract Changes

### 3.1 Current celestia-core ShareProof.Validate()
```
- ShareProofs count == RowProof.RowRoots count
- Total shares in ShareProofs == len(Data)
- All proofs have valid Start/End ranges
- RowProof.Validate(root) passes
```

### 3.2 New celestia-app ShareProof.Validate()
```
- Data is not empty
- ShareProofs is not empty
- Each proof.Proof bytes is not empty
- CommitmentProof is not nil
- CommitmentProof.ColumnProofs is not empty
- len(ColumnProofs) == len(ColumnIndices)
- If root provided and RootCommitment provided, they must match
- VerifyProof() passes
```

---

## 4. Import Dependencies Added in celestia-app

### New Cryptography Libraries
- `github.com/DataAvailabilityLayerNovel/rlnc-rsmt2d` (KZG/CDA framework)
- `github.com/DataAvailabilityLayerNovel/rlnc-rsmt2d/cda` (CDA provider)
- `github.com/DataAvailabilityLayerNovel/rlnc-rsmt2d/rlnc` (RLNC codec)
- `github.com/consensys/gnark-crypto/ecc/bls12-381/kzg` (BLS12-381 KZG)
- `github.com/consensys/gnark-crypto/ecc/bn254` (BN254 curve)
- `github.com/consensys/gnark-crypto/ecc/bn254/fr` (BN254 scalar field)

### Existing Dependencies (also in core)
- `github.com/cometbft/cometbft/crypto/merkle`
- `github.com/celestiaorg/go-square/v4`
- `github.com/celestiaorg/go-square/v4/share`

---

## 5. Test File Structure in celestia-app

Key test files to understand:
- `proof_test.go` - ShareProof unit tests
- `proto_interop_test.go` - Proto marshaling/unmarshaling tests
- `kzg_range_proof_row_batch_test.go` - Row batch verification tests

---

## 6. Function Dependencies Chain

```
NewShareInclusionProof
  ↓
NewShareInclusionProofFromEDS
  ├─→ da.ExtendShares()
  ├─→ da.NewDataAvailabilityHeader()
  ├─→ buildKZGProofContext()
  └─→ NewKZGRangeProofForRangeFromEDS()
      ├─→ dah.NamespaceRange()
      ├─→ merkle.ProofsFromByteSlices()
      └─→ [Internal KZG cell/column proofs]
        └─→ shareProofFromKZGRange()

ShareProof.Validate()
  ├─→ verifyBasicShape()
  └─→ VerifyProof()

ShareProof.VerifyProofWithKZGRange()
  ├─→ verifyBasicShape()
  ├─→ VerifyKZGRangeProof()
  └─→ Consistency checks against rangeProof
```

---

## 7. Breaking Changes Summary

| Aspect | Old (celestia-core) | New (celestia-app) |
|--------|-------------------|------------------|
| ShareProof.ShareProofs | NMTProof[] | KZGMultiProof[] |
| ShareProof.RowProof | RowProof{RowRoots, Proofs, ...} | CommitmentProof{ColumnProofs, ColumnIndices, RootCommitment} |
| Validation Logic | Row-based NMT proofs | Column-based KZG commitments |
| Proof Generation | Uses NMT trees + Merkle trees | Uses KZG commitments + Merkle trees |
| Dependencies | nmt package, basic crypto | rlnc-rsmt2d, gnark-crypto, CDA provider |

