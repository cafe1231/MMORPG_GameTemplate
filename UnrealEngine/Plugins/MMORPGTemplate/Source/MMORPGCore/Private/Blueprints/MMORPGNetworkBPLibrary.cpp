// Copyright (c) 2024 MMORPG Template Project

#include "Blueprints/MMORPGNetworkBPLibrary.h"
#include "MMORPGCore.h"
#include "Network/MMORPGNetworkManager.h"
#include "Dom/JsonObject.h"
#include "Serialization/JsonSerializer.h"
#include "Serialization/JsonReader.h"

void UMMORPGNetworkBPLibrary::ConnectToServer(const FString& Host, int32 Port)
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            NetworkManager->Connect(Host, Port);
        }
        else
        {
            MMORPG_LOG(Error, TEXT("NetworkManager is not initialized"));
        }
    }
}

void UMMORPGNetworkBPLibrary::DisconnectFromServer()
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            NetworkManager->Disconnect();
        }
    }
}

bool UMMORPGNetworkBPLibrary::IsConnected()
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            return NetworkManager->IsConnected();
        }
    }
    return false;
}

void UMMORPGNetworkBPLibrary::TestConnection(const FOnHttpRequestComplete& OnComplete)
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            NetworkManager->TestConnection([OnComplete](bool bSuccess, const FString& Response)
            {
                OnComplete.ExecuteIfBound(bSuccess, Response);
            });
        }
        else
        {
            OnComplete.ExecuteIfBound(false, TEXT("NetworkManager not initialized"));
        }
    }
    else
    {
        OnComplete.ExecuteIfBound(false, TEXT("MMORPG module not available"));
    }
}

void UMMORPGNetworkBPLibrary::GetHealthStatus(const FOnHttpRequestComplete& OnComplete)
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            NetworkManager->GetHealthStatus([OnComplete](bool bSuccess, const FString& Response)
            {
                OnComplete.ExecuteIfBound(bSuccess, Response);
            });
        }
        else
        {
            OnComplete.ExecuteIfBound(false, TEXT("NetworkManager not initialized"));
        }
    }
    else
    {
        OnComplete.ExecuteIfBound(false, TEXT("MMORPG module not available"));
    }
}

void UMMORPGNetworkBPLibrary::SendGetRequest(const FString& Endpoint, const FOnHttpRequestComplete& OnComplete)
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            NetworkManager->SendGetRequest(Endpoint, [OnComplete](bool bSuccess, const FString& Response)
            {
                OnComplete.ExecuteIfBound(bSuccess, Response);
            });
        }
        else
        {
            OnComplete.ExecuteIfBound(false, TEXT("NetworkManager not initialized"));
        }
    }
}

void UMMORPGNetworkBPLibrary::SendPostRequest(const FString& Endpoint, const FString& JsonData, const FOnHttpRequestComplete& OnComplete)
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
            TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonData);
            
            if (FJsonSerializer::Deserialize(Reader, JsonObject))
            {
                NetworkManager->SendPostRequest(Endpoint, JsonObject, [OnComplete](bool bSuccess, const FString& Response)
                {
                    OnComplete.ExecuteIfBound(bSuccess, Response);
                });
            }
            else
            {
                OnComplete.ExecuteIfBound(false, TEXT("Invalid JSON data"));
            }
        }
        else
        {
            OnComplete.ExecuteIfBound(false, TEXT("NetworkManager not initialized"));
        }
    }
}

void UMMORPGNetworkBPLibrary::SetAuthToken(const FString& Token)
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            NetworkManager->SetAuthToken(Token);
        }
    }
}

FString UMMORPGNetworkBPLibrary::GetServerURL()
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            return NetworkManager->GetServerURL();
        }
    }
    return TEXT("Not connected");
}

void UMMORPGNetworkBPLibrary::TestAPI(const FOnHttpRequestComplete& OnComplete)
{
    SendGetRequest(TEXT("/api/v1/test"), OnComplete);
}

void UMMORPGNetworkBPLibrary::EchoTest(const FString& Message, const FOnHttpRequestComplete& OnComplete)
{
    if (FMMORPGCoreModule::IsAvailable())
    {
        auto NetworkManager = FMMORPGCoreModule::Get().GetNetworkManager();
        if (NetworkManager.IsValid())
        {
            TSharedPtr<FJsonObject> JsonObject = MakeShareable(new FJsonObject);
            JsonObject->SetStringField(TEXT("message"), Message);
            JsonObject->SetNumberField(TEXT("timestamp"), FDateTime::Now().ToUnixTimestamp());
            JsonObject->SetStringField(TEXT("client"), TEXT("Unreal Engine 5.6"));
            
            NetworkManager->SendPostRequest(TEXT("/api/v1/echo"), JsonObject, [OnComplete](bool bSuccess, const FString& Response)
            {
                OnComplete.ExecuteIfBound(bSuccess, Response);
            });
        }
        else
        {
            OnComplete.ExecuteIfBound(false, TEXT("NetworkManager not initialized"));
        }
    }
}