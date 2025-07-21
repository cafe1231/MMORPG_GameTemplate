# MMORPG Error Handling Guide

## Overview

The MMORPG Template includes a comprehensive error handling framework designed to provide consistent error reporting, logging, and recovery mechanisms throughout your game. The system supports multiple severity levels, categories, retry logic, and user-friendly error messages.

## Architecture

### Core Components

1. **UMMORPGErrorHandler** - Central error management system
2. **FMMORPGError** - Error data structure
3. **UMMORPGErrorUtils** - Blueprint utilities
4. **UMMORPGRetryHandler** - Retry logic implementation

### Error Flow

```
Error Occurs → Create Error → Report to Handler → Log/Notify/Telemetry → Delegate Broadcast
```

## Error Structure

```cpp
struct FMMORPGError
{
    int32 ErrorCode;              // Unique error identifier
    EMMORPGErrorSeverity Severity; // Info/Warning/Error/Critical/Fatal
    EMMORPGErrorCategory Category; // Network/Auth/Game/Protocol/System
    FString Message;              // Human-readable message
    FString Details;              // Technical details
    FDateTime Timestamp;          // When error occurred
    FString Context;              // Where error occurred
    FString UserAction;           // Suggested user action
    bool bCanRetry;              // If error is retryable
    int32 MaxRetries;            // Maximum retry attempts
};
```

## Error Severity Levels

### Info
- Informational messages
- No action required
- Logged but not displayed to user

### Warning
- Potential issues
- May affect gameplay
- Displayed in development builds

### Error
- Operation failed
- User should be notified
- May be recoverable

### Critical
- Serious error affecting game state
- Immediate attention required
- Always logged and notified

### Fatal
- Unrecoverable error
- Game will terminate
- Full error dump created

## Error Categories

### Network (1000-1999)
```cpp
NetworkConnectionFailed = 1001
NetworkTimeout = 1002
NetworkDisconnected = 1003
NetworkInvalidResponse = 1004
NetworkServerUnreachable = 1005
```

### Authentication (2000-2999)
```cpp
AuthInvalidCredentials = 2001
AuthTokenExpired = 2002
AuthAccountLocked = 2003
AuthSessionInvalid = 2004
AuthPermissionDenied = 2005
```

### Protocol (3000-3999)
```cpp
ProtocolVersionMismatch = 3001
ProtocolInvalidMessage = 3002
ProtocolSerializationFailed = 3003
ProtocolDeserializationFailed = 3004
```

### Game Logic (4000-4999)
```cpp
GameInvalidOperation = 4001
GameStateCorrupted = 4002
GameResourceNotFound = 4003
GameActionNotAllowed = 4004
```

### System (5000-5999)
```cpp
SystemOutOfMemory = 5001
SystemFileNotFound = 5002
SystemPermissionDenied = 5003
SystemInitializationFailed = 5004
```

## Usage Examples

### C++ Error Reporting

#### Simple Error
```cpp
MMORPG_ERROR(NetworkConnectionFailed, "Failed to connect to game server");
```

#### Detailed Error
```cpp
MMORPG_ERROR_DETAILED(
    NetworkTimeout, 
    "Connection timed out",
    FString::Printf(TEXT("Server: %s, Timeout: %ds"), *ServerURL, TimeoutSeconds)
);
```

#### Using Error Builder
```cpp
UMMORPGErrorHandler* Handler = FMMORPGCoreModule::Get().GetErrorHandler();
if (Handler)
{
    Handler->CreateError(AuthInvalidCredentials)
        .WithMessage("Invalid username or password")
        .WithSeverity(EMMORPGErrorSeverity::Error)
        .WithCategory(EMMORPGErrorCategory::Auth)
        .WithUserAction("Please check your credentials and try again")
        .CanRetry(true, 3)
        .Report();
}
```

### Blueprint Error Reporting

#### Simple Error
```blueprint
Report Error
├─ Error Code: 1001
├─ Message: "Connection failed"
└─ Severity: Error
```

#### Network Error
```blueprint
Report Network Error
├─ Error Code: 1002
├─ Message: "Request timed out"
└─ Can Retry: True
```

#### From HTTP Response
```blueprint
HTTP Response → Create Error From Response → Report Error
```

## Error Handling Patterns

### Try-Catch Pattern
```cpp
void LoginUser(const FString& Username, const FString& Password)
{
    try
    {
        // Attempt login
        AuthService->Login(Username, Password);
    }
    catch (const std::exception& e)
    {
        UMMORPGErrorHandler* Handler = GetErrorHandler();
        Handler->ReportException("LoginUser", FString(e.what()));
    }
}
```

