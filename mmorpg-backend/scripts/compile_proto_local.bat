@echo off
REM Protocol Buffer Compilation Script for Windows
REM Uses local protoc installation

setlocal enabledelayedexpansion

REM Colors (using echo with escape sequences)
set "RED=[31m"
set "GREEN=[32m"
set "YELLOW=[33m"
set "NC=[0m"

REM Get script directory and project root
set "SCRIPT_DIR=%~dp0"
set "PROJECT_ROOT=%SCRIPT_DIR%.."
set "PROTO_DIR=%PROJECT_ROOT%\pkg\proto"
set "GO_OUT_DIR=%PROJECT_ROOT%\pkg\proto"
set "CPP_OUT_DIR=%PROJECT_ROOT%\..\UnrealEngine\Plugins\MMORPGTemplate\Source\MMORPGCore\Public\Proto"
set "PROTOC_PATH=%PROJECT_ROOT%\..\..\tools\protoc\bin\protoc.exe"

REM Check if protoc is installed
if not exist "%PROTOC_PATH%" (
    echo Error: protoc is not installed at %PROTOC_PATH%
    echo Please run install_protoc.bat first
    exit /b 1
)

REM Check if Go protoc plugin is installed
where protoc-gen-go >nul 2>nul
if %errorlevel% neq 0 (
    echo Installing Go protoc plugin...
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
)

REM Check if Go gRPC plugin is installed
where protoc-gen-go-grpc >nul 2>nul
if %errorlevel% neq 0 (
    echo Installing Go gRPC plugin...
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
)

echo Starting Protocol Buffer compilation...

REM Create output directories if they don't exist
if not exist "%GO_OUT_DIR%" mkdir "%GO_OUT_DIR%"
if not exist "%CPP_OUT_DIR%" mkdir "%CPP_OUT_DIR%"

REM Compile for Go
echo Compiling for Go...
cd /d "%PROJECT_ROOT%"

for %%f in ("%PROTO_DIR%\*.proto") do (
    echo   Compiling %%~nxf...
    "%PROTOC_PATH%" ^
        --go_out="%GO_OUT_DIR%" ^
        --go_opt=paths=source_relative ^
        --go-grpc_out="%GO_OUT_DIR%" ^
        --go-grpc_opt=paths=source_relative ^
        -I "%PROTO_DIR%" ^
        "%%f"
    
    if !errorlevel! neq 0 (
        echo Error compiling %%~nxf
        exit /b 1
    )
)

echo Go compilation complete!

REM Compile for C++ (Unreal Engine)
echo Compiling for C++ (Unreal Engine)...

for %%f in ("%PROTO_DIR%\*.proto") do (
    echo   Compiling %%~nxf...
    "%PROTOC_PATH%" ^
        --cpp_out="%CPP_OUT_DIR%" ^
        -I "%PROTO_DIR%" ^
        "%%f"
    
    if !errorlevel! neq 0 (
        echo Error compiling %%~nxf
        exit /b 1
    )
)

echo C++ compilation complete!

REM Generate a summary header for C++
echo Generating C++ include header...
(
echo // Auto-generated header file
echo // Include all Protocol Buffer headers
echo.
echo #pragma once
echo.
echo #include "base.pb.h"
echo #include "auth.pb.h"
echo #include "character.pb.h"
echo #include "world.pb.h"
echo #include "game.pb.h"
echo #include "chat.pb.h"
echo.
echo namespace MMORPG
echo {
echo     // Type aliases for easier use in Unreal Engine
echo     using namespace mmorpg;
echo }
) > "%CPP_OUT_DIR%\MMORPGProto.h"

echo Protocol Buffer compilation completed successfully!
echo.
echo Generated files:
echo   - Go files: %GO_OUT_DIR%
echo   - C++ files: %CPP_OUT_DIR%
echo.
echo To use in your code:
echo   - Go: import "github.com/mmorpg-template/backend/pkg/proto"
echo   - C++: #include "Proto/MMORPGProto.h"

endlocal
pause