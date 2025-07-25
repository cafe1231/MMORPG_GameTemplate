using UnrealBuildTool;

public class MMORPGNetwork : ModuleRules
{
    public MMORPGNetwork(ReadOnlyTargetRules Target) : base(Target)
    {
        PCHUsage = PCHUsageMode.UseExplicitOrSharedPCHs;

        PublicDependencyModuleNames.AddRange(new string[] 
        { 
            "Core", 
            "CoreUObject", 
            "Engine", 
            "HTTP",
            "Json",
            "JsonUtilities",
            "WebSockets",
            "MMORPGCore"
        });

        PrivateDependencyModuleNames.AddRange(new string[] 
        { 
            "Slate", 
            "SlateCore"
        });

        // Public include paths
        PublicIncludePaths.AddRange(new string[] 
        {
            "MMORPGNetwork/Public",
            "MMORPGNetwork/Public/Http",
            "MMORPGNetwork/Public/WebSocket"
        });

        // Private include paths
        PrivateIncludePaths.AddRange(new string[] 
        {
            "MMORPGNetwork/Private"
        });
    }
}