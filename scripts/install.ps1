#Requires -Version 5.1
<#
.SYNOPSIS
    Instalador de odachi para Windows: baja el binario de una release, verifica su
    sha256 y recién ahí lo instala.

.DESCRIPTION
    Si el checksum no coincide, no se instala nada: un instalador que deja un
    binario a medio verificar es peor que no tener instalador.

    Variables de entorno (las tres son opcionales):
      ODACHI_VERSION   versión a instalar (default: la última release publicada)
      ODACHI_BASE_URL  base de descarga (default: la URL de releases de GitHub)
      ODACHI_BIN_DIR   carpeta de destino (default: %LOCALAPPDATA%\odachi\bin)

.NOTES
    Equivalente Windows de scripts/install.sh. Los pasos son los mismos: detectar
    plataforma, resolver versión, bajar, verificar sha256, extraer, instalar y
    decir qué quedó y cómo llegar a él.
#>

$ErrorActionPreference = 'Stop'

# La salida de PowerShell usa la code page de la consola: sin esta línea, al
# redirigir a un archivo o a un log de CI "agregué" se convierte en "agregu?".
# El archivo en sí ya es UTF-8; lo que estaba mal era la salida.
$OutputEncoding = [Console]::OutputEncoding = [Text.Encoding]::UTF8

$repo = 'biggs-100/odachi'

function Fail([string] $message) {
    Write-Host "odachi: $message" -ForegroundColor Red
    exit 1
}

# El sha256 de un archivo, con dos caminos y el mismo resultado. Get-FileHash
# viene con PowerShell 4, pero un Windows PowerShell 5.1 puede llegar sin el
# módulo: se midió una máquina donde el cmdlet no resuelve y el instalador moría
# antes de verificar el checksum, que es justo el paso que no se puede saltear.
# certutil -hashfile trae SHA256 desde Windows 7 y es el respaldo. El camino se
# elige mirando si el cmdlet existe, no atrapando el error: así un checksum que no
# coincide sigue siendo un checksum que no coincide. Sin banderas ni variables
# nuevas.
function Get-Sha256([string] $path) {
    if (Get-Command -Name Get-FileHash -ErrorAction SilentlyContinue) {
        return (Get-FileHash -LiteralPath $path -Algorithm SHA256).Hash
    }

    if (-not (Get-Command -Name certutil.exe -ErrorAction SilentlyContinue)) {
        Fail "no pude calcular el sha256 de ${path}: esta máquina no tiene Get-FileHash ni certutil; sin hash no verifico el archivo y no instalé nada"
    }

    $lines = & certutil.exe -hashfile $path SHA256
    if ($LASTEXITCODE -ne 0) {
        Fail "no pude calcular el sha256 de $path con certutil (exit $LASTEXITCODE); sin hash no verifico el archivo y no instalé nada"
    }
    # La salida de certutil está traducida —un encabezado, el hash, un pie— así que
    # no se busca una palabra: la única línea que son 64 caracteres hexadecimales
    # es el hash, en cualquier idioma.
    foreach ($line in $lines) {
        $candidate = $line -replace '\s', ''
        if ($candidate -match '^[0-9a-fA-F]{64}$') {
            return $candidate.ToLowerInvariant()
        }
    }
    Fail "certutil no devolvió un sha256 legible para $path; sin hash no verifico el archivo y no instalé nada"
}

# 1 · plataforma. Lo que no está acá se dice, no se adivina.
$arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { 'amd64' }
    'ARM64' { 'arm64' }
    default { $null }
}
if (-not $arch) {
    Fail "no sé instalar en la arquitectura $($env:PROCESSOR_ARCHITECTURE); las releases cubren AMD64 y ARM64"
}

# 2 · versión: la que pidió el humano, o la última release.
$version = $env:ODACHI_VERSION
if ([string]::IsNullOrWhiteSpace($version)) {
    try {
        $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest" -Headers @{ 'User-Agent' = 'odachi-installer' }
        $version = $latest.tag_name
    }
    catch {
        Fail "no pude consultar la última release en GitHub ($($_.Exception.Message)); pasá ODACHI_VERSION=<versión> y volvé a correr esto"
    }
}
# El tag lleva "v" (v1.2.3); el archivo se llama odachi_1.2.3_windows_amd64.zip
$version = $version -replace '^v', ''
if ([string]::IsNullOrWhiteSpace($version)) {
    Fail 'la versión llegó vacía; pasá ODACHI_VERSION=<versión>'
}

$asset = "odachi_${version}_windows_${arch}.zip"

# 3 · base de descarga configurable, igual que en el instalador de shell.
$base = $env:ODACHI_BASE_URL
if ([string]::IsNullOrWhiteSpace($base)) {
    $base = "https://github.com/$repo/releases/download/v$version"
}
$base = $base.TrimEnd('/')

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ('odachi-install.' + [guid]::NewGuid().ToString('n'))
New-Item -ItemType Directory -Path $tmp -Force | Out-Null

