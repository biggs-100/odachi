package undo_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/biggs-100/odachi/internal/app/undo"
)

// dir is the throwaway project every test runs in. Paths inside it are built
// with filepath.Join, because the app touches the disk through the port and the
// port speaks the platform's separator.
const dir = "proyecto"

// The four artifacts, relative to the project root, in the shape the messages
// show them.
const (
	briefRel   = "brief.md"
	profileRel = ".odachi/profile.toml"
	fenceRel   = ".odachi/fence.toml"
	taskRel    = "odd/tasks/0001-reserva-de-turno.md"
)

// artifacts are the only paths undo may read, back up or write.
var artifacts = []string{briefRel, profileRel, fenceRel, taskRel}

// at is a fixed date so ordering never depends on the clock.
func at(day, hour, minute int) time.Time {
	return time.Date(2026, 10, day, hour, minute, 0, 0, time.UTC)
}

// digest is the name portion fsx.Backup appends: the first eight hex of the
// content's SHA-256. The fake mirrors it so the tests can predict the message.
func digest(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])[:8]
}

// backupName is the path fsx.Backup would write for content at path.
func backupName(path, content string) string {
	return path + ".bak-" + digest(content)
}

// storeName is where a new backup of rel lives under the project store.
func storeName(rel, content string) string {
	return backupName(filepath.ToSlash(filepath.Join(".odachi", "backups", filepath.FromSlash(rel))), content)
}

// fakeFiles is the disk in memory with a ledger: a test asserts what was read,
// backed up and written, and in which order.
type fakeFiles struct {
	files map[string][]byte
	mod   map[string]time.Time
	clock time.Time

	order   []string
	wrote   []string
	backups []string
	reads   []string
	touched []string
}

func newFakeFiles() *fakeFiles {
	return &fakeFiles{
		files: map[string][]byte{},
		mod:   map[string]time.Time{},
		clock: at(22, 0, 0),
	}
}

// put seeds a file with its date.
func (f *fakeFiles) put(rel, content string, when time.Time) *fakeFiles {
	path := filepath.Join(dir, filepath.FromSlash(rel))
	f.files[path] = []byte(content)
	f.mod[path] = when
	return f
}

// tick advances the fake clock, so a backup taken now is newer than any seeded
// one.
func (f *fakeFiles) tick() time.Time {
	f.clock = f.clock.Add(time.Minute)
	return f.clock
}

func (f *fakeFiles) Exists(path string) (bool, error) {
	f.touched = append(f.touched, path)
	_, ok := f.files[path]
	return ok, nil
}

func (f *fakeFiles) List(listDir string) ([]string, error) {
	f.touched = append(f.touched, listDir)
	var names []string
	for path := range f.files {
		if filepath.Dir(path) == listDir {
			names = append(names, filepath.Base(path))
		}
	}
	slices.Sort(names)
	return names, nil
}

func (f *fakeFiles) Read(path string) ([]byte, error) {
	f.reads = append(f.reads, path)
	f.touched = append(f.touched, path)
	data, ok := f.files[path]
	if !ok {
		return nil, fmt.Errorf("archivo inexistente: %s", path)
	}
	return data, nil
}

func (f *fakeFiles) Write(path string, data []byte) error {
	f.wrote = append(f.wrote, path)
	f.order = append(f.order, "write "+path)
	f.touched = append(f.touched, path)
	f.files[path] = append([]byte(nil), data...)
	f.mod[path] = f.tick()
	return nil
}

func (f *fakeFiles) Remove(path string) error {
	f.order = append(f.order, "remove "+path)
	f.touched = append(f.touched, path)
	if _, ok := f.files[path]; !ok {
		return fmt.Errorf("archivo inexistente: %s", path)
	}
	delete(f.files, path)
	delete(f.mod, path)
	return nil
}

