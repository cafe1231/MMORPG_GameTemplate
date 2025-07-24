using UnrealBuildTool;

public class MMORPGUI : ModuleRules
{
	public MMORPGUI(ReadOnlyTargetRules Target) : base(Target)
	{
		PCHUsage = PCHUsageMode.UseExplicitOrSharedPCHs;
		
		PublicDependencyModuleNames.AddRange(new string[] {
			"Core",
			"CoreUObject",
			"Engine",
			"MMORPGCore",
			"UMG",
			"Slate",
			"SlateCore",
			"InputCore"
		});
	}
}