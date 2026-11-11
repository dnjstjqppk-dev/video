#Requires -Version 5.1
# Register-GvpTasks.ps1 — persist the API + tunnel across logons via Scheduled
# Tasks (current user, no admin needed). Also starts both right now.
$ErrorActionPreference = 'Stop'
$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$ps = Join-Path $env:SystemRoot 'System32\WindowsPowerShell\v1.0\powershell.exe'

function Register-GvpTask($name, $script) {
  $action = "`"$ps`" -NoProfile -ExecutionPolicy Bypass -File `"$script`""
  schtasks.exe /Create /TN $name /TR $action /SC ONLOGON /RL LIMITED /F 2>&1 | Out-Null
  schtasks.exe /Run /TN $name 2>&1 | Out-Null
  Write-Output "task registered and started: $name"
}

Register-GvpTask 'GVP-API' (Join-Path $here 'Start-GvpApi.ps1')
Start-Sleep -Seconds 8
Register-GvpTask 'GVP-Tunnel' (Join-Path $here 'Start-GvpTunnel.ps1')
