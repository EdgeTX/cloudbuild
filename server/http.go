package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/gin-gonic/gin"
	ginlogrus "github.com/toorop/gin-logrus"
)

type Application struct {
	artifactory *artifactory.Artifactory
	server      *http.Server
}

func New(artifactory *artifactory.Artifactory) *Application {
	return &Application{
		artifactory: artifactory,
		server:      nil,
	}
}

func (application *Application) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Healthy",
	})
}

func (application *Application) root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "/",
	})
}

func (application *Application) createBuildJob(c *gin.Context) {
	decoder := json.NewDecoder(c.Request.Body)
	var req CreateBuildJobRequest
	err := decoder.Decode(&req)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusUnprocessableEntity,
			NewErrorResponse("Failed to decode your request"),
		)
		return
	}

	errs := req.Validate()
	if len(errs) > 0 {
		c.AbortWithStatusJSON(
			http.StatusUnprocessableEntity,
			NewValidationErrorResponse("Request is not valid", errs),
		)
		return
	}

	job, err := application.artifactory.CreateBuildJob(c.ClientIP(), req.CommitHash, req.Flags)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusServiceUnavailable,
			NewErrorResponse(fmt.Sprintf("Failed to create build job: %s", err)),
		)
		return
	}

	c.JSON(http.StatusCreated, job)
}

func (application *Application) buildJobStatus(c *gin.Context) {
	decoder := json.NewDecoder(c.Request.Body)
	var req GetBuildStatusRequest
	err := decoder.Decode(&req)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusUnprocessableEntity,
			NewErrorResponse("Failed to decode your request"),
		)
		return
	}

	errs := req.Validate()
	if len(errs) > 0 {
		c.AbortWithStatusJSON(
			http.StatusUnprocessableEntity,
			NewValidationErrorResponse("Request is not valid", errs),
		)
		return
	}

	job, err := application.artifactory.GetBuild(req.CommitHash, req.Flags)
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusServiceUnavailable,
			NewErrorResponse(fmt.Sprintf("Failed to check build job status: %s", err)),
		)
		return
	}

	if job == nil {
		c.AbortWithStatusJSON(
			http.StatusNotFound,
			NewErrorResponse("Failed to find job with requested params"),
		)
		return
	}

	c.JSON(http.StatusOK, job)
}

func (application *Application) Start(port int) error {
	gin.DebugPrintRouteFunc = func(httpMethod, absolutePath, handlerName string, nuHandlers int) {
		entry := log.WithFields(log.Fields{
			"method":      httpMethod,
			"path":        absolutePath,
			"handlerName": handlerName,
			"nuHandlers":  nuHandlers,
		})
		entry.Debugf("endpoint")
	}
	router := gin.New()
	router.Use(ginlogrus.Logger(log.New()))
	router.Use(gin.Recovery())
	router.GET("/", application.root)
	router.GET("/healthz", application.healthz)
	router.POST("/jobs", application.createBuildJob)
	router.POST("/status", application.buildJobStatus)

	server := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", port),
		Handler:           router,
		ReadHeaderTimeout: 30,
	}
	application.server = server
	return application.server.ListenAndServe()
}

func (application *Application) Stop(ctx context.Context) error {
	if application.server == nil {
		return nil
	}
	return application.server.Shutdown(ctx)
}
