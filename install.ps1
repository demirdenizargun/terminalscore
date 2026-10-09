# terminalscore kurulum betiği (Windows).
#
#   irm https://raw.githubusercontent.com/demirdenizargun/terminalscore/main/install.ps1 | iex
#
# Programı %LOCALAPPDATA%\Programs\terminalscore klasörüne kurar ve bu klasörü
# kullanıcının PATH'ine ekler (yönetici izni gerekmez).

$ErrorActionPreference = 'Stop'
$repo = 'demirdenizargun/terminalscore'

function Install-TerminalScore {
    $arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }

    $version = $env:TERMINALSCORE_VERSION
    if (-not $version) {
        try {
            $release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
        } catch {
            throw 'En son sürüm bulunamadı (henüz sürüm yayınlanmamış olabilir).'
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
        Write-Host "terminalscore $version indiriliyor (windows/$arch)..."
        Invoke-WebRequest "$base/$archive" -OutFile (Join-Path $tmp $archive) -UseBasicParsing
        Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt') -UseBasicParsing

        $line = Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -match " $([regex]::Escape($archive))$" }
        if (-not $line) { throw "$archive için sağlama bilgisi yok." }
        $expected = ($line -split ' ')[0]
        $actual = (Get-FileHash (Join-Path $tmp $archive) -Algorithm SHA256).Hash.ToLower()
        if ($actual -ne $expected) { throw 'İndirilen dosya bozuk (sağlama tutmadı).' }

        $installDir = Join-Path $env:LOCALAPPDATA 'Programs\terminalscore'
        New-Item -ItemType Directory -Force -Path $installDir | Out-Null
        Expand-Archive (Join-Path $tmp $archive) -DestinationPath $tmp -Force
        Copy-Item (Join-Path $tmp 'terminalscore.exe') $installDir -Force
        Write-Host "Kuruldu: $installDir\terminalscore.exe"

        $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
        if (-not (($userPath -split ';') -contains $installDir)) {
            [Environment]::SetEnvironmentVariable('Path', ($userPath.TrimEnd(';') + ';' + $installDir), 'User')
            Write-Host 'Klasör PATH''e eklendi. Yeni bir terminal açıp şunu yaz: terminalscore'
        } else {
            Write-Host 'Başlatmak için: terminalscore'
        }
    } finally {
        Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
    }
}

Install-TerminalScore
