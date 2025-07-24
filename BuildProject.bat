@echo off
echo Building MMORPGTemplate...
"C:\Program Files\Epic Games\UE_5.6\Engine\Build\BatchFiles\Build.bat" MMORPGTemplateEditor Win64 Development -Project="%cd%\MMORPGTemplate\MMORPGTemplate.uproject" -WaitMutex -FromMsBuild
echo.
echo Build Exit Code: %ERRORLEVEL%
pause