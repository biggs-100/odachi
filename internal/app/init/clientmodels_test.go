package init_test

import (
	"bytes"
	"strings"
	"testing"

	initapp "github.com/biggs-100/odachi/internal/app/init"
)

// dryFiles is the rehearsal's disk: reads are real and every write is dropped,
// the way the CLI's setupFiles answers. Backup reports the copy it would take
// without touching anything, so a test can assert both the plan and the bytes.
type dryFiles struct {
	real *fakeFiles
}

func (f dryFiles) Exists(path string) (bool, error) { return f.real.Exists(path) }
func (f dryFiles) Read(path string) ([]byte, error) { return f.real.Read(path) }
func (f dryFiles) Write(string, []byte) error       { return nil }
func (f dryFiles) Backup(path string) (string, error) {
	return path + ".bak-test", nil
}
func (f dryFiles) Remove(string) error         { return nil }
func (f dryFiles) MakeExecutable(string) error { return nil }

const (
	opencodeConfig = ".config/opencode/opencode.json"
	claudeSettings = ".claude/settings.json"
)

// TestInitWritesTheModelWhereEachClientReadsIt: one flag per client, each file
// keeps what it already had, and the report names both destinations.
func TestInitWritesTheModelWhereEachClientReadsIt(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	files.files[homeFile(home, opencodeConfig)] = []byte("{\n  // my config\n  \"$schema\": \"https://opencode.ai/config.json\"\n}\n")
	files.files[homeFile(home, claudeSettings)] = []byte("{\n  \"permissions\": {\"allow\": [\"Bash\"]}\n}\n")

	got := initapp.Run(initapp.Options{
		Dir:    dir,
		Home:   fakeHome{dir: home},
		Global: true,
		Models: map[string]string{"opencode": "anthropic/claude#fast", "claude": "sonnet"},
	}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}

	openCode := string(files.files[homeFile(home, opencodeConfig)])
	if !strings.Contains(openCode, "// my config") || !strings.Contains(openCode, `"$schema"`) {
		t.Errorf("the edit lost what the file had:\n%s", openCode)
	}
	if !strings.Contains(openCode, `"model": "anthropic/claude#fast"`) {
		t.Errorf("the opencode model is missing:\n%s", openCode)
	}
	claude := string(files.files[homeFile(home, claudeSettings)])
	if !strings.Contains(claude, `"permissions"`) {
		t.Errorf("the claude settings lost their keys:\n%s", claude)
	}
	if !strings.Contains(claude, `"model": "sonnet"`) {
		t.Errorf("the claude model is missing:\n%s", claude)
	}
	for _, want := range []string{"~/" + opencodeConfig, "~/" + claudeSettings, "anthropic/claude#fast", "sonnet"} {
		if !strings.Contains(strings.Join(got.Lines, "\n"), want) {
			t.Errorf("the report does not name %q:\n%s", want, strings.Join(got.Lines, "\n"))
		}
	}
}

// TestAModelIsWrittenForTheClientTheFlagNames: the flag names a client, so
// --agents does not decide where a model goes.
func TestAModelIsWrittenForTheClientTheFlagNames(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	initapp.Run(initapp.Options{
		Dir:    dir,
		Home:   fakeHome{dir: home},
		Global: true,
		Agents: "opencode",
		Models: map[string]string{"claude": "sonnet"},
	}, files, &fakeSkills{})

	claude, ok := files.files[homeFile(home, claudeSettings)]
	if !ok {
		t.Fatal("the claude model was not written")
	}
	if !bytes.Contains(claude, []byte(`"model": "sonnet"`)) {
		t.Errorf("the claude settings do not name the model:\n%s", claude)
	}
	if openCode, ok := files.files[homeFile(home, opencodeConfig)]; ok && bytes.Contains(openCode, []byte(`"model"`)) {
		t.Errorf("a model was written for a client the flag did not name:\n%s", openCode)
	}
}

