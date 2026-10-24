# main only takes a change that came through a pull request whose gate passed.
#
# The policy is .github/main-ruleset.json; this script applies it by name, so
# running it twice updates the same ruleset instead of adding a second one. The
# four required checks are the job names in .github/workflows/gate.yml: rename a
# job there and rename it here, because a required check with no run behind it
# keeps every pull request blocked.
#
# Needs gh authenticated with admin on the repository:
#   gh auth status
#
$ErrorActionPreference = 'Stop'

$root = (git rev-parse --show-toplevel).Trim()
$policy = Join-Path $root '.github/main-ruleset.json'
$body = Get-Content -Path $policy -Raw
$name = ($body | ConvertFrom-Json).name

$repo = (gh repo view --json nameWithOwner -q .nameWithOwner).Trim()
$filter = ".[] | select(.name == `"$name`") | .id"
$id = (gh api "repos/$repo/rulesets" --jq $filter).Trim()
if ($LASTEXITCODE -ne 0) {
    throw "could not list the rulesets of $repo; check gh auth status"
}

# gh reads --input as bytes; a BOM on Windows would make the API reject the
# document, so the temp file is written without one.
$tmp = [IO.Path]::GetTempFileName()
try {
    [IO.File]::WriteAllText($tmp, $body, [Text.UTF8Encoding]::new($false))

    if ($id) {
        gh api -X PUT "repos/$repo/rulesets/$id" --input $tmp | Out-Null
        Write-Output "updated the ruleset `"$name`" on $repo"
    } else {
        $id = (gh api -X POST "repos/$repo/rulesets" --input $tmp --jq .id).Trim()
        Write-Output "created the ruleset `"$name`" on $repo"
    }
} finally {
    Remove-Item $tmp -ErrorAction SilentlyContinue
}

$state = gh api "repos/$repo/rulesets/$id" | ConvertFrom-Json
$rules = ($state.rules | ForEach-Object { $_.type }) -join ', '
$required = ($state.rules |
    Where-Object { $_.type -eq 'required_status_checks' } |
    ForEach-Object { $_.parameters.required_status_checks.context }) -join ', '

Write-Output "rules: $rules"
if ($required) {
    Write-Output "required: $required"
} else {
    Write-Output "required: none; check that .github/main-ruleset.json carries the rule"
}
