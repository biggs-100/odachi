# config

## Requirements (verbatim, from the user)
R1 · "Parse reads a line of key-value pairs into a map, with two accepted separators and an error for a pair with no key or no value"

## Done when
- `a=b` and `a:b` both answer `{"a":"b"}`
- `a=b:c` keeps `b:c` as the value
- `a=` and `=b` answer an error that names the pair
- an empty line answers an empty map
