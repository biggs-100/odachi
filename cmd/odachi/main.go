// Command odachi is the whole product: one static binary, no daemon, no
// network. `odachi` with no arguments is `odachi start`.
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/biggz/odachi/internal/app/start"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run runs the CLI in the current directory. It returns the exit code instead
// of calling os.Exit so the behaviour is testable.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(stderr, "odachi: no pude leer el directorio actual: "+oneLine(err.Error()))
		return start.CodeBlocked
	}
	return runIn(dir, args, stdin, stdout, stderr)
}

// runIn dispatches one command. `odachi` and `odachi --flag ...` are both
// `odachi start`; an unknown command lists what exists.
func runIn(dir string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return runStart(dir, args, stdin, stdout, stderr)
	}

	switch args[0] {
	case "start":
		return runStart(dir, args[1:], stdin, stdout, stderr)
	default:
		fmt.Fprintf(stderr, "odachi: comando desconocido: %s; comandos disponibles: start (por ejemplo odachi start)\n", args[0])
		return start.CodeBlocked
	}
}

// runStart parses the flags, wires the two adapters and reports the result.
func runStart(dir string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("odachi start", flag.ContinueOnError)
	// Flag errors are translated into one Spanish line below, so the standard
	// library's own usage dump is discarded.
	flags.SetOutput(io.Discard)

	var (
		opts    start.Options
		done    stringList
		exclude stringList
		learn   stringList
		brief   string
		force   bool
		asJSON  bool
	)

	flags.StringVar(&opts.Outcome, "outcome", "", "respuesta 1: qué tiene que existir cuando terminemos")
	flags.StringVar(&opts.Audience, "audience", "", "respuesta 2: phone | cpu | both")
	flags.Var(&done, "done", "respuesta 3: cómo vas a saber que funcionó (repetible)")
	flags.Var(&exclude, "exclude", "respuesta 4: no-goal por clave del catálogo o texto libre (repetible)")
	flags.StringVar(&opts.Sensitive, "sensitive", "", "respuesta 5: none | names | health | money")
	flags.StringVar(&opts.Deadline, "deadline", "", "respuesta 6: today | week | relaxed")
	flags.Var(&learn, "learn", "respuesta 7: qué querés aprender (repetible)")
	flags.StringVar(&brief, "brief", "", "leer un brief existente en vez de preguntar")
	flags.BoolVar(&force, "force", false, "reescribir un brief.md existente")
	flags.BoolVar(&asJSON, "json", false, "salida para máquinas")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return start.CodeOK
		}
		fmt.Fprintln(stderr, "odachi: flag inválido: "+oneLine(err.Error())+"; corré odachi start -h para ver los flags")
		return start.CodeBlocked
	}
	if flags.NArg() > 0 {
		fmt.Fprintf(stderr, "odachi: sobra un argumento: %s; odachi start solo toma flags (corré odachi start -h)\n", flags.Arg(0))
		return start.CodeBlocked
	}

	opts.Dir = dir
	opts.Done = done
	opts.Exclude = exclude
	opts.Learn = learn
	opts.BriefPath = brief
	opts.Force = force
	// No terminal (or machine output) means no human to ask: the run uses the
	// flags or exits 2 naming them.
	opts.Machine = asJSON || !isTerminal(stdin)

	loaded, code, message := loadProfile(dir)
	if code != start.CodeOK {
		return report(start.Result{Code: code, Message: message}, asJSON, stdout, stderr)
	}
	opts.Profile = loaded

	var interview start.Interview
	if !opts.Machine {
		interview = &terminalInterview{in: bufio.NewReader(stdin), out: stdout}
	}

	return report(start.Run(opts, interview, osFiles{}), asJSON, stdout, stderr)
}

// report prints the result and returns the exit code. Machine output is JSON;
// human output is one line, on stderr when the run is blocked.
func report(result start.Result, asJSON bool, stdout, stderr io.Writer) int {
	if asJSON {
		files := result.Files
		if files == nil {
			files = []string{}
		}
		payload := struct {
			Code    int      `json:"code"`
			Message string   `json:"message"`
			Profile string   `json:"profile,omitempty"`
			Task    string   `json:"task,omitempty"`
			Files   []string `json:"files"`
		}{
			Code:    result.Code,
			Message: result.Message,
			Profile: result.Profile,
			Task:    result.Task,
			Files:   files,
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, "odachi: no pude armar la salida JSON: "+oneLine(err.Error()))
			return start.CodeBlocked
		}
		fmt.Fprintln(stdout, string(data))
		return result.Code
	}

	if result.Code == start.CodeOK {
		fmt.Fprintln(stdout, result.Message)
	} else {
		fmt.Fprintln(stderr, "odachi: "+result.Message)
	}
	return result.Code
}

// usage is the Spanish help a human reads with `odachi start -h`.
func usage(out io.Writer) {
	fmt.Fprint(out, `odachi · dos manos: vos decidís, el agente construye

uso:
  odachi                      lo mismo que odachi start
  odachi start [flags]

flags:
  --outcome s     respuesta 1: qué tiene que existir cuando terminemos
  --audience s    respuesta 2: phone | cpu | both
  --done s        respuesta 3 (repetible): cómo vas a saber que funcionó
  --exclude s     respuesta 4 (repetible): no-goal por clave o texto libre
  --sensitive s   respuesta 5: none | names | health | money
  --deadline s    respuesta 6: today | week | relaxed
  --learn s       respuesta 7 (repetible): qué querés aprender
  --brief path    leer un brief existente en vez de preguntar
  --force         reescribir un brief.md existente (se respalda primero)
  --json          salida para máquinas

códigos de salida: 0 artefactos escritos · 1 bloqueado · 2 falta una respuesta humana
`)
}

// isTerminal reports whether the input is a terminal. Anything else (a pipe, a
// file, a test) means no human is there to answer.
func isTerminal(in io.Reader) bool {
	file, ok := in.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// stringList is a repeatable string flag: --done a --done b.
type stringList []string

func (l *stringList) String() string {
	return strings.Join(*l, ", ")
}

func (l *stringList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

// oneLine keeps a CLI message on one line whatever an adapter said.
func oneLine(message string) string {
	return strings.Join(strings.Fields(message), " ")
}
