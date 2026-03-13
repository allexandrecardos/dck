param(
  [string]$Repo = "allexandrecardos/dck",
  [string]$InstallDir = "$env:ProgramFiles\dck",
  [string]$Version = ""
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($Version)) {
  $Version = $env:DCK_VERSION
}

if ([string]::IsNullOrWhiteSpace($Version)) {
  try {
    $release = Invoke-RestMethod -Headers @{ "User-Agent" = "dck-installer" } -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $release.tag_name
  } catch {
    Write-Host "[ERROR] Failed to resolve latest version. Use -Version vX.Y.Z or set DCK_VERSION" -ForegroundColor Red
    throw
  }
}

$asset = "dck_${Version}_windows_amd64.exe"
$url = "https://github.com/$Repo/releases/download/$Version/$asset"

Write-Host "[INFO] Downloading $url" -ForegroundColor Cyan
try {
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
} catch {
  $fallback = Join-Path $env:USERPROFILE "dck"
  Write-Host "[WARNING] Cannot write to $InstallDir. Falling back to $fallback" -ForegroundColor Yellow
  $InstallDir = $fallback
  New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

$dest = Join-Path $InstallDir "dck.exe"
$downloaded = $false

try {
  Invoke-WebRequest -Uri $url -OutFile $dest -UseBasicParsing
  $downloaded = $true
} catch {
  Write-Host "[WARNING] Download failed with Invoke-WebRequest. Trying BITS..." -ForegroundColor Yellow
  try {
    Start-BitsTransfer -Source $url -Destination $dest
    $downloaded = $true
  } catch {
    Write-Host "[ERROR] Failed to download. Please check your connection and try again." -ForegroundColor Red
    throw
  }
}

if (-not $downloaded) {
  throw "Download failed"
}

$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$InstallDir*") {
  [Environment]::SetEnvironmentVariable("Path", "$path;$InstallDir", "User")
  Write-Host "[INFO] Added to PATH (User). Restart terminal." -ForegroundColor Cyan
}

Write-Host "[INFO] Installed dck to $dest" -ForegroundColor Cyan
Write-Host "[INFO] Try: dck version" -ForegroundColor Cyan
