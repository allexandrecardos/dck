param(
  [Parameter(Mandatory = $true)]
  [string]$Version
)

$ErrorActionPreference = "Stop"

$dist = "dist"
New-Item -ItemType Directory -Force -Path $dist | Out-Null

$ldflags = "-X github.com/allexandrecardos/dck/cmd.version=$Version"

Write-Host "[INFO] Building $Version" -ForegroundColor Cyan

$env:GOOS = "windows"; $env:GOARCH = "amd64"
go build -ldflags $ldflags -o (Join-Path $dist "dck_${Version}_windows_amd64.exe") .

$env:GOOS = "linux"; $env:GOARCH = "amd64"
go build -ldflags $ldflags -o (Join-Path $dist "dck_${Version}_linux_amd64") .

$env:GOOS = "linux"; $env:GOARCH = "arm64"
go build -ldflags $ldflags -o (Join-Path $dist "dck_${Version}_linux_arm64") .

Remove-Item Env:GOOS; Remove-Item Env:GOARCH

Write-Host "[INFO] Done. Outputs in dist/" -ForegroundColor Cyan