func (f *fakeFiles) Rename(from, to string) error {
	f.order = append(f.order, "rename "+from+" -> "+to)
	f.touched = append(f.touched, from, to)
	data, ok := f.files[from]
	if !ok {
		return fmt.Errorf("archivo inexistente: %s", from)
	}
	if _, exists := f.files[to]; exists {
		return fmt.Errorf("ya existe: %s", to)
	}
	f.files[to] = append([]byte(nil), data...)
	f.mod[to] = f.mod[from]
	delete(f.files, from)
	delete(f.mod, from)
	return nil
}

func (f *fakeFiles) Backup(path string) (string, error) {
	data, ok := f.files[path]
	if !ok {
		return "", fmt.Errorf("no se pudo leer %s para respaldarlo", path)
	}
	backup := backupName(path, string(data))
	f.backups = append(f.backups, path)
	f.order = append(f.order, "backup "+path)
	f.touched = append(f.touched, path)
	f.files[backup] = append([]byte(nil), data...)
	f.mod[backup] = f.tick()
	return backup, nil
}

func (f *fakeFiles) ModTime(path string) (time.Time, error) {
	f.touched = append(f.touched, path)
	when, ok := f.mod[path]
	if !ok {
		return time.Time{}, fmt.Errorf("archivo inexistente: %s", path)
	}
	return when, nil
}

// content reads a file the fake holds, failing when it is not there.
func (f *fakeFiles) content(t *testing.T, rel string) string {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	data, ok := f.files[path]
	if !ok {
		t.Fatalf("%s no existe; archivos: %v", rel, f.paths())
	}
	return string(data)
}

func (f *fakeFiles) paths() []string {
	out := make([]string, 0, len(f.files))
	for path := range f.files {
		out = append(out, path)
	}
	slices.Sort(out)
	return out
}

// assertOnlyTheFourArtifacts is the boundary Contract X draws around undo: no
// path outside the allow-list is read, backed up or written.
func (f *fakeFiles) assertOnlyTheFourArtifacts(t *testing.T) {
	t.Helper()
	allowed := func(path string) bool {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return false
		}
		rel = filepath.ToSlash(rel)
		for _, artifact := range artifacts {
			if rel == artifact || strings.HasPrefix(rel, artifact+".bak-") {
				return true
			}
		}
		// Listing the directories that hold the artifacts is how the backups
		// are found; the directories themselves are not artifacts. The store
		// is undo's own directory now, and its subdirectories mirror the
		// artifact paths.
		if rel == ".odachi/backups" || strings.HasPrefix(rel, ".odachi/backups/") {
			return true
		}
		return rel == "." || rel == ".odachi" || rel == "odd/tasks"
	}
	for _, path := range f.touched {
		if !allowed(path) {
			t.Errorf("undo touched %q, which is outside the four artifacts", path)
		}
	}
}

// TestRestoreBacksUpTheCurrentStateFirst is the rule the whole safety story
// rests on: the state being replaced is saved before it is replaced, so the undo
// is itself undoable and a hand edit never disappears.
func TestRestoreBacksUpTheCurrentStateFirst(t *testing.T) {
	const (
		written = "lo que escribió odachi\n"
		hand    = "el humano lo editó a mano\n"
	)

	files := newFakeFiles()
	files.put(briefRel, hand, at(21, 23, 0))
	files.put(briefRel+".bak-"+digest(written), written, at(21, 20, 31))

	res := undo.Run(undo.Options{Dir: dir}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}

	briefPath := filepath.Join(dir, briefRel)
	if want := []string{"backup " + briefPath, "write " + briefPath}; !slices.Equal(files.order, want) {
		t.Errorf("order = %v, want %v (backup first, then the copy)", files.order, want)
	}

	if got, want := files.content(t, briefRel), written; got != want {
		t.Errorf("%s = %q, want the restored bytes %q", briefRel, got, want)
	}

	saved := briefRel + ".bak-" + digest(hand)
	if got := files.content(t, saved); got != hand {
		t.Errorf("the hand edit was not saved in %s: %q", saved, got)
	}

	want := "restauré " + briefRel + " desde " + briefRel + ".bak-" + digest(written) +
		"; lo que había quedado en " + saved
	if res.Message != want {
		t.Errorf("Message = %q, want %q", res.Message, want)
	}
	if !strings.Contains(strings.Join(res.Lines, "\n"), "git revert <sha>") {
		t.Errorf("Lines = %v; the undo output must say that code is git revert <sha>", res.Lines)
	}
	if len(res.Files) != 0 && !slices.Contains(res.Files, briefRel) {
		t.Errorf("Files = %v, want %s in it", res.Files, briefRel)
	}

	files.assertOnlyTheFourArtifacts(t)
}

