@echo off
REM Install Protocol Buffers Compiler for Windows

setlocal enabledelayedexpansion

echo Installing Protocol Buffers Compiler...

REM Check if protoc is already installed
where protoc >nul 2>nul
if %errorlevel% equ 0 (
    echo protoc is already installed:
    protoc --version
    echo.
    echo To reinstall, delete protoc.exe from your PATH first.
    exit /b 0
)

REM Set version and URLs
set PROTOC_VERSION=25.1
set PROTOC_ZIP=protoc-%PROTOC_VERSION%-win64.zip
set DOWNLOAD_URL=https://github.com/protocolbuffers/protobuf/releases/download/v%PROTOC_VERSION%/%PROTOC_ZIP%

REM Create tools directory
set TOOLS_DIR=%~dp0..\..\tools
if not exist "%TOOLS_DIR%" mkdir "%TOOLS_DIR%"
cd /d "%TOOLS_DIR%"

echo Downloading protoc v%PROTOC_VERSION%...
powershell -Command "Invoke-WebRequest -Uri '%DOWNLOAD_URL%' -OutFile '%PROTOC_ZIP%'"

if not exist %PROTOC_ZIP% (
    echo Error: Failed to download protoc
    exit /b 1
)

echo Extracting protoc...
powershell -Command "Expand-Archive -Path '%PROTOC_ZIP%' -DestinationPath 'protoc' -Force"

if not exist protoc\bin\protoc.exe (
    echo Error: Failed to extract protoc
    exit /b 1
)

echo.
echo Protocol Buffers Compiler installed successfully!
echo.
echo To use protoc, add the following to your PATH:
echo %CD%\protoc\bin
echo.
echo Or copy protoc.exe to a directory already in your PATH.
echo.

REM Try to add to PATH for current session
set PATH=%CD%\protoc\bin;%PATH%

REM Verify installation
protoc --version

echo.
echo Installation complete!
echo.
echo Next steps:
echo 1. Add %CD%\protoc\bin to your system PATH
echo 2. Restart your terminal
echo 3. Run 'compile_proto.bat' to compile the protocol buffers

endlocal