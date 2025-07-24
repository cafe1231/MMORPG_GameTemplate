# Phase 1B - Frontend Authentication Implementation Guide

## Overview
This guide provides step-by-step instructions for implementing the frontend authentication system in Unreal Engine 5.6. It's designed for beginners and includes detailed explanations for each step.

## Prerequisites
- Phase 1A backend authentication must be running
- Unreal Engine 5.6 installed
- Visual Studio 2022 (Windows) or Xcode (macOS)
- Basic understanding of Unreal Engine interface

## Part 1: C++ Implementation (Core Systems)

### Step 1: Authentication Data Structures
The authentication data structures have been created in:
- `MMORPGTemplate/Source/MMORPGCore/Public/Types/FAuthTypes.h`

These structures include:
- `FLoginRequest` - Data sent when logging in
- `FLoginResponse` - Data received after successful login
- `FRegisterRequest` - Data sent when registering
- `FRegisterResponse` - Data received after successful registration
- `FUserInfo` - Basic user information

### Step 2: Authentication Subsystem
The authentication subsystem has been created in:
- `MMORPGTemplate/Source/MMORPGCore/Public/Subsystems/UMMORPGAuthSubsystem.h`
- `MMORPGTemplate/Source/MMORPGCore/Private/Subsystems/MMORPGAuthSubsystem.cpp`

This subsystem provides:
- Login functionality
- Registration functionality
- Token management
- Auto-refresh capabilities
- Session persistence

### Step 3: Save Game System
The save game class has been created in:
- `MMORPGTemplate/Source/MMORPGCore/Public/SaveGame/UMMORPGAuthSaveGame.h`
- `MMORPGTemplate/Source/MMORPGCore/Private/SaveGame/MMORPGAuthSaveGame.cpp`

This handles:
- Saving authentication tokens
- Persisting user preferences
- Auto-login functionality

## Part 2: Blueprint Implementation (UI Creation)

### Step 1: Compile the C++ Code
1. Close Unreal Engine if it's open
2. Right-click on `MMORPGTemplate.uproject`
3. Select "Generate Visual Studio project files"
4. Open the `.sln` file in Visual Studio
5. Build the project (Ctrl+Shift+B)
6. Launch Unreal Engine

### Step 2: Create Main Menu Game Mode

#### 2.1 Create Folder Structure
1. In Content Browser, create folders:
   ```
   Content/
   ├── Blueprints/
   │   ├── GameModes/
   │   ├── PlayerControllers/
   │   └── SaveGames/
   └── UI/
       └── Auth/
   ```

#### 2.2 Create Main Menu Game Mode
1. Right-click in `Content/Blueprints/GameModes/`
2. Select `Blueprint Class`
3. Search for and select `GameModeBase`
4. Name it `BP_MainMenuGameMode`
5. Open the blueprint
6. In Class Defaults:
   - Set `Default Pawn Class` to `None`
   - Set `HUD Class` to `None`
7. Compile and Save

#### 2.3 Create Main Menu Player Controller
1. Right-click in `Content/Blueprints/PlayerControllers/`
2. Select `Blueprint Class`
3. Search for and select `PlayerController`
4. Name it `BP_MainMenuPlayerController`
5. Open the blueprint
6. In Event Graph, add `Event BeginPlay`
7. From BeginPlay, add node `Set Show Mouse Cursor`
   - Check the `Show Mouse Cursor` checkbox
8. Add node `Set Input Mode UI Only`
9. Compile and Save

#### 2.4 Link Game Mode and Player Controller
1. Open `BP_MainMenuGameMode`
2. In Class Defaults:
   - Set `Player Controller Class` to `BP_MainMenuPlayerController`
3. Compile and Save

### Step 3: Create Login Widget

#### 3.1 Create the Widget Blueprint
1. Right-click in `Content/UI/Auth/`
2. Select `User Interface` > `Widget Blueprint`
3. Name it `WBP_LoginScreen`
4. Open the widget

#### 3.2 Design the Login UI
1. **Add Canvas Panel**:
   - In Palette, drag `Canvas Panel` to the hierarchy
   - In Details panel, set anchors to full screen (preset in top-left)

2. **Add Background**:
   - Add an `Image` widget as child of Canvas Panel
   - Set anchors to full screen
   - Set a dark color or background image
   - Name it `BackgroundImage`

3. **Add Central Container**:
   - Add `Vertical Box` to Canvas Panel
   - Set anchors to center
   - Set position to (0, 0)
   - Set alignment to 0.5, 0.5
   - Name it `LoginContainer`
   - Set padding to 20 all around

4. **Add Title**:
   - In LoginContainer, add `Text` widget
   - Set text to "Login to MMORPG"
   - Set font size to 36
   - Set horizontal alignment to center
   - Name it `TitleText`

5. **Add Email Input**:
   - Add `Editable Text Box`
   - Set hint text to "Email"
   - Set font size to 16
   - Set padding to 10
   - Name it `EmailInput`

6. **Add Password Input**:
   - Add `Editable Text Box`
   - Set hint text to "Password"
   - Set `Is Password` to true
   - Set font size to 16
   - Set padding to 10
   - Name it `PasswordInput`

7. **Add Error Text**:
   - Add `Text` widget
   - Set text color to red
   - Set visibility to `Collapsed`
   - Name it `ErrorText`

8. **Add Login Button**:
   - Add `Button` widget
   - Add `Text` as child of button
   - Set button text to "Login"
   - Set button size to 200x40
   - Name button `LoginButton`
   - Name text `LoginButtonText`

