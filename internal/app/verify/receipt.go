package verify

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
	domainreview "github.com/biggs-100/odachi/internal/domain/review"
)

// The canonical runtime receipt: one object that says who ran, under which
// one-shot grant, in which fresh runtime session, over which bytes, and how it
// ended. Every field is a fact the tool observed or issued; a field a model wrote
// about itself is not one of them.
//
// The receipt is the seal of a managed run. It is validated fail-closed before it
// is stamped into a verification record or exported, and a field the tool could
// not observe makes the whole receipt invalid instead of absent.

// maxReceiptSession is the domain's byte bound for a session identity. The
// grammar below mirrors internal/domain/grant, which keeps its validator
// unexported.
const maxReceiptSession = 256

// maxReceiptRuntime bounds a runtime name. A runtime is a small identifier, not
// free text.
const maxReceiptRuntime = 64

// Outcome is the terminal fact a sealed runtime receipt carries. A run that did
// not complete seals nothing, so the only sealed outcome is completed.
type Outcome string

// OutcomeCompleted says the actor ran to a zero exit over a stable snapshot under
// one bound grant and session.
const OutcomeCompleted Outcome = "completed"

// SealedRuntimeReceipt is the tool-derived runtime fact of one managed run. Every
// field is observed or issued by Go; none of them is read from the actor's answer.
// Tree is the candidate the actor was handed, Base the candidate's base, and
// OutputTree the tree the run left behind.
type SealedRuntimeReceipt struct {
	Grant        domaingrant.ID
	Session      domaingrant.SessionID
	Runtime      string
	Model        string
	SessionModel string
	Role         string
	Tree         string
	Base         string
	OutputTree   string
	Snapshot     string
	StartedAt    string
	FinishedAt   string
	Outcome      Outcome
}

// receiptWire is the canonical JSON spelling. The typed identities are strings
// here so a receipt reads and writes as one stable object.
type receiptWire struct {
	Grant        string `json:"grant"`
	Session      string `json:"session"`
	Runtime      string `json:"runtime"`
	Model        string `json:"model"`
	SessionModel string `json:"session_model"`
	Role         string `json:"role"`
	Tree         string `json:"tree"`
	Base         string `json:"base"`
	OutputTree   string `json:"output_tree"`
	Snapshot     string `json:"snapshot"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at"`
	Outcome      string `json:"outcome"`
}

// MarshalJSON writes the canonical spelling of every fact.
func (receipt SealedRuntimeReceipt) MarshalJSON() ([]byte, error) {
	return json.Marshal(receiptWire{
		Grant:        receipt.Grant.String(),
		Session:      string(receipt.Session),
		Runtime:      receipt.Runtime,
		Model:        receipt.Model,
		SessionModel: receipt.SessionModel,
		Role:         receipt.Role,
		Tree:         receipt.Tree,
		Base:         receipt.Base,
		OutputTree:   receipt.OutputTree,
		Snapshot:     receipt.Snapshot,
		StartedAt:    receipt.StartedAt,
		FinishedAt:   receipt.FinishedAt,
		Outcome:      string(receipt.Outcome),
	})
}

// UnmarshalJSON reads one strict receipt. The grant is parsed into its canonical
// identity before anything else, and the parsed receipt is validated, so a
// receipt that reaches a caller is already a sealed fact or an error.
func (receipt *SealedRuntimeReceipt) UnmarshalJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var wire receiptWire
	if err := decoder.Decode(&wire); err != nil {
		return fmt.Errorf("the runtime receipt is not valid JSON: %s", oneLine(err.Error()))
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("the runtime receipt is not a single JSON object: seal one receipt")
	}

	if wire.Grant == "" {
		return errors.New("the runtime receipt carries no grant: a receipt that names no one-shot authority attests nothing")
	}
	grant, err := domaingrant.ParseID(wire.Grant)
	if err != nil {
		return fmt.Errorf("the runtime receipt names %q as its grant, which is not a Go-issued identity", wire.Grant)
	}
	*receipt = SealedRuntimeReceipt{
		Grant:        grant,
		Session:      domaingrant.SessionID(wire.Session),
		Runtime:      wire.Runtime,
		Model:        wire.Model,
		SessionModel: wire.SessionModel,
		Role:         wire.Role,
		Tree:         wire.Tree,
		Base:         wire.Base,
		OutputTree:   wire.OutputTree,
		Snapshot:     wire.Snapshot,
		StartedAt:    wire.StartedAt,
		FinishedAt:   wire.FinishedAt,
		Outcome:      Outcome(wire.Outcome),
	}
	return receipt.Validate()
}

