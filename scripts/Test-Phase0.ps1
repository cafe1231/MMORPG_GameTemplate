# Phase 0 Test Script
# Requires PowerShell 5.0+

$ErrorActionPreference = "Continue"

Write-Host "`n=====================================" -ForegroundColor Cyan
Write-Host "   MMORPG TEMPLATE - PHASE 0 TEST" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan
Write-Host ""

# Test results storage
$testResults = @{
    Environment = @{}
    Backend = @{}
    Plugin = @{}
    Docker = @{}
}

# Function to test command availability
function Test-Command {
    param($Command)
    $exists = $null -ne (Get-Command $Command -ErrorAction SilentlyContinue)
    return $exists
}

# Function to print test result
function Print-TestResult {
    param(
        [string]$TestName,
        [bool]$Success,
        [string]$Message = ""
    )
    
    if ($Success) {
        Write-Host "[✓] $TestName" -ForegroundColor Green
    } else {
        Write-Host "[✗] $TestName" -ForegroundColor Red
    }
    
    if ($Message) {
        Write-Host "    $Message" -ForegroundColor Gray
    }
}

# 1. Environment Check
Write-Host "`n[1] Environment Check" -ForegroundColor Yellow
Write-Host "--------------------" -ForegroundColor Gray

# Check Go
$hasGo = Test-Command "go"
$testResults.Environment.Go = $hasGo
Print-TestResult "Go Installation" $hasGo
if ($hasGo) {
    $goVersion = go version
    Write-Host "    $goVersion" -ForegroundColor Gray
}

# Check Docker
$hasDocker = Test-Command "docker"
$testResults.Environment.Docker = $hasDocker
Print-TestResult "Docker Installation" $hasDocker
if ($hasDocker) {
    $dockerVersion = docker --version
    Write-Host "    $dockerVersion" -ForegroundColor Gray
}

# Check Git
$hasGit = Test-Command "git"
$testResults.Environment.Git = $hasGit
Print-TestResult "Git Installation" $hasGit

# Check Protoc
$hasProtoc = Test-Command "protoc"
$testResults.Environment.Protoc = $hasProtoc
Print-TestResult "Protocol Buffers Compiler" $hasProtoc

# 2. Backend Tests
Write-Host "`n[2] Backend Tests" -ForegroundColor Yellow
Write-Host "-----------------" -ForegroundColor Gray

# Check if backend can compile
if ($hasGo) {
    Push-Location mmorpg-backend
    
    # Test build
    Write-Host "Building gateway..." -ForegroundColor Gray
    $buildResult = go build -o bin\gateway.exe cmd\gateway\main.go 2>&1
    $buildSuccess = $LASTEXITCODE -eq 0
    $testResults.Backend.Build = $buildSuccess
    Print-TestResult "Backend Compilation" $buildSuccess
    
    # Check go.mod
    $goModExists = Test-Path "go.mod"
    $testResults.Backend.GoMod = $goModExists
    Print-TestResult "Go Modules" $goModExists
    
    # Check proto files
    $protoFiles = Get-ChildItem -Path "pkg\proto" -Filter "*.proto" -ErrorAction SilentlyContinue
    $hasProtos = $protoFiles.Count -gt 0
    $testResults.Backend.Protos = $hasProtos
    Print-TestResult "Protocol Buffer Files" $hasProtos "Found $($protoFiles.Count) proto files"
    
    Pop-Location
}

# 3. Plugin Tests
Write-Host "`n[3] Unreal Plugin Tests" -ForegroundColor Yellow
Write-Host "-----------------------" -ForegroundColor Gray

$pluginPath = "UnrealEngine\Plugins\MMORPGTemplate"
$pluginExists = Test-Path "$pluginPath\MMORPGTemplate.uplugin"
$testResults.Plugin.Exists = $pluginExists
Print-TestResult "Plugin Structure" $pluginExists

if ($pluginExists) {
    # Check source files
    $cppFiles = Get-ChildItem -Path "$pluginPath\Source" -Filter "*.cpp" -Recurse -ErrorAction SilentlyContinue
    $headerFiles = Get-ChildItem -Path "$pluginPath\Source" -Filter "*.h" -Recurse -ErrorAction SilentlyContinue
    
    $testResults.Plugin.SourceFiles = ($cppFiles.Count -gt 0) -and ($headerFiles.Count -gt 0)
    Print-TestResult "C++ Source Files" $testResults.Plugin.SourceFiles "$($cppFiles.Count) .cpp, $($headerFiles.Count) .h files"
    
    # Check build files
    $buildFiles = Get-ChildItem -Path "$pluginPath\Source" -Filter "*.Build.cs" -Recurse -ErrorAction SilentlyContinue
    $testResults.Plugin.BuildFiles = $buildFiles.Count -gt 0
    Print-TestResult "Build Configuration" $testResults.Plugin.BuildFiles "$($buildFiles.Count) build files"
}

