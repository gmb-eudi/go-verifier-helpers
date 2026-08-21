// Package handoffwire defines the wire contract for a result handoff over a
// Valkey key/queue: one producer enqueues an Envelope, one consumer reads and
// delivers it. Stdlib-only by design — bring your own Valkey client.
package handoffwire

import "time"

// Valkey key/queue contract. RespCodeKeyFmt indexes a one-time response_code
// to its session id; the consumer reads it once via GETDEL. Producer and
// consumer never share a key literal beyond these constants.
const (
	QueueKey       = "vc:handoff:queue"
	PayloadKeyFmt  = "vc:handoff:payload:%s"
	RespCodeKeyFmt = "vc:respcode:%s" // response_code -> session id (read once via GETDEL)

	// Key-provider ids shared by producer and consumer. KeyWebhookSigning is
	// the kid a client verifies webhook signatures against (published via JWKS).
	KeyHandoffEnc     = "handoff-enc"
	KeyWebhookSigning = "webhook-signing"
)

// Envelope is the consumption contract carried on the queue. Every field
// EXCEPT ResultJWE is structural/identifying only — no attribute value ever
// appears here. ResultJWE is a compact JWE (ECDH-ES + A256GCM to the
// handoff-enc key) whose plaintext is the Result below; the claim VALUES live
// only inside that ciphertext.
type Envelope struct {
	Version       int       `json:"version"`
	SessionID     string    `json:"session_id"`
	ClientID      string    `json:"client_id"`
	CorrelationID string    `json:"correlation_id"`
	WebhookURL    string    `json:"webhook_url"`
	ResultJWE     string    `json:"result_jwe"`
	EnqueuedAt    time.Time `json:"enqueued_at"`
	ExpiresAt     time.Time `json:"expires_at"`
}

// Result is the decoded plaintext of Envelope.ResultJWE. NOTE: CheckResult's
// check-name wire tag is "name", not "check" — pinned by round-trip tests.
type Result struct {
	SessionID   string             `json:"session_id"`
	Outcome     string             `json:"outcome"`
	Report      *Report            `json:"report"`
	Credentials []ResultCredential `json:"credentials,omitempty"`
}

// ResultCredential carries one credential's disclosed claim VALUES — the
// sole value-bearing type in this contract.
type ResultCredential struct {
	QueryCredentialID string         `json:"query_credential_id"`
	Format            string         `json:"format"`
	DoctypeOrVCT      string         `json:"doctype_or_vct"`
	Claims            map[string]any `json:"claims"`
}

// Report is the value-free verification report (claim NAMES only).
type Report struct {
	SessionID   string              `json:"session_id"`
	Outcome     string              `json:"outcome"` // verified | failed
	FailCode    string              `json:"fail_code,omitempty"`
	Checks      []CheckResult       `json:"checks"`
	Credentials []CredentialSummary `json:"credentials,omitempty"`
	Policy      map[string]bool     `json:"policy"`
}

// CheckResult is one row of the verification report.
type CheckResult struct {
	Check   string `json:"name"`
	Outcome string `json:"outcome"` // pass | fail | skipped
	Code    string `json:"code,omitempty"`
	SpecRef string `json:"spec_ref,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

// CredentialSummary is the value-free per-credential report entry.
type CredentialSummary struct {
	QueryCredID      string   `json:"query_credential_id"`
	Format           string   `json:"format"`
	DoctypeOrVCT     string   `json:"doctype_or_vct"`
	IssuerCountry    string   `json:"issuer_country,omitempty"`
	ClaimNames       []string `json:"claim_names"`
	StatusProvenance string   `json:"status_provenance,omitempty"`
	DecoyDigests     int      `json:"decoy_digests,omitempty"`
}
