# KZG-first Alignment Plan for celestia-core

## 1. Muc tieu

Dong bo celestia-core voi thay doi moi tu celestia-app theo huong KZG-first:
- ShareProof tren wire khong con phu thuoc RowProof/NMTProof.
- DAH proto mo rong bang commitment fields moi.
- Luong RPC va validation trong core nhat quan voi schema moi.

## 2. Pham vi thay doi

### 2.1 Proof wire schema
Nguon thay doi tu app:
- proto KZG-first: `celestia-app/proto/celestia/core/v1/proof/proof.proto`
- runtime validation moi: `celestia-app/pkg/proof/share_proof.go`

Can cap nhat core:
- `celestia-core/proto/tendermint/types/types.proto`
- generated: `celestia-core/proto/tendermint/types/types.pb.go`
- runtime: `celestia-core/types/share_proof.go`
- runtime legacy lien quan: `celestia-core/types/row_proof.go`

### 2.2 DAH proto mo rong
Nguon thay doi tu app:
- `celestia-app/proto/celestia/core/v1/da/data_availability_header.proto`

Can cap nhat core (neu core con trao doi/consume DAH theo proto cu):
- cap nhat message DAH de co:
  - `piece_commitments`
  - `column_commitments`
  - `namespace_index`
- danh dau deprecated:
  - `row_roots`
  - `column_roots`

## 3. Delta mapping (App -> Core)

### 3.1 ShareProof
App moi:
- `share_proofs`: `repeated KZGMultiProof`
- `commitment_proof`: `CommitmentProof`
- bo row-based fields khoi wire contract moi

Core hien tai:
- `share_proofs`: `repeated NMTProof`
- `row_proof`: `RowProof`

Action:
1. Doi schema `ShareProof` trong `types.proto` sang KZG-first.
2. Them message `KZGMultiProof` va `CommitmentProof` trong `types.proto`.
3. Xoa/giu legacy message `RowProof`, `NMTProof` theo chien luoc migration:
   - Option A (clean break): loai bo khoi proto.
   - Option B (compatible phase): giu message cu nhung khong dung trong API moi.

### 3.2 Validation contract
App moi yeu cau:
- `share_proofs` khong rong.
- moi `proof` bytes khong rong.
- `commitment_proof` bat buoc.
- `len(column_proofs) == len(column_indices)`.
- neu co `root_commitment`, phai match root input verifier.

Action:
1. Refactor `celestia-core/types/share_proof.go` theo contract tren.
2. Bo coupling den `RowProof.Validate` trong luong moi.
3. Neu can backward compatibility, tach validate legacy thanh duong rieng (feature flag hoac version gate).

### 3.3 RPC va serialization
Diem can dong bo:
- `celestia-core/rpc/core/tx.go`
- `celestia-core/rpc/core/types/responses.go`
- `celestia-core/rpc/client/http/http.go`
- `celestia-core/rpc/client/local/local.go`

Action:
1. Dam bao marshal/unmarshal `ShareProof` dung shape moi.
2. Giu ten endpoint neu can, nhung payload va semantics theo schema moi.
3. Neu can migration, bo sung endpoint/versioned response (`prove_shares_v3`) de tranh breaking ngay.

## 4. Ke hoach trien khai theo giai doan

### Phase 0 - Chuan bi migration
- Chot migration policy:
  - hard break hay dual-mode.
- Chot version matrix tuong thich:
  - new producer -> new consumer
  - new producer -> old consumer
  - old producer -> new consumer

Deliverable:
- migration note trong release notes.

### Phase 1 - Proto va generated code
- Cap nhat `types.proto` theo KZG-first.
- Regenerate pb go files.
- Sua compile errors lien quan.

Deliverable:
- build xanh o module proto/types.

### Phase 2 - Runtime types + validation
- Refactor `types/share_proof.go`.
- Danh gia lai vai tro `types/row_proof.go`:
  - deprecate/legacy-only.
- Cap nhat convertors ToProto/FromProto.

Deliverable:
- unit tests cho validation contract moi pass.

### Phase 3 - RPC layer
- Cap nhat `prove_shares` path va responses.
- Dam bao backward behavior ro rang (tra loi loi co nghia neu payload cu).

Deliverable:
- integration tests RPC pass.

### Phase 4 - Test matrix va interop
- Them test roundtrip binary payload.
- Them golden vectors:
  - valid payload
  - empty proof
  - mismatched column sizes
  - root mismatch
- Chay matrix version tests.

Deliverable:
- CI co test interop va regressions.

## 5. Danh sach test can cap nhat

Hien tai dang gan chat row/NMT va can doi:
- `celestia-core/types/share_proof_test.go`
- `celestia-core/types/row_proof_test.go`
- `celestia-core/types/row_proof_overflow_test.go`

Them moi:
- `celestia-core/types/share_proof_kzg_test.go`
- `celestia-core/types/proto_interop_test.go` (tuong tu huong app)

## 6. Rui ro va giam thieu

Rui ro:
- Client ben ngoai con parse row-based wire schema.
- Break RPC consumers khong nang cap dong bo.
- Test vectors chua phu edge cases KZG commitments.

Giam thieu:
- Chay dual validation trong giai doan transition.
- Cong bo migration guide ro rang.
- Them negative vectors vao CI.

## 7. Definition of Done

1. `ShareProof` trong core dung KZG-first schema.
2. Validation trong core khop contract moi tu app.
3. RPC prove shares tra payload hop le theo schema moi.
4. Interop tests (marshal/unmarshal + semantic validation) pass.
5. Co migration note cho teams tich hop.

## 8. Thu tu thuc hien de xuat (thuc te)

1. Chinh `types.proto` + regenerate.
2. Refactor `types/share_proof.go`.
3. Dieu chinh RPC parse/response.
4. Sua tests cu + them tests moi.
5. Chay test package lien quan va tong hop ket qua.
