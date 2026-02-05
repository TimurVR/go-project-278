package handler

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go-project-278/Internal/dto"
	"go-project-278/Internal/repository"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type App struct {
	Ctx  context.Context
	Repo repository.PostRepository
}

type ValidationErrorResponse struct {
	Errors map[string]string `json:"errors"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func isValidURL(urlStr string) bool {
	u, err := url.ParseRequestURI(urlStr)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func isValidShortName(shortName string) bool {
	if shortName == "" {
		return true
	}
	if len(shortName) < 3 || len(shortName) > 32 {
		return false
	}
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", shortName)
	return matched
}

func respondWithValidationError(c *gin.Context, field, message string) {
	c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
		Errors: map[string]string{field: message},
	})
}

func respondWithValidationErrors(c *gin.Context, errorsMap map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, ValidationErrorResponse{
		Errors: errorsMap,
	})
}

func respondWithBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		Error: message,
	})
}

func JSONValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				c.Next()
				return
			}
			bodyBytes, _ := c.GetRawData()
			if len(bodyBytes) == 0 {
				c.Next()
				return
			}
			if !json.Valid(bodyBytes) {
				respondWithBadRequest(c, "invalid request")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func GenerateUniqueString() string {
	return uuid.New().String()
}

func GenerateShortCode(url string) string {
	hash := sha256.Sum256([]byte(url + time.Now().String()))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	return encoded[:8]
}

func (a *App) Routes(r *gin.Engine) {
	r.GET("/api/r/:code", a.Redirect)
	api := r.Group("/api")
	{
		api.GET("/links", a.GetLinks)
		api.POST("/links", a.CreateLinks)
		api.GET("/links/:id", a.HandleLink)
		api.PUT("/links/:id", a.HandleLink)
		api.DELETE("/links/:id", a.HandleLink)
		api.GET("/link_visits", a.GetVisits)
	}
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "404 Not Found",
			"message": "Запрашиваемый ресурс не найден",
			"path":    c.Request.URL.Path,
			"method":  c.Request.Method,
		})
	})
}

func (a *App) Redirect(c *gin.Context) {
	code := c.Param("code")
	link, err := a.Repo.GetLinkByShortName(a.Ctx, code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Link not found"})
		return
	}

	visit := dto.Visit{
		LinkID:    link.Id,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Status:    http.StatusFound,
		CreatedAt: time.Now(),
	}
	err = a.Repo.RecordVisit(a.Ctx, visit)
	if err != nil {
		fmt.Print(err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err})
		return
	}
	c.Redirect(http.StatusFound, link.Original_url)
}

func (a *App) HandleLink(rw *gin.Context) {
	req := rw.Param("id")
	switch rw.Request.Method {
	case "GET":
		id, err2 := strconv.Atoi(req)
		if err2 != nil {
			rw.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		link, err := a.Repo.GetLinkByID(a.Ctx, id)
		if err != nil {
			rw.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		rw.JSON(http.StatusOK, link)

	case "PUT":
		id, err2 := strconv.Atoi(req)
		if err2 != nil {
			respondWithBadRequest(rw, "invalid id")
			return
		}
		currentLink, err := a.Repo.GetLinkByID(a.Ctx, id)
		if err != nil {
			rw.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		var request dto.LinkRequest
		if err := rw.ShouldBindJSON(&request); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				errorsMap := make(map[string]string)
				for _, e := range ve {
					field := strings.ToLower(e.Field())
					switch e.Tag() {
					case "required":
						errorsMap[field] = "обязательное поле"
					default:
						errorsMap[field] = "некорректное значение"
					}
				}
				respondWithValidationErrors(rw, errorsMap)
				return
			}
			respondWithBadRequest(rw, "invalid request")
			return
		}
		validationErrors := make(map[string]string)
		if request.Original_url == "" {
			validationErrors["original_url"] = "обязательное поле"
		} else if !isValidURL(request.Original_url) {
			validationErrors["original_url"] = "некорректный URL"
		}
		if request.Short_name != "" && !isValidShortName(request.Short_name) {
			if len(request.Short_name) < 3 || len(request.Short_name) > 32 {
				validationErrors["short_name"] = "длина должна быть от 3 до 32 символов"
			} else {
				validationErrors["short_name"] = "может содержать только буквы, цифры, дефисы и подчеркивания"
			}
		}
		if len(validationErrors) > 0 {
			respondWithValidationErrors(rw, validationErrors)
			return
		}
		shortName := request.Short_name
		if shortName == "" {
			shortName = currentLink.Short_name
		} else if shortName != currentLink.Short_name {
			exists, err := a.Repo.CheckShortNameExists(a.Ctx, shortName)
			if err != nil {
				rw.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
			if exists {
				respondWithValidationError(rw, "short_name", "уже существует")
				return
			}
		}
		host := rw.Request.Host
		scheme := "https"
		if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
			scheme = "http"
		}
		baseURL := fmt.Sprintf("%s://%s", scheme, host)
		responce := dto.LinkResponce{
			Id:           id,
			Original_url: request.Original_url,
			Short_name:   shortName,
			Short_url:    baseURL + "/api/r/" + shortName,
		}
		err1 := a.Repo.UpdateLink(a.Ctx, responce)
		if err1 != nil {
			if strings.Contains(err1.Error(), "unique constraint") ||
				strings.Contains(err1.Error(), "duplicate") {
				respondWithValidationError(rw, "short_name", "уже существует")
				return
			}
			rw.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		rw.JSON(http.StatusOK, responce)

	case "DELETE":
		id, err2 := strconv.Atoi(req)
		if err2 != nil {
			respondWithBadRequest(rw, "invalid id")
			return
		}
		err := a.Repo.DeleteLinkByID(a.Ctx, id)
		if err != nil {
			rw.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		rw.Status(204)
	}
}

func (a *App) CreateLinks(rw *gin.Context) {
	var request dto.LinkRequest
	if err := rw.ShouldBindJSON(&request); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			errorsMap := make(map[string]string)
			for _, e := range ve {
				field := strings.ToLower(e.Field())
				switch e.Tag() {
				case "required":
					errorsMap[field] = "обязательное поле"
				default:
					errorsMap[field] = "некорректное значение"
				}
			}
			respondWithValidationErrors(rw, errorsMap)
			return
		}
		respondWithBadRequest(rw, "invalid request")
		return
	}

	validationErrors := make(map[string]string)
	if request.Original_url == "" {
		validationErrors["original_url"] = "обязательное поле"
	} else if !isValidURL(request.Original_url) {
		validationErrors["original_url"] = "некорректный URL"
	}

	if request.Short_name != "" && !isValidShortName(request.Short_name) {
		if len(request.Short_name) < 3 || len(request.Short_name) > 32 {
			validationErrors["short_name"] = "длина должна быть от 3 до 32 символов"
		} else {
			validationErrors["short_name"] = "может содержать только буквы, цифры, дефисы и подчеркивания"
		}
	}

	if len(validationErrors) > 0 {
		respondWithValidationErrors(rw, validationErrors)
		return
	}

	if request.Short_name != "" {
		exists, err := a.Repo.CheckShortNameExists(a.Ctx, request.Short_name)
		if err != nil {
			rw.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if exists {
			respondWithValidationError(rw, "short_name", "уже существует")
			return
		}
	}

	shortName := request.Short_name
	if shortName == "" {
		shortName = GenerateUniqueString()
	}

	host := rw.Request.Host
	scheme := "https"
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") {
		scheme = "http"
	}

	baseURL := fmt.Sprintf("%s://%s", scheme, host)
	responce := dto.LinkResponce1{
		Original_url: request.Original_url,
		Short_name:   shortName,
		Short_url:    baseURL + "/r/" + shortName,
	}

	id, err1 := a.Repo.CreateLink(a.Ctx, responce)
	if err1 != nil {
		if strings.Contains(err1.Error(), "unique constraint") ||
			strings.Contains(err1.Error(), "duplicate") {
			respondWithValidationError(rw, "short_name", "уже существует")
			return
		}
		rw.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	responce1 := dto.LinkResponce{
		Id:           id,
		Original_url: request.Original_url,
		Short_name:   shortName,
		Short_url:    baseURL + "/api/r/" + shortName,
	}
	rw.JSON(http.StatusCreated, responce1)
}

func (a *App) GetLinks(rw *gin.Context) {
	allLinks, _ := a.Repo.ListLinks(a.Ctx)
	total := len(allLinks)
	rangeParam := rw.Query("range")
	if rangeParam == "" {
		rw.Header("Content-Range", fmt.Sprintf("links 0-%d/%d", total-1, total))
		rw.JSON(http.StatusOK, allLinks)
		return
	}
	rangeParam = strings.Trim(rangeParam, "[]")
	parts := strings.Split(rangeParam, ",")
	if len(parts) != 2 {
		respondWithBadRequest(rw, "range must be in format [start,end]")
		return
	}
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		respondWithBadRequest(rw, "range values must be integers")
		return
	}
	if start < 0 || end < 0 || start > end {
		respondWithBadRequest(rw, "invalid range values")
		return
	}
	responce, err := a.Repo.ListLinksLimited(a.Ctx, start, end)
	if err != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	rw.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", start, end, total))
	rw.JSON(http.StatusOK, responce)
}

func (a *App) GetVisits(rw *gin.Context) {
	allVisits, _ := a.Repo.ListVisits(a.Ctx)
	total := len(allVisits)
	rangeParam := rw.Query("range")
	if rangeParam == "" {
		rw.Header("Content-Range", fmt.Sprintf("visits 0-%d/%d", total-1, total))
		rw.JSON(http.StatusOK, allVisits)
		return
	}
	rangeParam = strings.Trim(rangeParam, "[]")
	parts := strings.Split(rangeParam, ",")
	if len(parts) != 2 {
		respondWithBadRequest(rw, "range must be in format [start,end]")
		return
	}
	start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil {
		respondWithBadRequest(rw, "range values must be integers")
		return
	}
	if start < 0 || end < 0 || start > end {
		respondWithBadRequest(rw, "invalid range values")
		return
	}
	responce, err := a.Repo.ListVisitsLimited(a.Ctx, start, end)
	if err != nil {
		fmt.Print(err)
		rw.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	rw.Header("Content-Range", fmt.Sprintf("visits %d-%d/%d", start, end, total))
	rw.JSON(http.StatusOK, responce)
}
