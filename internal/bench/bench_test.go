package bench

import (
	"strings"
	"testing"
)

// TestJourneysAreTheCorpusContractXVAsksFor: eight recorridos, each one a name,
// a reason and at least one measured invocation. A journey with no step would
// measure nothing and pass for the wrong reason.
func TestJourneysAreTheCorpusContractXVAsksFor(t *testing.T) {
	journeys := Journeys()
	if len(journeys) != 8 {
		t.Fatalf("Contract XV fixes eight journeys, got %d", len(journeys))
	}

	want := []string{
		"repo-nuevo-start",
		"start-dos-veces",
		"check-sin-prueba",
		"check-run-verde",
		"archivo-prohibido",
		"plan-y-undo",
		"review-solo-docs",
		"repo-sin-commits",
	}

	seen := map[string]bool{}
	for i, journey := range journeys {
		if journey.Name != want[i] {
			t.Fatalf("journey %d is %q, want %q", i, journey.Name, want[i])
		}
		if seen[journey.Name] {
			t.Fatalf("duplicate journey %q", journey.Name)
		}
		seen[journey.Name] = true
		if strings.TrimSpace(journey.Why) == "" {
			t.Fatalf("journey %q says nothing about why it exists", journey.Name)
		}
		if len(journey.Steps) == 0 {
			t.Fatalf("journey %q measures nothing", journey.Name)
		}
		for _, step := range journey.Steps {
			if len(step.Args) == 0 {
				t.Fatalf("journey %q has a step with no arguments", journey.Name)
			}
		}
	}
}

// TestNamedFilters is `--journeys`: a name selects, a typo is reported with the
// list instead of being ignored.
func TestNamedFilters(t *testing.T) {
	all := Journeys()

	selected, unknown := Named(all, []string{"plan-y-undo", "repo-sin-commits"})
	if len(unknown) != 0 {
		t.Fatalf("unexpected unknown names: %v", unknown)
	}
	if len(selected) != 2 || selected[0].Name != "plan-y-undo" || selected[1].Name != "repo-sin-commits" {
		t.Fatalf("the selection must follow the corpus order, got %v", names(selected))
	}

	selected, unknown = Named(all, []string{"inventado"})
	if len(selected) != 0 || len(unknown) != 1 || unknown[0] != "inventado" {
		t.Fatalf("a typo must be reported: selected %v, unknown %v", names(selected), unknown)
	}

	all8, unknown := Named(all, nil)
	if len(all8) != 8 || len(unknown) != 0 {
		t.Fatalf("no filter means every journey, got %d", len(all8))
	}
}

// names is the journey names, for a readable failure.
func names(journeys []Journey) []string {
	out := make([]string, 0, len(journeys))
	for _, journey := range journeys {
		out = append(out, journey.Name)
	}
	return out
}

// TestCommandFindsTheContinuation is the heart of the instrument: a blocked
// message has to name something runnable, and finding it is how the bench decides
// what to run.
func TestCommandFindsTheContinuation(t *testing.T) {
	cases := []struct {
		name    string
		message string
		want    []string
	}{
		{
			name:    "the gate prints its next line",
			message: "odachi: falta la prueba de este cambio\nsiguiente: odachi check --proof <archivo>\n",
			want:    []string{"odachi", "check", "--proof", "<archivo>"},
		},
		{
			name:    "a sentence, not a line",
			message: "odachi: ya existe brief.md y no quiero pisarlo sin permiso; volvé a correr odachi start --force (o borrá brief.md)\n",
			want:    []string{"odachi", "start", "--force"},
		},
		{
			name:    "the answer is on the previous line",
			message: "odachi: fuera de alcance: pagos: internal/payments/pay.go\nsiguiente: odachi ask\n",
			want:    []string{"odachi", "ask"},
		},
		{
			name:    "a quoted command stays one argument",
			message: "siguiente: odachi check --run \"go test ./...\"\n",
			want:    []string{"odachi", "check", "--run", "go test ./..."},
		},
		{
			name:    "a flag's value is part of the command",
			message: "no pude leer el perfil; volvé a correr odachi level --to senior con uno de esos nombres\n",
			want:    []string{"odachi", "level", "--to", "senior"},
		},
		{
			name:    "an explicit lens is a value",
			message: "borrá el archivo y corré odachi review --lens reliability\n",
			want:    []string{"odachi", "review", "--lens", "reliability"},
		},
		{
			name:    "prose after the command is not part of it",
			message: "corré odachi init y pegá su bloque en el hook\n",
			want:    []string{"odachi", "init"},
		},
		{
			name:    "no odachi command at all",
			message: "odachi: no pude leer el árbol de git: corré odachi check dentro de un repositorio git\n",
			want:    []string{"odachi", "check"},
		},
		{
			name:    "a command that is not odachi is not this instrument's",
			message: "no hay .git: no hay hook posible; corré git init\n",
			want:    nil,
		},
		{
			name:    "nothing named",
			message: "odachi: la prueba falló: sin salida\nsiguiente: odachi check\n",
			want:    []string{"odachi", "check"},
		},
		{
			name:    "empty",
			message: "",
			want:    nil,
		},
		{
			name:    "the bare name is not a command",
			message: "instalá odachi y volvé\n",
			want:    nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Command(c.message)
			if len(got) != len(c.want) {
				t.Fatalf("Command() = %v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Fatalf("Command() = %v, want %v", got, c.want)
				}
			}
		})
	}
}

