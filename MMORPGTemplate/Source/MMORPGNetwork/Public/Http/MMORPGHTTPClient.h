#pragma once

#include "CoreMinimal.h"
#include "Http.h"
#include "Kismet/BlueprintAsyncActionBase.h"
#include "MMORPGHTTPClient.generated.h"

DECLARE_DYNAMIC_MULTICAST_DELEGATE_TwoParams(FOnHTTPResponse, const FString&, ResponseContent, int32, ResponseCode);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_OneParam(FOnHTTPError, const FString&, ErrorMessage);

/**
 * HTTP request types
 */
UENUM(BlueprintType)
enum class EMMORPGHTTPVerb : uint8
{
	GET		UMETA(DisplayName = "GET"),
	POST	UMETA(DisplayName = "POST"),
	PUT		UMETA(DisplayName = "PUT"),
	DELETE	UMETA(DisplayName = "DELETE")
};

/**
 * Async HTTP request for Blueprint use
 */
UCLASS()
class MMORPGNETWORK_API UMMORPGHTTPRequest : public UBlueprintAsyncActionBase
{
	GENERATED_BODY()

public:
	// Success delegate
	UPROPERTY(BlueprintAssignable)
	FOnHTTPResponse OnSuccess;

	// Error delegate
	UPROPERTY(BlueprintAssignable)
	FOnHTTPError OnError;

	/**
	 * Create an async HTTP request
	 * @param URL The URL to request
	 * @param Verb The HTTP verb to use
	 * @param Headers Additional headers
	 * @param Body Request body (for POST/PUT)
	 * @return The async action
	 */
	UFUNCTION(BlueprintCallable, meta = (BlueprintInternalUseOnly = "true", Category = "MMORPG|Network", WorldContext = "WorldContextObject"))
	static UMMORPGHTTPRequest* MakeHTTPRequest(
		UObject* WorldContextObject,
		const FString& URL,
		EMMORPGHTTPVerb Verb,
		const TMap<FString, FString>& Headers,
		const FString& Body
	);

	// UBlueprintAsyncActionBase interface
	virtual void Activate() override;

private:
	// Request parameters
	FString RequestURL;
	EMMORPGHTTPVerb RequestVerb;
	TMap<FString, FString> RequestHeaders;
	FString RequestBody;

	// World context
	TWeakObjectPtr<UObject> WorldContextObject;

	// Process the HTTP response
	void OnResponseReceived(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful);

	// Convert enum to string
	FString VerbToString(EMMORPGHTTPVerb Verb) const;
};

/**
 * HTTP client utility functions
 */
UCLASS()
class MMORPGNETWORK_API UMMORPGHTTPClient : public UBlueprintFunctionLibrary
{
	GENERATED_BODY()

public:
	/**
	 * Build URL with query parameters
	 * @param BaseURL The base URL
	 * @param QueryParams Query parameters to append
	 * @return Complete URL with query string
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Network", meta = (DisplayName = "Build URL"))
	static FString BuildURL(const FString& BaseURL, const TMap<FString, FString>& QueryParams);

	/**
	 * Parse JSON response to a struct
	 * @param JsonString The JSON string to parse
	 * @param OutStruct The struct to populate
	 * @return True if parsing was successful
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Parse JSON Response", CallInEditor = "true"))
	static bool ParseJsonResponse(const FString& JsonString, UStruct* StructDefinition, void* OutStruct);

	/**
	 * Encode struct to JSON string
	 * @param InStruct The struct to encode
	 * @return JSON string representation
	 */
	UFUNCTION(BlueprintCallable, Category = "MMORPG|Network", meta = (DisplayName = "Encode to JSON", CallInEditor = "true"))
	static FString EncodeStructToJson(UStruct* StructDefinition, const void* InStruct);

	/**
	 * Create authorization header
	 * @param Token The authorization token
	 * @return Header value
	 */
	UFUNCTION(BlueprintPure, Category = "MMORPG|Network", meta = (DisplayName = "Create Auth Header"))
	static FString CreateAuthHeader(const FString& Token);
};