// TestUndoingTheUndoPutsTheHandEditBack: the previous test left the hand edit in
// a backup; a second undo has to find it, because it is now the newest one.
func TestUndoingTheUndoPutsTheHandEditBack(t *testing.T) {
	const (
		written = "lo que escribió odachi\n"
		hand    = "el humano lo editó a mano\n"
	)

	files := newFakeFiles()
	files.put(briefRel, hand, at(21, 23, 0))
	files.put(briefRel+".bak-"+digest(written), written, at(21, 20, 31))

	if res := undo.Run(undo.Options{Dir: dir}, files, nil); res.Code != undo.CodeOK {
		t.Fatalf("first undo: Code = %d (message: %s)", res.Code, res.Message)
	}

	second := undo.Run(undo.Options{Dir: dir}, files, nil)
	if second.Code != undo.CodeOK {
		t.Fatalf("second undo: Code = %d (message: %s)", second.Code, second.Message)
	}
	if got := files.content(t, briefRel); got != hand {
		t.Errorf("%s = %q, want the hand edit back (%q)", briefRel, got, hand)
	}
	if want := "restauré " + briefRel + " desde " + briefRel + ".bak-" + digest(hand) + "; lo que había quedado en " +
		briefRel + ".bak-" + digest(written); second.Message != want {
		t.Errorf("Message = %q, want %q", second.Message, want)
	}
}

// TestUndoRefusesToGuessBetweenArtifacts: with no --file and more than one
// candidate, the newest backup belongs to whichever artifact was written last, so
// the only honest answer is to ask. Guessing once restored fence.toml when the
// human meant the brief.
func TestUndoRefusesToGuessBetweenArtifacts(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel, "brief nuevo", at(21, 23, 0))
	files.put(briefRel+".bak-"+digest("brief viejo"), "brief viejo", at(20, 9, 0))
	files.put(taskRel, "documento nuevo", at(21, 23, 0))
	files.put(taskRel+".bak-"+digest("documento viejo"), "documento viejo", at(22, 9, 30))

	res := undo.Run(undo.Options{Dir: dir}, files, nil)

	if res.Code == undo.CodeOK {
		t.Fatalf("Code = %d: adivinó cuál artefacto restaurar", res.Code)
	}
	if got := files.content(t, taskRel); got != "documento nuevo" {
		t.Errorf("%s = %q, want it untouched", taskRel, got)
	}
	if got := files.content(t, briefRel); got != "brief nuevo" {
		t.Errorf("%s = %q, want it untouched", briefRel, got)
	}
	if !strings.Contains(res.Message, "--file") {
		t.Errorf("Message = %q, want it to name --file", res.Message)
	}
}

// TestFileFiltersToOneArtifact: `--file brief` ignores a newer backup of another
// artifact, which is the reason the flag exists.
func TestFileFiltersToOneArtifact(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel, "brief nuevo", at(21, 23, 0))
	files.put(briefRel+".bak-"+digest("brief viejo"), "brief viejo", at(20, 9, 0))
	files.put(taskRel, "documento nuevo", at(21, 23, 0))
	files.put(taskRel+".bak-"+digest("documento viejo"), "documento viejo", at(22, 9, 30))

	res := undo.Run(undo.Options{Dir: dir, File: "brief"}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	if got := files.content(t, briefRel); got != "brief viejo" {
		t.Errorf("%s = %q, want the brief backup", briefRel, got)
	}
	if got := files.content(t, taskRel); got != "documento nuevo" {
		t.Errorf("%s = %q, want it untouched: --file brief is not about task docs", taskRel, got)
	}
}

