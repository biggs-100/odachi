package design

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	domain "github.com/biggs-100/odachi/internal/domain/design"
)

// Source is one file the project declares its design in: a stylesheet, a token
// file, a theme. Sync reads them and writes nothing back: a tool that edits the
// sources it learns from is a tool nobody can trust twice.
type Source struct {
	Path    string
	Content string
}

// Token is one CSS custom property found in a source, with the file it came from.
type Token struct {
	Name   string
	Value  string
	Source string
}

// Font is one font family a source declares.
type Font struct {
	Family string
	Source string
}

var (
	customProperty = regexp.MustCompile(`--([a-zA-Z0-9_-]+)\s*:\s*([^;{}]+);`)
	fontFamily     = regexp.MustCompile(`font-family\s*:\s*([^;{}]+);`)
)

// ReadSources collects the tokens and fonts the sources declare, in the order the
// sources are given. A name declared twice keeps the first value: the order is the
// walker's, and the report names the source of each entry so a person can judge it.
func ReadSources(sources []Source) (tokens []Token, fonts []Font) {
	seenToken := map[string]bool{}
	seenFont := map[string]bool{}

	for _, source := range sources {
		for _, match := range customProperty.FindAllStringSubmatch(source.Content, -1) {
			name := strings.TrimSpace(match[1])
			value := strings.TrimSpace(match[2])
			if seenToken[name] || value == "" {
				continue
			}
			seenToken[name] = true
			tokens = append(tokens, Token{Name: name, Value: value, Source: source.Path})
		}
		for _, match := range fontFamily.FindAllStringSubmatch(source.Content, -1) {
			family := strings.TrimSpace(match[1])
			if seenFont[family] || family == "" {
				continue
			}
			seenFont[family] = true
			fonts = append(fonts, Font{Family: family, Source: source.Path})
		}
	}

	sort.SliceStable(tokens, func(i, j int) bool { return tokens[i].Name < tokens[j].Name })
	sort.SliceStable(fonts, func(i, j int) bool { return fonts[i].Family < fonts[j].Family })
	return tokens, fonts
}

// Sync writes the design system from the sources: design/system.md names every
// token and font with the file it came from, and design/system.html renders them so
// the canvas can show the system as a board.
//
// A project with no source gets a refusal, never an invented system: a token the
// project does not have is a token the next artboard will use on a screen that
// never had it.
func Sync(sources []Source, files Files, dir string) Result {
	tokens, fonts := ReadSources(sources)
	if len(tokens) == 0 && len(fonts) == 0 {
		return blocked("no design source found: sync reads CSS custom properties and font families, and this project declares none; say what the system is and the skill writes it by hand")
	}

	root := filepath.Join(dir, domain.Dir)
	markdown := RenderSystemMarkdown(tokens, fonts, sources)
	if err := writeFile(files, filepath.Join(root, "system.md"), markdown); err != nil {
		return blocked("could not write " + filepath.Join(root, "system.md") + ": " + oneLine(err.Error()))
	}
	html := RenderSystemHTML(tokens, fonts)
	if err := writeFile(files, filepath.Join(root, "system.html"), html); err != nil {
		return blocked("could not write " + filepath.Join(root, "system.html") + ": " + oneLine(err.Error()))
	}

	result := Result{
		Code:  CodeOK,
		Files: []string{filepath.ToSlash(filepath.Join(root, "system.md")), filepath.ToSlash(filepath.Join(root, "system.html"))},
		Lines: []string{
			fmt.Sprintf("%d token%s and %d font famil%s from %d source%s",
				len(tokens), plural(len(tokens), "", "s"), len(fonts), plural(len(fonts), "y", "ies"), len(sources), plural(len(sources), "", "s")),
			"design/system.md and design/system.html written; no source was edited",
		},
	}
	result.Message = result.Lines[0]
	return result
}

// writeFile creates the design folder and writes one file, with a backup when the
// file is there: the system is regenerated, and an older one is what a person reads
// when the new one surprises them.
func writeFile(files Files, path, content string) error {
	if exists, err := files.Exists(path); err != nil {
		return err
	} else if exists {
		if _, err := files.Backup(path); err != nil {
			return err
		}
	}
	return files.Write(path, []byte(content))
}

