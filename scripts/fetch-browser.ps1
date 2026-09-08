<#
.SYNOPSIS
  Downloads a portable Chromium into the `browser/` folder so it can be shipped
  alongside hettix.exe (Burp-style embedded browser).

.DESCRIPTION
  Fetches the latest official Chromium snapshot for Win_x64 and extracts it to
  <repo>/browser/chrome-win/, which is where the launcher (pkg/chrome) looks for
  a bundled browser next to the executable. Run before building the installer.

.PARAMETER Revision
  A specific Chromium snapshot revision. Defaults to the latest (LAST_CHANGE).
#>
param(
  [string]$Revision = ""
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$browserDir = Join-Path $repoRoot "browser"
$base = "https://storage.googleapis.com/chromium-browser-snapshots/Win_x64"

if ([string]::IsNullOrEmpty($Revision)) {
  Write-Host "Resolving latest Chromium revision..."
  $Revision = (Invoke-WebRequest -UseBasicParsing "$base/LAST_CHANGE").Content.Trim()
}

Write-Host "Using Chromium revision $Revision"

$zipUrl = "$base/$Revision/chrome-win.zip"
$tmpZip = Join-Path $env:TEMP "hettix-chrome-win-$Revision.zip"

Write-Host "Downloading $zipUrl ..."
Invoke-WebRequest -UseBasicParsing $zipUrl -OutFile $tmpZip

if (Test-Path $browserDir) {
  Remove-Item -Recurse -Force $browserDir
}
New-Item -ItemType Directory -Force -Path $browserDir | Out-Null

Write-Host "Extracting to $browserDir ..."
Expand-Archive -Path $tmpZip -DestinationPath $browserDir -Force
Remove-Item -Force $tmpZip

$chrome = Join-Path $browserDir "chrome-win\chrome.exe"
if (-not (Test-Path $chrome)) {
  throw "Expected chrome.exe at $chrome but it was not found."
}

Write-Host "Chromium ready at $chrome"
