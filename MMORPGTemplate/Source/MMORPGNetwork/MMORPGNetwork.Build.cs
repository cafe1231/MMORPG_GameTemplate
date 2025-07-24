using UnrealBuildTool;

public class MMORPGNetwork : ModuleRules
{
	public MMORPGNetwork(ReadOnlyTargetRules Target) : base(Target)
	{
		PCHUsage = PCHUsageMode.UseExplicitOrSharedPCHs;
		
		PublicDependencyModuleNames.AddRange(new string[] {
			"Core",
			"CoreUObject",
			"Engine",
			"MMORPGCore",
			"MMORPGProto",
			"HTTP",
			"WebSockets",
			"Json",
			"JsonUtilities"
		});
	}
}