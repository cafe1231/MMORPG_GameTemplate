package character

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/auth"
	"github.com/mmorpg-template/backend/internal/domain/character"
	portsCharacter "github.com/mmorpg-template/backend/internal/ports/character"
	"github.com/mmorpg-template/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockCharacterService is a mock implementation of the character service
type MockCharacterService struct {
	mock.Mock
}

func (m *MockCharacterService) CreateCharacter(ctx context.Context, req *portsCharacter.CreateCharacterRequest) (*character.Character, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Character), args.Error(1)
}

func (m *MockCharacterService) GetCharacter(ctx context.Context, characterID string) (*character.Character, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Character), args.Error(1)
}

func (m *MockCharacterService) GetCharacterByName(ctx context.Context, name string) (*character.Character, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Character), args.Error(1)
}

func (m *MockCharacterService) ListCharactersByUser(ctx context.Context, userID string) ([]*character.Character, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*character.Character), args.Error(1)
}

func (m *MockCharacterService) DeleteCharacter(ctx context.Context, characterID string, userID string) error {
	args := m.Called(ctx, characterID, userID)
	return args.Error(0)
}

func (m *MockCharacterService) RestoreCharacter(ctx context.Context, characterID string, userID string) error {
	args := m.Called(ctx, characterID, userID)
	return args.Error(0)
}

func (m *MockCharacterService) GetAppearance(ctx context.Context, characterID string) (*character.Appearance, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Appearance), args.Error(1)
}

func (m *MockCharacterService) UpdateAppearance(ctx context.Context, characterID string, req *portsCharacter.UpdateAppearanceRequest) (*character.Appearance, error) {
	args := m.Called(ctx, characterID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Appearance), args.Error(1)
}

func (m *MockCharacterService) GetStats(ctx context.Context, characterID string) (*character.Stats, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Stats), args.Error(1)
}

func (m *MockCharacterService) AllocateStatPoint(ctx context.Context, characterID string, stat string) (*character.Stats, error) {
	args := m.Called(ctx, characterID, stat)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Stats), args.Error(1)
}

func (m *MockCharacterService) GetPosition(ctx context.Context, characterID string) (*character.Position, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Position), args.Error(1)
}

func (m *MockCharacterService) UpdatePosition(ctx context.Context, characterID string, req *portsCharacter.UpdatePositionRequest) (*character.Position, error) {
	args := m.Called(ctx, characterID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Position), args.Error(1)
}

func (m *MockCharacterService) TeleportToSafePosition(ctx context.Context, characterID string) (*character.Position, error) {
	args := m.Called(ctx, characterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*character.Position), args.Error(1)
}

func (m *MockCharacterService) ValidateCharacterOwnership(ctx context.Context, characterID string, userID string) error {
	args := m.Called(ctx, characterID, userID)
	return args.Error(0)
}

func (m *MockCharacterService) CanCreateCharacter(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockCharacterService) SelectCharacter(ctx context.Context, characterID string, userID string, sessionID string) error {
	args := m.Called(ctx, characterID, userID, sessionID)
	return args.Error(0)
}

func setupTestRouter(t *testing.T) (*gin.Engine, *MockCharacterService, string) {
	gin.SetMode(gin.TestMode)
	
	// Create JWT config and middleware
	jwtConfig := &JWTConfig{
		AccessSecret: "test-secret",
		Issuer:       "mmorpg-auth",
	}
	jwtMiddleware := NewJWTMiddleware(jwtConfig, logger.NewNoop())
	
	// Create mock service
	mockService := new(MockCharacterService)
	
	// Create handler
	handler := NewHTTPHandler(mockService, jwtMiddleware, logger.NewNoop())
	
	// Setup router
	router := gin.New()
	handler.RegisterRoutes(router)
	
	// Create valid JWT token for testing
	claims := &auth.Claims{
		UserID:    "test-user-123",
		SessionID: "session-456",
		Email:     "test@example.com",
		Username:  "testuser",
		Roles:     []string{"player"},
		DeviceID:  "device-789",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mmorpg-auth",
			Subject:   "test-user-123",
		},
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)
	
	return router, mockService, tokenString
}

