// Copyright (c) 2024 MMORPG Template Project

#pragma once

#include "CoreMinimal.h"
#include "Modules/ModuleManager.h"

class FToolBarBuilder;
class FMenuBuilder;
class SDockTab;
class SWidget;

/**
 * Editor module for MMORPG Template plugin
 * Provides editor tools and utilities for MMORPG development
 */
class FMMORPGEditorModule : public IModuleInterface
{
public:
    /** IModuleInterface implementation */
    virtual void StartupModule() override;
    virtual void ShutdownModule() override;
    
    /** This function will be bound to Command */
    void OpenMMORPGDashboard();
    
    /** Opens the server connection test window */
    void OpenConnectionTestWindow();
    
    /** Opens the protocol buffer viewer */
    void OpenProtocolViewer();
    
private:
    /** Register menu extensions */
    void RegisterMenuExtensions();
    
    /** Add toolbar extension */
    void AddToolbarExtension(FToolBarBuilder& Builder);
    
    /** Add menu extension */
    void AddMenuExtension(FMenuBuilder& Builder);
    
    /** Create the MMORPG dashboard widget */
    TSharedRef<SWidget> CreateMMORPGDashboard();
    
    /** Create the connection test widget */
    TSharedRef<SWidget> CreateConnectionTestWidget();
    
    /** Create the protocol viewer widget */
    TSharedRef<SWidget> CreateProtocolViewerWidget();
    
    /** Dashboard tab spawner */
    TSharedRef<SDockTab> OnSpawnDashboardTab(const class FSpawnTabArgs& SpawnTabArgs);
    
private:
    TSharedPtr<class FUICommandList> PluginCommands;
    
    /** Tab manager */
    static const FName MMORPGDashboardTabName;
    static const FName ConnectionTestTabName;
    static const FName ProtocolViewerTabName;
};