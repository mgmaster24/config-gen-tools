#Requires -Version 5.1
# Installs the latest (or $env:NVIMFORGE_VERSION-pinned) nvimforge release
# binary for Windows. This script only downloads, verifies, and places the
# binary — all of nvimforge's actual logic lives in the Go binary itself.
$ErrorActionPreference = "Stop"

$Repo = "mgmaster24/config-gen-tools"
# config-gen-tools holds several tools, each released under its own
# tool-scoped tag (e.g. nvimforge/v1.2.3), so releases must be filtered by
# this prefix rather than taking the repo's newest release.
$Tool = "nvimforge"
$InstallDir = if ($env:NVIMFORGE_INSTALL_DIR) { $env:NVIMFORGE_INSTALL_DIR } else { "$env:LOCALAPPDATA\nvimforge\bin" }
$Version = $env:NVIMFORGE_VERSION

function Write-Info($msg) { Write-Host "==> $msg" -ForegroundColor Cyan }
function Write-ErrAndExit($msg) { Write-Host "error: $msg" -ForegroundColor Red; exit 1 }

if (-not [Environment]::Is64BitOperatingSystem) {
    Write-ErrAndExit "unsupported architecture (only amd64 is published for Windows)"
}
$Arch = "amd64"

if (-not $Version) {
    Write-Info "Resolving latest $Tool release..."
    # /releases/latest would return whichever tool released most recently, so
    # list releases (newest first) and take the first tagged for this tool.
    $releases = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases?per_page=100"
    $Version = ($releases | Where-Object { $_.tag_name -like "$Tool/*" } | Select-Object -First 1).tag_name
    if (-not $Version) { Write-ErrAndExit "could not resolve the latest $Tool release" }
}

# Accept either a bare version (v1.2.3) or a fully-qualified tag
# (nvimforge/v1.2.3) in $env:NVIMFORGE_VERSION.
if ($Version.StartsWith("$Tool/")) { $Version = $Version.Substring("$Tool/".Length) }
$VersionNum = $Version.TrimStart("v")
$Archive = "nvimforge_${VersionNum}_windows_${Arch}.zip"
$BaseUrl = "https://github.com/$Repo/releases/download/$Tool/$Version"

$WorkDir = Join-Path $env:TEMP ([System.Guid]::NewGuid())
New-Item -ItemType Directory -Path $WorkDir | Out-Null

try {
    $ArchivePath = Join-Path $WorkDir $Archive
    $ChecksumsPath = Join-Path $WorkDir "checksums.txt"

    Write-Info "Downloading $Archive ($Version)..."
    Invoke-WebRequest -Uri "$BaseUrl/$Archive" -OutFile $ArchivePath
    Invoke-WebRequest -Uri "$BaseUrl/checksums.txt" -OutFile $ChecksumsPath

    Write-Info "Verifying checksum..."
    $expectedLine = Select-String -Path $ChecksumsPath -Pattern ([Regex]::Escape($Archive))
    if (-not $expectedLine) { Write-ErrAndExit "no checksum entry found for $Archive" }
    $expected = ($expectedLine.Line -split '\s+')[0]
    $actual = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLower()
    if ($expected -ne $actual) { Write-ErrAndExit "checksum mismatch for $Archive (expected $expected, got $actual)" }

    Write-Info "Extracting..."
    Expand-Archive -Path $ArchivePath -DestinationPath $WorkDir -Force

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Path (Join-Path $WorkDir "nvimforge.exe") -Destination (Join-Path $InstallDir "nvimforge.exe") -Force

    Write-Info "Installed nvimforge $Version to $InstallDir\nvimforge.exe"

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
        Write-Info "Added $InstallDir to your User PATH. Restart your terminal for it to take effect."
    }

    Write-Info "Run 'nvimforge install' to get started."
}
finally {
    Remove-Item -Path $WorkDir -Recurse -Force -ErrorAction SilentlyContinue
}
