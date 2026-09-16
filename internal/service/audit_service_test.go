package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockAuditRepo struct {
	mock.Mock
}

func (m *mockAuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *mockAuditRepo) FindAll(ctx context.Context, page, limit int, entityType, action string) ([]model.AuditLog, int64, error) {
	args := m.Called(ctx, page, limit, entityType, action)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]model.AuditLog), args.Get(1).(int64), args.Error(2)
}

func TestAuditService_Log_Success(t *testing.T) {
	repo := new(mockAuditRepo)
	svc := service.NewAuditService(repo)

	userID := uint64(1)
	entityID := uint64(42)
	oldVal := map[string]any{"status": "PENDING"}
	newVal := map[string]any{"status": "CONFIRMED"}

	repo.On("Create", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return *l.UserID == userID &&
			l.Action == "UPDATE_STATUS" &&
			l.EntityType == "order" &&
			*l.EntityID == entityID &&
			l.OldValue != nil &&
			l.NewValue != nil &&
			*l.IPAddress == "127.0.0.1"
	})).Return(nil).Once()

	err := svc.Log(context.Background(), service.AuditEntry{
		UserID:     &userID,
		Action:     "UPDATE_STATUS",
		EntityType: "order",
		EntityID:   &entityID,
		OldValue:   oldVal,
		NewValue:   newVal,
		IPAddress:  "127.0.0.1",
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuditService_Log_NullableFields(t *testing.T) {
	repo := new(mockAuditRepo)
	svc := service.NewAuditService(repo)

	repo.On("Create", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return l.UserID == nil &&
			l.Action == "SYSTEM_EVENT" &&
			l.EntityType == "system" &&
			l.EntityID == nil &&
			l.OldValue == nil &&
			l.NewValue == nil &&
			l.IPAddress == nil
	})).Return(nil).Once()

	err := svc.Log(context.Background(), service.AuditEntry{
		UserID:     nil,
		Action:     "SYSTEM_EVENT",
		EntityType: "system",
		EntityID:   nil,
		OldValue:   nil,
		NewValue:   nil,
		IPAddress:  "",
	})

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuditService_Log_NeverFailsBusinessFlow_WhenRepoFails(t *testing.T) {
	repo := new(mockAuditRepo)
	svc := service.NewAuditService(repo)

	// Repo returns a database error
	repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db connection lost")).Once()

	err := svc.Log(context.Background(), service.AuditEntry{
		Action:     "CREATE",
		EntityType: "product",
		NewValue:   map[string]any{"name": "Bread"},
	})

	// Crucial rule: AuditService.Log must NEVER return error
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuditService_Log_NeverFailsBusinessFlow_WhenMarshalFails(t *testing.T) {
	repo := new(mockAuditRepo)
	svc := service.NewAuditService(repo)

	// Channel cannot be marshaled into JSON
	unmarshalable := make(chan int)

	err := svc.Log(context.Background(), service.AuditEntry{
		Action:     "TEST",
		EntityType: "channel",
		OldValue:   unmarshalable,
	})

	// Crucial rule: must not panic and must return nil
	assert.NoError(t, err)
	// Repo.Create should not even be called if marshalling fails
	// Test when NewValue fails to marshal
	err = svc.Log(context.Background(), service.AuditEntry{
		Action:     "TEST",
		EntityType: "channel",
		NewValue:   unmarshalable,
	})
	assert.NoError(t, err)
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestAuditService_Log_RawJSONOrStringValue(t *testing.T) {
	repo := new(mockAuditRepo)
	svc := service.NewAuditService(repo)

	rawJSON := `{"key":"value"}`
	repo.On("Create", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return l.NewValue != nil && *l.NewValue == rawJSON
	})).Return(nil).Once()

	err := svc.Log(context.Background(), service.AuditEntry{
		Action:     "TEST_RAW",
		EntityType: "test",
		NewValue:   rawJSON,
	})
	assert.NoError(t, err)

	// Test with []byte raw JSON
	rawBytes := []byte(`{"byteKey":"byteValue"}`)
	repo.On("Create", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return l.NewValue != nil && *l.NewValue == string(rawBytes)
	})).Return(nil).Once()

	err = svc.Log(context.Background(), service.AuditEntry{
		Action:     "TEST_BYTES",
		EntityType: "test",
		NewValue:   rawBytes,
	})
	assert.NoError(t, err)

	// Test with plain string that is not JSON
	plainString := "plain string"
	repo.On("Create", mock.Anything, mock.MatchedBy(func(l *model.AuditLog) bool {
		return l.NewValue != nil && *l.NewValue == `"plain string"`
	})).Return(nil).Once()

	err = svc.Log(context.Background(), service.AuditEntry{
		Action:     "TEST_PLAIN",
		EntityType: "test",
		NewValue:   plainString,
	})
	assert.NoError(t, err)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAuditService_List_Success(t *testing.T) {
	repo := new(mockAuditRepo)
	svc := service.NewAuditService(repo)

	uid := uint64(1)
	eid := uint64(10)
	oldVal := `{"name":"Old"}`
	newVal := `{"name":"New"}`
	ip := "127.0.0.1"
	now := time.Now()

	dbLogs := []model.AuditLog{
		{
			ID:         1,
			UserID:     &uid,
			Action:     "UPDATE",
			EntityType: "product",
			EntityID:   &eid,
			OldValue:   &oldVal,
			NewValue:   &newVal,
			IPAddress:  &ip,
			CreatedAt:  now,
		},
	}

	repo.On("FindAll", mock.Anything, 1, 20, "product", "UPDATE").
		Return(dbLogs, int64(1), nil).Once()

	resp, meta, err := svc.List(context.Background(), 1, 20, "product", "UPDATE")
	require.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, uint64(1), resp[0].ID)
	assert.Equal(t, "UPDATE", resp[0].Action)
	assert.Equal(t, "product", resp[0].EntityType)
	assert.Equal(t, int64(1), meta.TotalRows)
	assert.Equal(t, 1, meta.Page)
	assert.Equal(t, 20, meta.Limit)

	// Verify OldValue and NewValue are valid JSON representations
	oldBytes, err := json.Marshal(resp[0].OldValue)
	require.NoError(t, err)
	assert.JSONEq(t, oldVal, string(oldBytes))

	repo.AssertExpectations(t)
}

func TestAuditService_List_RepositoryError(t *testing.T) {
	repo := new(mockAuditRepo)
	svc := service.NewAuditService(repo)

	repo.On("FindAll", mock.Anything, 1, 20, "", "").
		Return(nil, int64(0), errors.New("db query error")).Once()

	resp, meta, err := svc.List(context.Background(), 1, 20, "", "")
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Zero(t, meta.TotalRows)

	repo.AssertExpectations(t)
}
