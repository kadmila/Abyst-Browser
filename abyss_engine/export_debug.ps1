Write-Output "Deleting \AbyssUI\AbyssCLI"
Remove-Item ..\AbyssUI\AbyssCLI\* -Recurse

Write-Output "Deleting \AbyssUIBuild\AbyssCLI"
Remove-Item ..\AbyssUIBuild\AbyssCLI\* -Recurse

Write-Output "Copying \Debug to \AbyssUI\AbyssCLI"
Copy-Item -Path .\bin\Debug\net8.0\* -Destination ..\AbyssUI\AbyssCLI -Recurse

Write-Output "Copying \Debug to \AbyssUIBuild\AbyssCLI"
Copy-Item -Path .\bin\Debug\net8.0\* -Destination ..\AbyssUIBuild\AbyssCLI -Recurse

Write-Output "Copying \ABI to \AbyssUI\Assets\Host\ABI"
Copy-Item -Path .\ABI\* -Destination ..\AbyssUI\Assets\Host\ABI -Recurse