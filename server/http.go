package server

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/auth"
	"github.com/edgetx/cloudbuild/processor"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	ginlogrus "github.com/toorop/gin-logrus"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
)

type Application struct {
	artifactory *artifactory.Artifactory
	auth        *auth.AuthTokenDB
	workers     *processor.WorkerDB
}

func New(art *artifactory.Artifactory,
	auth *auth.AuthTokenDB,
	workers *processor.WorkerDB,
) *Application {
	return &Application{
		artifactory: art,
		auth:        auth,
		workers:     workers,
	}
}

func bindQuery(c *gin.Context, query interface{}) error {
	if err := c.ShouldBindQuery(query); err != nil {
		BadRequestResponse(c, err)
		return err
	}
	return nil
}

func bindBuildRequest(c *gin.Context) (*BuildRequest, error) {
	req := &BuildRequest{}
	if err := c.ShouldBindBodyWith(req, binding.JSON); err != nil {
		UnprocessableEntityResponse(c, err.Error())
		return nil, err
	}
	if errs := req.Validate(); len(errs) > 0 {
		c.AbortWithStatusJSON(
			http.StatusUnprocessableEntity,
			NewValidationErrorResponse("Request is not valid", errs),
		)
		return nil, ErrInvalidRequest
	}
	return req, nil
}

func (app *Application) metrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
	})
}

func (app *Application) root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "/",
	})
}

func (app *Application) listBuildJobs(c *gin.Context) {
	var query artifactory.JobQuery
	if bindQuery(c, &query) != nil {
		return
	}

	if err := query.Validate(); err != nil {
		BadRequestResponse(c, err)
		return
	}

	jobs, err := app.artifactory.ListJobs(&query)
	if err != nil {
		ServiceUnavailableResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (app *Application) listWorkers(c *gin.Context) {
	workers, err := app.workers.List()
	if err != nil {
		ServiceUnavailableResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, processor.WorkersDtoFromModels(workers))
}

func (app *Application) createBuildJob(c *gin.Context) {
	req, err := bindBuildRequest(c)
	if err != nil {
		return
	}
	job, err := app.artifactory.CreateBuildJob(c.ClientIP(), req.CommitHash, req.Flags)
	if err != nil {
		ServiceUnavailableResponse(c, err)
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (app *Application) buildJobStatus(c *gin.Context) {
	req, err := bindBuildRequest(c)
	if err != nil {
		return
	}
	job, err := app.artifactory.GetBuild(req.CommitHash, req.Flags)
	if err != nil {
		ServiceUnavailableResponse(c, err)
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

func (app *Application) authenticated(handler gin.HandlerFunc) gin.HandlerFunc {
	return BearerAuth(app.auth, handler)
}

func (app *Application) addAPIRoutes(rg *gin.RouterGroup) {
	// authenticated
	rg.GET("/metrics", app.authenticated(app.metrics))
	rg.GET("/jobs", app.authenticated(app.listBuildJobs))
	rg.GET("/workers", app.authenticated(app.listWorkers))
	// public
	rg.POST("/jobs", app.createBuildJob)
	rg.POST("/status", app.buildJobStatus)
}

func debugRoutes(method, path, _ string, _ int) {
	log.WithFields(log.Fields{
		"method": method,
		"path":   path,
	}).Debugf("endpoint")
}

func (app *Application) Start(listen string) error {
	gin.DebugPrintRouteFunc = debugRoutes
	router := gin.New()
	router.Use(ginlogrus.Logger(log.New()))
	router.Use(gin.Recovery())

	// this should be a config parameter (in case behind CF)
	router.SetTrustedProxies(nil) //nolint:errcheck

	// later this should server static content (dashboard app?)
	router.GET("/", app.root)

	api := router.Group("/api")
	app.addAPIRoutes(api)

	return router.Run(listen)
}