// RenderSystemMarkdown is the system a person reads: the tokens grouped by what they
// are for, each with its source.
func RenderSystemMarkdown(tokens []Token, fonts []Font, sources []Source) string {
	var out strings.Builder
	out.WriteString("# Design system\n\n")
	out.WriteString("Captured by `odachi design sync` from the project's own sources. Every value names the\n")
	out.WriteString("file it came from, and this file is the only thing a design run is allowed to follow.\n\n")

	out.WriteString("## Sources read\n\n")
	for _, source := range sources {
		out.WriteString("- `" + source.Path + "`\n")
	}

	colors, spacing, other := []Token{}, []Token{}, []Token{}
	for _, token := range tokens {
		lower := strings.ToLower(token.Name)
		switch {
		case strings.HasPrefix(lower, "color") || strings.HasPrefix(lower, "bg") ||
			strings.HasPrefix(lower, "fg") || strings.HasPrefix(lower, "accent") ||
			strings.HasPrefix(lower, "border") || looksLikeColor(token.Value):
			colors = append(colors, token)
		case strings.Contains(lower, "space") || strings.Contains(lower, "gap") ||
			strings.Contains(lower, "padding") || strings.Contains(lower, "radius") ||
			strings.Contains(lower, "shadow") || strings.Contains(lower, "size"):
			spacing = append(spacing, token)
		default:
			other = append(other, token)
		}
	}

	section := func(title string, list []Token) {
		if len(list) == 0 {
			return
		}
		out.WriteString("\n## " + title + "\n\n")
		out.WriteString("| Token | Value | Source |\n| --- | --- | --- |\n")
		for _, token := range list {
			out.WriteString("| `--" + token.Name + "` | `" + token.Value + "` | `" + token.Source + "` |\n")
		}
	}
	section("Colors", colors)
	section("Spacing, radius and shadows", spacing)
	section("Other tokens", other)

	if len(fonts) > 0 {
		out.WriteString("\n## Typography\n\n")
		out.WriteString("| Family | Source |\n| --- | --- |\n")
		for _, font := range fonts {
			out.WriteString("| `" + font.Family + "` | `" + font.Source + "` |\n")
		}
	}

	out.WriteString("\n## Rules\n\n")
	out.WriteString("- A design run uses only these tokens. A colour or a family that is not here is a question\n")
	out.WriteString("  for the human, not a decision the agent takes.\n")
	out.WriteString("- The system is captured from the sources and never written back into them.\n")
	return out.String()
}

// RenderSystemHTML draws the system as one self-contained board: the swatches, the
// type scale and the spacing, so the canvas can show what the project already is.
func RenderSystemHTML(tokens []Token, fonts []Font) string {
	var out strings.Builder
	out.WriteString("<!doctype html>\n<html lang=\"en\"><head><meta charset=\"utf-8\"><title>Design system</title>\n")
	out.WriteString("<style>body{font:14px/1.5 ui-sans-serif,system-ui,sans-serif;margin:24px;color:#111;background:#fff}")
	out.WriteString("h1{font-size:20px;margin:0 0 4px}p{color:#666;margin:0 0 24px}")
	out.WriteString("h2{font-size:13px;text-transform:uppercase;letter-spacing:.08em;color:#666;margin:28px 0 10px}")
	out.WriteString(".grid{display:flex;flex-wrap:wrap;gap:12px}.swatch{width:150px;border:1px solid #e5e5e5;border-radius:8px;overflow:hidden}")
	out.WriteString(".swatch .chip{height:56px}.swatch .meta{padding:6px 8px;font:11px/1.4 ui-monospace,monospace;word-break:break-all}")
	out.WriteString("table{border-collapse:collapse;width:100%}td,th{border-bottom:1px solid #eee;text-align:left;padding:6px 8px;font-size:12px}")
	out.WriteString("code{font-family:ui-monospace,monospace}</style></head><body>\n")
	out.WriteString("<h1>Design system</h1><p>Captured by odachi design sync from the project's sources.</p>\n")

	colors, rest := []Token{}, []Token{}
	for _, token := range tokens {
		if looksLikeColor(token.Value) {
			colors = append(colors, token)
			continue
		}
		rest = append(rest, token)
	}

	if len(colors) > 0 {
		out.WriteString("<h2>Colors</h2><div class=\"grid\">\n")
		for _, token := range colors {
			out.WriteString("<div class=\"swatch\"><div class=\"chip\" style=\"background:" + escapeAttr(token.Value) + "\"></div>")
			out.WriteString("<div class=\"meta\">--" + escapeAttr(token.Name) + "<br>" + escapeAttr(token.Value) + "</div></div>\n")
		}
		out.WriteString("</div>\n")
	}

	if len(rest) > 0 {
		out.WriteString("<h2>Tokens</h2><table><tr><th>Token</th><th>Value</th><th>Source</th></tr>\n")
		for _, token := range rest {
			out.WriteString("<tr><td><code>--" + escapeAttr(token.Name) + "</code></td><td><code>" + escapeAttr(token.Value) + "</code></td><td><code>" + escapeAttr(token.Source) + "</code></td></tr>\n")
		}
		out.WriteString("</table>\n")
	}

	if len(fonts) > 0 {
		out.WriteString("<h2>Typography</h2><table><tr><th>Family</th><th>Source</th></tr>\n")
		for _, font := range fonts {
			out.WriteString("<tr><td style=\"font-family:" + escapeAttr(font.Family) + "\">" + escapeAttr(font.Family) + "</td><td><code>" + escapeAttr(font.Source) + "</code></td></tr>\n")
		}
		out.WriteString("</table>\n")
	}

	out.WriteString("</body></html>\n")
	return out.String()
}

// looksLikeColor asks the value, not the name: a token called brand that holds a hex
// is a colour, and one called shadow that holds a blur is not.
func looksLikeColor(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(value, "#") {
		return true
	}
	for _, prefix := range []string{"rgb(", "rgba(", "hsl(", "hsla(", "oklch(", "oklab(", "color("} {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

// escapeAttr keeps a value inside an HTML attribute: the values come from the
// project's own files, and a stray quote must not break the board.
func escapeAttr(value string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "'", "&#39;")
	return replacer.Replace(value)
}