// TestTheMissingClaudeFileIsTheSmallestDocumentItReads: there is no human layout
// to preserve in a file that does not exist, so the first run writes one line.
func TestTheMissingClaudeFileIsTheSmallestDocumentItReads(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	initapp.Run(initapp.Options{
		Dir:    dir,
		Home:   fakeHome{dir: home},
		Models: map[string]string{"claude": "sonnet"},
	}, files, &fakeSkills{})

	got, ok := files.files[homeFile(home, claudeSettings)]
	if !ok {
		t.Fatal("the claude settings file was not created")
	}
	if want := "{\"model\": \"sonnet\"}\n"; string(got) != want {
		t.Errorf("settings = %q, want %q", got, want)
	}
}

// TestAChangedModelTouchesOnlyTheValueAndBacksUpFirst: the model is one string,
// and the rest of the human's file is not re-encoded.
func TestAChangedModelTouchesOnlyTheValueAndBacksUpFirst(t *testing.T) {
	home := "hogar"
	path := homeFile(home, opencodeConfig)
	original := "{\n  \"model\": \"old/name#a\",\n  \"mcp\": {}\n}\n"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	files.files[path] = []byte(original)

	initapp.Run(initapp.Options{
		Dir:    dir,
		Home:   fakeHome{dir: home},
		Models: map[string]string{"opencode": "new/name#b"},
	}, files, &fakeSkills{})

	want := strings.Replace(original, `"old/name#a"`, `"new/name#b"`, 1)
	if string(files.files[path]) != want {
		t.Errorf("the config moved more than the value:\n%q\nwant\n%q", files.files[path], want)
	}
	backup, write := indexOf(files.events, "backup "+slashed(dir, path)), indexOf(files.events, "write "+slashed(dir, path))
	if backup < 0 || write < 0 || backup > write {
		t.Errorf("events = %v, want a backup before the write", files.events)
	}
}

// TestAModelRerunTouchesNothing: the same flag twice is the same file.
func TestAModelRerunTouchesNothing(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	options := initapp.Options{
		Dir:    dir,
		Home:   fakeHome{dir: home},
		Global: true,
		Models: map[string]string{"opencode": "anthropic/claude#fast", "claude": "sonnet"},
	}

	initapp.Run(options, files, &fakeSkills{})
	files.events = nil
	got := initapp.Run(options, files, &fakeSkills{})

	if len(files.events) != 0 {
		t.Errorf("the second run touched the home: %v", files.events)
	}
	if joined := strings.Join(got.Lines, "\n"); !strings.Contains(joined, "already the model") {
		t.Errorf("the second run does not say the model is already there:\n%s", joined)
	}
}

// TestADryRunNamesTheModelItWouldWriteAndWritesNothing: a rehearsal reports the
// destination and leaves every byte as it was.
func TestADryRunNamesTheModelItWouldWriteAndWritesNothing(t *testing.T) {
	home := "hogar"
	real := newFakeFiles(map[string]string{})
	dry := dryFiles{real: real}

	got := initapp.RunSetup(initapp.Options{
		Home:   fakeHome{dir: home},
		Agents: "both",
		DryRun: true,
		Models: map[string]string{"claude": "sonnet"},
	}, dry, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	if _, exists := real.files[homeFile(home, claudeSettings)]; exists {
		t.Error("a dry run wrote the claude settings")
	}
	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "would write sonnet as the model") {
		t.Errorf("the rehearsal does not name what it would write:\n%s", joined)
	}
	if !slicesContains(got.Files, "~/.claude/settings.json") {
		t.Errorf("Files = %v, want the rehearsal destination", got.Files)
	}
}

