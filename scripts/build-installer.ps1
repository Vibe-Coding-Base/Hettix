<#
.SYNOPSIS
  Builds the full Hettix Windows installer: admin UI, binary, bundled browser
  and the Inno Setup package.

.DESCRIPTION
  Runs end to end:
    1. Builds the admin SPA and embeds it into the binary.
    2. Compiles hettix.exe (CGO-free).
    3. Fetches a portable Chromium into browser/ (unless -SkipBrowser).
    4. Compiles installer\hettix.iss with Inno Setup (ISCC.exe).
  The resulting installer is written to installer\out\.

.PARAMETER Version
  Version stamped into the installer filename and metadata.

.PARAMETER SkipBrowser
  Reuse an already-downloaded browser/ folder instead of fetching it again.
#>
param(
  [string]$Version = "0.0.0",
  [switch]$SkipBrowser
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
Set-Location $repoRoot

Write-Host "==> Building admin UI"
Push-Location admin
npm ci
npm run build
Pop-Location

Write-Host "==> Embedding admin UI"
Remove-Item -Recurse -Force cmd\hettix\admin -ErrorAction SilentlyContinue
Copy-Item -Recurse admin\dist cmd\hettix\admin

Write-Host "==> Building hettix.exe"
$env:CGO_ENABLED = "0"
go build -o hettix.exe ./cmd/hettix

if (-not $SkipBrowser) {
  Write-Host "==> Fetching portable Chromium"
  & "$PSScriptRoot\fetch-browser.ps1"
}

Write-Host "==> Compiling installer"
$iscc = Get-Command ISCC.exe -ErrorAction SilentlyContinue
if (-not $iscc) {
  foreach ($p in @(
    "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
    "${env:ProgramFiles}\Inno Setup 6\ISCC.exe"
  )) {
    if (Test-Path $p) { $iscc = $p; break }
  }
}
if (-not $iscc) {
  throw "Inno Setup (ISCC.exe) not found. Install it: winget install JRSoftware.InnoSetup"
}

& $iscc "/DMyAppVersion=$Version" "installer\hettix.iss"

Write-Host "==> Done. Installer is in installer\out\"
