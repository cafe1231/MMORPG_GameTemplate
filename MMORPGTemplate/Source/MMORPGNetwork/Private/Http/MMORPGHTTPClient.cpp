#include "Http/MMORPGHTTPClient.h"
#include "HttpModule.h"
#include "Interfaces/IHttpResponse.h"

UMMORPGHTTPClient::UMMORPGHTTPClient()
{
    BaseURL = TEXT("http://localhost:3000");
    Timeout = 10.0f;
}

void UMMORPGHTTPClient::SendGetRequest(const FString& URL)
{
    TMap<FString, FString> EmptyHeaders;
    SendGetRequestWithHeaders(URL, EmptyHeaders);
}

void UMMORPGHTTPClient::SendGetRequestWithHeaders(const FString& URL, const TMap<FString, FString>& Headers)
{
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request = FHttpModule::Get().CreateRequest();
    
    FString FullURL = URL.StartsWith(TEXT("http")) ? URL : BaseURL + URL;
    Request->SetURL(FullURL);
    Request->SetVerb(TEXT("GET"));
    Request->SetTimeout(Timeout);
    
    // Set headers
    for (const auto& Header : Headers)
    {
        Request->SetHeader(Header.Key, Header.Value);
    }
    
    ProcessRequest(Request);
}

void UMMORPGHTTPClient::SendPostRequest(const FString& URL, const FString& ContentString)
{
    TMap<FString, FString> EmptyHeaders;
    SendPostRequestWithHeaders(URL, ContentString, EmptyHeaders);
}

void UMMORPGHTTPClient::SendPostRequestWithHeaders(const FString& URL, const FString& ContentString, const TMap<FString, FString>& Headers)
{
    TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request = FHttpModule::Get().CreateRequest();
    
    FString FullURL = URL.StartsWith(TEXT("http")) ? URL : BaseURL + URL;
    Request->SetURL(FullURL);
    Request->SetVerb(TEXT("POST"));
    Request->SetTimeout(Timeout);
    Request->SetContentAsString(ContentString);
    
    // Set default content type if not provided
    bool bHasContentType = false;
    for (const auto& Header : Headers)
    {
        Request->SetHeader(Header.Key, Header.Value);
        if (Header.Key.Equals(TEXT("Content-Type"), ESearchCase::IgnoreCase))
        {
            bHasContentType = true;
        }
    }
    
    if (!bHasContentType)
    {
        Request->SetHeader(TEXT("Content-Type"), TEXT("application/json"));
    }
    
    ProcessRequest(Request);
}

void UMMORPGHTTPClient::SetBaseURL(const FString& InBaseURL)
{
    BaseURL = InBaseURL;
}

void UMMORPGHTTPClient::SetTimeout(float InTimeout)
{
    Timeout = FMath::Max(1.0f, InTimeout);
}

void UMMORPGHTTPClient::ProcessRequest(TSharedRef<IHttpRequest, ESPMode::ThreadSafe> Request)
{
    Request->OnProcessRequestComplete().BindUObject(this, &UMMORPGHTTPClient::OnResponseReceived);
    Request->ProcessRequest();
}

void UMMORPGHTTPClient::OnResponseReceived(FHttpRequestPtr Request, FHttpResponsePtr Response, bool bWasSuccessful)
{
    int32 ResponseCode = 0;
    FString ResponseContent = TEXT("");
    
    if (bWasSuccessful && Response.IsValid())
    {
        ResponseCode = Response->GetResponseCode();
        ResponseContent = Response->GetContentAsString();
    }
    
    OnRequestComplete.Broadcast(bWasSuccessful, ResponseCode, ResponseContent);
}