# 4. Docker Tests
Write-Host "`n[4] Docker Tests" -ForegroundColor Yellow
Write-Host "----------------" -ForegroundColor Gray

if ($hasDocker) {
    # Check if Docker daemon is running
    $dockerInfo = docker info 2>&1
    $dockerRunning = $LASTEXITCODE -eq 0
    $testResults.Docker.Running = $dockerRunning
    Print-TestResult "Docker Daemon" $dockerRunning
    
    if ($dockerRunning) {
        # Check docker-compose file
        $composeExists = Test-Path "mmorpg-backend\docker-compose.yml"
        $testResults.Docker.Compose = $composeExists
        Print-TestResult "Docker Compose File" $composeExists
        
        # Check if any containers are running
        $containers = docker ps --format "table {{.Names}}" | Select-Object -Skip 1
        if ($containers) {
            Write-Host "    Running containers:" -ForegroundColor Gray
            foreach ($container in $containers) {
                Write-Host "    - $container" -ForegroundColor Gray
            }
        }
    } else {
        Write-Host "    Docker Desktop is not running!" -ForegroundColor Yellow
        Write-Host "    Please start Docker Desktop to run backend services." -ForegroundColor Yellow
    }
}

# 5. Documentation Tests
Write-Host "`n[5] Documentation Tests" -ForegroundColor Yellow
Write-Host "-----------------------" -ForegroundColor Gray

$docFiles = @(
    "README.md",
    "docs\phases\phase0\PHASE0_COMPLETION_REPORT.md",
    "docs\guides\QUICKSTART.md",
    "docs\guides\DEVELOPMENT_SETUP.md"
)

$allDocsExist = $true
foreach ($doc in $docFiles) {
    $exists = Test-Path $doc
    if (-not $exists) { $allDocsExist = $false }
    Print-TestResult "$doc" $exists
}
$testResults.Documentation = $allDocsExist

# 6. Git Repository Test
Write-Host "`n[6] Git Repository Tests" -ForegroundColor Yellow
Write-Host "------------------------" -ForegroundColor Gray

if ($hasGit) {
    $isGitRepo = Test-Path ".git"
    $testResults.Git.Repository = $isGitRepo
    Print-TestResult "Git Repository" $isGitRepo
    
    if ($isGitRepo) {
        $gitStatus = git status --porcelain
        $hasChanges = $gitStatus.Length -gt 0
        Print-TestResult "Working Directory" (-not $hasChanges) $(if ($hasChanges) { "Has uncommitted changes" } else { "Clean" })
        
        $currentBranch = git branch --show-current
        Write-Host "    Current branch: $currentBranch" -ForegroundColor Gray
        
        $remoteUrl = git remote get-url origin 2>$null
        if ($remoteUrl) {
            Write-Host "    Remote: $remoteUrl" -ForegroundColor Gray
        }
    }
}

# Summary
Write-Host "`n=====================================" -ForegroundColor Cyan
Write-Host "   TEST SUMMARY" -ForegroundColor Cyan
Write-Host "=====================================" -ForegroundColor Cyan

$totalTests = 0
$passedTests = 0

foreach ($category in $testResults.Keys) {
    foreach ($test in $testResults[$category].Keys) {
        $totalTests++
        if ($testResults[$category][$test]) {
            $passedTests++
        }
    }
}

$successRate = [math]::Round(($passedTests / $totalTests) * 100, 0)
$color = if ($successRate -ge 80) { "Green" } elseif ($successRate -ge 60) { "Yellow" } else { "Red" }

Write-Host "`nTests Passed: $passedTests/$totalTests ($successRate%)" -ForegroundColor $color

# Next Steps
Write-Host "`n[Next Steps]" -ForegroundColor Cyan
Write-Host "------------" -ForegroundColor Gray

if (-not $testResults.Docker.Running) {
    Write-Host "1. Start Docker Desktop" -ForegroundColor Yellow
}

Write-Host "2. Start backend services:" -ForegroundColor White
Write-Host "   cd mmorpg-backend" -ForegroundColor Gray
Write-Host "   docker-compose up -d" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Run the gateway:" -ForegroundColor White
Write-Host "   go run cmd/gateway/main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "4. Test the connection:" -ForegroundColor White
Write-Host "   curl http://localhost:8090/api/v1/test" -ForegroundColor Gray
Write-Host ""
Write-Host "5. Open Unreal Engine and test the plugin" -ForegroundColor White
Write-Host ""
Write-Host "For detailed testing instructions, see: PHASE0_TEST_GUIDE.md" -ForegroundColor Cyan
Write-Host ""