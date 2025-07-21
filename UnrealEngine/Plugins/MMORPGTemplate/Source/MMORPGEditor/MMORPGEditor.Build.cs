// Copyright (c) 2024 MMORPG Template Project

using UnrealBuildTool;
using System.IO;

public class MMORPGEditor : ModuleRules
{
    public MMORPGEditor(ReadOnlyTargetRules Target) : base(Target)
    {
        PCHUsage = ModuleRules.PCHUsageMode.UseExplicitOrSharedPCHs;
        
        PublicIncludePaths.AddRange(
            new string[] {
                Path.Combine(ModuleDirectory, "Public")
            }
        );
        
        PrivateIncludePaths.AddRange(
            new string[] {
                Path.Combine(ModuleDirectory, "Private")
            }
        );
        
        PublicDependencyModuleNames.AddRange(
            new string[]
            {
                "Core",
                "CoreUObject",
                "Engine",
                "UnrealEd",
                "MMORPGCore"
            }
        );
        
        PrivateDependencyModuleNames.AddRange(
            new string[]
            {
                "EditorSubsystem",
                "EditorWidgets",
                "Slate",
                "SlateCore",
                "ToolMenus",
                "ToolWidgets",
                "EditorStyle",
                "Projects",
                "InputCore",
                "LevelEditor",
                "Settings",
                "SettingsEditor",
                "PropertyEditor",
                "DetailCustomizations",
                "AssetTools",
                "ClassViewer",
                "KismetWidgets",
                "KismetCompiler",
                "BlueprintGraph",
                "BlueprintEditorModule",
                "Kismet",
                "KismetCompiler",
                "ToolMenus",
                "Blutility",
                "UMGEditor",
                "Json",
                "JsonUtilities",
                "DesktopPlatform",
                "WorkspaceMenuStructure"
            }
        );
        
        // Set C++ standard
        CppStandard = CppStandardVersion.Cpp17;
    }
}