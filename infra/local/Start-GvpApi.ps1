#Requires -Version 5.1
# Start-GvpApi.ps1 — launch the Go API (18092) and media worker (8081).
# Idempotent: skips binaries already listening. Logs to %LOCALAPPDATA%\gvp\logs.
$ErrorActionPreference = 'Stop'
$bin = Join-Path $env:LOCALAPPDATA 'gvp\bin'
$logs = Join-Path $env:LOCALAPPDATA 'gvp\logs'
$media = Join-Path $env:LOCALAPPDATA 'gvp\media'
New-Item -ItemType Directory -Force $logs, $media | Out-Null

function Test-Listening($port) {
  $c = Get-NetTCPConnection -ErrorAction SilentlyContinue | Where-Object { $_.State -eq 'Listen' -and $_.LocalPort -eq $port }
  return $null -ne $c
}

if (-not (Test-Listening 18092)) {
  $env:PORT = '18092'
  $env:GVP_ENV = 'test'
  $env:GVP_ALLOW_DEV_TOKENS = 'true'
  $env:GVP_LOCAL_MEDIA_DIR = $media
  $env:MEDIA_WORKER_URL = 'http://localhost:8081'
  Start-Process -FilePath (Join-Path $bin 'gvp-api.exe') -WorkingDirectory $bin `
    -RedirectStandardOutput (Join-Path $logs 'gvp-api.out.log') `
    -RedirectStandardError (Join-Path $logs 'gvp-api.err.log') `
    -WindowStyle Hidden
  Write-Output 'gvp-api starting on :18092'
} else {
  Write-Output 'gvp-api already listening on :18092'
}

if (-not (Test-Listening 8081)) {
  # Children inherit this process env; re-point PORT at the worker.
  $env:PORT = '8081'
  Start-Process -FilePath (Join-Path $bin 'gvp-media-worker.exe') -WorkingDirectory $bin `
    -RedirectStandardOutput (Join-Path $logs 'gvp-media-worker.out.log') `
    -RedirectStandardError (Join-Path $logs 'gvp-media-worker.err.log') `
    -WindowStyle Hidden
  Write-Output 'gvp-media-worker starting on :8081'
} else {
  Write-Output 'gvp-media-worker already listening on :8081'
}
