# Authentication System Error Codes

## Error Code Ranges
- **1000-1999**: Network errors
- **2000-2999**: Authentication errors  
- **3000-3999**: Protocol/Parsing errors
- **4000-4999**: Game logic errors
- **5000-5999**: System errors

## Specific Error Codes Used in Phase 1B

### Network Errors (1000-1999)
- **1001**: Network subsystem not available
- **1002**: Failed to create HTTP client

### Authentication Errors (2000-2999)
- **2001**: No refresh token available
- **2002**: No saved credentials found

### Protocol Errors (3000-3999)
- **3001**: Failed to parse login response
- **3002**: Failed to parse register response
- **3003**: Failed to parse refresh response

## Usage Example
```cpp
// Creating an error
FMMORPGError Error(1001, "Network subsystem not available", EMMORPGErrorCategory::Network);

// In Blueprint
if (Error.Code >= 1000 && Error.Code < 2000) 
{
    // Handle network error
}
```