9. **Add Register Link**:
   - Add `Text` widget
   - Set text to "Don't have an account? Register"
   - Set font size to 14
   - Name it `RegisterLink`

#### 3.3 Add Login Functionality

1. **Switch to Graph mode** (top-right tabs)

2. **Add Variables**:
   - Click `+` in Variables section
   - Add variable `AuthSubsystem` of type `MMORPG Auth Subsystem` (Object Reference)
   - Add variable `IsLoading` of type `Boolean`

3. **Initialize Auth Subsystem**:
   ```
   Event Construct → Get Game Instance → Get Subsystem (MMORPGAuthSubsystem) → Set AuthSubsystem
   ```

4. **Create Login Function**:
   - Create new function called `AttemptLogin`
   - Add logic:
     ```
     Get EmailInput → GetText → ToString → Local Variable "Email"
     Get PasswordInput → GetText → ToString → Local Variable "Password"
     
     // Validate inputs
     If Email is empty OR Password is empty:
         Set ErrorText → SetText "Please fill in all fields"
         Set ErrorText → SetVisibility (Visible)
         Return
     
     // Show loading state
     Set IsLoading = true
     Set LoginButton → SetIsEnabled (false)
     Set ErrorText → SetVisibility (Collapsed)
     
     // Call login
     AuthSubsystem → Login:
         Email: Email variable
         Password: Password variable
         On Success: Call "OnLoginSuccess"
         On Failure: Call "OnLoginFailure"
     ```

5. **Create Success Handler**:
   - Create function `OnLoginSuccess`
   - Add parameter `LoginResponse` of type `FLoginResponse`
   - Logic:
     ```
     Print String "Login successful!"
     // TODO: Open character selection screen
     ```

6. **Create Failure Handler**:
   - Create function `OnLoginFailure`
   - Add parameter `Error` of type `FMMORPGError`
   - Logic:
     ```
     Set IsLoading = false
     Set LoginButton → SetIsEnabled (true)
     Set ErrorText → SetText (Error.Message)
     Set ErrorText → SetVisibility (Visible)
     ```

7. **Wire Login Button**:
   - Select LoginButton in Designer
   - In Details panel, find Events section
   - Click `+` next to `On Clicked`
   - In the graph: `OnClicked (LoginButton) → AttemptLogin`

### Step 4: Create Register Widget

#### 4.1 Duplicate Login Widget
1. Right-click `WBP_LoginScreen`
2. Select `Duplicate`
3. Name it `WBP_RegisterScreen`

#### 4.2 Modify for Registration
1. Change title to "Create Account"
2. Add `Editable Text Box` for Username (between Email and Password)
3. Add `Editable Text Box` for Confirm Password (after Password)
4. Change button text to "Create Account"
5. Change link text to "Already have an account? Login"

#### 4.3 Update Functionality
1. Rename `AttemptLogin` to `AttemptRegister`
2. Update to use `AuthSubsystem → Register` instead of Login
3. Add password confirmation validation

### Step 5: Create Main Menu Level

#### 5.1 Create the Level
1. File → New Level → Empty Level
2. Save as `Content/Maps/MainMenu`

#### 5.2 Set Game Mode
1. World Settings (Window → World Settings)
2. Set `GameMode Override` to `BP_MainMenuGameMode`

#### 5.3 Add Login Widget to Screen
1. Open `BP_MainMenuPlayerController`
2. In BeginPlay, after Set Input Mode:
   ```
   Create Widget:
       Class: WBP_LoginScreen
   → Add to Viewport
   ```

### Step 6: Test the Login System

#### 6.1 Set Default Map
1. Edit → Project Settings
2. Maps & Modes
3. Set `Editor Startup Map` to `MainMenu`
4. Set `Game Default Map` to `MainMenu`

#### 6.2 Ensure Backend is Running
```bash
cd mmorpg-backend
docker-compose up -d
go run cmd/gateway/main.go
go run cmd/auth/main.go
```

#### 6.3 Test Login
1. Play in Editor (PIE)
2. Try logging in with test credentials
3. Check console for success/error messages

## Part 3: Character System (After Login Works)

### Step 1: Character Selection Widget
*To be implemented after basic login is working*

### Step 2: Character Creation Widget
*To be implemented after character selection*

## Troubleshooting

### Common Issues

1. **"Cannot find MMORPGAuthSubsystem"**
   - Make sure you've compiled the C++ code
   - Restart Unreal Engine after compilation

2. **Login button does nothing**
   - Check that backend services are running
   - Verify the API URLs in the auth subsystem
   - Check Output Log for errors

3. **"Connection refused" errors**
   - Ensure Docker is running
   - Check that auth service is running on port 8081
   - Verify gateway is running on port 8080

### Debug Tips

1. **Enable verbose logging**:
   - Edit → Project Settings → Engine → General Settings
   - Set Log Level to `VeryVerbose`

2. **Use Print String liberally**:
   - Add Print String nodes to track execution flow
   - Different colors for different states

3. **Check the Output Log**:
   - Window → Developer Tools → Output Log
   - Filter by "MMORPG" to see relevant logs

## Next Steps

Once basic login is working:
1. Implement character selection screen
2. Add auto-login functionality
3. Implement logout
4. Add loading indicators
5. Polish UI with animations

---
**Last Updated**: 2025-07-24
**Guide Version**: 1.0