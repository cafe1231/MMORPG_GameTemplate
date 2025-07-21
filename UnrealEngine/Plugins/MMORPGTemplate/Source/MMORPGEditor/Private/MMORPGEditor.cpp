// Copyright (c) 2024 MMORPG Template Project

#include "MMORPGEditor.h"
#include "MMORPGCore.h"
#include "Widgets/Docking/SDockTab.h"
#include "Widgets/Layout/SBox.h"
#include "Widgets/Text/STextBlock.h"
#include "Widgets/Input/SButton.h"
#include "Widgets/Layout/SScrollBox.h"
#include "Framework/MultiBox/MultiBoxBuilder.h"
#include "LevelEditor.h"
#include "ToolMenus.h"
#include "Misc/MessageDialog.h"

static const FName MMORPGEditorTabName("MMORPGEditor");

const FName FMMORPGEditorModule::MMORPGDashboardTabName(TEXT("MMORPGDashboard"));
const FName FMMORPGEditorModule::ConnectionTestTabName(TEXT("MMORPGConnectionTest"));
const FName FMMORPGEditorModule::ProtocolViewerTabName(TEXT("MMORPGProtocolViewer"));

#define LOCTEXT_NAMESPACE "FMMORPGEditorModule"

void FMMORPGEditorModule::StartupModule()
{
    // Register tab spawners
    FGlobalTabmanager::Get()->RegisterNomadTabSpawner(
        MMORPGDashboardTabName,
        FOnSpawnTab::CreateRaw(this, &FMMORPGEditorModule::OnSpawnDashboardTab))
        .SetDisplayName(LOCTEXT("MMORPGDashboardTitle", "MMORPG Dashboard"))
        .SetMenuType(ETabSpawnerMenuType::Hidden);
    
    // Register menu extensions
    RegisterMenuExtensions();
    
    UE_LOG(LogTemp, Log, TEXT("MMORPG Editor Module Started"));
}

void FMMORPGEditorModule::ShutdownModule()
{
    // Unregister tab spawners
    FGlobalTabmanager::Get()->UnregisterNomadTabSpawner(MMORPGDashboardTabName);
    
    // Clean up menu extensions
    UToolMenus::UnRegisterStartupCallback(this);
    UToolMenus::UnregisterOwner(this);
    
    UE_LOG(LogTemp, Log, TEXT("MMORPG Editor Module Shutdown"));
}

void FMMORPGEditorModule::RegisterMenuExtensions()
{
    // Register with the ToolMenus system
    UToolMenus::RegisterStartupCallback(
        FSimpleMulticastDelegate::FDelegate::CreateLambda([this]()
        {
            UToolMenus* ToolMenus = UToolMenus::Get();
            
            // Add to Window menu
            UToolMenu* Menu = ToolMenus->ExtendMenu("LevelEditor.MainMenu.Window");
            FToolMenuSection& Section = Menu->FindOrAddSection("MMORPGTools");
            Section.Label = LOCTEXT("MMORPGToolsLabel", "MMORPG Tools");
            
            Section.AddMenuEntry(
                "OpenMMORPGDashboard",
                LOCTEXT("OpenMMORPGDashboardLabel", "MMORPG Dashboard"),
                LOCTEXT("OpenMMORPGDashboardTooltip", "Open the MMORPG Template dashboard"),
                FSlateIcon(),
                FUIAction(FExecuteAction::CreateRaw(this, &FMMORPGEditorModule::OpenMMORPGDashboard))
            );
            
            Section.AddMenuEntry(
                "OpenConnectionTest",
                LOCTEXT("OpenConnectionTestLabel", "Connection Test"),
                LOCTEXT("OpenConnectionTestTooltip", "Test connection to MMORPG backend"),
                FSlateIcon(),
                FUIAction(FExecuteAction::CreateRaw(this, &FMMORPGEditorModule::OpenConnectionTestWindow))
            );
            
            Section.AddMenuEntry(
                "OpenProtocolViewer",
                LOCTEXT("OpenProtocolViewerLabel", "Protocol Viewer"),
                LOCTEXT("OpenProtocolViewerTooltip", "View Protocol Buffer definitions"),
                FSlateIcon(),
                FUIAction(FExecuteAction::CreateRaw(this, &FMMORPGEditorModule::OpenProtocolViewer))
            );
            
            // Add toolbar button
            UToolMenu* ToolbarMenu = ToolMenus->ExtendMenu("LevelEditor.LevelEditorToolBar");
            FToolMenuSection& ToolbarSection = ToolbarMenu->FindOrAddSection("MMORPG");
            
            ToolbarSection.AddEntry(
                FToolMenuEntry::InitToolBarButton(
                    "MMORPGDashboard",
                    FUIAction(FExecuteAction::CreateRaw(this, &FMMORPGEditorModule::OpenMMORPGDashboard)),
                    LOCTEXT("MMORPGToolbarLabel", "MMORPG"),
                    LOCTEXT("MMORPGToolbarTooltip", "Open MMORPG Dashboard"),
                    FSlateIcon()
                )
            );
        })
    );
}