// Validate refuses a receipt the tool cannot stand behind: a missing fact, an
// identity that is not safe to store, a tree or snapshot that is not canonical, a
// clock that runs backwards, a role or outcome outside the closed vocabulary, or
// two models from different providers in one run.
func (receipt SealedRuntimeReceipt) Validate() error {
	if err := receipt.Grant.Validate(); err != nil {
		return errors.New("the runtime receipt carries no grant: a receipt that names no one-shot authority attests nothing")
	}
	if receipt.Session == "" {
		return errors.New("the runtime receipt carries no session: a receipt that names no fresh runtime session attests nothing")
	}
	if !safeSessionID(string(receipt.Session)) {
		return fmt.Errorf("the runtime receipt names %q as its session, which is not a safe identity", receipt.Session)
	}
	if receipt.Runtime == "" {
		return errors.New("the runtime receipt carries no runtime: a receipt that names no runtime attests nothing")
	}
	if !canonicalRuntime(receipt.Runtime) {
		return fmt.Errorf("the runtime receipt names %q as its runtime, which is not a canonical runtime name", receipt.Runtime)
	}
	if receipt.Model == "" {
		return errors.New("the runtime receipt carries no planned model: a receipt that names no model attests nothing")
	}
	if err := canonicalModel(receipt.Model); err != nil {
		return fmt.Errorf("the runtime receipt names %q as its planned model, which is not a model identity", receipt.Model)
	}
	if receipt.SessionModel == "" {
		return errors.New("the runtime receipt carries no observed session model: a receipt that names no observed model attests nothing")
	}
	if err := canonicalModel(receipt.SessionModel); err != nil {
		return fmt.Errorf("the runtime receipt names %q as its observed session model, which is not a model identity", receipt.SessionModel)
	}
	if receipt.Role != VerifierRole {
		return fmt.Errorf("the runtime receipt names %q as its role; the known roles are %q", receipt.Role, VerifierRole)
	}
	if receipt.Tree == "" {
		return errors.New("the runtime receipt carries no candidate tree: a receipt that names no bytes attests nothing")
	}
	if !canonicalTreeHash(receipt.Tree) {
		return fmt.Errorf("the runtime receipt names %q as its candidate tree, which is not a canonical tree hash", receipt.Tree)
	}
	if receipt.Base == "" {
		return errors.New("the runtime receipt carries no candidate base: the tree is read against a base")
	}
	if !canonicalTreeHash(receipt.Base) {
		return fmt.Errorf("the runtime receipt names %q as its candidate base, which is not a canonical tree hash", receipt.Base)
	}
	if receipt.OutputTree == "" {
		return errors.New("the runtime receipt carries no output tree: a receipt that names no output attests nothing")
	}
	if !canonicalTreeHash(receipt.OutputTree) {
		return fmt.Errorf("the runtime receipt names %q as its output tree, which is not a canonical tree hash", receipt.OutputTree)
	}
	if receipt.Snapshot == "" {
		return errors.New("the runtime receipt carries no snapshot: a receipt that names no content identity attests nothing")
	}
	if !canonicalSnapshot(receipt.Snapshot) {
		return fmt.Errorf("the runtime receipt names %q as its snapshot digest, which is not a SHA-256 content identity", receipt.Snapshot)
	}
	if receipt.StartedAt == "" {
		return errors.New("the runtime receipt carries no start: the tool seals both instants")
	}
	started, err := canonicalInstant(receipt.StartedAt)
	if err != nil {
		return fmt.Errorf("the runtime receipt says started_at %q, which is not canonical RFC3339 UTC", receipt.StartedAt)
	}
	if receipt.FinishedAt == "" {
		return errors.New("the runtime receipt carries no finish: the tool seals both instants")
	}
	finished, err := canonicalInstant(receipt.FinishedAt)
	if err != nil {
		return fmt.Errorf("the runtime receipt says finished_at %q, which is not canonical RFC3339 UTC", receipt.FinishedAt)
	}
	if finished.Before(started) {
		return fmt.Errorf("the runtime receipt finished at %s before it started at %s: a run cannot end before it begins", receipt.FinishedAt, receipt.StartedAt)
	}
	if receipt.Outcome != OutcomeCompleted {
		return fmt.Errorf("the runtime receipt names %q as its outcome; the only sealed outcome is %q", receipt.Outcome, OutcomeCompleted)
	}
	if receipt.Role == VerifierRole && receipt.OutputTree != receipt.Tree {
		return fmt.Errorf("the runtime receipt names the output tree %q and the candidate tree %q: a verifier freezes one tree and cannot move it", receipt.OutputTree, receipt.Tree)
	}
	if planned, observed := modelProvider(receipt.Model), modelProvider(receipt.SessionModel); planned != "" && observed != "" && planned != observed {
		return fmt.Errorf("the runtime receipt names the planned model's provider %q and the observed session model's provider %q in one run: a receipt does not span another provider", planned, observed)
	}
	return nil
}

