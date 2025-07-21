// Copyright (c) 2024 MMORPG Template Project

using UnrealBuildTool;
using System.IO;

public class MMORPGCore : ModuleRules
{
    public MMORPGCore(ReadOnlyTargetRules Target) : base(Target)
    {
        PCHUsage = ModuleRules.PCHUsageMode.UseExplicitOrSharedPCHs;
        bEnforceIWYU = true;
        
        // Optimize for performance
        OptimizeCode = CodeOptimization.InShippingBuildsOnly;
        
        PublicIncludePaths.AddRange(
            new string[] {
                Path.Combine(ModuleDirectory, "Public"),
                Path.Combine(ModuleDirectory, "Public/Network"),
                Path.Combine(ModuleDirectory, "Public/Authentication"),
                Path.Combine(ModuleDirectory, "Public/Data"),
                Path.Combine(ModuleDirectory, "Public/Gameplay"),
                Path.Combine(ModuleDirectory, "Public/UI"),
                Path.Combine(ModuleDirectory, "Public/Utils"),
                Path.Combine(ModuleDirectory, "Public/Proto")
            }
        );
        
        PrivateIncludePaths.AddRange(
            new string[] {
                Path.Combine(ModuleDirectory, "Private"),
                Path.Combine(ModuleDirectory, "Private/Network"),
                Path.Combine(ModuleDirectory, "Private/Authentication"),
                Path.Combine(ModuleDirectory, "Private/Data"),
                Path.Combine(ModuleDirectory, "Private/Gameplay"),
                Path.Combine(ModuleDirectory, "Private/UI"),
                Path.Combine(ModuleDirectory, "Private/Utils")
            }
        );
        
        PublicDependencyModuleNames.AddRange(
            new string[]
            {
                "Core",
                "CoreUObject",
                "Engine",
                "InputCore",
                "Http",
                "Json",
                "JsonUtilities",
                "WebSockets",
                "Networking",
                "Sockets",
                "OnlineSubsystem",
                "OnlineSubsystemUtils",
                "UMG",
                "Slate",
                "SlateCore",
                "ProceduralMeshComponent",
                "NavigationSystem",
                "AIModule"
            }
        );
        
        PrivateDependencyModuleNames.AddRange(
            new string[]
            {
                "RenderCore",
                "RHI",
                "Projects",
                "DeveloperSettings",
                "GameplayTags",
                "GameplayTasks",
                "GameplayAbilities"
            }
        );
        
        // Add Protocol Buffers support
        if (Target.Platform == UnrealTargetPlatform.Win64)
        {
            // Windows-specific protobuf configuration
            string ProtobufPath = Path.Combine(ModuleDirectory, "ThirdParty", "protobuf", "Win64");
            
            PublicIncludePaths.Add(Path.Combine(ProtobufPath, "include"));
            PublicAdditionalLibraries.Add(Path.Combine(ProtobufPath, "lib", "libprotobuf.lib"));
            
            RuntimeDependencies.Add(Path.Combine(ProtobufPath, "bin", "libprotobuf.dll"));
        }
        else if (Target.Platform == UnrealTargetPlatform.Mac)
        {
            // Mac-specific protobuf configuration
            string ProtobufPath = Path.Combine(ModuleDirectory, "ThirdParty", "protobuf", "Mac");
            
            PublicIncludePaths.Add(Path.Combine(ProtobufPath, "include"));
            PublicAdditionalLibraries.Add(Path.Combine(ProtobufPath, "lib", "libprotobuf.a"));
        }
        else if (Target.Platform == UnrealTargetPlatform.Linux)
        {
            // Linux-specific protobuf configuration
            string ProtobufPath = Path.Combine(ModuleDirectory, "ThirdParty", "protobuf", "Linux");
            
            PublicIncludePaths.Add(Path.Combine(ProtobufPath, "include"));
            PublicAdditionalLibraries.Add(Path.Combine(ProtobufPath, "lib", "libprotobuf.a"));
        }
        
        // Development-only modules
        if (Target.Configuration != UnrealTargetConfiguration.Shipping)
        {
            PrivateDependencyModuleNames.AddRange(
                new string[]
                {
                    "UnrealEd",
                    "EditorSubsystem"
                }
            );
        }
        
        // Preprocessor definitions
        PublicDefinitions.AddRange(
            new string[]
            {
                "MMORPG_TEMPLATE_VERSION=1",
                "WITH_MMORPG_TEMPLATE=1"
            }
        );
        
        // Enable exceptions for protobuf
        bEnableExceptions = true;
        
        // Enable RTTI for protobuf
        bUseRTTI = true;
        
        // Set C++ standard
        CppStandard = CppStandardVersion.Cpp17;
    }
}