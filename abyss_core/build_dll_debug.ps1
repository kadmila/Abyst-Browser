Write-Output "building abyssnet.dll"
go build -tags=debug -o abyssnet.dll -buildmode=c-shared .\native_dll\.

Write-Output "Copying abyssnet.dll to \abyss_engine\bin\Debug\net8.0\"
Copy-Item abyssnet.dll ..\abyss_engine\bin\Debug\net8.0\