# MMORPG Developer Console Guide

## Overview

The MMORPG Template includes a built-in developer console that provides essential debugging and development tools directly in-game. The console is designed to be extensible, allowing you to add custom commands for your specific game needs.

## Accessing the Console

### Default Key Binding
- Press **F1** to toggle the console visibility
- Or use the console command: `mmorpg.console`

### Blueprint Access
```blueprint
Get MMORPG Core Module → Get Developer Console → Toggle Console
```

### C++ Access
```cpp
#include "Console/MMORPGDeveloperConsole.h"

// Get console instance
UMMORPGDeveloperConsole* Console = FMMORPGCoreModule::Get().GetDeveloperConsole();
if (Console)
{
    Console->ToggleConsole();
}
```

## Built-in Commands

### System Commands
- `help` - Show all available commands
- `help <command>` - Show help for a specific command
- `clear` - Clear console output
- `quit` - Exit the game

### Network Commands
- `connect <host> <port>` - Connect to MMORPG server
- `disconnect` - Disconnect from server
- `test` - Run connection test
- `netstats` - Show network statistics

### Debug Commands
- `status` - Show system status
- `memstats` - Show memory statistics
- `fps` - Toggle FPS display

## Adding Custom Commands

### Blueprint Method
1. Create a Blueprint that inherits from `Object`
2. In the Construction Script:
```blueprint
Get MMORPG Core Module → Get Developer Console → 
Register Command("mycommand", "Description", OnCommandExecuted)
```

### C++ Method
```cpp
// In your initialization code
UMMORPGDeveloperConsole* Console = FMMORPGCoreModule::Get().GetDeveloperConsole();
if (Console)
{
    Console->RegisterCommand(
        TEXT("mycommand"),
        TEXT("My custom command description"),
        [](const TArray<FString>& Args)
        {
            // Command implementation
            UE_LOG(LogTemp, Log, TEXT("My command executed with %d args"), Args.Num());
        }
    );
}
```

### Using the Macro
```cpp
// In any .cpp file
MMORPG_CONSOLE_COMMAND(MyCommand, "My command description", 
    [](const TArray<FString>& Args)
    {
        // Command implementation
    }
);
```

## Creating the Console UI Widget

The console requires a UI widget to be created in the editor. Here's how to set it up:

### 1. Create Console Widget Blueprint
1. In Content Browser, navigate to `Plugins/MMORPGTemplate/Content/UI/`
2. Right-click → User Interface → Widget Blueprint
3. Name it `WBP_DeveloperConsole`
4. Set parent class to `MMORPGConsoleWidget`

### 2. Design the Console UI
Essential elements to include:
- **Output ScrollBox** - For displaying console output
- **Input TextBox** - For entering commands
- **Background Panel** - Semi-transparent background

### 3. Implement Required Functions
Override these Blueprint events:
- `AddOutputLine(Text, Color)` - Add text to output with color
- `ClearOutput()` - Clear all output text
- `SetKeyboardFocus()` - Focus the input field

### 4. Handle Input
In the Input TextBox:
- **OnTextCommitted** → Call `OnCommandSubmitted`
- Handle **Up/Down Arrow** keys → Call `NavigateHistory`
- Handle **Tab** key → Call `GetAutocompleteSuggestions`

## Console Features

### Command History
- Use **Up/Down arrows** to navigate through command history
- History is saved between sessions
- Maximum 100 commands stored

### Auto-completion
- Press **Tab** to auto-complete commands
- Shows suggestions for partial matches

### Color-coded Output
- **White** - Normal output
- **Green** - Success messages
- **Yellow** - Warnings
- **Red** - Errors
- **Cyan** - Command echo
- **Gray** - Debug information

### Output Management
- Maximum 500 lines of output
- Automatic scrolling to newest content
- Timestamps on all messages

## Configuration

### Config File
Edit `Config/DefaultMMORPG.ini`:
```ini
[/Script/MMORPGCore.MMORPGDeveloperConsole]
MaxHistorySize=100
MaxOutputLines=500
EnableTimestamps=true
DefaultTextColor=(R=1.0,G=1.0,B=1.0,A=1.0)
```

### Runtime Configuration
```cpp
Console->SetMaxHistorySize(200);
Console->SetMaxOutputLines(1000);
```

## Best Practices

### Command Naming
- Use lowercase with no spaces
- Use dots for namespacing: `game.spawn.enemy`
- Keep names short but descriptive

