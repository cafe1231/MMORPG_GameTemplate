# Phase 1B Link Error Fix

## Problem
Live Coding compilation fails with unresolved external symbols when MMORPGUI module tries to link with MMORPGCore module.

## Error Messages
```
error LNK2019: unresolved external symbol "__declspec(dllimport) class UClass * __cdecl Z_Construct_UClass_UMMORPGAuthSubsystem_NoRegister(void)"
error LNK2019: unresolved external symbol "__declspec(dllimport) public: void __cdecl UMMORPGAuthSubsystem::Login(struct FLoginRequest const &)"
error LNK2019: unresolved external symbol "__declspec(dllimport) public: void __cdecl UMMORPGAuthSubsystem::Register(struct FRegisterRequest const &)"
```

## Root Cause
Live Coding sometimes has issues with inter-module dependencies when new classes are added or when module dependencies change significantly.

## Solution

### Step 1: Close Unreal Editor
Close the Unreal Editor completely.

### Step 2: Clean the Project
Delete these folders:
- `MMORPGTemplate/Binaries`
- `MMORPGTemplate/Intermediate`
- `MMORPGTemplate/Saved/Cooked` (if exists)

### Step 3: Regenerate Project Files
Right-click on `MMORPGTemplate.uproject` and select "Generate Visual Studio project files"

### Step 4: Rebuild in Visual Studio
1. Open `MMORPGTemplate.sln` in Visual Studio
2. Set configuration to "Development Editor"
3. Set platform to "Win64"
4. Right-click on MMORPGTemplate in Solution Explorer
5. Select "Rebuild"

### Step 5: Launch from Visual Studio
Press F5 to launch Unreal Editor with debugger attached.

## Alternative Quick Fix (if above doesn't work)

Add explicit module loading order in `MMORPGTemplate.uproject`:
```json
"Modules": [
    {
        "Name": "MMORPGCore",
        "Type": "Runtime",
        "LoadingPhase": "Default"
    },
    {
        "Name": "MMORPGNetwork",
        "Type": "Runtime",
        "LoadingPhase": "Default"
    },
    {
        "Name": "MMORPGUI",
        "Type": "Runtime",
        "LoadingPhase": "PostDefault"  // Load after Core
    },
    {
        "Name": "MMORPGTemplate",
        "Type": "Runtime",
        "LoadingPhase": "PostDefault"
    }
]
```

## Prevention
1. Always do a full rebuild when adding new modules or changing module dependencies
2. Use forward declarations where possible to reduce header dependencies
3. Ensure proper module export macros (MODULENAME_API) are used

## Files Updated
- `MMORPGUI.Build.cs` - Added MMORPGCore/Public to include paths
- Widget CPP files - Fixed include paths for UMMORPGAuthSubsystem