// TestIsTemplated: a command with a placeholder cannot be run alone, and counting
// it as a dead end would blame the message for the bench's own limitation.
func TestIsTemplated(t *testing.T) {
	cases := []struct {
		command []string
		want    bool
	}{
		{[]string{"odachi", "check", "--proof", "<archivo>"}, true},
		{[]string{"odachi", "level", "--to", "<perfil>"}, true},
		{[]string{"odachi", "start", "--force"}, false},
		{[]string{"odachi", "check", "--run", "go test ./..."}, false},
		{nil, false},
	}
	for _, c := range cases {
		if got := IsTemplated(c.command); got != c.want {
			t.Fatalf("IsTemplated(%v) = %v, want %v", c.command, got, c.want)
		}
	}
}

// TestKindOf is the classification table: four answers and nothing else.
func TestKindOf(t *testing.T) {
	cases := []struct {
		name    string
		command []string
		exit    int
		want    Kind
	}{
		{"names a command that works", []string{"odachi", "review"}, 0, InBand},
		{"names a command that does not work", []string{"odachi", "borrar"}, 1, DeadEnd},
		{"names a command that needs a human", []string{"odachi", "start"}, 2, Interactive},
		{"names no command", nil, 0, OutOfBand},
		{"names a placeholder", []string{"odachi", "check", "--proof", "<archivo>"}, -1, Templated},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := KindOf(c.command, c.exit); got != c.want {
				t.Fatalf("KindOf() = %q, want %q", got, c.want)
			}
		})
	}
}

// TestSandboxEnvIsAControlledHome: the journey must not read or write the home
// of whoever ran the bench, and git must not read their config.
func TestSandboxEnvIsAControlledHome(t *testing.T) {
	base := []string{
		"PATH=/usr/bin",
		"HOME=/home/ana",
		"USERPROFILE=C:/Users/ana",
		"GIT_CONFIG_GLOBAL=/home/ana/.gitconfig",
		"KEEP=me",
	}
	env := SandboxEnv(base, "/tmp/journey/home")

	got := map[string]string{}
	for _, entry := range env {
		name, value, found := strings.Cut(entry, "=")
		if !found {
			t.Fatalf("environment entries are KEY=VALUE, got %q", entry)
		}
		if _, seen := got[name]; seen {
			t.Fatalf("%s appears twice: the sandbox would read the human's or the journey's home, never both", name)
		}
		got[name] = value
	}

	for _, name := range []string{"HOME", "USERPROFILE"} {
		if got[name] != "/tmp/journey/home" {
			t.Fatalf("%s = %q, want the sandbox home", name, got[name])
		}
	}
	if got["PATH"] != "/usr/bin" {
		t.Fatalf("PATH must be inherited: git has to be reachable, got %q", got["PATH"])
	}
	if got["KEEP"] != "me" {
		t.Fatalf("the rest of the environment is inherited")
	}
	if !strings.HasPrefix(got["GIT_CONFIG_GLOBAL"], "/tmp/journey/home") {
		t.Fatalf("git must read its config inside the sandbox, got %q", got["GIT_CONFIG_GLOBAL"])
	}
}
