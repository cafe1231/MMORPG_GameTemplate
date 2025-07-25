@echo off
echo ========================================
echo MMORPG Template Clean and Rebuild Script
echo ========================================
echo.

echo Step 1: Cleaning project files...
if exist "MMORPGTemplate\Binaries" (
    echo Deleting Binaries folder...
    rmdir /s /q "MMORPGTemplate\Binaries"
)
if exist "MMORPGTemplate\Intermediate" (
    echo Deleting Intermediate folder...
    rmdir /s /q "MMORPGTemplate\Intermediate"
)
if exist "MMORPGTemplate\Saved\Cooked" (
    echo Deleting Cooked folder...
    rmdir /s /q "MMORPGTemplate\Saved\Cooked"
)
echo Clean complete!
echo.

echo Step 2: Looking for Unreal Engine installation...
set UE_PATH=
if exist "C:\Program Files\Epic Games\UE_5.6" (
    set "UE_PATH=C:\Program Files\Epic Games\UE_5.6"
) else if exist "D:\Epic Games\UE_5.6" (
    set "UE_PATH=D:\Epic Games\UE_5.6"
) else if exist "E:\Epic Games\UE_5.6" (
    set "UE_PATH=E:\Epic Games\UE_5.6"
) else (
    echo ERROR: Could not find Unreal Engine 5.6 installation!
    echo Please edit this script to set the correct UE_PATH
    pause
    exit /b 1
)
echo Found Unreal Engine at: %UE_PATH%
echo.

echo Step 3: Generating Visual Studio project files...
"%UE_PATH%\Engine\Build\BatchFiles\GenerateProjectFiles.bat" "%cd%\MMORPGTemplate\MMORPGTemplate.uproject" -game -rocket -progress
if %errorlevel% neq 0 (
    echo ERROR: Failed to generate project files!
    pause
    exit /b 1
)
echo Project files generated!
echo.

echo Step 4: Building the project...
"%UE_PATH%\Engine\Build\BatchFiles\Build.bat" MMORPGTemplateEditor Win64 Development "%cd%\MMORPGTemplate\MMORPGTemplate.uproject" -waitmutex
if %errorlevel% neq 0 (
    echo ERROR: Build failed!
    pause
    exit /b 1
)
echo.

echo ========================================
echo Build completed successfully!
echo ========================================
echo.
echo You can now open MMORPGTemplate.uproject in Unreal Engine.
echo.
pause