// TestListShowsOneLinePerArtifactNewestFirst pins the shape of the --list line.
func TestListShowsOneLinePerArtifactNewestFirst(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel, "brief actual", at(21, 23, 0))
	files.put(briefRel+".bak-"+digest("brief 1"), "brief 1", at(20, 10, 0))
	files.put(briefRel+".bak-"+digest("brief 2"), "brief 2", at(21, 20, 31))
	files.put(profileRel, "perfil actual", at(21, 23, 0))
	files.put(profileRel+".bak-"+digest("perfil 1"), "perfil 1", at(22, 8, 15))

	res := undo.Run(undo.Options{Dir: dir, List: true}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}

	want := []string{
		profileRel + " · 1 backup · último 2026-10-22 08:15",
		briefRel + " · 2 backups · último 2026-10-21 20:31",
	}
	if !slices.Equal(res.Lines, want) {
		t.Errorf("Lines = %v, want %v", res.Lines, want)
	}
	if len(files.wrote) != 0 || len(files.backups) != 0 {
		t.Errorf("--list wrote something: wrote=%v backups=%v", files.wrote, files.backups)
	}
	files.assertOnlyTheFourArtifacts(t)
}

// TestListFileFiltersToOneArtifact: the flag works for the listing too.
func TestListFileFiltersToOneArtifact(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel+".bak-"+digest("brief 1"), "brief 1", at(20, 10, 0))
	files.put(profileRel+".bak-"+digest("perfil 1"), "perfil 1", at(22, 8, 15))

	res := undo.Run(undo.Options{Dir: dir, List: true, File: "profile"}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	if want := []string{profileRel + " · 1 backup · último 2026-10-22 08:15"}; !slices.Equal(res.Lines, want) {
		t.Errorf("Lines = %v, want %v", res.Lines, want)
	}
}

// TestStoreBackupIsRestored: the store is the new home, but a backup is still a
// backup. Undo reads it and names the path it used.
func TestStoreBackupIsRestored(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel, "actual", at(21, 23, 0))
	old := storeName(briefRel, "viejo")
	files.put(old, "viejo", at(21, 20, 31))

	res := undo.Run(undo.Options{Dir: dir}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	if got := files.content(t, briefRel); got != "viejo" {
		t.Errorf("%s = %q, want the store backup", briefRel, got)
	}
	if !strings.Contains(res.Message, old) {
		t.Errorf("message = %q, want the store path %s", res.Message, old)
	}
}

// TestListReadsLegacyAndStoreBackups: a project halfway through the migration
// sees both, and the newest wins regardless of where it lives.
func TestListReadsLegacyAndStoreBackups(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel+".bak-"+digest("brief viejo"), "brief viejo", at(20, 10, 0))
	files.put(storeName(profileRel, "perfil viejo"), "perfil viejo", at(22, 8, 15))

	res := undo.Run(undo.Options{Dir: dir, List: true}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	want := []string{
		profileRel + " · 1 backup · último 2026-10-22 08:15",
		briefRel + " · 1 backup · último 2026-10-20 10:00",
	}
	if !slices.Equal(res.Lines, want) {
		t.Errorf("Lines = %v, want %v", res.Lines, want)
	}
}

