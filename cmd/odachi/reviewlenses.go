package main

import reviewassets "github.com/biggs-100/odachi/internal/assets/review"

// reviewLenses is the Lenses port for `odachi review`: the four specialist
// prompts compiled into the binary. They live in internal/assets/review as
// content, not as code, so a review needs no network and no repository.
type reviewLenses struct{}

// Lens returns one lens prompt verbatim. The domain strips the frontmatter when
// it builds the prompt; the file stays complete so a human can read it.
func (reviewLenses) Lens(name string) (string, error) { return reviewassets.Lens(name) }