### Command Implementation
```cpp
void HandleSpawnCommand(const TArray<FString>& Args)
{
    // Validate arguments
    if (Args.Num() < 2)
    {
        Console->WriteOutput("Usage: spawn <type> <count>", FColor::Red);
        return;
    }
    
    // Parse arguments
    FString Type = Args[0];
    int32 Count = FCString::Atoi(*Args[1]);
    
    // Execute command
    for (int32 i = 0; i < Count; i++)
    {
        // Spawn logic here
    }
    
    // Provide feedback
    Console->WriteOutput(
        FString::Printf(TEXT("Spawned %d %s"), Count, *Type), 
        FColor::Green
    );
}
```

### Error Handling
Always provide clear error messages:
```cpp
if (!IsValid(Target))
{
    Console->WriteOutput("Error: No valid target selected", FColor::Red);
    return;
}
```

## Debugging with Console

### Logging Integration
All console output is also logged to:
- Output Log (with `LogMMORPGConsole` category)
- Log files in `Saved/Logs/`

### Performance Monitoring
```
// Enable performance stats
stat fps
stat unit
stat game

// Custom stats
mmorpg.stats.network
mmorpg.stats.memory
```

### Network Debugging
```
// Test connection
connect localhost 8090
test

// Monitor network
netstats
mmorpg.net.verbose 1
```

## Security Considerations

### Shipping Builds
- Console is automatically disabled in shipping builds
- Remove sensitive commands before shipping
- Consider role-based command access

### Command Validation
Always validate input:
```cpp
// Sanitize string input
FString SafeInput = Args[0].Replace(TEXT("'"), TEXT(""));

// Validate numeric input
int32 Value;
if (!FDefaultValueHelper::ParseInt(Args[1], Value))
{
    Console->WriteOutput("Invalid number format", FColor::Red);
    return;
}

// Range validation
Value = FMath::Clamp(Value, 0, 100);
```

## Extending the Console

### Custom Output Formatting
```cpp
// Table output
Console->WriteOutput("=== Player Stats ===", FColor::Yellow);
Console->WriteOutput("Health:  100/100", FColor::Green);
Console->WriteOutput("Mana:    50/75", FColor::Cyan);
Console->WriteOutput("Level:   12", FColor::White);
```

### Interactive Commands
```cpp
Console->RegisterCommand("confirm", "Confirm previous action",
    [this](const TArray<FString>& Args)
    {
        if (PendingAction.IsValid())
        {
            ExecutePendingAction();
            Console->WriteOutput("Action confirmed", FColor::Green);
        }
        else
        {
            Console->WriteOutput("No pending action", FColor::Yellow);
        }
    }
);
```

### Command Aliases
```cpp
// Register same handler for multiple commands
auto ConnectHandler = [](const TArray<FString>& Args) { /* ... */ };
Console->RegisterCommand("connect", "Connect to server", ConnectHandler);
Console->RegisterCommand("c", "Alias for connect", ConnectHandler);
```

## Troubleshooting

### Console Not Appearing
1. Check if console widget is created
2. Verify F1 key binding is not overridden
3. Ensure not in shipping build

### Commands Not Working
1. Check command is registered
2. Verify argument count and format
3. Check log for error messages

### Performance Issues
1. Limit output lines: `Console->SetMaxOutputLines(100)`
2. Clear old output: `clear`
3. Disable verbose logging in production

## Examples

### Game-Specific Commands
```cpp
// Teleport command
MMORPG_CONSOLE_COMMAND(teleport, "Teleport to coordinates",
    [](const TArray<FString>& Args)
    {
        if (Args.Num() < 3) return;
        
        float X = FCString::Atof(*Args[0]);
        float Y = FCString::Atof(*Args[1]);
        float Z = FCString::Atof(*Args[2]);
        
        // Teleport player to coordinates
        FVector Location(X, Y, Z);
        // Implementation...
    }
);

// Give item command
MMORPG_CONSOLE_COMMAND(giveitem, "Give item to player",
    [](const TArray<FString>& Args)
    {
        if (Args.Num() < 2) return;
        
        FString ItemID = Args[0];
        int32 Quantity = FCString::Atoi(*Args[1]);
        
        // Add item to inventory
        // Implementation...
    }
);
```

This developer console provides a powerful foundation for debugging and development. Extend it with your own commands to streamline your development workflow!