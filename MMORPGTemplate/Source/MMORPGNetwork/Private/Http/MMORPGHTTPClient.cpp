#include "Http/MMORPGHTTPClient.h"
#include "Dom/JsonObject.h"
#include "Serialization/JsonSerializer.h"
#include "Serialization/JsonWriter.h"
#include "Serialization/JsonReader.h"
#include "Engine/World.h"
#include "MMORPGNetwork.h"

UMMORPGHTTPRequest* UMMORPGHTTPRequest::MakeHTTPRequest(
	UObject* WorldContextObject,
	const FString& URL,
	EMMORPGHTTPVerb Verb,
	const TMap<FString, FString>& Headers,
	const FString& Body)
{
	UMMORPGHTTPRequest* BlueprintNode = NewObject<UMMORPGHTTPRequest>();
	BlueprintNode->WorldContextObject = WorldContextObject;
	BlueprintNode->RequestURL = URL;
	BlueprintNode->RequestVerb = Verb;
	BlueprintNode->RequestHeaders = Headers;
	BlueprintNode->RequestBody = Body;
	return BlueprintNode;
}

void UMMORPGHTTPRequest::Activate()
{
	// Create HTTP request
	TSharedRef<IHttpRequest, ESPMode::ThreadSafe> HttpRequest = FHttpModule::Get().CreateRequest();

	// Set URL and verb
	HttpRequest->SetURL(RequestURL);
	HttpRequest->SetVerb(VerbToString(RequestVerb));

	// Set headers
	for (const auto& Header : RequestHeaders)
	{
		HttpRequest->SetHeader(Header.Key, Header.Value);
	}

	// Set content type if not specified
	if (!RequestHeaders.Contains(TEXT("Content-Type")) && !RequestBody.IsEmpty())
	{
		HttpRequest->SetHeader(TEXT("Content-Type"), TEXT("application/json"));
	}

	// Set body for POST/PUT
	if (RequestVerb == EMMORPGHTTPVerb::POST || RequestVerb == EMMORPGHTTPVerb::PUT)
	{
		HttpRequest->SetContentAsString(RequestBody);
	}

	// Bind response callback
	HttpRequest->OnProcessRequestComplete().BindUObject(this, &UMMORPGHTTPRequest::OnResponseReceived);

	// Send request
	if (!HttpRequest->ProcessRequest())
	{
		OnError.Broadcast(TEXT("Failed to send HTTP request"));
		SetReadyToDestroy();
	}

	UE_LOG(LogMMORPGNetwork, Log, TEXT("HTTP Request sent: %s %s"), *VerbToString(RequestVerb), *RequestURL);
}

void UMMORPGHTTPRequest::OnResponseReceived(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful)
{
	if (bWasSuccessful && Response.IsValid())
	{
		int32 ResponseCode = Response->GetResponseCode();
		FString ResponseContent = Response->GetContentAsString();

		UE_LOG(LogMMORPGNetwork, Log, TEXT("HTTP Response received: Code=%d"), ResponseCode);

		if (ResponseCode >= 200 && ResponseCode < 300)
		{
			OnSuccess.Broadcast(ResponseContent, ResponseCode);
		}
		else
		{
			FString ErrorMessage = FString::Printf(TEXT("HTTP Error: %d - %s"), ResponseCode, *ResponseContent);
			OnError.Broadcast(ErrorMessage);
		}
	}
	else
	{
		FString ErrorMessage = TEXT("HTTP Request failed: No response");
		if (Request.IsValid())
		{
			ErrorMessage = FString::Printf(TEXT("HTTP Request failed: %s"), *Request->GetURL());
		}
		
		UE_LOG(LogMMORPGNetwork, Error, TEXT("%s"), *ErrorMessage);
		OnError.Broadcast(ErrorMessage);
	}

	SetReadyToDestroy();
}

FString UMMORPGHTTPRequest::VerbToString(EMMORPGHTTPVerb Verb) const
{
	switch (Verb)
	{
		case EMMORPGHTTPVerb::GET: return TEXT("GET");
		case EMMORPGHTTPVerb::POST: return TEXT("POST");
		case EMMORPGHTTPVerb::PUT: return TEXT("PUT");
		case EMMORPGHTTPVerb::DELETE: return TEXT("DELETE");
		default: return TEXT("GET");
	}
}

// UMMORPGHTTPClient implementation

FString UMMORPGHTTPClient::BuildURL(const FString& BaseURL, const TMap<FString, FString>& QueryParams)
{
	if (QueryParams.Num() == 0)
	{
		return BaseURL;
	}

	FString Result = BaseURL;
	Result += TEXT("?");

	bool bFirst = true;
	for (const auto& Param : QueryParams)
	{
		if (!bFirst)
		{
			Result += TEXT("&");
		}
		Result += FString::Printf(TEXT("%s=%s"), 
			*FPlatformHttp::UrlEncode(Param.Key),
			*FPlatformHttp::UrlEncode(Param.Value));
		bFirst = false;
	}

	return Result;
}

bool UMMORPGHTTPClient::ParseJsonResponse(const FString& JsonString, UStruct* StructDefinition, void* OutStruct)
{
	if (!StructDefinition || !OutStruct)
	{
		return false;
	}

	TSharedPtr<FJsonObject> JsonObject;
	TSharedRef<TJsonReader<>> Reader = TJsonReaderFactory<>::Create(JsonString);

	if (!FJsonSerializer::Deserialize(Reader, JsonObject) || !JsonObject.IsValid())
	{
		return false;
	}

	// Use FJsonObjectConverter when available
	// For now, return false as we need proper JSON conversion
	// This will be implemented when we have specific message types
	return false;
}

FString UMMORPGHTTPClient::EncodeStructToJson(UStruct* StructDefinition, const void* InStruct)
{
	if (!StructDefinition || !InStruct)
	{
		return TEXT("{}");
	}

	// Use FJsonObjectConverter when available
	// For now, return empty object
	// This will be implemented when we have specific message types
	return TEXT("{}");
}

FString UMMORPGHTTPClient::CreateAuthHeader(const FString& Token)
{
	return FString::Printf(TEXT("Bearer %s"), *Token);
}