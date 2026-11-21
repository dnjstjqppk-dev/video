#Requires -Version 5.1
# Start-GvpApi.ps1 — launch the Go API (18092) and media worker (8081).
# Idempotent: skips binaries already listening. Logs to %LOCALAPPDATA%\gvp\logs.
$ErrorActionPreference = 'Stop'
$bin = Join-Path $env:LOCALAPPDATA 'gvp\bin'
$logs = Join-Path $env:LOCALAPPDATA 'gvp\logs'
$media = Join-Path $env:LOCALAPPDATA 'gvp\media'
New-Item -ItemType Directory -Force $logs, $media | Out-Null
$repoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$toolsBin = Join-Path $repoRoot 'tools\bin'
$toolsData = Join-Path $repoRoot 'tools\data'

function Test-Listening($port) {
  $c = Get-NetTCPConnection -ErrorAction SilentlyContinue | Where-Object { $_.State -eq 'Listen' -and $_.LocalPort -eq $port }
  return $null -ne $c
}

# SeaweedFS S3 (:9002) backs R2_ENDPOINT for uploads and HLS publishing.
# Paths must be absolute: scheduled tasks start with no repo working directory.
if (-not (Test-Listening 9002)) {
  $env:debug = $null # weed parses $env:debug as bool; a stray value aborts boot
  Start-Process -FilePath (Join-Path $toolsBin 'weed.exe') -ArgumentList (
    'server -dir=' + (Join-Path $toolsData 'seaweed') +
    ' -master.port=19333 -volume.port=18083 -s3 -s3.port=9002 -s3.config=' + (Join-Path $toolsBin 's3config.json')
  ) -WorkingDirectory $repoRoot `
    -RedirectStandardOutput (Join-Path $logs 'gvp-seaweed.out.log') `
    -RedirectStandardError (Join-Path $logs 'gvp-seaweed.err.log') `
    -WindowStyle Hidden
  Write-Output 'seaweed starting on :9002'
} else {
  Write-Output 'seaweed already listening on :9002'
}

if (-not (Test-Listening 18092)) {
  $env:PORT = '18092'
  $env:GVP_ENV = 'test'
  $env:GVP_ALLOW_DEV_TOKENS = 'true'
  $env:GVP_LOCAL_MEDIA_DIR = $media
  $env:MEDIA_WORKER_URL = 'http://localhost:8081'
  # Pin chunk relay: browsers cannot reach the loopback SeaweedFS endpoint
  # that presigned S3 PUTs would target, so uploads relay through the API
  # while publishing still goes to the HLS bucket.
  $env:UPLOAD_MODE = 'local'
  # Object storage: the API only publishes artifacts when R2 keys are set
  # explicitly, so read the local SeaweedFS identities (untracked,
  # operator-local; endpoints default to localhost:9002).
  $s3cfgPath = Join-Path $toolsBin 's3config.json'
  if (Test-Path $s3cfgPath) {
    $s3ident = (Get-Content $s3cfgPath -Raw | ConvertFrom-Json).identities[0].credentials[0]
    if ($s3ident.accessKey -and $s3ident.secretKey) {
      $env:R2_ACCESS_KEY_ID = $s3ident.accessKey
      $env:R2_SECRET_ACCESS_KEY = $s3ident.secretKey
    }
  }
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
  # Transcode toolchain lives in the repo (tools/bin); the worker only falls
  # back to PATH, so pin the binaries explicitly (ffmpeg/ffprobe/packager are
  # not on PATH on the operator PC).
  $env:FFMPEG_BIN = Join-Path $toolsBin 'ffmpeg.exe'
  $env:FFPROBE_BIN = Join-Path $toolsBin 'ffprobe.exe'
  $env:PACKAGER_BIN = Join-Path $toolsBin 'packager.exe'
  Start-Process -FilePath (Join-Path $bin 'gvp-media-worker.exe') -WorkingDirectory $bin `
    -RedirectStandardOutput (Join-Path $logs 'gvp-media-worker.out.log') `
    -RedirectStandardError (Join-Path $logs 'gvp-media-worker.err.log') `
    -WindowStyle Hidden
  Write-Output 'gvp-media-worker starting on :8081'
} else {
  Write-Output 'gvp-media-worker already listening on :8081'
}
