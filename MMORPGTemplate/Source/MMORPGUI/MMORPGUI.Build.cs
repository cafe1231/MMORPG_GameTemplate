using UnrealBuildTool;

public class MMORPGUI : ModuleRules
{
    public MMORPGUI(ReadOnlyTargetRules Target) : base(Target)
    {
        PCHUsage = PCHUsageMode.UseExplicitOrSharedPCHs;

        PublicDependencyModuleNames.AddRange(new string[] 
        { 
            "Core", 
            "CoreUObject", 
            "Engine", 
            "UMG",
            "Slate",
            "SlateCore",
            "MMORPGCore",
            "MMORPGNetwork"
        });

        PrivateDependencyModuleNames.AddRange(new string[] 
        { 
            "InputCore"
        });

        // Public include paths
        PublicIncludePaths.AddRange(new string[] 
        {
            "MMORPGUI/Public",
            "MMORPGUI/Public/Auth"
        });

        // Private include paths
        PrivateIncludePaths.AddRange(new string[] 
        {
            "MMORPGUI/Private"
        });

        // Include paths from other modules
        PublicIncludePaths.AddRange(new string[]
        {
            "MMORPGCore/Public"
        });
    }
}