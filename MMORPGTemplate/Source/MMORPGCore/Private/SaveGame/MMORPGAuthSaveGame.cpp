#include "SaveGame/UMMORPGAuthSaveGame.h"

UMMORPGAuthSaveGame::UMMORPGAuthSaveGame()
{
    // Initialize default values
    RefreshToken = TEXT("");
    UserInfo = FUserInfo();
    bRememberMe = false;
    LastLoginTime = FDateTime::MinValue();
    SaveGameVersion = 1;
}

void UMMORPGAuthSaveGame::ClearData()
{
    RefreshToken = TEXT("");
    UserInfo = FUserInfo();
    bRememberMe = false;
    LastLoginTime = FDateTime::MinValue();
}

bool UMMORPGAuthSaveGame::HasValidAuthData() const
{
    // Check if we have a refresh token and remember me is enabled
    return bRememberMe && !RefreshToken.IsEmpty() && !UserInfo.Id.IsEmpty();
}