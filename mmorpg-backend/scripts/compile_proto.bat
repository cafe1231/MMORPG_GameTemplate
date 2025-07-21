@echo off
REM Protocol Buffer Compilation Script for Windows
REM Compiles .proto files for both Go backend and C++ Unreal Engine client

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

REM Check if protoc is installed
where protoc >nul 2>nul
if %errorlevel% neq 0 (
    echo %RED%Error: protoc is not installed%NC%
    echo Please install Protocol Buffers compiler:
    echo   - Download from https://github.com/protocolbuffers/protobuf/releases
    echo   - Add to PATH environment variable
    exit /b 1
)

REM Check if Go protoc plugin is installed
where protoc-gen-go >nul 2>nul
if %errorlevel% neq 0 (
    echo %YELLOW%Installing Go protoc plugin...%NC%
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
)

REM Check if Go gRPC plugin is installed
where protoc-gen-go-grpc >nul 2>nul
if %errorlevel% neq 0 (
    echo %YELLOW%Installing Go gRPC plugin...%NC%
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
)

echo %GREEN%Starting Protocol Buffer compilation...%NC%

REM Create output directories if they don't exist
if not exist "%GO_OUT_DIR%" mkdir "%GO_OUT_DIR%"
if not exist "%CPP_OUT_DIR%" mkdir "%CPP_OUT_DIR%"

REM Compile for Go
echo %YELLOW%Compiling for Go...%NC%
cd /d "%PROJECT_ROOT%"

for %%f in ("%PROTO_DIR%\*.proto") do (
    echo   Compiling %%~nxf...
    protoc ^
        --go_out="%GO_OUT_DIR%" ^
        --go_opt=paths=source_relative ^
        --go-grpc_out="%GO_OUT_DIR%" ^
        --go-grpc_opt=paths=source_relative ^
        -I "%PROTO_DIR%" ^
        "%%f"
    
    if !errorlevel! neq 0 (
        echo %RED%Error compiling %%~nxf%NC%
        exit /b 1
    )
)

echo %GREEN%Go compilation complete!%NC%

REM Compile for C++ (Unreal Engine)
echo %YELLOW%Compiling for C++ (Unreal Engine)...%NC%

for %%f in ("%PROTO_DIR%\*.proto") do (
    echo   Compiling %%~nxf...
    protoc ^
        --cpp_out="%CPP_OUT_DIR%" ^
        -I "%PROTO_DIR%" ^
        "%%f"
    
    if !errorlevel! neq 0 (
        echo %RED%Error compiling %%~nxf%NC%
        exit /b 1
    )
)

echo %GREEN%C++ compilation complete!%NC%

REM Generate a summary header for C++
echo %YELLOW%Generating C++ include header...%NC%
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

echo %GREEN%Protocol Buffer compilation completed successfully!%NC%
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