// TestCollectMovesLegacyBackupsIntoTheStore: the migration is a move, not a
// copy: the bytes and the date survive, and a second run has nothing to do.
func TestCollectMovesLegacyBackupsIntoTheStore(t *testing.T) {
	const (
		oldBrief = "brief viejo"
		oldTask  = "task viejo"
	)
	files := newFakeFiles()
	files.put(briefRel, "brief actual", at(21, 23, 0))
	files.put(briefRel+".bak-"+digest(oldBrief), oldBrief, at(20, 10, 0))
	files.put(taskRel, "task actual", at(21, 23, 0))
	files.put(taskRel+".bak-"+digest(oldTask), oldTask, at(20, 11, 0))

	res := undo.Run(undo.Options{Dir: dir, Collect: true}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	if !strings.Contains(res.Message, "moví 2 backups") {
		t.Errorf("message = %q, want the count", res.Message)
	}

	briefStore := storeName(briefRel, oldBrief)
	taskStore := storeName(taskRel, oldTask)
	if got := files.content(t, briefStore); got != oldBrief {
		t.Errorf("%s = %q, want %q", briefStore, got, oldBrief)
	}
	if got := files.content(t, taskStore); got != oldTask {
		t.Errorf("%s = %q, want %q", taskStore, got, oldTask)
	}
	if !files.mod[filepath.Join(dir, filepath.FromSlash(briefStore))].Equal(at(20, 10, 0)) {
		t.Errorf("the migration lost the backup's date")
	}
	if _, ok := files.files[filepath.Join(dir, filepath.FromSlash(briefRel+".bak-"+digest(oldBrief)))]; ok {
		t.Errorf("the legacy brief backup still exists after collect")
	}

	second := undo.Run(undo.Options{Dir: dir, Collect: true}, files, nil)
	if second.Code != undo.CodeOK {
		t.Fatalf("second collect: Code = %d (message: %s)", second.Code, second.Message)
	}
	if !strings.Contains(second.Message, "no hay backups viejos") {
		t.Errorf("second collect message = %q, want a no-op", second.Message)
	}
	if len(files.wrote) != 0 || len(files.backups) != 0 {
		t.Errorf("collect wrote an artifact or made a backup: wrote=%v backups=%v", files.wrote, files.backups)
	}
}

// TestCollectNeverOverwritesADifferentBackup: the target name is content
// addressed, so a different file there means the store is not ours to trust.
func TestCollectNeverOverwritesADifferentBackup(t *testing.T) {
	const old = "viejo"
	legacy := briefRel + ".bak-" + digest(old)
	target := storeName(briefRel, old)

	files := newFakeFiles()
	files.put(briefRel, "actual", at(21, 23, 0))
	files.put(legacy, old, at(20, 10, 0))
	files.put(target, "otro contenido", at(20, 9, 0))

	res := undo.Run(undo.Options{Dir: dir, Collect: true}, files, nil)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeBlocked, res.Message)
	}
	if !strings.Contains(res.Message, "no piso un backup") {
		t.Errorf("message = %q, want the refusal", res.Message)
	}
	for _, rel := range []string{legacy, target} {
		if _, ok := files.files[filepath.Join(dir, filepath.FromSlash(rel))]; !ok {
			t.Errorf("%s disappeared after a refused collect", rel)
		}
	}
}

// TestCollectRefusesTheOtherForms: it moves everything; asking for one artifact
// or for a listing at the same time is a different question.
func TestCollectRefusesTheOtherForms(t *testing.T) {
	for _, opts := range []undo.Options{
		{Dir: dir, Collect: true, List: true},
		{Dir: dir, Collect: true, File: "brief"},
	} {
		res := undo.Run(opts, newFakeFiles(), nil)
		if res.Code != undo.CodeBlocked {
			t.Errorf("Run(%+v).Code = %d, want %d", opts, res.Code, undo.CodeBlocked)
		}
	}
}

