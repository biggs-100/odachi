// Package rules renders the odachi block of AGENTS.md — the one text an agent
// obeys on every session — with the voice of the profile in force as its last
// section.
//
// One owner for one text. `odachi init` writes the whole block and `odachi level
// --to <profile>` rewrites the voice section in place, and both go through this
// package: a second copy of the rules is a second set of rules, and an agent
// that reads one while the binary enforces the other is being lied to.
//
// The block is the artifact, not a summary of it: it says what odachi is, the
// four hard rules the gate enforces, the six verbs, and how the agent talks. It
// has to make sense to an agent that knows nothing about this project, which is
// why the text is pinned byte for byte by test instead of asserted by pieces.
//
// It is pure: no filesystem, no os. The app layer decides where the project is
// and writes the bytes.
package rules

import (
	"strings"

	"github.com/biggs-100/odachi/internal/domain/persona"
)

const (
	// File is where the block lives, relative to the project root. Every client
	// odachi knows reads AGENTS.md; that is the whole reason the voice moved
	// here from a per-client file nobody's selector listed.
	File = "AGENTS.md"
	// Begin and End are the markers the block lives between, so replacing it
	// never touches a line the human wrote around it.
	Begin = "<!-- odachi:begin -->"
	End   = "<!-- odachi:end -->"
	// Section is the heading of the block's last section: how the agent talks.
	// It is the one part of the block that changes with the profile, and the
	// only part `odachi level --to` rewrites.
	Section = "## How you speak"
	// SectionES is the read-only alias for blocks written before the language
	// migration. New blocks use Section; old ones still rewrite their voice.
	SectionES = "## Cómo hablás"
)

// headTop, headRest and headTail are the block up to and including the blank line
// that opens the voice section, split where the answer depends on where the craft
// lives. Binding them in Body is what keeps the two voices from differing anywhere
// but in that section: everything here is the same bytes whatever voice is in
// force, and only the two sentences about the skills change with the scope.
const (
	headTop = `
# Project rules (odachi)

odachi is a binary and four plain-text artifacts —brief, task doc, proof and
profile—. The binary enforces the gate; the artifacts are the memory. The human
decides the vision and what is non-negotiable; the agent executes with craft. If
you are unsure where the context lives: ` + "`brief.md`" + ` for the vision, ` + "`odd/tasks/`" + ` for the plan,
`
	headRest = `
Four hard rules. They are not advice: ` + "`odachi check`" + ` enforces them.

1. Brief before code. If ` + "`brief.md`" + ` is not complete, run ` + "`odachi start`" + ` first.
   Without a written vision, nothing is built.
2. Nothing outside the fence. If something is in "Not included", do not touch it:
   ask with ` + "`odachi ask \"<question>\"`" + `. ` + "`check`" + ` enforces the fence.
3. Do not say "done" without proof. Done is ` + "`odachi check --run \"<command>\"`" + ` green,
   with the output saved in ` + "`proof/`" + `. Without that, it is not done.
4. Small pieces, one commit each. One task-doc step, one commit.

One more, about folders. This project is this repository. Another folder is another
project —its own brief, its own fence, its own store, its own memory—, and odachi is
never read or run there, and its files are never touched, unless the human asks for it.

And one about asking, because the human is in the loop. A question is answered, not
implemented: an answer is not a proposal, and work starts when the human asks for it or
approves one. An irreversible action —push, pull request, merge, tag or release, an
install, another folder— names its exact target and waits for a yes. Ask at the gate,
once: a question that would not change what you do next is noise.

One more, about scratch. The note you write while you work —a commit message, a pull
request body, a plan— lives in ` + "`.git/odachi/scratch/`" + `, never in the tree and never
loose in the git directory: the tree must stay clean, and a file with no folder is
one nobody can find or empty again. ` + "`odachi clean --apply`" + ` empties that folder.

One more, about closing. When the human says the session is over —"cerramos sesión",
"seguimos mañana"— a project that keeps ` + "`docs/next-session.md`" + ` rewrites that file before
the agent says it is closed: the newest feature, what was measured, and the decision that
waits. The chat is gone tomorrow, and the file is what the next session reads; ` + "`odachi doctor`" + ` says
when it went stale.
`
	headTail = `
## The six verbs

- ` + "`odachi start`" + ` — the interview: writes brief, profile, fence and task doc.
- ` + "`odachi ask \"<question>\"`" + ` — records an open question; with ` + "`--answer`" + ` it answers it.
- ` + "`odachi level`" + ` — shows your facts and proposes senior when earned; with ` + "`--to <profile>`" + ` it applies the profile and this voice.
- ` + "`odachi summary`" + ` — what is done in each feature.
- ` + "`odachi check`" + ` — is it ready? brief, fence, open questions, planned requirements and proof; with ` + "`--run`" + ` it runs the proof.
- ` + "`odachi init`" + ` — writes this block, the hook, the skills and the profile voice.

` + Section + `

`
)