// TestAMalformedConfigIsPendedAndKept: a file the editor cannot read is never
// rewritten, and the line names the member to add by hand.
func TestAMalformedConfigIsPendedAndKept(t *testing.T) {
	home := "hogar"
	path := homeFile(home, opencodeConfig)
	original := `{"mcp": {`
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})
	files.files[path] = []byte(original)

	got := initapp.Run(initapp.Options{
		Dir:    dir,
		Home:   fakeHome{dir: home},
		Models: map[string]string{"opencode": "anthropic/claude#fast"},
	}, files, &fakeSkills{})

	if string(files.files[path]) != original {
		t.Errorf("a malformed config was rewritten:\n%s", files.files[path])
	}
	joined := strings.Join(got.Lines, "\n")
	if !strings.Contains(joined, "pending") || !strings.Contains(joined, "by hand") {
		t.Errorf("the report does not say what to do:\n%s", joined)
	}
}

// TestSetupWritesTheModelsTheMachineAskedFor: setup is the machine verb, and the
// same writers answer it.
func TestSetupWritesTheModelsTheMachineAskedFor(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{})

	got := initapp.RunSetup(initapp.Options{
		Home:   fakeHome{dir: home},
		Agents: "both",
		Models: map[string]string{"opencode": "anthropic/claude#fast", "claude": "sonnet"},
	}, files, &fakeSkills{})

	if got.Code != 0 {
		t.Fatalf("Code = %d, want 0 (message %q)", got.Code, got.Message)
	}
	for path, want := range map[string]string{
		homeFile(home, opencodeConfig): `"model": "anthropic/claude#fast"`,
		homeFile(home, claudeSettings): `"model": "sonnet"`,
	} {
		data, ok := files.files[path]
		if !ok {
			t.Errorf("%s was not written", path)
			continue
		}
		if !bytes.Contains(data, []byte(want)) {
			t.Errorf("%s does not carry %s:\n%s", path, want, data)
		}
	}
}

// TestTheModelFlagRefusals: the flag grammar has four ways to be wrong, and each
// one is refused by name before a byte is read.
func TestTheModelFlagRefusals(t *testing.T) {
	cases := []struct {
		name  string
		value string
		wants []string
	}{
		{"an unknown client", "hermes=sonnet", []string{"hermes", "opencode", "claude"}},
		{"a value without an equals sign", "opencode", []string{"client=model"}},
		{"an empty model", "opencode=", []string{"empty"}},
		{"a model that is not an identifier", "opencode=sonnet 4", []string{"identifier"}},
		{"a shell fragment", "claude=x;rm", []string{"identifier"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := initapp.ParseModels([]string{tc.value})
			if err == nil {
				t.Fatalf("ParseModels(%q) accepted it", tc.value)
			}
			for _, want := range tc.wants {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want it to name %q", err, want)
				}
			}
		})
	}
}

// TestParseModelsReadsRepeatedFlags: one occurrence per client, and the map is
// what the run writes.
func TestParseModelsReadsRepeatedFlags(t *testing.T) {
	got, err := initapp.ParseModels([]string{"opencode=anthropic/claude#fast", "claude=sonnet"})
	if err != nil {
		t.Fatalf("ParseModels: %v", err)
	}
	if got["opencode"] != "anthropic/claude#fast" || got["claude"] != "sonnet" {
		t.Errorf("models = %v", got)
	}
}

// TestRunRefusesAnUnknownModelClient: a caller that builds the map directly hits
// the same refusal as the flag.
func TestRunRefusesAnUnknownModelClient(t *testing.T) {
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	got := initapp.Run(initapp.Options{
		Dir:    dir,
		Home:   fakeHome{dir: "hogar"},
		Models: map[string]string{"hermes": "sonnet"},
	}, files, &fakeSkills{})

	if got.Code == 0 {
		t.Fatal("an unknown client was accepted")
	}
	for _, want := range []string{"hermes", "opencode", "claude"} {
		if !strings.Contains(got.Message, want) {
			t.Errorf("Message = %q, want it to name %q", got.Message, want)
		}
	}
}