try {
    $archive = Join-Path $tmp $asset
    $checksums = Join-Path $tmp 'checksums.txt'
    $bundle = Join-Path $tmp 'checksums.txt.sigstore.json'

    # 4 · descarga y verificación, en un directorio temporal: el destino no se
    # toca hasta que el archivo demuestra ser el que la release publicó.
    try {
        Invoke-WebRequest -Uri "$base/$asset" -OutFile $archive -UseBasicParsing
    }
    catch {
        Fail "no pude descargar $base/$asset ($($_.Exception.Message)); revisá la versión y tu conexión"
    }
    try {
        Invoke-WebRequest -Uri "$base/checksums.txt" -OutFile $checksums -UseBasicParsing
    }
    catch {
        Fail "no pude descargar $base/checksums.txt; sin checksums no instalo nada"
    }

    try {
        Invoke-WebRequest -Uri "$base/checksums.txt.sigstore.json" -OutFile $bundle -UseBasicParsing
    }
    catch {
        Fail "no pude descargar $base/checksums.txt.sigstore.json; sin la firma, el checksum y el binario vendrían del mismo lugar y eso no prueba nada"
    }

    # La firma es lo que hace confiable al checksum: los dos bajan del mismo host, y
    # sin verificarla la cadena termina en GitHub, no en la release. Se exige cosign:
    # un chequeo que corre sólo a veces es una degradación silenciosa.
    $cosign = if ($env:ODACHI_COSIGN) { $env:ODACHI_COSIGN } else { 'cosign' }
    if (-not (Get-Command -Name $cosign -ErrorAction SilentlyContinue) -and -not (Test-Path -LiteralPath $cosign)) {
        Fail "me falta cosign para verificar la firma de checksums.txt; instalalo desde https://docs.sigstore.dev/cosign/system_config/installation/ y volvé a correr esto"
    }
    # cosign is a native program and prints "Verified OK" on stderr. Under
    # $ErrorActionPreference 'Stop' PowerShell 5.1 turns a native program's stderr
    # into a terminating error, so the check aborted exactly when it succeeded.
    # The preference is relaxed for the call and restored: the exit code is the
    # answer, not the stream it wrote to.
    $previous = $ErrorActionPreference
    $ErrorActionPreference = 'Continue'
    & $cosign verify-blob --bundle $bundle --certificate-identity "https://github.com/$repo/.github/workflows/release.yml@refs/tags/v$version" --certificate-oidc-issuer "https://token.actions.githubusercontent.com" $checksums 2>&1 | Out-Null
    $verified = $LASTEXITCODE
    $ErrorActionPreference = $previous
    if ($verified -ne 0) {
        Fail "la firma de checksums.txt no verifica contra este workflow y esta versión: el archivo no es el que la release publicó, y no instalé nada"
    }

    $expected = $null
    foreach ($line in Get-Content -LiteralPath $checksums) {
        $fields = $line.Trim() -split '\s+'
        if ($fields.Count -ge 2 -and $fields[1].TrimStart('*') -eq $asset) {
            $expected = $fields[0]
            break
        }
    }
    if ([string]::IsNullOrWhiteSpace($expected)) {
        Fail "checksums.txt no trae el hash de $asset; no instalo algo que la release no certifica"
    }

    $actual = Get-Sha256 $archive
    if ($expected.ToLowerInvariant() -ne $actual.ToLowerInvariant()) {
        Fail "el checksum de $asset no coincide (esperaba $expected, obtuve $actual): el archivo está corrupto o no es el que dice ser, y no instalé nada"
    }

    $out = Join-Path $tmp 'out'
    Expand-Archive -LiteralPath $archive -DestinationPath $out -Force
    $source = Join-Path $out 'odachi.exe'
    if (-not (Test-Path -LiteralPath $source)) {
        Fail "$asset no trae el binario odachi.exe"
    }

    # 5 · instalación: %LOCALAPPDATA%\odachi\bin (o lo que diga ODACHI_BIN_DIR).
    $binDir = $env:ODACHI_BIN_DIR
    if ([string]::IsNullOrWhiteSpace($binDir)) {
        $binDir = Join-Path $env:LOCALAPPDATA 'odachi\bin'
    }
    New-Item -ItemType Directory -Path $binDir -Force | Out-Null
    $target = Join-Path $binDir 'odachi.exe'
    Copy-Item -LiteralPath $source -Destination $target -Force

    # 6 · qué quedó instalado, y el PATH del usuario si falta esa carpeta.
    Write-Host "odachi $version instalado en $target"

    $userPath = [Environment]::GetEnvironmentVariable('PATH', 'User')
    $entries = @()
    if (-not [string]::IsNullOrWhiteSpace($userPath)) {
        $entries = @($userPath -split ';' | Where-Object { $_ -ne '' })
    }
    $alreadyThere = $false
    foreach ($entry in $entries) {
        if ($entry.TrimEnd('\') -ieq $binDir.TrimEnd('\')) {
            $alreadyThere = $true
            break
        }
    }
    if (-not $alreadyThere) {
        $newPath = (@($entries) + $binDir) -join ';'
        [Environment]::SetEnvironmentVariable('PATH', $newPath, 'User')
        Write-Host "odachi: agregué $binDir al PATH de tu usuario; abrí una terminal nueva y probá 'odachi --version'"
    }
    else {
        Write-Host "odachi: $binDir ya estaba en tu PATH"
    }

    # 7 - la primera conversación tiene que encontrar un agente que sepa odachi.
    if ($env:ODACHI_NO_SETUP) {
        Write-Host "odachi: ODACHI_NO_SETUP=1; instalé solo el binario."
    }
    else {
        & $target setup
        if ($LASTEXITCODE -ne 0) {
            Write-Host "odachi: el binario quedó instalado, pero el setup global no terminó (exit $LASTEXITCODE); corré odachi setup otra vez." -ForegroundColor Yellow
        }
    }
}
finally {
    Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
}