// skillsParagraph and skillsPlace are the two sentences that depend on where the
// craft lives: a project that carries its own copy reads one folder, and a project
// that keeps it in the machine reads the other. Naming both and letting the reader
// choose was tried and it lies whenever a stale copy survives -the run says the
// machine copy is the live one and the block pointed at the copy nothing refreshes.
func skillsParagraph(scope string) string {
	if scope == persona.SkillsHome {
		return "the skills for the craft: " + "`~/.agents/skills/`" + ", the copy this machine keeps.\n"
	}
	return "the skills for the craft: " + "`.agents/skills/`" + ", this project's copy.\n"
}

// skillsPlace is the line the agent reads before acting, in the same two cases.
func skillsPlace(scope string) string {
	if scope == persona.SkillsHome {
		return "Skills: the ones in " + "`~/.agents/skills/`" + ", the copy this machine keeps; inside this project, load " + "`drive`" + " before acting: it is the manual.\n"
	}
	return "Skills: the ones in " + "`.agents/skills/`" + "; inside this project, load " + "`drive`" + " before acting: it is the manual.\n"
}

// Body returns the text of the block without its markers, trailing newline
// included. It is what a native agent file mirrors: the agent must read the same
// words as the block, and the markers only exist so AGENTS.md can be edited in
// place.
//
// Body is the block for a voice in the default scope: the project's own copy.
// BodyFor is the same text for the scope the run resolved.
func Body(voice string) (string, error) { return BodyFor(voice, persona.SkillsProject) }

// BodyFor is Body with the scope the craft lives in: persona.SkillsHome names the
// machine's copy, and anything else -including a project that never chose- names
// the project's.
func BodyFor(voice, scope string) (string, error) {
	body, err := persona.Voice(voice)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(headTop, "\n") + skillsParagraph(scope) + headRest + "\n" + skillsPlace(scope) + headTail + body, nil
}

// Block is the whole block in the default scope; BlockFor names the scope.
func Block(voice string) (string, error) { return BlockFor(voice, persona.SkillsProject) }

// BlockFor returns the whole block for one voice and one scope, trailing newline
// included, with the markers that let WithBlock replace it in place.
func BlockFor(voice, scope string) (string, error) {
	body, err := BodyFor(voice, scope)
	if err != nil {
		return "", err
	}
	return Begin + "\n" + body + End + "\n", nil
}

// BodyIn returns the text between the markers of a complete block, or false when
// the file has no complete block. It is the mirror's source of truth: a native
// agent file written from this text cannot disagree with the block the agent
// reads, even after the human edited the rules by hand.
// fold is what a reader sees: CRLF becomes LF. A checkout is not an edit -a Windows clone
// with core.autocrlf=true gives every line a CR- and every reader below parses by line, so
// the endings are folded once, here, instead of in each caller. The writer never folds: a
// file keeps the line endings its platform gave it.
func fold(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}

func BodyIn(content string) (string, bool) {
	content = fold(content)
	begin := lineStart(content, Begin)
	if begin < 0 {
		return "", false
	}
	offset := lineStart(content[begin:], End)
	if offset < 0 {
		return "", false
	}
	body := content[begin+len(Begin) : begin+offset]
	return strings.TrimPrefix(body, "\n"), true
}

// WithBlock returns content with the block replaced in place, or appended when
// the file has no block yet. It is `odachi init`'s write: the block is generated
// whole, so an upgrade refreshes every rule at once.
//
// A file with an opening marker and no closing one is treated as a truncated
// block: it is replaced from the marker to the end instead of gaining a second
// one.
func WithBlock(content, voice string) (string, error) {
	return WithBlockFor(content, voice, persona.SkillsProject)
}

