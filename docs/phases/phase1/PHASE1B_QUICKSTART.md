# Phase 1B Quick Start Guide

## Compile and Setup

1. **Open Project in Unreal Engine 5.6**
   - Double-click `MMORPGTemplate.uproject`
   - Let Unreal compile the C++ code

2. **Create Blueprint Widgets**

### Create Login Widget (WBP_Login)
1. Content Browser > Add > User Interface > Widget Blueprint
2. Name it `WBP_Login`
3. Parent class: `MMORPGLoginWidget`
4. Design the widget with:
   - EmailTextBox (Editable Text Box)
   - PasswordTextBox (Editable Text Box, Is Password = true)
   - LoginButton (Button)
   - RegisterButton (Button)
   - ErrorText (Text Block)

### Create Register Widget (WBP_Register)
1. Create `WBP_Register` with parent `MMORPGRegisterWidget`
2. Add components:
   - EmailTextBox
   - UsernameTextBox
   - PasswordTextBox (Is Password = true)
   - ConfirmPasswordTextBox (Is Password = true)
   - RegisterButton
   - BackToLoginButton
   - ErrorText

### Create Main Auth Widget (WBP_Auth)
1. Create `WBP_Auth` with parent `MMORPGAuthWidget`
2. Add components:
   - AuthSwitcher (Widget Switcher)
   - Add WBP_Login as child of switcher
   - Add WBP_Register as child of switcher
3. Set variables:
   - LoginWidget = WBP_Login instance
   - RegisterWidget = WBP_Register instance

## Create Game Mode

1. **Create BP_AuthGameMode**
   - Parent: `AMMORPGAuthGameMode`
   - Set Auth Widget Class to `WBP_Auth`

2. **Create BP_AuthPlayerController**
   - Parent: `AMMORPGAuthPlayerController`

3. **Configure Game Mode**
   - Set Player Controller Class to `BP_AuthPlayerController`

## Create Auth Level

1. **New Level**
   - File > New Level > Basic
   - Save as `AuthLevel`

2. **World Settings**
   - Game Mode Override: `BP_AuthGameMode`

3. **Project Settings**
   - Maps & Modes > Default Maps
   - Editor Startup Map: `AuthLevel`
   - Game Default Map: `AuthLevel`

## Test Without Backend

To test UI flow without a backend:

1. In `UMMORPGAuthSubsystem::Initialize()`, change:
   ```cpp
   ServerURL = TEXT("http://localhost:3000");
   ```
   to a mock server or comment out HTTP calls

2. Or implement mock responses in the handlers:
   ```cpp
   void UMMORPGAuthSubsystem::HandleLoginResponse(...)
   {
       FAuthResponse MockResponse;
       MockResponse.bSuccess = true;
       MockResponse.Message = TEXT("Mock login successful");
       OnLoginResponse.Broadcast(MockResponse);
   }
   ```

## Run and Test

1. **Play in Editor**
   - Click Play button
   - Should see login screen
   - Test form validation
   - Test switching between login/register

2. **Check Logs**
   - Window > Developer Tools > Output Log
   - Look for any errors or warnings

## Common Issues

1. **Widgets not showing**: Check widget hierarchy and bindings
2. **Buttons not working**: Ensure button bindings in NativeConstruct
3. **HTTP errors**: Check server URL and ensure backend is running
4. **Compilation errors**: Regenerate project files and rebuild

## Next Steps

1. Style the widgets with your game's theme
2. Add loading indicators during API calls  
3. Implement remember me functionality
4. Add password strength indicator
5. Connect to actual backend server