#   irm https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.ps1 | iex
#
# Installs to %LOCALAPPDATA%\Programs\terminalscore and adds it to the user PATH.

$ErrorActionPreference = 'Stop'
$repo = 'demirdenizargun/terminalscore'

function Install-TerminalScore {
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }

    $version = $env:TERMINALSCORE_VERSION
    if (-not $version) {
        try {
            $release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
        } catch {
            throw "Couldn't find the latest release."
        }
        $version = $release.tag_name
    }
    $version = $version.TrimStart('v')

    $base = $env:TERMINALSCORE_BASE_URL
    if (-not $base) { $base = "https://github.com/$repo/releases/download/v$version" }

    $archive = "terminalscore_${version}_windows_${arch}.zip"
    $tmp = Join-Path ([IO.Path]::GetTempPath()) ("terminalscore-" + [guid]::NewGuid())
    New-Item -ItemType Directory -Path $tmp | Out-Null

    try {
        Write-Host "Downloading terminalscore $version (windows/$arch)..."
        Invoke-WebRequest "$base/$archive" -OutFile (Join-Path $tmp $archive) -UseBasicParsing
        Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt') -UseBasicParsing

        $line = Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -match " $([regex]::Escape($archive))$" }
        if (-not $line) { throw "No checksum for $archive." }
        $expected = ($line -split ' ')[0]
        $actual = (Get-FileHash (Join-Path $tmp $archive) -Algorithm SHA256).Hash.ToLower()
        if ($actual -ne $expected) { throw 'Checksum mismatch, the download is corrupted.' }

        $installDir = Join-Path $env:LOCALAPPDATA 'Programs\terminalscore'
        New-Item -ItemType Directory -Force -Path $installDir | Out-Null
        Expand-Archive (Join-Path $tmp $archive) -DestinationPath $tmp -Force
        Copy-Item (Join-Path $tmp 'terminalscore.exe') $installDir -Force
        Write-Host "Installed: $installDir\terminalscore.exe"

        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if (-not $userPath) { $userPath = '' }
        if (-not (($userPath -split ';') -contains $installDir)) {
            [Environment]::SetEnvironmentVariable('Path', (($userPath.TrimEnd(';') + ';' + $installDir).TrimStart(';')), 'User')
            Write-Host 'Added to PATH. Open a new terminal and run: terminalscore'
        } else {
            Write-Host 'Run it with: terminalscore'
        }
    } finally {
        Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
    }
}

Install-TerminalScore