// WithBlockFor is WithBlock for the scope the run resolved.
func WithBlockFor(content, voice, scope string) (string, error) {
	block, err := BlockFor(voice, scope)
	if err != nil {
		return content, err
	}

	begin := strings.Index(content, Begin)
	if begin < 0 {
		return appendBlock(content, block), nil
	}

	end := strings.Index(content[begin:], End)
	if end < 0 {
		return content[:begin] + block, nil
	}
	rest := content[begin+end+len(End):]
	if newline := strings.IndexByte(rest, '\n'); newline >= 0 {
		rest = rest[newline+1:]
	} else {
		rest = ""
	}
	return content[:begin] + block + rest, nil
}

// appendBlock puts the block at the end of a file that does not have one,
// separated by one blank line, and keeps every byte that was already there.
func appendBlock(content, block string) string {
	if strings.TrimSpace(content) == "" {
		return block
	}
	return strings.TrimRight(content, "\n") + "\n\n" + block
}

// WithVoice returns content with the block's voice section rewritten in place.
// It is `odachi level --to`'s write, and it owns exactly one section: every other
// byte of the file —the rules, the verbs, whatever the human wrote above and
// below— is preserved, the way the .gitignore block preserves its file.
//
// A block written before the voice moved into it has no section, and it gains one
// as the last section of the block. A file with no block is returned untouched:
// writing the whole block is init's job, and level never invents a block for a
// file odachi did not write.
//
// changed is false when the bytes are already what they should be, so a second
// run writes nothing at all.
func WithVoice(content, voice string) (string, bool, error) {
	body, err := persona.Voice(voice)
	if err != nil {
		return content, false, err
	}

	start, end, insert, ok := voiceSpan(content)
	if !ok {
		return content, false, nil
	}
	if insert {
		added := content[:start] + "\n" + Section + "\n\n" + body + content[start:]
		return added, true, nil
	}

	updated := content[:start] + "\n" + body + content[end:]
	return updated, updated != content, nil
}

// VoiceIn returns the voice a file already publishes, and whether it publishes
// one at all. A block written before the voice moved into it has no section, and
// that is a state to report, not a voice to guess. The text comes back trimmed,
// because a human's trailing blank line is not a different voice.
func VoiceIn(content string) (string, bool) {
	content = fold(content)
	start, end, insert, ok := voiceSpan(content)
	if !ok || insert {
		return "", false
	}
	return strings.TrimSpace(content[start:end]), true
}

// HasBlock reports whether content carries a complete odachi block. A truncated
// block is not one: `odachi init` repairs it, and a check that called it complete
// would leave the agent without rules.
func HasBlock(content string) bool {
	_, _, _, ok := voiceSpan(fold(content))
	return ok
}

// voiceSpan locates the bytes the voice section's body owns. insert says the
// block has no voice section yet, and then start == end is the line it belongs
// before: the end of the block, so the voice is always the last section.
func voiceSpan(text string) (start, end int, insert, ok bool) {
	begin := lineStart(text, Begin)
	if begin < 0 {
		return 0, 0, false, false
	}

	offset := lineStart(text[begin:], End)
	if offset < 0 {
		return 0, 0, false, false
	}
	endMarker := begin + offset

	heading, headingText := sectionHeading(text[begin:endMarker])
	if heading < 0 {
		return endMarker, endMarker, true, true
	}

	start = begin + heading + len(headingText)
	if strings.HasPrefix(text[start:], "\n") {
		start++
	}

	// The body stops at the next heading line, so a section that is not last —
	// a human moved it, or an older block grew one after it — keeps what follows.
	end = endMarker
	for from := start; from < endMarker; {
		newline := strings.IndexByte(text[from:endMarker], '\n')
		if newline < 0 {
			break
		}
		lineStart := from + newline + 1
		if lineStart >= endMarker {
			break
		}
		if text[lineStart] == '#' {
			end = lineStart
			break
		}
		from = lineStart
	}
	return start, end, false, true
}

// lineStart is the index of the marker when it owns its own line, and -1 when it
// does not: a marker inside another line is not a block this tool wrote.
// sectionHeading finds the voice heading, preferring the English one. The
// Spanish alias keeps a block written before the language migration readable
// and rewritable.
func sectionHeading(text string) (int, string) {
	if heading := lineStart(text, Section); heading >= 0 {
		return heading, Section
	}
	if heading := lineStart(text, SectionES); heading >= 0 {
		return heading, SectionES
	}
	return -1, ""
}

func lineStart(text, marker string) int {
	for from := 0; from < len(text); {
		index := strings.Index(text[from:], marker)
		if index < 0 {
			return -1
		}
		index += from
		if index == 0 || text[index-1] == '\n' {
			return index
		}
		from = index + len(marker)
	}
	return -1
}
