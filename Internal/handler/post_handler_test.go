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
func (m *MockRepository) CheckShortNameExists(ctx context.Context, shortName string) (bool, error) {
    args := m.Called(ctx, shortName)
    return args.Bool(0), args.Error(1)
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
func (m *MockRepository) ListLinksLimited(ctx context.Context, start, end int) ([]*dto.LinkResponce, error) {
    args := m.Called(ctx, start, end)
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

func setupTestRouter(app *handler.App) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	app.Routes(router)
	return router
}
func (m *MockRepository) GetLinkByShortName(ctx context.Context, shortName string) (*dto.LinkResponce, error) {
	args := m.Called(ctx, shortName)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).(*dto.LinkResponce), args.Error(1)
}

func (m *MockRepository) RecordVisit(ctx context.Context, visit dto.Visit) error {
	args := m.Called(ctx, visit)
	return args.Error(0)
}

func (m *MockRepository) ListVisits(ctx context.Context) ([]*dto.Visit, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*dto.Visit), args.Error(1)
}

func (m *MockRepository) ListVisitsLimited(ctx context.Context, start, limit int) ([]*dto.Visit, error) {
	args := m.Called(ctx, start, limit)
	if args.Get(0) == nil { return nil, args.Error(1) }
	return args.Get(0).([]*dto.Visit), args.Error(1)
}
func TestCreateLinks_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CheckShortNameExists", mock.Anything, "test-short").
		Return(false, nil)
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
	
	assert.Equal(t, http.StatusCreated, w.Code)
	mockRepo.AssertCalled(t, "CheckShortNameExists", mock.Anything, "test-short")
	mockRepo.AssertExpectations(t)
}

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
	
	assert.Equal(t, http.StatusCreated, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestCreateLinks_BadRequest_InvalidJSON(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{invalid json`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateLinks_InternalServerError(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CheckShortNameExists", mock.Anything, "test-short").
		Return(false, nil)
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

func TestCreateLinks_ValidationError_EmptyOriginalUrl(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{
		"original_url": "",
		"short_name": "test"
	}`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateLinks_ValidationError_InvalidUrl(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{
		"original_url": "invalid-url",
		"short_name": "test"
	}`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateLinks_ValidationError_ShortNameTooShort(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{
		"original_url": "https://example.com",
		"short_name": "ab"
	}`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateLinks_ValidationError_ShortNameTooLong(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{
		"original_url": "https://example.com",
		"short_name": "this-is-a-very-long-short-name-that-exceeds"
	}`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestCreateLinks_ValidationError_ShortNameDuplicate(t *testing.T) {
	mockRepo := &MockRepository{}
	mockRepo.On("CheckShortNameExists", mock.Anything, "existing").
		Return(true, nil)

	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	jsonData := `{
		"original_url": "https://example.com",
		"short_name": "existing"
	}`
	
	c.Request = httptest.NewRequest("POST", "/api/links", bytes.NewBufferString(jsonData))
	c.Request.Header.Set("Content-Type", "application/json")
	
	app.CreateLinks(c)
	
	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	mockRepo.AssertExpectations(t)
}

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
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestHandleLink_PUT_Success(t *testing.T) {
    mockRepo := &MockRepository{}
    mockRepo.On("CheckShortNameExists", mock.Anything, "updated-name").
        Return(false, nil)  
    mockRepo.On("UpdateLink", mock.Anything, mock.AnythingOfType("dto.LinkResponce")).
        Return(nil)

    app := &handler.App{
        Ctx:  context.Background(),
        Repo: mockRepo,
    }
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Params = gin.Params{gin.Param{Key: "id", Value: "1"}}
    
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

func TestHandleLink_DELETE_InvalidId(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid"}}
	c.Request = httptest.NewRequest("DELETE", "/api/links/invalid", nil)
	
	app.HandleLink(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

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
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

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

func TestGetLinksLimited_Success(t *testing.T) {
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
    mockRepo.On("ListLinksLimited", mock.Anything, 0, 10).
        Return(expectedLinks, nil)  
    app := &handler.App{
        Ctx:  context.Background(),
        Repo: mockRepo,
    }
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/links?range=[0,10]", nil)    
    app.GetLinks(c)  
    assert.Equal(t, http.StatusOK, w.Code)
    var links []*dto.LinkResponce
    err := json.Unmarshal(w.Body.Bytes(), &links)
    assert.NoError(t, err)
    assert.Len(t, links, 2)
    assert.Equal(t, "https://example1.com", links[0].Original_url)
    assert.Equal(t, "test1", links[0].Short_name)
    assert.Equal(t, "https://example2.com", links[1].Original_url)
    assert.Equal(t, "test2", links[1].Short_name)
    mockRepo.AssertCalled(t, "ListLinks", mock.Anything)
    mockRepo.AssertCalled(t, "ListLinksLimited", mock.Anything, 0, 10)
    mockRepo.AssertExpectations(t)
}

func TestGetLinksLimited_EmptyList(t *testing.T) {
    mockRepo := &MockRepository{}
    mockRepo.On("ListLinks", mock.Anything).
        Return([]*dto.LinkResponce{}, nil)
    
    mockRepo.On("ListLinksLimited", mock.Anything, 5, 15).
        Return([]*dto.LinkResponce{}, nil)
    app := &handler.App{
        Ctx:  context.Background(),
        Repo: mockRepo,
    }
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/links?range=[5,15]", nil)
    app.GetLinks(c)
    assert.Equal(t, http.StatusOK, w.Code)
    var links []*dto.LinkResponce
    err := json.Unmarshal(w.Body.Bytes(), &links)
    assert.NoError(t, err)
    assert.Empty(t, links)
    mockRepo.AssertCalled(t, "ListLinks", mock.Anything)
    mockRepo.AssertCalled(t, "ListLinksLimited", mock.Anything, 5, 15)
    mockRepo.AssertExpectations(t)
}

func TestGetLinksLimited_DatabaseError(t *testing.T) {
    mockRepo := &MockRepository{}
    mockRepo.On("ListLinks", mock.Anything).
        Return([]*dto.LinkResponce{
            {Id: 1, Original_url: "https://example.com", Short_name: "test"},
        }, nil)
    mockRepo.On("ListLinksLimited", mock.Anything, 0, 10).
        Return(nil, errors.New("database error"))
    app := &handler.App{
        Ctx:  context.Background(),
        Repo: mockRepo,
    }
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("GET", "/api/links?range=[0,10]", nil)
    app.GetLinks(c)
    assert.Equal(t, http.StatusInternalServerError, w.Code)
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Contains(t, response, "error")
    mockRepo.AssertCalled(t, "ListLinks", mock.Anything)
    mockRepo.AssertCalled(t, "ListLinksLimited", mock.Anything, 0, 10)
    mockRepo.AssertExpectations(t)
}

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

func TestRedirect_Success(t *testing.T) {
	mockRepo := &MockRepository{}
	expectedLink := &dto.LinkResponce{
		Id:           1,
		Original_url: "https://example.com",
		Short_name:   "testcode",
	}
	mockRepo.On("GetLinkByShortName", mock.Anything, "testcode").Return(expectedLink, nil)
	mockRepo.On("RecordVisit", mock.Anything, mock.AnythingOfType("dto.Visit")).Return(nil)
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}
	router := setupTestRouter(app)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/r/testcode", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Location"))
	mockRepo.AssertExpectations(t)
}

func TestGetVisits_Success_WithRange(t *testing.T) {
	mockRepo := &MockRepository{}
	allVisits := []*dto.Visit{{Id: 1}, {Id: 2}, {Id: 3}}
	mockRepo.On("ListVisits", mock.Anything).Return(allVisits, nil)
	limitedVisits := []*dto.Visit{{Id: 1}, {Id: 2}}
	mockRepo.On("ListVisitsLimited", mock.Anything, 0, 1).Return(limitedVisits, nil)
	app := &handler.App{
		Ctx:  context.Background(),
		Repo: mockRepo,
	}
	router := setupTestRouter(app)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/link_visits?range=[0,1]", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "visits 0-1/3", w.Header().Get("Content-Range"))
	mockRepo.AssertExpectations(t)
}

func TestNoRoute_ReturnsJSON(t *testing.T) {
	mockRepo := &MockRepository{}
	app := &handler.App{Repo: mockRepo}
	router := setupTestRouter(app)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/undefined-route", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Запрашиваемый ресурс не найден")
}