// TestNoBackupsExplainsWhatOdachiWrites: nothing to undo is a blocked run that
// says what odachi writes and which command writes it.
func TestNoBackupsExplainsWhatOdachiWrites(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel, "sin backups", at(21, 23, 0))

	res := undo.Run(undo.Options{Dir: dir}, files, nil)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, undo.CodeBlocked)
	}
	if strings.Contains(res.Message, "\n") {
		t.Errorf("the message must be one line: %q", res.Message)
	}
	for _, want := range []string{"brief.md", ".odachi/profile.toml", ".odachi/fence.toml", "odd/tasks/*.md", "odachi start"} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("Message = %q, want it to name %q", res.Message, want)
		}
	}
	if len(files.wrote) != 0 || len(files.backups) != 0 {
		t.Errorf("a blocked undo wrote something: wrote=%v backups=%v", files.wrote, files.backups)
	}
}

// TestUnknownFileListsTheFour: a name outside the allow-list is refused before
// anything is read, and the message says what the four are.
func TestUnknownFileListsTheFour(t *testing.T) {
	files := newFakeFiles()
	files.put(briefRel+".bak-"+digest("brief 1"), "brief 1", at(20, 10, 0))

	res := undo.Run(undo.Options{Dir: dir, File: "inventado"}, files, nil)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d", res.Code, undo.CodeBlocked)
	}
	for _, want := range []string{"inventado", "brief", "profile", "fence", "task"} {
		if !strings.Contains(res.Message, want) {
			t.Errorf("Message = %q, want it to name %q", res.Message, want)
		}
	}
	if len(files.touched) != 0 {
		t.Errorf("a refused --file touched the disk: %v", files.touched)
	}
}

// TestABackupOfSomeOtherFileIsNotUndoable: a `.bak-` next to a file odachi never
// writes is housekeeping, not an undo.
func TestABackupOfSomeOtherFileIsNotUndoable(t *testing.T) {
	files := newFakeFiles()
	files.put("README.md", "readme", at(21, 23, 0))
	files.put("README.md.bak-"+digest("readme viejo"), "readme viejo", at(22, 9, 0))
	files.put(".odachi/otro.toml", "otro", at(21, 23, 0))
	files.put(".odachi/otro.toml.bak-"+digest("otro viejo"), "otro viejo", at(22, 9, 30))
	files.put("odd/tasks/notas.txt", "notas", at(21, 23, 0))
	files.put("odd/tasks/notas.txt.bak-"+digest("notas viejas"), "notas viejas", at(22, 10, 0))

	res := undo.Run(undo.Options{Dir: dir}, files, nil)

	if res.Code != undo.CodeBlocked {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeBlocked, res.Message)
	}
	if !strings.Contains(res.Message, "odachi start") {
		t.Errorf("Message = %q, want it to name odachi start", res.Message)
	}
	if len(files.wrote) != 0 || len(files.backups) != 0 {
		t.Errorf("a blocked undo wrote something: wrote=%v backups=%v", files.wrote, files.backups)
	}
}

// TestRestoringADeletedArtifactNeedsNoBackup: when the current file is gone
// there is nothing to save first, and refusing to restore would leave the human
// with no way back at all.
func TestRestoringADeletedArtifactNeedsNoBackup(t *testing.T) {
	files := newFakeFiles()
	fencePath := filepath.Join(dir, filepath.FromSlash(fenceRel))
	files.put(fenceRel+".bak-"+digest("reglas viejas"), "reglas viejas", at(20, 10, 0))

	res := undo.Run(undo.Options{Dir: dir}, files, nil)

	if res.Code != undo.CodeOK {
		t.Fatalf("Code = %d, want %d (message: %s)", res.Code, undo.CodeOK, res.Message)
	}
	if got := files.content(t, fenceRel); got != "reglas viejas" {
		t.Errorf("%s = %q, want the backup's bytes", fenceRel, got)
	}
	if want := []string{"write " + fencePath}; !slices.Equal(files.order, want) {
		t.Errorf("order = %v, want only the write: there was nothing to back up", files.order)
	}
	if !strings.Contains(res.Message, "no existía") {
		t.Errorf("Message = %q, want it to say the artifact was not there", res.Message)
	}
}
