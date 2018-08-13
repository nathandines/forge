$ErrorActionPreference = 'Stop'

$fileName  = "forge_v${Env:ChocolateyPackageVersion}.zip"
$toolsPath = Split-Path -Parent $MyInvocation.MyCommand.Definition
$zip_path = "$toolsPath\$fileName"
Remove-Item $toolsPath\* -Recurse -Force -Exclude $fileName

$packageArgs = @{
    PackageName  = 'forge'
    FileFullPath = $zip_path
    Destination  = $toolsPath
}
Get-ChocolateyUnzip @packageArgs
Remove-Item $zip_path -ea 0

if ((Get-OSArchitectureWidth 64) -and ($Env:ChocolateyForceX86 -ne 'true')) {
    Write-Verbose "Removing x32 version"
    Remove-Item "$toolsPath/forge32.exe" -ea 0
    Move-Item "$toolsPath/forge64.exe" "$toolsPath/forge.exe" -Force
} else {
    Write-Verbose "Removing x64 version"
    Remove-Item "$toolsPath/forge64.exe" -ea 0
    Move-Item "$toolsPath/forge32.exe" "$toolsPath/forge.exe" -Force
}
