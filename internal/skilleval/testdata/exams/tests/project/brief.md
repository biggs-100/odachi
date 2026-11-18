# config

The service reads its settings from one line of key-value pairs. The first
version parses that line and nothing else: no files, no environment, no defaults.

## Not included (first version)
- reading a file or the environment: the line is given.
- nested keys: a key is a word.
- quoting: a value runs to the next separator.
