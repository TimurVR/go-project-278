package handler

import (
	"context"
	"go-project-278/Internal/dto"
	"go-project-278/Internal/repository"
	"net/http"
	"strconv"
	"crypto/md5"
	"encoding/base64"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type App struct {
	Ctx       context.Context
	Repo      repository.PostRepository
}
func GenerateUniqueString() string {
    return uuid.New().String()
}
func GenerateShortCode(url string) string {
	hash := md5.Sum([]byte(url))
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	shortCode := encoded[:6]
	return strings.TrimRight(shortCode, "=")
}

func (a *App) Routes(r *gin.Engine) {
	r.POST("/api/links", a.CreateLinks)
	r.GET("/api/links", a.GetLinks)
	r.GET("/api/links/:id",  a.HandleLink)
	r.PUT("/api/links/:id",  a.HandleLink)
	r.DELETE("/api/links/:id",  a.HandleLink)
	r.NoRoute(func(c *gin.Context) {
        c.JSON(http.StatusNotFound, gin.H{
            "error":   "404 Not Found",
            "message": "Запрашиваемый ресурс не найден",
            "path":    c.Request.URL.Path,
            "method":  c.Request.Method,
        })
    })
}
func (a *App) HandleLink(rw *gin.Context) {
	req := rw.Param("id")
    switch rw.Request.Method {
    case "GET":
		id,err2:=strconv.Atoi(req)
		if err2!=nil{
			rw.JSON(http.StatusBadRequest, gin.H{"error": err2.Error()})
			return
		}
        link,err:=a.Repo.GetLinkByID(a.Ctx,id)
		if err!=nil{
			rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rw.JSON(http.StatusOK, link)
    case "PUT":
		id,err2:=strconv.Atoi(req)
		if err2!=nil{
			rw.JSON(http.StatusBadRequest, gin.H{"error": err2.Error()})
			return
		}
        var request dto.LinkRequest
		err := rw.ShouldBindJSON(&request)
		if err != nil {
			rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var responce dto.LinkResponce
		if (request.Short_name==""){
			responce.Short_name=GenerateUniqueString() 
		}
		responce.Id=id
		responce.Original_url=request.Original_url
		responce.Short_name=request.Short_name
		responce.Short_url=GenerateShortCode(request.Original_url)
		err1:=a.Repo.UpdateLink(a.Ctx,responce)
		if err1 != nil {
			rw.JSON(http.StatusBadRequest, gin.H{"error": err1.Error()})
		return
		}
		rw.JSON(http.StatusOK, responce)
    case "DELETE":
		id,err2:=strconv.Atoi(req)
		if err2!=nil{
			rw.JSON(http.StatusBadRequest, gin.H{"error": err2.Error()})
			return
		}
        err:=a.Repo.DeleteLinkByID(a.Ctx,id)
		if err!=nil{
			rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		rw.Status(204) 
    }
}
func (a *App) CreateLinks(rw *gin.Context) {
	var request dto.LinkRequest
	err := rw.ShouldBindJSON(&request)
	if err != nil {
		rw.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var responce dto.LinkResponce
	if (request.Short_name==""){
		responce.Short_name=GenerateUniqueString() 
	}
	responce.Original_url=request.Original_url
	responce.Short_name=request.Short_name
	responce.Short_url=GenerateShortCode(request.Original_url)
	err1:=a.Repo.CreateLink(a.Ctx, responce)
	if err1 != nil {
		rw.JSON(http.StatusInternalServerError, gin.H{"error": err1.Error()})
		return
	}
	rw.Status(201) 
}

func (a *App) GetLinks(rw *gin.Context) {
	rangeParam := rw.Query("range")
	if rangeParam == "" {
           responce, err := a.Repo.ListLinks(a.Ctx)
		   if err != nil {
				rw.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			rw.JSON(http.StatusOK, responce)
    }
	rangeParam = strings.Trim(rangeParam, "[]")
    parts := strings.Split(rangeParam, ",")
    if len(parts) != 2 {
            rw.JSON(400, gin.H{"error": "range must be in format [start,end]"})
            return
    }
    start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
    end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
    if err1 != nil || err2 != nil {
            rw.JSON(400, gin.H{"error": "range values must be integers"})
            return
     }
    if start < 0 || end < 0 || start > end {
            rw.JSON(400, gin.H{"error": "invalid range values"})
            return
    }
	responce, err := a.Repo.ListLinksLimited(a.Ctx,start,end)
	if err != nil {
				rw.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
	}
	rw.JSON(http.StatusOK, responce)

}
