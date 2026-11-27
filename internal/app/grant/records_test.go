package grant

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"

	domain "github.com/biggs-100/odachi/internal/domain/grant"
)

// probeRecord is the smallest record envelope the reader must accept. The
// storage primitives never interpret a record; T3-T5 supply their own envelope
// and this struct proves the reader's contract is generic.
type probeRecord struct {
	Version int `json:"version"`
}

func TestNewFileStoreMakesTheGrantNamespaceDurable(t *testing.T) {
	parent := t.TempDir()
	runtimeRoot := filepath.Join(parent, "runtime", "nested")
	var synced []string
	store, err := newFileStore(runtimeRoot, func(ctx context.Context, path string) error {
		return ensureDurableDirectoryWith(ctx, path, func(_ context.Context, directory string) error {
			synced = append(synced, directory)
			return nil
		})
	})
	if err != nil {
		t.Fatalf("newFileStore() error = %v", err)
	}
	if store.root != runtimeRoot {
		t.Fatalf("store root = %q, want %q", store.root, runtimeRoot)
	}
	// Every directory that appears is synced, and its parent is synced after it,
	// so a crash cannot leave a grant namespace that a reader cannot find.
	want := []string{
		filepath.Join(parent, "runtime"), parent,
		runtimeRoot, filepath.Join(parent, "runtime"),
		filepath.Join(runtimeRoot, namespaceName), runtimeRoot,
	}
	if !slices.Equal(synced, want) {
		t.Fatalf("synced directories = %q, want %q", synced, want)
	}
	info, err := os.Stat(filepath.Join(runtimeRoot, namespaceName))
	if err != nil {
		t.Fatalf("Stat(namespace) error = %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("namespace %q is not a directory", info.Name())
	}
}

func TestNewFileStoreCreatesTheNamespaceOnDisk(t *testing.T) {
	runtimeRoot := filepath.Join(t.TempDir(), "runtime", "nested")
	store, err := NewFileStore(runtimeRoot)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}
	if store.root != runtimeRoot {
		t.Fatalf("store root = %q, want %q", store.root, runtimeRoot)
	}
	info, err := os.Stat(filepath.Join(runtimeRoot, namespaceName))
	if err != nil {
		t.Fatalf("Stat(namespace) error = %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("%q is not a directory", info.Name())
	}
	opened, err := store.open()
	if err != nil {
		t.Fatalf("open() error = %v", err)
	}
	if err := opened.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func TestNewFileStoreFailsClosedWithoutDirectoryDurability(t *testing.T) {
	for _, target := range []string{"runtime root", namespaceName} {
		t.Run(target, func(t *testing.T) {
			runtimeRoot := filepath.Join(t.TempDir(), "runtime")
			unsupported := runtimeRoot
			if target == namespaceName {
				unsupported = filepath.Join(runtimeRoot, namespaceName)
			}
			_, err := newFileStore(runtimeRoot, func(ctx context.Context, path string) error {
				return ensureDurableDirectoryWith(ctx, path, func(_ context.Context, directory string) error {
					if directory == unsupported {
						return ErrDirectorySyncUnsupported
					}
					return nil
				})
			})
			if !errors.Is(err, ErrStorageFailure) || !errors.Is(err, ErrDirectorySyncUnsupported) {
				t.Fatalf("newFileStore() error = %v, want a storage failure that wraps unsupported directory sync", err)
			}
		})
	}
}

func TestNewFileStoreRejectsAnUnusableRuntimeRoot(t *testing.T) {
	if _, err := NewFileStore("   "); !errors.Is(err, ErrStorageFailure) {
		t.Fatalf("NewFileStore(blank) error = %v, want ErrStorageFailure", err)
	}
	file := filepath.Join(t.TempDir(), "runtime")
	if err := os.WriteFile(file, []byte("not a directory"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if _, err := NewFileStore(file); !errors.Is(err, ErrStorageFailure) {
		t.Fatalf("NewFileStore(file) error = %v, want ErrStorageFailure", err)
	}
}

func TestRecordPathRefusesAnInvalidIdentityOrSuffix(t *testing.T) {
	store := &FileStore{root: t.TempDir()}
	if _, err := store.recordPath(domain.ID{}, ".json"); !errors.Is(err, domain.ErrInvalidID) {
		t.Fatalf("recordPath(zero ID) error = %v, want ErrInvalidID", err)
	}
	id, err := domain.NewID()
	if err != nil {
		t.Fatalf("NewID() error = %v", err)
	}
	for _, suffix := range []string{"json", "", ".json/../escape", `\json`} {
		if _, err := store.recordPath(id, suffix); !errors.Is(err, ErrStorageFailure) {
			t.Fatalf("recordPath(suffix %q) error = %v, want ErrStorageFailure", suffix, err)
		}
	}
	path, err := store.recordPath(id, ".json")
	if err != nil {
		t.Fatalf("recordPath() error = %v", err)
	}
	if want := filepath.Join(namespaceName, id.String()+".json"); path != want {
		t.Fatalf("recordPath() = %q, want %q", path, want)
	}
}

func TestPublishRecordCopiesTheCompleteBytesIntoTheDestinationDirectory(t *testing.T) {
	dir, root := newRecordRoot(t)
	name := filepath.Join(namespaceName, "grant.json")
	final := filepath.Join(dir, name)
	payload := []byte(`{"version":1}`)
	staged := false
	err := publishRecordWith(t.Context(), root, name, payload, publisher{
		syncFile: func(file *os.File) error {
			staged = true
			staging := file.Name()
			if !strings.HasSuffix(staging, filepath.Join(namespaceName, "grant.json"+stagingSuffix)) {
				t.Errorf("staging record = %q, want it beside %q", staging, name)
			}
			written, err := os.ReadFile(staging)
			if err != nil {
				t.Fatalf("ReadFile(staging) error = %v", err)
			}
			if !bytes.Equal(written, payload) {
				t.Fatalf("staging record = %q, want the complete payload %q", written, payload)
			}
			if _, err := os.Stat(final); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("final record was visible before publication: %v", err)
			}
			return file.Sync()
		},
		syncParent: syncParentDirectory,
	})
	if !staged {
		t.Fatal("publication never synced the staging copy")
	}
	if err != nil {
		t.Fatalf("publishRecordWith() error = %v", err)
	}
	got, err := os.ReadFile(final)
	if err != nil {
		t.Fatalf("ReadFile(final) error = %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("published record = %q, want %q", got, payload)
	}
	assertOnlyRecord(t, filepath.Dir(final), filepath.Base(final))
}

func TestPublishRecordSyncsTheDirectoryAfterTheRecordIsVisible(t *testing.T) {
	dir, root := newRecordRoot(t)
	name := filepath.Join(namespaceName, "grant.json")
	final := filepath.Join(dir, name)
	payload := []byte(`{"version":1}`)
	var steps []string
	err := publishRecordWith(t.Context(), root, name, payload, publisher{
		syncFile: func(file *os.File) error {
			steps = append(steps, "file")
			return file.Sync()
		},
		syncParent: func(ctx context.Context, opened *os.Root, synced string) error {
			if synced != name {
				t.Errorf("parent sync name = %q, want the published record %q", synced, name)
			}
			data, err := os.ReadFile(final)
			if err != nil {
				t.Fatalf("ReadFile(final) during parent sync: %v", err)
			}
			if !bytes.Equal(data, payload) {
				t.Fatalf("record during parent sync = %q, want %q", data, payload)
			}
			if _, err := os.Stat(final + stagingSuffix); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("staging copy survived into the directory sync: %v", err)
			}
			steps = append(steps, "parent")
			return syncParentDirectory(ctx, opened, synced)
		},
	})
	if err != nil {
		t.Fatalf("publishRecordWith() error = %v", err)
	}
	if !slices.Equal(steps, []string{"file", "parent"}) {
		t.Fatalf("durability order = %q, want the record synced before the directory", steps)
	}
}

func TestPublishRecordHidesTheRecordUntilItsBytesAreComplete(t *testing.T) {
	dir, root := newRecordRoot(t)
	name := filepath.Join(namespaceName, "grant.json")
	final := filepath.Join(dir, name)
	payload := []byte(`{"version":1}`)

	staging := make(chan struct{})
	release := make(chan struct{})
	unblock := sync.OnceFunc(func() { close(release) })
	t.Cleanup(unblock)

	done := make(chan error, 1)
	go func() {
		done <- publishRecordWith(context.WithoutCancel(t.Context()), root, name, payload, publisher{
			syncFile: func(file *os.File) error {
				close(staging)
				<-release
				return file.Sync()
			},
			syncParent: syncParentDirectory,
		})
	}()
	<-staging

	stop := make(chan struct{})
	var readers sync.WaitGroup
	badObservation := make(chan string, 1)
	report := func(observation string) {
		select {
		case badObservation <- observation:
		default:
		}
	}
	readers.Go(func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			data, err := os.ReadFile(final)
			if err != nil {
				if !errors.Is(err, fs.ErrNotExist) {
					report(err.Error())
					return
				}
				runtime.Gosched()
				continue
			}
			if !bytes.Equal(data, payload) {
				report(fmt.Sprintf("reader observed an incomplete record %q", data))
				return
			}
		}
	})

	if _, err := os.Stat(final); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("final record existed while its bytes were still being staged: %v", err)
	}
	unblock()
	if err := <-done; err != nil {
		t.Fatalf("publishRecordWith() error = %v", err)
	}
	close(stop)
	readers.Wait()

	select {
	case observation := <-badObservation:
		t.Fatal(observation)
	default:
	}
	got, err := os.ReadFile(final)
	if err != nil {
		t.Fatalf("ReadFile(final) error = %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("published record = %q, want %q", got, payload)
	}
}

func TestPublishRecordKeepsThePreviousCompleteRecord(t *testing.T) {
	dir, root := newRecordRoot(t)
	name := filepath.Join(namespaceName, "grant.json")
	final := filepath.Join(dir, name)
	original := []byte(`{"version":1}`)
	if err := os.WriteFile(final, original, 0o600); err != nil {
		t.Fatalf("WriteFile(original) error = %v", err)
	}

	err := publishRecord(t.Context(), root, name, []byte(`{"version":2}`))
	if !errors.Is(err, ErrStorageFailure) || !errors.Is(err, fs.ErrExist) {
		t.Fatalf("publishRecord() error = %v, want a storage failure that wraps fs.ErrExist", err)
	}
	got, err := os.ReadFile(final)
	if err != nil {
		t.Fatalf("ReadFile(final) error = %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("record after refused publication = %q, want the untouched original %q", got, original)
	}
	assertOnlyRecord(t, filepath.Dir(final), filepath.Base(final))
}

func TestPublishRecordRemovesTheStagingCopyWhenDurabilityFails(t *testing.T) {
	t.Run("record sync", func(t *testing.T) {
		dir, root := newRecordRoot(t)
		name := filepath.Join(namespaceName, "grant.json")
		failure := errors.New("the device refused the flush")
		err := publishRecordWith(t.Context(), root, name, []byte(`{"version":1}`), publisher{
			syncFile:   func(*os.File) error { return failure },
			syncParent: syncParentDirectory,
		})
		if !errors.Is(err, ErrStorageFailure) || !errors.Is(err, failure) {
			t.Fatalf("publishRecordWith() error = %v, want a storage failure that wraps the sync failure", err)
		}
		assertOnlyRecord(t, filepath.Join(dir, namespaceName))
		if _, err := os.Stat(filepath.Join(dir, name)); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("a record survived a failed durability sync: %v", err)
		}
	})

	t.Run("directory sync", func(t *testing.T) {
		dir, root := newRecordRoot(t)
		name := filepath.Join(namespaceName, "grant.json")
		payload := []byte(`{"version":1}`)
		err := publishRecordWith(t.Context(), root, name, payload, publisher{
			syncFile:   (*os.File).Sync,
			syncParent: func(context.Context, *os.Root, string) error { return ErrDirectorySyncUnsupported },
		})
		if !errors.Is(err, ErrStorageFailure) || !errors.Is(err, ErrDirectorySyncUnsupported) {
			t.Fatalf("publishRecordWith() error = %v, want a storage failure that wraps unsupported directory sync", err)
		}
		// The record is complete and may already be read: deleting it would be a
		// second unproven mutation. It is reported as an unconfirmed publication
		// and the temporary file is gone.
		got, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("ReadFile(final) error = %v", err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("record after failed directory sync = %q, want %q", got, payload)
		}
		assertOnlyRecord(t, filepath.Join(dir, namespaceName), filepath.Base(name))
	})
}

func TestPublishRecordRefusesEmptyAndOversizedPayloads(t *testing.T) {
	dir, root := newRecordRoot(t)
	name := filepath.Join(namespaceName, "grant.json")
	for label, payload := range map[string][]byte{
		"empty":     {},
		"oversized": bytes.Repeat([]byte("x"), maxRecordSize+1),
	} {
		t.Run(label, func(t *testing.T) {
			err := publishRecord(t.Context(), root, name, payload)
			if !errors.Is(err, ErrStorageFailure) {
				t.Fatalf("publishRecord(%s) error = %v, want ErrStorageFailure", label, err)
			}
			assertOnlyRecord(t, filepath.Join(dir, namespaceName))
		})
	}
}

func TestReadRecordDistinguishesAbsenceFromCorruption(t *testing.T) {
	dir, root := newRecordRoot(t)
	name := filepath.Join(namespaceName, "grant.json")
	final := filepath.Join(dir, name)

	if err := readRecord(t.Context(), root, name, &probeRecord{}); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("readRecord(missing) error = %v, want fs.ErrNotExist", err)
	}

	for label, content := range map[string]string{
		"empty":          "",
		"truncated":      `{"version":1`,
		"unknown field":  `{"version":1,"extra":true}`,
		"trailing value": `{"version":1}{"version":2}`,
		"not an object":  `"record"`,
	} {
		t.Run(label, func(t *testing.T) {
			if err := os.WriteFile(final, []byte(content), 0o600); err != nil {
				t.Fatalf("WriteFile() error = %v", err)
			}
			if err := readRecord(t.Context(), root, name, &probeRecord{}); !errors.Is(err, ErrCorruptRecord) {
				t.Fatalf("readRecord(%s) error = %v, want ErrCorruptRecord", label, err)
			}
		})
	}

	t.Run("oversized", func(t *testing.T) {
		if err := os.WriteFile(final, bytes.Repeat([]byte("x"), maxRecordSize+1), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		if err := readRecord(t.Context(), root, name, &probeRecord{}); !errors.Is(err, ErrCorruptRecord) {
			t.Fatalf("readRecord(oversized) error = %v, want ErrCorruptRecord", err)
		}
	})

	t.Run("directory", func(t *testing.T) {
		if err := os.Remove(final); err != nil {
			t.Fatalf("Remove() error = %v", err)
		}
		if err := os.Mkdir(final, 0o700); err != nil {
			t.Fatalf("Mkdir() error = %v", err)
		}
		if err := readRecord(t.Context(), root, name, &probeRecord{}); !errors.Is(err, ErrCorruptRecord) {
			t.Fatalf("readRecord(directory) error = %v, want ErrCorruptRecord", err)
		}
		if _, err := recordExists(t.Context(), root, name); !errors.Is(err, ErrCorruptRecord) {
			t.Fatalf("recordExists(directory) error = %v, want ErrCorruptRecord", err)
		}
	})
}

func TestReadRecordAcceptsOneCompleteEnvelope(t *testing.T) {
	dir, root := newRecordRoot(t)
	name := filepath.Join(namespaceName, "grant.json")
	final := filepath.Join(dir, name)
	if err := os.WriteFile(final, []byte(`{"version":1}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if exists, err := recordExists(t.Context(), root, name); err != nil || !exists {
		t.Fatalf("recordExists() = %v, %v, want true, nil", exists, err)
	}
	var record probeRecord
	if err := readRecord(t.Context(), root, name, &record); err != nil {
		t.Fatalf("readRecord() error = %v", err)
	}
	if record.Version != 1 {
		t.Fatalf("decoded version = %d, want 1", record.Version)
	}
	if exists, err := recordExists(t.Context(), root, filepath.Join(namespaceName, "absent.json")); err != nil || exists {
		t.Fatalf("recordExists(absent) = %v, %v, want false, nil", exists, err)
	}
}

// newRecordRoot opens a temp root that already holds the grant namespace, so a
// test exercises one publication instead of store setup.
func newRecordRoot(t *testing.T) (string, *os.Root) {
	t.Helper()
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatalf("OpenRoot() error = %v", err)
	}
	t.Cleanup(func() { _ = root.Close() })
	if err := root.Mkdir(namespaceName, storeMode); err != nil {
		t.Fatalf("Mkdir(%s) error = %v", namespaceName, err)
	}
	return dir, root
}

// assertOnlyRecord asserts the directory holds no temporary or unexpected file.
func assertOnlyRecord(t *testing.T, dir string, want ...string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%s) error = %v", dir, err)
	}
	got := make([]string, 0, len(entries))
	for _, entry := range entries {
		got = append(got, entry.Name())
	}
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("directory %s holds %q, want %q (no staging copy)", dir, got, want)
	}
}