// TestReadClientModelsAnswersWhatEachFileCarries: the section shows the model
// each client reads today, and the value comes from the file the writer edits
// (0337).
func TestReadClientModelsAnswersWhatEachFileCarries(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{})
	files.files[homeFile(home, opencodeConfig)] = []byte("{\n  // my config\n  \"model\": \"anthropic/claude#fast\"\n}\n")
	files.files[homeFile(home, claudeSettings)] = []byte("{\n  \"model\": \"sonnet\",\n  \"permissions\": {\"allow\": [\"Bash\"]}\n}\n")

	got, _, err := initapp.ReadClientModels(home, files)
	if err != nil {
		t.Fatalf("ReadClientModels: %v", err)
	}
	if got["opencode"] != "anthropic/claude#fast" || got["claude"] != "sonnet" {
		t.Fatalf("models = %v, want the value each file carries", got)
	}
	if len(files.events) != 0 {
		t.Errorf("reading wrote: %v", files.events)
	}
}

// TestReadClientModelsNamesNoModelWhenTheKeyIsAbsent: a client whose config says
// nothing about a model is a row with none, not an empty string in a map.
func TestReadClientModelsNamesNoModelWhenTheKeyIsAbsent(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{})
	files.files[homeFile(home, opencodeConfig)] = []byte("{\"$schema\": \"https://opencode.ai/config.json\"}\n")

	got, _, err := initapp.ReadClientModels(home, files)
	if err != nil {
		t.Fatalf("ReadClientModels: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("models = %v, want none: neither file names one", got)
	}
}

// TestReadClientModelsFindsTheConfigOpenCodeAlreadyHas: an existing .jsonc is the
// human's choice, and the reader looks where the writer writes.
func TestReadClientModelsFindsTheConfigOpenCodeAlreadyHas(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{})
	jsonc := strings.TrimSuffix(opencodeConfig, ".json") + ".jsonc"
	files.files[homeFile(home, jsonc)] = []byte("{\n  // mine\n  \"model\": \"cheap/model\"\n}\n")

	got, _, err := initapp.ReadClientModels(home, files)
	if err != nil {
		t.Fatalf("ReadClientModels: %v", err)
	}
	if got["opencode"] != "cheap/model" {
		t.Fatalf("models = %v, want the value from the .jsonc", got)
	}
}

// TestReadClientModelsNamesAFileItCannotRead: a broken config is named where the
// screen can show it, and does not fail the call: "no model" is a fact about the
// file and "could not read it" is another, and neither is a broken project (0337).
func TestReadClientModelsNamesAFileItCannotRead(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{})
	files.files[homeFile(home, claudeSettings)] = []byte(`{"model": `)

	_, unreadable, err := initapp.ReadClientModels(home, files)
	if err != nil {
		t.Fatalf("ReadClientModels: %v, want the broken file named in its own map", err)
	}
	if !strings.Contains(unreadable["claude"], "settings.json") {
		t.Errorf("unreadable = %v, want the file named", unreadable)
	}
}

// TestTheWriterAndTheReaderAgreeOnTheFilePath: the value a run writes is the
// value the screen reads back, because both halves use the same path (0337).
func TestTheWriterAndTheReaderAgreeOnTheFilePath(t *testing.T) {
	home := "hogar"
	files := newFakeFiles(map[string]string{".git/HEAD": "ref: refs/heads/main\n"})

	initapp.RunSetup(initapp.Options{
		Home:   fakeHome{dir: home},
		Agents: "both",
		Models: map[string]string{"opencode": "anthropic/claude#fast", "claude": "sonnet"},
	}, files, &fakeSkills{})

	got, _, err := initapp.ReadClientModels(home, files)
	if err != nil {
		t.Fatalf("ReadClientModels: %v", err)
	}
	if got["opencode"] != "anthropic/claude#fast" || got["claude"] != "sonnet" {
		t.Fatalf("models = %v, want what the run just wrote", got)
	}
}

// slicesContains answers whether a slice names one destination. It is local to
// this file so the assertions read the same way as the other init tests.
func slicesContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
