#Requires -Version 5.1
# Start-GvpTunnel.ps1 — quick tunnel localhost:18092 -> Cloudflare, then point
# the api-proxy worker at the fresh tunnel URL.
# Reads the Cloudflare API token from %APPDATA%\gvp\cf_api_token.txt (created
# once by the operator; never stored in the repo).
$ErrorActionPreference = 'Stop'
$logs = Join-Path $env:LOCALAPPDATA 'gvp\logs'
New-Item -ItemType Directory -Force $logs | Out-Null
$tokenFile = Join-Path $env:APPDATA 'gvp\cf_api_token.txt'
if (-not (Test-Path $tokenFile)) { throw "Missing token file: $tokenFile" }
$cfToken = (Get-Content $tokenFile -Raw).Trim()
$cfAccount = 'f5c363135efe9b22ca66eb39f2f8e1da'
$repoRoot = 'C:\Users\dnjst\Desktop\project\video'

$cloudflared = 'C:\Program Files (x86)\cloudflared\cloudflared.exe'
if (-not (Test-Path $cloudflared)) { $cloudflared = (Get-Command cloudflared -ErrorAction Stop).Source }

$tunnelLog = Join-Path $logs 'gvp-tunnel.out.log'
if (Test-Path $tunnelLog) { Remove-Item $tunnelLog -Force }
$tunnelErr = Join-Path $logs 'gvp-tunnel.err.log'
if (Test-Path $tunnelErr) { Remove-Item $tunnelErr -Force }

Start-Process -FilePath $cloudflared -ArgumentList 'tunnel', '--no-autoupdate', '--url', 'http://localhost:18092' `
  -RedirectStandardOutput $tunnelLog -RedirectStandardError $tunnelErr -WindowStyle Hidden
Write-Output 'cloudflared quick tunnel starting...'

$tunnelUrl = $null
for ($i = 0; $i -lt 36; $i++) {
  Start-Sleep -Seconds 5
  $lines = @()
  if (Test-Path $tunnelLog) { $lines += Get-Content $tunnelLog -ErrorAction SilentlyContinue }
  if (Test-Path $tunnelErr) { $lines += Get-Content $tunnelErr -ErrorAction SilentlyContinue }
  $hit = $lines | Select-String -Pattern 'https://[a-z0-9-]+\.trycloudflare\.com' | Select-Object -First 1
  if ($hit) { $tunnelUrl = $hit.Matches.Value; break }
}
if (-not $tunnelUrl) { throw 'Timed out waiting for tunnel URL (see gvP-tunnel logs)' }
Write-Output "tunnel URL: $tunnelUrl"
Set-Content (Join-Path $logs 'gvp-tunnel.url') $tunnelUrl

# Point the stable api-proxy worker at the fresh tunnel URL.
$env:CLOUDFLARE_API_TOKEN = $cfToken
$env:CLOUDFLARE_ACCOUNT_ID = $cfAccount
Push-Location (Join-Path $repoRoot 'apps\api-proxy')
try {
  $tunnelUrl | npx --yes wrangler@4.158.0 secret put API_UPSTREAM 2>&1 | Select-Object -Last 3
} finally {
  Pop-Location
  $env:CLOUDFLARE_API_TOKEN = $null
  $cfToken = $null
}
Write-Output 'api-proxy upstream updated'
