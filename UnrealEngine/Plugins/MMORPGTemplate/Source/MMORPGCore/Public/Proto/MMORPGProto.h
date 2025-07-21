// Auto-generated header file
// Include all Protocol Buffer headers

#pragma once

#include "base.pb.h"
#include "auth.pb.h"
#include "character.pb.h"
#include "world.pb.h"
#include "game.pb.h"
#include "chat.pb.h"

namespace MMORPG
{
    // Type aliases for easier use in Unreal Engine
    using namespace mmorpg;
    
    // Commonly used types
    using GameMessage = mmorpg::GameMessage;
    using ErrorCode = mmorpg::ErrorCode;
    using Vector3 = mmorpg::Vector3;
    using Transform = mmorpg::Transform;
    
    // Authentication types
    using LoginRequest = mmorpg::LoginRequest;
    using LoginResponse = mmorpg::LoginResponse;
    
    // Character types
    using CharacterInfo = mmorpg::CharacterInfo;
    using CharacterData = mmorpg::CharacterData;
    
    // World types
    using PlayerPositionUpdate = mmorpg::PlayerPositionUpdate;
    using AreaUpdate = mmorpg::AreaUpdate;
    
    // Game types
    using InventoryUpdate = mmorpg::InventoryUpdate;
    using QuestUpdate = mmorpg::QuestUpdate;
    
    // Chat types
    using ChatMessage = mmorpg::ChatMessage;
}