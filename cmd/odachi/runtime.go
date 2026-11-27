package main

// This file is the deterministic, opt-in OpenCode runtime-control harness. It is
// never wired to a verb: the ordinary CLI does not call it, and the test gate
// ODACHI_OPENCODE_RUNTIME_TEST=1 is what runs it. It builds a control executable
// in a temp directory, never starts the installed OpenCode binary and never calls
// a live model.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/biggs-100/odachi/internal/adapters/fsx"
	"github.com/biggs-100/odachi/internal/adapters/opencode"
	grantapp "github.com/biggs-100/odachi/internal/app/grant"
	verifyapp "github.com/biggs-100/odachi/internal/app/verify"
	domaingrant "github.com/biggs-100/odachi/internal/domain/grant"
)

const (
	// runtimeControlLog is where the control executable records one line per
	// spawn, beside itself.
	runtimeControlLog = "spawns.log"

	// The two canonical hashes the managed verifier binds. They name bytes the
	// harness owns; the receipt validates their shape, not this fixture.
	runtimeCandidateTree = "0123456789abcdef0123456789abcdef01234567"
	runtimeCandidateBase = "fedcba9876543210fedcba9876543210fedcba98"
)

// RuntimeControl names the deterministic control executable the harness spawns.
// The executable writes its spawn log and counter beside itself, so one temp
// directory isolates one proof run.
type RuntimeControl struct {
	Executable string
}

// RuntimeProof is what the harness observed: how many times each grant spawned
// the control executable, which sessions it received, whether a replay was
// refused without a spawn, whether a moved candidate was refused, and every argv
// the control executable saw.
type RuntimeProof struct {
	SpawnsPerGrant  []int
	Sessions        []string
	ReplayRefused   bool
	ReplaySpawns    int
	MutationBefore  string
	MutationAfter   string
	MutationRefused bool
	MutationSealed  bool
	Argv            [][]string
}

// RunOpenCodeRuntimeProofs runs the five runtime controls R1 promises over real
// adapters, a real durable grant store and a frozen candidate, with the control
// executable standing in for OpenCode. It returns an error when a proof could not
// be observed; the caller asserts the facts in RuntimeProof.
func RunOpenCodeRuntimeProofs(ctx context.Context, control RuntimeControl, workRoot string) (RuntimeProof, error) {
	if strings.TrimSpace(control.Executable) == "" {
		return RuntimeProof{}, errors.New("the runtime control names no executable")
	}
	if strings.TrimSpace(workRoot) == "" {
		return RuntimeProof{}, errors.New("the runtime proof has no work directory")
	}
	if err := os.MkdirAll(workRoot, 0o755); err != nil {
		return RuntimeProof{}, fmt.Errorf("prepare the runtime proof work directory: %w", err)
	}

	source := filepath.Join(workRoot, "candidate")
	if err := writeRuntimeCandidate(source); err != nil {
		return RuntimeProof{}, err
	}

	store, err := grantapp.NewFileStore(filepath.Join(workRoot, "grant-store"))
	if err != nil {
		return RuntimeProof{}, err
	}
	service := grantapp.NewService(store)
	adapter := opencode.New(opencode.ExecRunner{Binary: control.Executable})
	clock := &runtimeClock{now: time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)}
	freezer := &runtimeFreezer{source: source, root: filepath.Join(workRoot, "freezes")}

	request := verifyapp.ManagedVerify{
		Dir:      workRoot,
		Tree:     runtimeCandidateTree,
		Base:     runtimeCandidateBase,
		Runtime:  "opencode",
		Model:    "control/planned",
		Verifier: "odachi-verify",
		Prompt:   "verify the frozen candidate",
	}
	deps := verifyapp.ManagedVerifyDeps{
		Freezer:   freezer,
		Grants:    service,
		Sessions:  runtimeSessions{adapter: adapter},
		Executor:  runtimeExecutor{},
		Snapshots: runtimeSnapshots{},
		Clock:     clock,
	}

	proof := RuntimeProof{}
	var firstGrantID domaingrant.ID
	for run := 1; run <= 2; run++ {
		before, err := runtimeSpawnCount(control.logPath())
		if err != nil {
			return proof, err
		}
		result := verifyapp.RunManagedVerify(ctx, request, deps)
		if result.Code != verifyapp.CodeOK || result.Sealed == nil {
			return proof, fmt.Errorf("managed verify %d was refused: %s", run, result.Message)
		}
		after, err := runtimeSpawnCount(control.logPath())
		if err != nil {
			return proof, err
		}
		proof.SpawnsPerGrant = append(proof.SpawnsPerGrant, after-before)
		proof.Sessions = append(proof.Sessions, string(result.Sealed.Session))
		if run == 1 {
			firstGrantID = result.Sealed.Grant
		}
	}

	beforeReplay, err := runtimeSpawnCount(control.logPath())
	if err != nil {
		return proof, err
	}
	_, replayErr := service.Claim(ctx, firstGrantID)
	afterReplay, err := runtimeSpawnCount(control.logPath())
	if err != nil {
		return proof, err
	}
	proof.ReplayRefused = replayErr != nil
	proof.ReplaySpawns = afterReplay - beforeReplay
	if !proof.ReplayRefused {
		return proof, errors.New("replaying the completed grant was not refused")
	}

	mutating := &runtimeMutatingExecutor{}
	mutationDeps := deps
	mutationDeps.Executor = mutating
	beforeMutation, err := runtimeSpawnCount(control.logPath())
	if err != nil {
		return proof, err
	}
	mutation := verifyapp.RunManagedVerify(ctx, request, mutationDeps)
	afterMutation, err := runtimeSpawnCount(control.logPath())
	if err != nil {
		return proof, err
	}
	proof.SpawnsPerGrant = append(proof.SpawnsPerGrant, afterMutation-beforeMutation)
	proof.MutationBefore = mutating.before
	proof.MutationAfter = mutating.after
	proof.MutationRefused = mutation.Code != verifyapp.CodeOK && mutation.Sealed == nil
	proof.MutationSealed = mutation.Sealed != nil

	proof.Argv, err = runtimeSpawnArgv(control.logPath())
	if err != nil {
		return proof, err
	}
	return proof, nil
}

