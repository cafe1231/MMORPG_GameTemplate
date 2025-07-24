using UnrealBuildTool;
using System.IO;

public class MMORPGProto : ModuleRules
{
	public MMORPGProto(ReadOnlyTargetRules Target) : base(Target)
	{
		PCHUsage = PCHUsageMode.UseExplicitOrSharedPCHs;
		
		PublicDependencyModuleNames.AddRange(new string[] {
			"Core",
			"CoreUObject",
			"Engine",
			"MMORPGCore"
		});
		
		// Protocol Buffers will be added later
		// For now, we'll use JSON serialization
	}
}