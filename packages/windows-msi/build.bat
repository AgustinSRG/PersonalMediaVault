@echo off

cd "..\..\backend"

echo Building backend...

call build-production.bat

cd "..\backup-tool"

echo Building backup tool...

call build-production.bat

cd "..\launcher"

echo Building launcher...

call build-production.bat

cd "..\frontend"

echo Building frontend...

call npm install
call npm run build

cd "..\packages\windows-msi"

echo Packaging...

call powershell < configure.ps1

call make-wix.bat