// logPath is the spawn log the control executable writes beside itself.
func (control RuntimeControl) logPath() string {
	return filepath.Join(filepath.Dir(control.Executable), runtimeControlLog)
}

// BuildRuntimeControl writes and builds the deterministic control executable in
// dir. It compiles one stdlib-only program with an isolated module, so the
// harness never depends on the repository's module graph or on an installed
// OpenCode binary.
func BuildRuntimeControl(ctx context.Context, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("prepare the runtime control directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module odachi-runtime-control\n\ngo 1.20\n"), 0o644); err != nil {
		return "", fmt.Errorf("write the runtime control module: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(runtimeControlProgram), 0o644); err != nil {
		return "", fmt.Errorf("write the runtime control program: %w", err)
	}

	executable := filepath.Join(dir, "opencode-control")
	if runtime.GOOS == "windows" {
		executable += ".exe"
	}
	command := exec.CommandContext(ctx, "go", "build", "-o", executable, ".")
	command.Dir = dir
	command.Env = append(os.Environ(), "GO111MODULE=on", "GOWORK=off", "CGO_ENABLED=0", "GOFLAGS=")
	if output, err := command.CombinedOutput(); err != nil {
		return "", fmt.Errorf("build the runtime control executable: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return executable, nil
}

// runtimeSessions turns the fresh-only OpenCode adapter into the verifier's
// session boundary. It has no resume or fork method to call.
type runtimeSessions struct {
	adapter *opencode.Adapter
}

func (sessions runtimeSessions) CreateFreshSession(ctx context.Context) (verifyapp.VerifierSession, error) {
	session, err := sessions.adapter.CreateFreshSession(ctx)
	if err != nil {
		return verifyapp.VerifierSession{}, err
	}
	return verifyapp.VerifierSession{ID: session.ID, Model: session.Model.String()}, nil
}

// runtimeFreezer copies the fixture into a private run directory. The verifier
// never receives the worktree itself, so a mutation proof cannot damage it.
type runtimeFreezer struct {
	source string
	root   string
	next   int
}

func (freezer *runtimeFreezer) ArchiveRun(_, _ string) (string, string, error) {
	freezer.next++
	run := filepath.Join(freezer.root, fmt.Sprintf("run-%d", freezer.next))
	candidate := filepath.Join(run, "candidate")
	if err := copyRuntimeTree(freezer.source, candidate); err != nil {
		return "", "", err
	}
	return run, candidate, nil
}

func (freezer *runtimeFreezer) Discard(dir string) error {
	return os.RemoveAll(dir)
}

// runtimeSnapshots is the content identity the managed verifier compares before
// and after the run.
type runtimeSnapshots struct{}

func (runtimeSnapshots) Digest(dir string) (string, error) {
	return fsx.Snapshot(dir)
}

// runtimeClock is a deterministic clock: each read advances one second, so a
// receipt's start and finish are ordered and reproducible.
type runtimeClock struct {
	now time.Time
}

func (clock *runtimeClock) Now() time.Time {
	current := clock.now
	clock.now = clock.now.Add(time.Second)
	return current
}

// runtimeExecutor answers successfully without touching the candidate.
type runtimeExecutor struct{}

func (runtimeExecutor) Execute(context.Context, verifyapp.VerifierExecutionRequest) (verifyapp.VerifierExecution, error) {
	return verifyapp.VerifierExecution{Output: "verified\n", ExitCode: 0}, nil
}

// runtimeMutatingExecutor changes the frozen candidate before returning success.
// It records both snapshots, so the proof shows the change and the verifier's
// refusal to seal it.
type runtimeMutatingExecutor struct {
	before string
	after  string
}

func (executor *runtimeMutatingExecutor) Execute(_ context.Context, request verifyapp.VerifierExecutionRequest) (verifyapp.VerifierExecution, error) {
	before, err := fsx.Snapshot(request.Dir)
	if err != nil {
		return verifyapp.VerifierExecution{}, err
	}
	target := filepath.Join(request.Dir, "internal", "candidate.txt")
	if err := os.WriteFile(target, []byte("mutated candidate\n"), 0o644); err != nil {
		return verifyapp.VerifierExecution{}, err
	}
	after, err := fsx.Snapshot(request.Dir)
	if err != nil {
		return verifyapp.VerifierExecution{}, err
	}
	executor.before, executor.after = before, after
	return verifyapp.VerifierExecution{Output: "mutated\n", ExitCode: 0}, nil
}

// writeRuntimeCandidate creates the small frozen fixture the verifier reads.
func writeRuntimeCandidate(root string) error {
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		return fmt.Errorf("prepare the runtime candidate: %w", err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "candidate.txt"), []byte("stable candidate\n"), 0o644); err != nil {
		return fmt.Errorf("write the runtime candidate: %w", err)
	}
	return nil
}

// copyRuntimeTree copies the regular files and directories of a frozen fixture.
// A symlink or another special entry is refused instead of followed.
func copyRuntimeTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, item fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, rel)
		if item.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !item.Type().IsRegular() {
			return fmt.Errorf("the runtime fixture entry %s is not a regular file or a directory", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// runtimeSpawnCount is how many times the control executable has been spawned.
func runtimeSpawnCount(path string) (int, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read the runtime spawn log: %w", err)
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return 0, nil
	}
	return len(strings.Split(text, "\n")), nil
}

// runtimeSpawnArgv is every argv the control executable recorded, in order.
func runtimeSpawnArgv(path string) ([][]string, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read the runtime spawn log: %w", err)
	}
	var calls [][]string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var argv []string
		if err := json.Unmarshal([]byte(line), &argv); err != nil {
			return nil, fmt.Errorf("read one runtime spawn record: %w", err)
		}
		calls = append(calls, argv)
	}
	return calls, nil
}

