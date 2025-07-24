@echo off
echo ========================================
echo MMORPG Template Build Error Checker
echo ========================================
echo.
echo Building project and capturing detailed errors...
echo.

"C:\Program Files\Epic Games\UE_5.6\Engine\Build\BatchFiles\Build.bat" MMORPGTemplateEditor Win64 DebugGame -Project="%~dp0MMORPGTemplate\MMORPGTemplate.uproject" -WaitMutex -FromMsBuild -architecture=x64 > build_output.txt 2>&1

echo.
echo Build complete. Check build_output.txt for detailed error messages.
echo.
echo Last 50 lines of output:
echo ========================================
powershell -command "Get-Content build_output.txt -Tail 50"
echo.
pause