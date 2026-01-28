package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"go-project-278/Internal/dto"
	"go-project-278/Internal/handler"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateLink(ctx context.Context, link dto.LinkResponce) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockRepository) GetLinkByID(ctx context.Context, Id int) (*dto.LinkResponce, error) {
	args := m.Called(ctx, Id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LinkResponce), args.Error(1)
}

func (m *MockRepository) ListLinks(ctx context.Context) ([]*dto.LinkResponce, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*dto.LinkResponce), args.Error(1)
}

func (m *MockRepository) UpdateLink(ctx context.Context, link dto.LinkResponce) error {
	args := m.Called(ctx, link)
	return args.Error(0)
}

func (m *MockRepository) DeleteLinkByID(ctx context.Context, Id int) error {
	args := m.Called(ctx, Id)
	return args.Error(0)
}

// Вспомогательная функция для создания роутера
func setupTestRouter(app *handler.App) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	app.Routes(router)
	return router
}

// Тест на успешное создание ссылки
func TestCreateLinks_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CreateLink", mock.Anything, mock.AnythingOfType("dto.LinkResponce")).
		Return(nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{
		"original_url": "https://example.com",
		"short_name": "test-short"
	}`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

// Тест на создание ссылки с автогенерацией short_name
func TestCreateLinks_Success_AutoGenerateShortName(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CreateLink", mock.Anything, mock.AnythingOfType("dto.LinkResponce")).
		Return(nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{
		"original_url": "https://example.com"
	}`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

// Тест на ошибку валидации JSON
func TestCreateLinks_BadRequest_InvalIdJSON(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{invalId json`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на ошибку базы данных при создании
func TestCreateLinks_InternalServerError(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CreateLink", mock.Anything, mock.AnythingOfType("dto.LinkResponce")).
		Return(errors.New("database error"))

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{
		"original_url": "https://example.com",
		"short_name": "test-short"
	}`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

// Тест на успешное получение ссылки по Id
func TestHandleLink_GET_Success(t *testing.T) {
    mockRepo := &MockRepository{}
    expectedLink := &dto.LinkResponce{
        Id:           1,
        Original_url: "https://example.com",
        Short_name:   "test",
        Short_url:    "abc123",
    }
    mockRepo.On("GetLinkByID", mock.Anything, 1).
        Return(expectedLink, nil)  
    app := &handler.App{
        Ctx:  context.Background(),
        Repo: mockRepo,
    }
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
    c.Request = httptest.NewRequest("GET", "/api/links/1", nil)
    app.HandleLink(c)
    assert.Equal(t, http.StatusOK, w.Code)
    var response dto.LinkResponce
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Equal(t, expectedLink.Id, response.Id)
    assert.Equal(t, expectedLink.Original_url, response.Original_url)
    mockRepo.AssertExpectations(t) 
}

// Тест на ошибку при получении ссылки (не найдена)
func TestHandleLink_GET_NotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	
	mockRepo.On("GetLinkByID", mock.Anything, 999).
		Return(nil, errors.New("link not found"))

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
	c.Request = httptest.NewRequest("GET", "/api/links/999", nil)
	
	app.HandleLink(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertExpectations(t)
}

// Тест на успешное обновление ссылки
func TestHandleLink_PUT_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	
	mockRepo.On("UpdateLink", mock.Anything, mock.AnythingOfType("dto.LinkResponce")).
		Return(nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "Id", Value: "1"}}
	
	jsonData := `{
		"original_url": "https://updated.com",
		"short_name": "updated-name"
	}`
	
	c.Request = httptest.NewRequest("PUT", "/api/links/1", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.HandleLink(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dto.LinkResponce
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "https://updated.com", response.Original_url)
	assert.Equal(t, "updated-name", response.Short_name)
	assert.NotEmpty(t, response.Short_url)
	
	mockRepo.AssertExpectations(t)
}
// Тест на ошибку валидации JSON при обновлении
func TestHandleLink_PUT_BadRequest_InvalIdJSON(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "Id", Value: "1"}}
	
	jsonData := `{invalId json`
	
	c.Request = httptest.NewRequest("PUT", "/api/links/1", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.HandleLink(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на ошибку при обновлении (неверный Id)
func TestHandleLink_PUT_InvalIdId(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalId"}}
	
	jsonData := `{
		"original_url": "https://example.com"
	}`
	
	c.Request = httptest.NewRequest("PUT", "/api/links/invalId", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.HandleLink(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на ошибку базы данных при обновлении
func TestHandleLink_PUT_UpdateError(t *testing.T) {
	mockRepo := &MockRepository{}
	
	mockRepo.On("UpdateLink", mock.Anything, mock.AnythingOfType("dto.LinkResponce")).
		Return(errors.New("update failed"))

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	
	jsonData := `{
		"original_url": "https://example.com",
		"short_name": "test"
	}`
	
	c.Request = httptest.NewRequest("PUT", "/api/links/1", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.HandleLink(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertExpectations(t)
}

// Тест на успешное удаление ссылки
func TestHandleLink_DELETE_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	
	mockRepo.On("DeleteLinkByID", mock.Anything, 1).
		Return(nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest("DELETE", "/api/links/1", nil)
	
	app.HandleLink(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
	
	mockRepo.AssertExpectations(t)
}

// Тест на ошибку при удалении (неверный Id)
func TestHandleLink_DELETE_InvalIdId(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalId"}}
	c.Request = httptest.NewRequest("DELETE", "/api/links/invalId", nil)
	
	app.HandleLink(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// Тест на ошибку базы данных при удалении
func TestHandleLink_DELETE_NotFound(t *testing.T) {
	mockRepo := &MockRepository{}
	
	mockRepo.On("DeleteLinkByID", mock.Anything, 999).
		Return(errors.New("link not found"))

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
	c.Request = httptest.NewRequest("DELETE", "/api/links/999", nil)
	
	app.HandleLink(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockRepo.AssertExpectations(t)
}

// Тест на успешное получение всех ссылок
func TestGetLinks_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	
	expectedLinks := []*dto.LinkResponce{
		{
			Id:           1,
			Original_url: "https://example1.com",
			Short_name:   "test1",
			Short_url:    "abc123",
		},
		{
			Id:           2,
			Original_url: "https://example2.com",
			Short_name:   "test2",
			Short_url:    "def456",
		},
	}
	
	mockRepo.On("ListLinks", mock.Anything).
		Return(expectedLinks, nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Request = httptest.NewRequest("GET", "/api/links", nil)
	
	app.GetLinks(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response []dto.LinkResponce
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, expectedLinks[0].Id, response[0].Id)
	assert.Equal(t, expectedLinks[1].Id, response[1].Id)
	
	mockRepo.AssertExpectations(t)
}

// Тест на получение пустого списка ссылок
func TestGetLinks_EmptyList(t *testing.T) {
	mockRepo := &MockRepository{}
	
	mockRepo.On("ListLinks", mock.Anything).
		Return([]*dto.LinkResponce{}, nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Request = httptest.NewRequest("GET", "/api/links", nil)
	
	app.GetLinks(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response []dto.LinkResponce
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Empty(t, response)
	
	mockRepo.AssertExpectations(t)
}

// Тест на ошибку базы данных при получении списка ссылок
func TestGetLinks_DatabaseError(t *testing.T) {
	mockRepo := &MockRepository{}
	
	mockRepo.On("ListLinks", mock.Anything).
		Return([]*dto.LinkResponce{}, errors.New("database error"))

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Request = httptest.NewRequest("GET", "/api/links", nil)
	
	app.GetLinks(c)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
	
	mockRepo.AssertExpectations(t)
}

// Тест на обработку несуществующего маршрута
func TestRoutes_NoRoute(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	router := setupTestRouter(app)
	
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/nonexistent", nil)
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "404 Not Found", response["error"])
	assert.Equal(t, "Запрашиваемый ресурс не найден", response["message"])
	assert.Equal(t, "/nonexistent", response["path"])
	assert.Equal(t, "GET", response["method"])
}