// runtimeControlProgram is the deterministic stand-in for OpenCode. It accepts
// exactly the fresh-session request and refuses any other argv, records one line
// per spawn, mints a distinct session identity from a counter, and reports a
// model. It is stdlib-only and is built into a temp directory by the harness.
const runtimeControlProgram = `package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	expected := []string{"api", "post", "/api/session", "--data", "{}"}
	args := os.Args[1:]
	if len(args) != len(expected) {
		fail("unexpected argv: " + strings.Join(args, " "))
	}
	for index := range expected {
		if args[index] != expected[index] {
			fail("unexpected argv: " + strings.Join(args, " "))
		}
	}

	executable, err := os.Executable()
	if err != nil {
		fail(err.Error())
	}
	base := filepath.Dir(executable)
	if err := appendSpawn(filepath.Join(base, "spawns.log"), args); err != nil {
		fail(err.Error())
	}
	count, err := nextCount(filepath.Join(base, "counter"))
	if err != nil {
		fail(err.Error())
	}
	id := fmt.Sprintf("ses_control_%d", count)
	model := fmt.Sprintf("model-%d", count)
	fmt.Printf("{\"data\":{\"id\":%q,\"model\":{\"providerID\":\"control\",\"id\":%q}}}", id, model)
}

func appendSpawn(path string, args []string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	line, err := json.Marshal(args)
	if err != nil {
		return err
	}
	_, err = file.Write(append(line, '\n'))
	return err
}

func nextCount(path string) (int, error) {
	count := 0
	data, err := os.ReadFile(path)
	switch {
	case err == nil:
		text := strings.TrimSpace(string(data))
		if text != "" {
			count, err = strconv.Atoi(text)
			if err != nil {
				return 0, err
			}
		}
	case !os.IsNotExist(err):
		return 0, err
	}
	count++
	if err := os.WriteFile(path, []byte(strconv.Itoa(count)), 0o600); err != nil {
		return 0, err
	}
	return count, nil
}

func fail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(2)
}
`
