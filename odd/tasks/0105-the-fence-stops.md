# the-fence-stops
Base: d9d6bd75d21fafd42ece448bdd8e305c3a92e061

## Requirements (verbatim, from the user)
R1 · "the fence stops protecting the CLI surface, so the feature after it can change a default: the policy moves in its own commit, before the code that depends on it"

## Done when
- this feature closes with no code, and the fence in force no longer names cmd/odachi/main.go
- the two rules that stay are the verdict shape and the proof format, unchanged
- the next feature can change the CLI default and CI reads a base fence that allows it
- a test drives nothing new here: the change is the policy, and check proves it by passing

## Steps
- [ ] T1 · the fence stops naming the CLI, and the brief says why   R1

## History
L1 · "the fence stops protecting the CLI surface, so the feature after it can change a default: the policy moves in its own commit, before the code that depends on it"
L2 · OFICIO · "the policy moves before the code: this feature carries no code, and the next one changes the CLI default under the fence this commit leaves behind" (porque CI judges a pull request with the fence of its base, so a default that lives in main.go can only change after the fence in main stops protecting it)