### Result Pattern
```cpp
struct FLoginResult
{
    bool bSuccess;
    FMMORPGError Error;
    FUserData UserData;
};

FLoginResult AttemptLogin()
{
    FLoginResult Result;
    
    if (!IsNetworkAvailable())
    {
        Result.bSuccess = false;
        Result.Error = CreateError(NetworkDisconnected)
            .WithMessage("No network connection")
            .Build();
        return Result;
    }
    
    // Continue with login...
}
```

### Callback Pattern
```cpp
void ConnectToServer(TFunction<void(bool, const FMMORPGError&)> OnComplete)
{
    NetworkManager->Connect(ServerURL, [OnComplete](bool bSuccess, int32 ErrorCode)
    {
        if (!bSuccess)
        {
            FMMORPGError Error;
            Error.ErrorCode = ErrorCode;
            Error.Message = GetErrorMessage(ErrorCode);
            Error.Category = EMMORPGErrorCategory::Network;
            Error.bCanRetry = true;
            
            OnComplete(false, Error);
        }
        else
        {
            OnComplete(true, FMMORPGError());
        }
    });
}
```

## Retry Logic

### Using Retry Handler
```cpp
UMMORPGRetryHandler* RetryHandler = NewObject<UMMORPGRetryHandler>();

FMMORPGRetryPolicy Policy;
Policy.MaxAttempts = 3;
Policy.InitialDelay = 1.0f;
Policy.MaxDelay = 30.0f;
Policy.bUseExponentialBackoff = true;

RetryHandler->ExecuteWithRetry(Policy,
    [this]() -> bool
    {
        // Attempt operation
        return NetworkManager->Connect();
    },
    [this](bool bSuccess)
    {
        if (bSuccess)
        {
            UE_LOG(LogTemp, Log, TEXT("Connected successfully"));
        }
        else
        {
            MMORPG_ERROR(NetworkConnectionFailed, "Failed to connect after retries");
        }
    }
);
```

### Manual Retry Pattern
```cpp
void ConnectWithRetry(int32 AttemptsLeft = 3)
{
    NetworkManager->Connect([this, AttemptsLeft](bool bSuccess)
    {
        if (!bSuccess && AttemptsLeft > 0)
        {
            // Retry after delay
            GetWorld()->GetTimerManager().SetTimer(RetryTimer,
                [this, AttemptsLeft]()
                {
                    ConnectWithRetry(AttemptsLeft - 1);
                },
                2.0f, // 2 second delay
                false
            );
        }
        else if (!bSuccess)
        {
            MMORPG_ERROR(NetworkConnectionFailed, "Connection failed after all retries");
        }
    });
}
```

## Error Notifications

### In-Game Notifications
Errors are automatically displayed to users based on severity:
- **Info**: No notification
- **Warning**: Brief toast notification
- **Error**: Modal dialog with user action
- **Critical**: Full-screen error dialog

### Developer Console
All errors are logged to the developer console:
```
[ERROR 1001] Failed to connect to server
[ERROR 2002] Authentication token expired
```

### Log Files
Errors are saved to:
```
ProjectDir/Saved/Logs/MMORPG/Errors_YYYYMMDD.log
```

## Custom Error Handlers

### Creating Custom Handler
```cpp
class FMyErrorHandler : public IErrorHandler
{
public:
    virtual void HandleError(const FMMORPGError& Error) override
    {
        // Custom error handling
        if (Error.Category == EMMORPGErrorCategory::Network)
        {
            ShowNetworkErrorUI(Error);
        }
    }
};

// Register handler
ErrorHandler->RegisterCustomHandler(MakeShareable(new FMyErrorHandler()));
```

### Blueprint Error Events
```blueprint
Event On Error Occurred BP
├─ Error (FMMORPGError)
└─ Switch on Error Category
    ├─ Network → Show Network Error UI
    ├─ Auth → Return to Login Screen
    └─ Game → Show In-Game Alert
```

## Localization

### Setting Up Localized Messages
```cpp
// In your localization file
LOCTEXT("Error_NetworkTimeout", "Connection timed out. Please check your internet connection.")
LOCTEXT("Error_AuthFailed", "Login failed. Please check your username and password.")

// In error handler
FText GetLocalizedError(int32 ErrorCode)
{
    switch (ErrorCode)
    {
    case NetworkTimeout:
        return LOCTEXT("Error_NetworkTimeout", "...");
    case AuthInvalidCredentials:
        return LOCTEXT("Error_AuthFailed", "...");
    default:
        return FText::Format(LOCTEXT("Error_Unknown", "Unknown error ({0})"), ErrorCode);
    }
}
```

## Best Practices

