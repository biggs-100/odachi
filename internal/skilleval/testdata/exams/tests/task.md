# T1 · config.Parse

`config.Parse` reads one line of key-value pairs and answers a map.

- A pair is `key=value` or `key:value`. The first separator on the line splits
  each pair; the other one is part of the value.
- A pair with no key, or with no value, is an error that names the pair.
- An empty line answers an empty map, not an error.

`Parse` lives in `config.go`. This is step T1 of the task document: work it, and
close it the way the project asks.
