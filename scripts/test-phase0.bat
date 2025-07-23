@echo off
echo ====================================
echo    PHASE 0 - TEST AUTOMATIQUE
echo ====================================
echo.

echo [1] Verification de l'environnement...
echo ------------------------------------
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [X] Go n'est pas installe
) else (
    echo [OK] Go est installe
    go version
)

where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo [X] Docker n'est pas installe
) else (
    echo [OK] Docker est installe
    docker --version
)

echo.
echo [2] Test de compilation du backend...
echo ------------------------------------
cd mmorpg-backend
go build -o bin\gateway.exe cmd\gateway\main.go
if %errorlevel% neq 0 (
    echo [X] Erreur de compilation
) else (
    echo [OK] Gateway compile avec succes
)
cd ..

echo.
echo [3] Verification des fichiers cles...
echo ------------------------------------
if exist "mmorpg-backend\bin\gateway.exe" (
    echo [OK] Gateway binary present
) else (
    echo [X] Gateway binary manquant
)

if exist "UnrealEngine\Plugins\MMORPGTemplate\MMORPGTemplate.uplugin" (
    echo [OK] Plugin UE5 present
) else (
    echo [X] Plugin UE5 manquant
)

if exist "docs\phases\phase0\PHASE0_COMPLETION_REPORT.md" (
    echo [OK] Documentation complete
) else (
    echo [X] Documentation incomplete
)

echo.
echo [4] Docker Status...
echo ------------------------------------
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo [!] Docker n'est pas lance. Veuillez demarrer Docker Desktop.
) else (
    echo [OK] Docker est operationnel
    echo.
    echo Pour demarrer les services:
    echo   cd mmorpg-backend
    echo   docker-compose up -d
)

echo.
echo ====================================
echo    RESULTATS DU TEST
echo ====================================
echo.
echo Pour un test complet, consultez:
echo   PHASE0_TEST_GUIDE.md
echo.
echo Prochaines etapes:
echo 1. Demarrer Docker Desktop
echo 2. cd mmorpg-backend
echo 3. docker-compose up -d
echo 4. go run cmd/gateway/main.go
echo.
pause