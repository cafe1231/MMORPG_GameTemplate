package nats

import (
	"context"
	"encoding/json"
	"testing"
	"time"
	
	"github.com/google/uuid"
	"github.com/mmorpg-template/backend/internal/domain/character"
	"github.com/mmorpg-template/backend/internal/ports"
	"github.com/mmorpg-template/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMessageQueue is a mock implementation of ports.MessageQueue
type MockMessageQueue struct {
	mock.Mock
	publishedMessages []struct {
		Subject string
		Data    []byte
	}
}

func (m *MockMessageQueue) Connect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockMessageQueue) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMessageQueue) Publish(ctx context.Context, subject string, data []byte) error {
	m.publishedMessages = append(m.publishedMessages, struct {
		Subject string
		Data    []byte
	}{Subject: subject, Data: data})
	args := m.Called(ctx, subject, data)
	return args.Error(0)
}

func (m *MockMessageQueue) PublishWithReply(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	args := m.Called(ctx, subject, data, timeout)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockMessageQueue) Subscribe(ctx context.Context, subject string, handler ports.MessageHandler) (ports.QueueSubscription, error) {
	args := m.Called(ctx, subject, handler)
	return args.Get(0).(ports.QueueSubscription), args.Error(1)
}

func (m *MockMessageQueue) QueueSubscribe(ctx context.Context, subject, queue string, handler ports.MessageHandler) (ports.QueueSubscription, error) {
	args := m.Called(ctx, subject, queue, handler)
	return args.Get(0).(ports.QueueSubscription), args.Error(1)
}

func (m *MockMessageQueue) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	args := m.Called(ctx, subject, data, timeout)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockMessageQueue) CreateStream(ctx context.Context, config ports.StreamConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *MockMessageQueue) DeleteStream(ctx context.Context, stream string) error {
	args := m.Called(ctx, stream)
	return args.Error(0)
}

func (m *MockMessageQueue) GetStreamInfo(ctx context.Context, stream string) (*ports.StreamInfo, error) {
	args := m.Called(ctx, stream)
	return args.Get(0).(*ports.StreamInfo), args.Error(1)
}

