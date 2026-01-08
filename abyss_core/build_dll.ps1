Write-Output "building abyssnet.dll"
go build -o abyssnet.dll -buildmode=c-shared .\windll\.

Write-Output "Copying abyssnet.dll to \abyss_engine\bin\Release\net8.0\"
Copy-Item abyssnet.dll ..\abyss_engine\bin\Release\net8.0\
Copy-Item abyssnet.h ..\abyss_engine\