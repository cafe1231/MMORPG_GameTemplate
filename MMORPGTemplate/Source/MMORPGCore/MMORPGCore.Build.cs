using UnrealBuildTool;

public class MMORPGCore : ModuleRules
{
    public MMORPGCore(ReadOnlyTargetRules Target) : base(Target)
    {
        PCHUsage = PCHUsageMode.UseExplicitOrSharedPCHs;

        PublicDependencyModuleNames.AddRange(new string[] 
        { 
            "Core", 
            "CoreUObject", 
            "Engine", 
            "HTTP",
            "Json",
            "JsonUtilities"
        });

        PrivateDependencyModuleNames.AddRange(new string[] 
        { 
            "Slate", 
            "SlateCore",
            "MMORPGNetwork"  // Added here to avoid circular dependency
        });

        // Public include paths
        PublicIncludePaths.AddRange(new string[] 
        {
            "MMORPGCore/Public",
            "MMORPGCore/Public/Types",
            "MMORPGCore/Public/Subsystems"
        });

        // Private include paths
        PrivateIncludePaths.AddRange(new string[] 
        {
            "MMORPGCore/Private"
        });
    }
}