### 1. Always Provide Context
```cpp
// Good
MMORPG_ERROR_DETAILED(
    GameResourceNotFound,
    "Failed to load character data",
    FString::Printf(TEXT("CharacterID: %s"), *CharacterID)
);

// Bad
MMORPG_ERROR(GameResourceNotFound, "Not found");
```

### 2. Use Appropriate Severity
- **Info**: Configuration loaded, cache cleared
- **Warning**: Fallback used, performance degraded
- **Error**: Operation failed, user action blocked
- **Critical**: Data corruption, security breach
- **Fatal**: Out of memory, critical asset missing

### 3. Provide User Actions
```cpp
Handler->CreateError(NetworkDisconnected)
    .WithMessage("Lost connection to server")
    .WithUserAction("Please check your internet connection and try reconnecting")
    .Report();
```

### 4. Make Errors Actionable
```cpp
// Provide enough information to debug
Error.Details = FString::Printf(
    TEXT("Endpoint: %s\nMethod: %s\nStatus: %d\nResponse: %s"),
    *Endpoint, *Method, StatusCode, *ResponseBody
);
```

### 5. Handle Cascading Errors
```cpp
// Prevent error spam
static FDateTime LastErrorTime;
if (FDateTime::Now() - LastErrorTime < FTimespan::FromSeconds(1))
{
    return; // Skip duplicate errors
}
LastErrorTime = FDateTime::Now();
```

## Integration with Other Systems

### Network Manager
```cpp
void FMMORPGNetworkManager::OnRequestFailed(int32 StatusCode, const FString& Response)
{
    FMMORPGError Error = UMMORPGErrorUtils::CreateErrorFromResponse(StatusCode, Response);
    GetErrorHandler()->ReportError(Error);
}
```

### Authentication
```cpp
void HandleLoginResponse(const FLoginResponse& Response)
{
    if (!Response.bSuccess)
    {
        UMMORPGErrorUtils::ReportAuthError(
            Response.ErrorCode,
            Response.ErrorMessage
        );
    }
}
```

### Game Systems
```cpp
bool EquipItem(const FItemData& Item)
{
    if (!CanEquipItem(Item))
    {
        GetErrorHandler()->CreateError(GameActionNotAllowed)
            .WithMessage("Cannot equip item")
            .WithDetails(FString::Printf(TEXT("Level requirement: %d"), Item.RequiredLevel))
            .WithUserAction("Reach the required level to equip this item")
            .Report();
        return false;
    }
    
    // Equip item...
    return true;
}
```

## Testing

### Unit Tests
```cpp
TEST_CASE("ErrorHandler")
{
    UMMORPGErrorHandler* Handler = NewObject<UMMORPGErrorHandler>();
    Handler->Initialize();
    
    // Test error reporting
    FMMORPGError TestError;
    TestError.ErrorCode = 1001;
    TestError.Message = "Test error";
    
    bool bDelegateCalled = false;
    Handler->OnErrorOccurred.AddLambda([&](const FMMORPGError& Error)
    {
        bDelegateCalled = true;
        CHECK(Error.ErrorCode == 1001);
    });
    
    Handler->ReportError(TestError);
    CHECK(bDelegateCalled);
}
```

### Integration Tests
```cpp
// Test retry logic
FMMORPGRetryPolicy Policy;
Policy.MaxAttempts = 3;
Policy.InitialDelay = 0.1f;

int32 AttemptCount = 0;
RetryHandler->ExecuteWithRetry(Policy,
    [&]() -> bool
    {
        AttemptCount++;
        return AttemptCount >= 3; // Succeed on third attempt
    },
    [](bool bSuccess)
    {
        CHECK(bSuccess);
        CHECK(AttemptCount == 3);
    }
);
```

## Performance Considerations

### Error Throttling
```cpp
// Limit error reporting rate
if (ErrorHistory.Num() > 0)
{
    const FMMORPGError& LastError = ErrorHistory.Last();
    if (LastError.ErrorCode == Error.ErrorCode &&
        (Error.Timestamp - LastError.Timestamp).GetTotalSeconds() < 1.0)
    {
        return; // Skip duplicate error
    }
}
```

### Async Error Handling
```cpp
// Don't block game thread for error handling
AsyncTask(ENamedThreads::AnyBackgroundThreadNormalTask, [Error]()
{
    // Log to file
    // Send telemetry
    // Process error data
});
```

## Debugging

### Console Commands
```
mmorpg.errors.show - Show recent errors
mmorpg.errors.clear - Clear error history
mmorpg.errors.test <code> - Generate test error
mmorpg.errors.export - Export error log
```

### Debug UI
Create a debug widget to display:
- Recent errors list
- Error statistics
- Retry status
- Network state

This comprehensive error handling system ensures your MMORPG can gracefully handle failures, provide meaningful feedback to users, and maintain stability even in adverse conditions.