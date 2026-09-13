<#
.SYNOPSIS
  Builds the full Hettix Windows installer: admin UI, native desktop app,
  bundled browser and the Inno Setup package.

.DESCRIPTION
  Runs end to end:
    1. Builds the admin SPA and embeds it into the shared adminui package.
    2. Compiles the native desktop app (Hettix.exe) with the Wails CLI.
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
yarn install --frozen-lockfile
yarn run build
Pop-Location

Write-Host "==> Embedding admin UI"
Remove-Item -Recurse -Force pkg\adminui\admin -ErrorAction SilentlyContinue
Copy-Item -Recurse admin\dist pkg\adminui\admin

Write-Host "==> Building Hettix.exe (desktop app)"
$env:CGO_ENABLED = "0"
$wails = Join-Path (go env GOPATH) "bin\wails.exe"
if (-not (Test-Path $wails)) {
  throw "Wails CLI not found at $wails. Install it: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
}
Push-Location cmd\hettix-desktop
& $wails build -s -skipbindings -ldflags "-X main.version=$Version" -o Hettix.exe
Pop-Location

# All build artifacts land in .\releases at the repo root.
New-Item -ItemType Directory -Force -Path releases | Out-Null
Copy-Item -Force cmd\hettix-desktop\build\bin\Hettix.exe releases\Hettix.exe

if (-not $SkipBrowser) {
  Write-Host "==> Fetching portable Chromium"
  & "$PSScriptRoot\fetch-browser.ps1"
}

Write-Host "==> Compiling installer"
$iscc = (Get-Command ISCC.exe -ErrorAction SilentlyContinue).Source
if (-not $iscc) {
  $candidates = @(
    "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
    "${env:ProgramFiles}\Inno Setup 6\ISCC.exe",
    "${env:LOCALAPPDATA}\Programs\Inno Setup 6\ISCC.exe"
  )
  $regKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\Inno Setup 6_is1"
  $regLoc = (Get-ItemProperty -Path $regKey -ErrorAction SilentlyContinue).InstallLocation
  if ($regLoc) { $candidates += (Join-Path $regLoc "ISCC.exe") }

  foreach ($p in $candidates) {
    if ($p -and (Test-Path $p)) { $iscc = $p; break }
  }
}
if (-not $iscc) {
  throw "Inno Setup (ISCC.exe) not found. Install it: winget install JRSoftware.InnoSetup"
}

& $iscc "/DMyAppVersion=$Version" "installer\hettix.iss"

Write-Host "==> Done. Artifacts are in .\releases"