func TestEventPublisher_PublishCharacterCreated(t *testing.T) {
	ctx := context.Background()
	mockMQ := new(MockMessageQueue)
	log := logger.New()
	
	// Setup expectations
	mockMQ.On("CreateStream", ctx, mock.AnythingOfType("ports.StreamConfig")).Return(nil)
	mockMQ.On("Publish", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
	
	publisher := NewEventPublisher(mockMQ, log)
	err := publisher.Initialize(ctx)
	assert.NoError(t, err)
	
	// Create test event
	event := &character.CharacterCreatedEvent{
		BaseEvent: character.BaseEvent{
			EventType:   character.EventCharacterCreated,
			CharacterID: uuid.New().String(),
			UserID:      uuid.New().String(),
		},
		Name:       "TestCharacter",
		ClassType:  character.ClassWarrior,
		Race:       character.RaceHuman,
		Gender:     character.GenderMale,
		Level:      1,
		SlotNumber: 1,
	}
	
	// Publish event
	err = publisher.PublishCharacterCreated(ctx, event)
	assert.NoError(t, err)
	
	// Verify event was published
	assert.Len(t, mockMQ.publishedMessages, 1)
	assert.Equal(t, string(character.EventCharacterCreated), mockMQ.publishedMessages[0].Subject)
	
	// Verify event data
	var publishedEvent character.CharacterCreatedEvent
	err = json.Unmarshal(mockMQ.publishedMessages[0].Data, &publishedEvent)
	assert.NoError(t, err)
	assert.Equal(t, event.Name, publishedEvent.Name)
	assert.Equal(t, event.ClassType, publishedEvent.ClassType)
	assert.NotEmpty(t, publishedEvent.EventID)
	assert.NotZero(t, publishedEvent.Timestamp)
}

func TestEventPublisher_PublishCharacterPositionUpdated(t *testing.T) {
	ctx := context.Background()
	mockMQ := new(MockMessageQueue)
	log := logger.New()
	
	// Setup expectations
	mockMQ.On("CreateStream", ctx, mock.AnythingOfType("ports.StreamConfig")).Return(nil)
	mockMQ.On("Publish", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
	
	publisher := NewEventPublisher(mockMQ, log)
	err := publisher.Initialize(ctx)
	assert.NoError(t, err)
	
	// Create test event
	charID := uuid.New()
	previousPos := &character.Position{
		CharacterID: charID,
		WorldID:     "world1",
		ZoneID:      "zone1",
		PositionX:   100.0,
		PositionY:   200.0,
		PositionZ:   50.0,
	}
	
	newPos := &character.Position{
		CharacterID: charID,
		WorldID:     "world1",
		ZoneID:      "zone1",
		PositionX:   150.0,
		PositionY:   250.0,
		PositionZ:   50.0,
	}
	
	event := &character.CharacterPositionUpdatedEvent{
		BaseEvent: character.BaseEvent{
			EventType:   character.EventCharacterPositionUpdated,
			CharacterID: charID.String(),
			UserID:      uuid.New().String(),
		},
		PreviousPosition: previousPos,
		NewPosition:      newPos,
		MovementType:     "walk",
	}
	
	// Publish event
	err = publisher.PublishCharacterPositionUpdated(ctx, event)
	assert.NoError(t, err)
	
	// Verify event was published
	assert.Len(t, mockMQ.publishedMessages, 1)
	assert.Equal(t, string(character.EventCharacterPositionUpdated), mockMQ.publishedMessages[0].Subject)
}

func TestEventPublisher_RetryOnFailure(t *testing.T) {
	ctx := context.Background()
	mockMQ := new(MockMessageQueue)
	log := logger.New()
	
	// Setup expectations - fail first 2 attempts, succeed on 3rd
	mockMQ.On("CreateStream", ctx, mock.AnythingOfType("ports.StreamConfig")).Return(nil)
	mockMQ.On("Publish", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).
		Return(ports.ErrMQPublish).Twice()
	mockMQ.On("Publish", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).
		Return(nil).Once()
	
	publisher := NewEventPublisher(mockMQ, log)
	err := publisher.Initialize(ctx)
	assert.NoError(t, err)
	
	// Create test event
	event := &character.CharacterDeletedEvent{
		BaseEvent: character.BaseEvent{
			EventType:   character.EventCharacterDeleted,
			CharacterID: uuid.New().String(),
			UserID:      uuid.New().String(),
		},
		Name:       "TestCharacter",
		SoftDelete: true,
	}
	
	// Publish event - should succeed after retries
	err = publisher.PublishCharacterDeleted(ctx, event)
	assert.NoError(t, err)
	
	// Verify publish was called 3 times
	mockMQ.AssertNumberOfCalls(t, "Publish", 3)
}

func TestEventPublisher_DeadLetterQueue(t *testing.T) {
	ctx := context.Background()
	mockMQ := new(MockMessageQueue)
	log := logger.New()
	
	// Setup expectations - all attempts fail
	mockMQ.On("CreateStream", ctx, mock.AnythingOfType("ports.StreamConfig")).Return(nil)
	mockMQ.On("Publish", ctx, mock.MatchedBy(func(subject string) bool {
		return subject == string(character.EventCharacterDeleted)
	}), mock.AnythingOfType("[]uint8")).Return(ports.ErrMQPublish)
	
	// Dead letter queue publish should succeed
	mockMQ.On("Publish", ctx, mock.MatchedBy(func(subject string) bool {
		return subject == "character.dlq.character.deleted"
	}), mock.AnythingOfType("[]uint8")).Return(nil)
	
	publisher := NewEventPublisher(mockMQ, log)
	err := publisher.Initialize(ctx)
	assert.NoError(t, err)
	
	// Create test event
	event := &character.CharacterDeletedEvent{
		BaseEvent: character.BaseEvent{
			EventType:   character.EventCharacterDeleted,
			CharacterID: uuid.New().String(),
			UserID:      uuid.New().String(),
		},
		Name:       "TestCharacter",
		SoftDelete: true,
	}
	
	// Publish event - should fail but send to DLQ
	err = publisher.PublishCharacterDeleted(ctx, event)
	assert.Error(t, err)
	
	// Verify DLQ publish was called
	mockMQ.AssertCalled(t, "Publish", ctx, "character.dlq.character.deleted", mock.AnythingOfType("[]uint8"))
}