void FMMORPGEditorModule::OpenMMORPGDashboard()
{
    FGlobalTabmanager::Get()->TryInvokeTab(MMORPGDashboardTabName);
}

void FMMORPGEditorModule::OpenConnectionTestWindow()
{
    // For now, show a message dialog
    FText Title = LOCTEXT("ConnectionTestTitle", "Connection Test");
    FText Message = LOCTEXT("ConnectionTestMessage", 
        "Connection test functionality will be implemented here.\n\n"
        "This will allow you to:\n"
        "- Test connection to backend services\n"
        "- Verify authentication\n"
        "- Check protocol compatibility");
    
    FMessageDialog::Open(EAppMsgType::Ok, Message, &Title);
}

void FMMORPGEditorModule::OpenProtocolViewer()
{
    // For now, show a message dialog
    FText Title = LOCTEXT("ProtocolViewerTitle", "Protocol Viewer");
    FText Message = LOCTEXT("ProtocolViewerMessage", 
        "Protocol viewer functionality will be implemented here.\n\n"
        "This will allow you to:\n"
        "- View all Protocol Buffer definitions\n"
        "- Test message serialization\n"
        "- Generate test data");
    
    FMessageDialog::Open(EAppMsgType::Ok, Message, &Title);
}

TSharedRef<SWidget> FMMORPGEditorModule::CreateMMORPGDashboard()
{
    return SNew(SVerticalBox)
        + SVerticalBox::Slot()
        .AutoHeight()
        .Padding(10)
        [
            SNew(STextBlock)
            .Text(LOCTEXT("DashboardTitle", "MMORPG Template Dashboard"))
            .Font(FCoreStyle::GetDefaultFontStyle("Bold", 16))
        ]
        + SVerticalBox::Slot()
        .AutoHeight()
        .Padding(10, 5)
        [
            SNew(STextBlock)
            .Text(LOCTEXT("DashboardVersion", "Version: 0.1.0"))
        ]
        + SVerticalBox::Slot()
        .Padding(10)
        .FillHeight(1.0f)
        [
            SNew(SScrollBox)
            + SScrollBox::Slot()
            [
                SNew(SVerticalBox)
                // Quick Actions
                + SVerticalBox::Slot()
                .AutoHeight()
                .Padding(0, 10)
                [
                    SNew(STextBlock)
                    .Text(LOCTEXT("QuickActionsLabel", "Quick Actions"))
                    .Font(FCoreStyle::GetDefaultFontStyle("Bold", 12))
                ]
                + SVerticalBox::Slot()
                .AutoHeight()
                .Padding(0, 5)
                [
                    SNew(SButton)
                    .Text(LOCTEXT("TestConnectionButton", "Test Backend Connection"))
                    .OnClicked_Lambda([this]() -> FReply {
                        OpenConnectionTestWindow();
                        return FReply::Handled();
                    })
                ]
                + SVerticalBox::Slot()
                .AutoHeight()
                .Padding(0, 5)
                [
                    SNew(SButton)
                    .Text(LOCTEXT("CompileProtoButton", "Compile Protocol Buffers"))
                    .OnClicked_Lambda([]() -> FReply {
                        FText Title = LOCTEXT("CompileProtoTitle", "Compile Protocol Buffers");
                        FText Message = LOCTEXT("CompileProtoMessage", 
                            "To compile Protocol Buffers:\n\n"
                            "1. Open a terminal\n"
                            "2. Navigate to the backend directory\n"
                            "3. Run: make proto (or scripts/compile_proto.bat on Windows)");
                        FMessageDialog::Open(EAppMsgType::Ok, Message, &Title);
                        return FReply::Handled();
                    })
                ]
                // Documentation
                + SVerticalBox::Slot()
                .AutoHeight()
                .Padding(0, 20, 0, 10)
                [
                    SNew(STextBlock)
                    .Text(LOCTEXT("DocumentationLabel", "Documentation"))
                    .Font(FCoreStyle::GetDefaultFontStyle("Bold", 12))
                ]
                + SVerticalBox::Slot()
                .AutoHeight()
                .Padding(0, 5)
                [
                    SNew(SButton)
                    .Text(LOCTEXT("GettingStartedButton", "Getting Started Guide"))
                    .OnClicked_Lambda([]() -> FReply {
                        FPlatformProcess::LaunchURL(TEXT("https://docs.mmorpg-template.com/getting-started"), nullptr, nullptr);
                        return FReply::Handled();
                    })
                ]
                + SVerticalBox::Slot()
                .AutoHeight()
                .Padding(0, 5)
                [
                    SNew(SButton)
                    .Text(LOCTEXT("APIReferenceButton", "API Reference"))
                    .OnClicked_Lambda([]() -> FReply {
                        FPlatformProcess::LaunchURL(TEXT("https://docs.mmorpg-template.com/api"), nullptr, nullptr);
                        return FReply::Handled();
                    })
                ]
                // Status
                + SVerticalBox::Slot()
                .AutoHeight()
                .Padding(0, 20, 0, 10)
                [
                    SNew(STextBlock)
                    .Text(LOCTEXT("StatusLabel", "Status"))
                    .Font(FCoreStyle::GetDefaultFontStyle("Bold", 12))
                ]
                + SVerticalBox::Slot()
                .AutoHeight()
                .Padding(0, 5)
                [
                    SNew(STextBlock)
                    .Text(LOCTEXT("StatusInfo", "Backend: Not Connected\nProtocol Version: 1\nPlugin Status: Active"))
                ]
            ]
        ];
}

TSharedRef<SWidget> FMMORPGEditorModule::CreateConnectionTestWidget()
{
    return SNew(SBox)
        .Padding(10)
        [
            SNew(STextBlock)
            .Text(LOCTEXT("ConnectionTestPlaceholder", "Connection test widget - To be implemented"))
        ];
}

TSharedRef<SWidget> FMMORPGEditorModule::CreateProtocolViewerWidget()
{
    return SNew(SBox)
        .Padding(10)
        [
            SNew(STextBlock)
            .Text(LOCTEXT("ProtocolViewerPlaceholder", "Protocol viewer widget - To be implemented"))
        ];
}

TSharedRef<SDockTab> FMMORPGEditorModule::OnSpawnDashboardTab(const FSpawnTabArgs& SpawnTabArgs)
{
    return SNew(SDockTab)
        .TabRole(ETabRole::NomadTab)
        [
            CreateMMORPGDashboard()
        ];
}

#undef LOCTEXT_NAMESPACE

IMPLEMENT_MODULE(FMMORPGEditorModule, MMORPGEditor)