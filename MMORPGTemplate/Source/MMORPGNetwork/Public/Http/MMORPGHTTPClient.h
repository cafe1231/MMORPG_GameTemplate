#pragma once

#include "CoreMinimal.h"
#include "Http.h"
#include "Engine/GameInstance.h"
#include "MMORPGHTTPClient.generated.h"

DECLARE_DYNAMIC_MULTICAST_DELEGATE_ThreeParams(FOnHttpRequestComplete, bool, bWasSuccessful, int32, ResponseCode, const FString&, ResponseContent);

UCLASS(BlueprintType, Blueprintable)
class MMORPGNETWORK_API UMMORPGHTTPClient : public UObject
{
    GENERATED_BODY()

public:
    UMMORPGHTTPClient();

    UPROPERTY(BlueprintAssignable, Category = "MMORPG|Network")
    FOnHttpRequestComplete OnRequestComplete;

    // Simple GET request without headers
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Send GET Request"))
    void SendGetRequest(const FString& URL);

    // GET request with headers
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Send GET Request With Headers"))
    void SendGetRequestWithHeaders(const FString& URL, const TMap<FString, FString>& Headers);

    // Simple POST request without headers
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Send POST Request"))
    void SendPostRequest(const FString& URL, const FString& ContentString);

    // POST request with headers
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Send POST Request With Headers"))
    void SendPostRequestWithHeaders(const FString& URL, const FString& ContentString, const TMap<FString, FString>& Headers);

    // Set base URL for all requests
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
    void SetBaseURL(const FString& InBaseURL);

    // Get the base URL
    UFUNCTION(BlueprintPure, Category = "MMORPG|Network")
    FString GetBaseURL() const { return BaseURL; }

    // Set default timeout
    UFUNCTION(BlueprintCallable, Category = "MMORPG|Network")
    void SetTimeout(float InTimeout);

private:
    void ProcessRequest(TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request);
    void OnResponseReceived(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful);

    FString BaseURL;
    float Timeout;
};