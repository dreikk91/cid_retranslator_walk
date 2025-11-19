@echo off
echo Building resource file...

REM Install rsrc if not present
where rsrc >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Installing rsrc tool...
    go install github.com/akavel/rsrc@latest
)

REM Generate .syso file from resources
echo Generating rsrc.syso from resources.rc...
rsrc -ico icon.ico -manifest multiplepages.exe.manifest -o rsrc.syso

echo Building executable...
go build -o cid_retranslator.exe .

echo Build complete!