// ParseRuntimeReceipt reads one receipt and refuses anything that is not a sealed
// fact. Decoding is strict: an unknown field, a second object or prose is not a
// receipt.
func ParseRuntimeReceipt(data []byte) (SealedRuntimeReceipt, error) {
	text := strings.TrimSpace(string(data))
	if text == "" {
		return SealedRuntimeReceipt{}, errors.New("the runtime receipt is empty: there is nothing to read")
	}
	var receipt SealedRuntimeReceipt
	if err := json.Unmarshal([]byte(text), &receipt); err != nil {
		return SealedRuntimeReceipt{}, err
	}
	return receipt, nil
}

// RenderRuntimeReceipt writes the canonical receipt: indented JSON with a
// trailing newline. Two renders of the same receipt are the same bytes.
func RenderRuntimeReceipt(receipt SealedRuntimeReceipt) ([]byte, error) {
	if err := receipt.Validate(); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		// Unreachable: every field is a string or a typed identity with a string
		// spelling.
		return nil, errors.New("could not encode the runtime receipt")
	}
	return append(data, '\n'), nil
}

// safeSessionID mirrors the grant domain's session grammar: a strict ASCII
// identity, so a format control or a space can never reach a path or a log line.
func safeSessionID(value string) bool {
	if len(value) == 0 || len(value) > maxReceiptSession {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '_', r == '.', r == ':':
		default:
			return false
		}
	}
	return true
}

// canonicalRuntime accepts a lowercase runtime identifier: it starts with a
// letter and continues with letters, digits, dashes, underscores or dots.
func canonicalRuntime(value string) bool {
	if len(value) == 0 || len(value) > maxReceiptRuntime {
		return false
	}
	for i, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case i > 0 && (r >= '0' && r <= '9' || r == '-' || r == '_' || r == '.'):
		default:
			return false
		}
	}
	return true
}

// canonicalModel accepts exactly the model identity a grant records, without
// surrounding whitespace, and refuses shell code the way {model} does.
func canonicalModel(value string) error {
	if value != strings.TrimSpace(value) {
		return errors.New("the model identity carries surrounding whitespace")
	}
	_, err := domainreview.NormalizeModel(value)
	return err
}

// canonicalTreeHash accepts the hex git writes for a tree or a commit: 40 or 64
// lowercase characters, whole.
func canonicalTreeHash(value string) bool {
	return canonicalHex(value, 40, 64)
}

// canonicalSnapshot accepts the content identity the snapshot reader returns: a
// SHA-256 digest in lowercase hex.
func canonicalSnapshot(value string) bool {
	return canonicalHex(value, 64)
}

func canonicalHex(value string, lengths ...int) bool {
	matched := false
	for _, length := range lengths {
		if len(value) == length {
			matched = true
			break
		}
	}
	if !matched {
		return false
	}
	for _, r := range value {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

// canonicalInstant parses the instant the tool sealed and refuses any other
// spelling of it: UTC, RFC3339, no fraction.
func canonicalInstant(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, err
	}
	if parsed.UTC().Format(time.RFC3339) != value {
		return time.Time{}, fmt.Errorf("%s is not the canonical UTC spelling", value)
	}
	return parsed, nil
}

// modelProvider is the provider part of a model identity, or empty when the value
// names no provider and cannot be compared.
func modelProvider(model string) string {
	identity := domainreview.ModelIdentity(model)
	provider, _, found := strings.Cut(identity, "/")
	if !found {
		return ""
	}
	return provider
}