func TestCharacterAPI_CreateCharacter(t *testing.T) {
	router, mockService, token := setupTestRouter(t)
	
	// Setup mock expectations
	expectedChar := &character.Character{
		ID:         uuid.New(),
		UserID:     uuid.MustParse("00000000-0000-0000-0000-000000000123"),
		Name:       "TestWarrior",
		SlotNumber: 1,
		Level:      1,
		Experience: 0,
		ClassType:  character.ClassWarrior,
		Race:       character.RaceHuman,
		Gender:     character.GenderMale,
		CreatedAt:  time.Now(),
	}
	
	mockService.On("CreateCharacter", mock.Anything, mock.MatchedBy(func(req *portsCharacter.CreateCharacterRequest) bool {
		return req.UserID == "test-user-123" &&
			req.Name == "TestWarrior" &&
			req.SlotNumber == 1 &&
			req.ClassType == character.ClassWarrior &&
			req.Race == character.RaceHuman &&
			req.Gender == character.GenderMale
	})).Return(expectedChar, nil)
	
	// Create request
	reqBody := CreateCharacterRequest{
		Name:       "TestWarrior",
		SlotNumber: 1,
		ClassType:  "warrior",
		Race:       "human",
		Gender:     "male",
	}
	
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPost, "/api/v1/characters", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response CharacterResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, expectedChar.ID.String(), response.ID)
	assert.Equal(t, expectedChar.Name, response.Name)
	assert.Equal(t, expectedChar.Level, response.Level)
	
	mockService.AssertExpectations(t)
}

func TestCharacterAPI_ListCharacters(t *testing.T) {
	router, mockService, token := setupTestRouter(t)
	
	// Setup mock expectations
	expectedChars := []*character.Character{
		{
			ID:         uuid.New(),
			UserID:     uuid.MustParse("00000000-0000-0000-0000-000000000123"),
			Name:       "Warrior1",
			SlotNumber: 1,
			Level:      10,
			ClassType:  character.ClassWarrior,
			Race:       character.RaceHuman,
			Gender:     character.GenderMale,
		},
		{
			ID:         uuid.New(),
			UserID:     uuid.MustParse("00000000-0000-0000-0000-000000000123"),
			Name:       "Mage1",
			SlotNumber: 2,
			Level:      5,
			ClassType:  character.ClassMage,
			Race:       character.RaceElf,
			Gender:     character.GenderFemale,
		},
	}
	
	mockService.On("ListCharactersByUser", mock.Anything, "test-user-123").Return(expectedChars, nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/characters", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string][]CharacterResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	characters := response["characters"]
	assert.Len(t, characters, 2)
	assert.Equal(t, "Warrior1", characters[0].Name)
	assert.Equal(t, "Mage1", characters[1].Name)
	
	mockService.AssertExpectations(t)
}

func TestCharacterAPI_Unauthorized(t *testing.T) {
	router, _, _ := setupTestRouter(t)
	
	// Create request without auth header
	req := httptest.NewRequest(http.MethodGet, "/api/v1/characters", nil)
	
	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	errorDetail := response["error"].(map[string]interface{})
	assert.Equal(t, "UNAUTHORIZED", errorDetail["code"])
}

func TestCharacterAPI_CheckNameAvailability(t *testing.T) {
	router, mockService, _ := setupTestRouter(t)
	
	tests := []struct {
		name          string
		queryName     string
		mockSetup     func()
		expectedCode  int
		expectedAvail bool
	}{
		{
			name:      "name available",
			queryName: "NewHero",
			mockSetup: func() {
				mockService.On("GetCharacterByName", mock.Anything, "NewHero").
					Return(nil, character.ErrCharacterNotFound)
			},
			expectedCode:  http.StatusOK,
			expectedAvail: true,
		},
		{
			name:      "name taken",
			queryName: "ExistingHero",
			mockSetup: func() {
				mockService.On("GetCharacterByName", mock.Anything, "ExistingHero").
					Return(&character.Character{Name: "ExistingHero"}, nil)
			},
			expectedCode:  http.StatusOK,
			expectedAvail: false,
		},
		{
			name:         "missing name parameter",
			queryName:    "",
			mockSetup:    func() {},
			expectedCode: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// Setup mock
			tt.mockSetup()
			
			// Create request
			url := "/api/v1/characters/check-name"
			if tt.queryName != "" {
				url += "?name=" + tt.queryName
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			
			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Assert response
			assert.Equal(t, tt.expectedCode, w.Code)
			
			if tt.expectedCode == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				
				assert.Equal(t, tt.expectedAvail, response["available"])
				assert.Equal(t, tt.queryName, response["name"])
				
				if !tt.expectedAvail {
					suggestions := response["suggestions"].([]interface{})
					assert.NotEmpty(t, suggestions)
				}
			}
			
			mockService.AssertExpectations(t)
		})
	}
}