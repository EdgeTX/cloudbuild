package server

import (
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/auth"
	"github.com/edgetx/cloudbuild/processor"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ginlogrus "github.com/toorop/gin-logrus"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrNotFound       = errors.New("object not found")
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
	RegisterMetrics()
	go art.RunMetrics(
		metricBuildRequestQueued,
		metricBuildRequestBuilding,
		metricBuildRequestFailed,
	)
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

func bindBuildRequest(c *gin.Context) (*artifactory.BuildRequest, error) {
	req := &artifactory.BuildRequest{}
	if err := c.ShouldBindBodyWith(req, binding.JSON); err != nil {
		UnprocessableEntityResponse(c, err.Error())
		return nil, err
	}
	if err := req.Validate(); err != nil {
		UnprocessableEntityResponse(c, err.Error())
		return nil, err
	}
	return req, nil
}

func metricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
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

func (app *Application) deleteBuildJob(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		BadRequestResponse(c, ErrInvalidRequest)
		return
	}
	err := app.artifactory.DeleteJob(jobID)
	if err != nil {
		ServiceUnavailableResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
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

	job, err := app.artifactory.CreateBuildJob(c.ClientIP(), req)
	if err != nil {
		ServiceUnavailableResponse(c, err)
		return
	}

	metricBuildRequestTotal.WithLabelValues(
		req.Release,
		req.Target,
	).Inc()

	c.JSON(http.StatusCreated, job)
}

func (app *Application) buildJobStatus(c *gin.Context) {
	req, err := bindBuildRequest(c)
	if err != nil {
		return
	}
	job, err := app.artifactory.GetBuild(req)
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
	// authenticated endpoints
	rg.GET("/jobs", app.authenticated(app.listBuildJobs))
	rg.DELETE("/job/:id", app.authenticated(app.deleteBuildJob))
	rg.GET("/workers", app.authenticated(app.listWorkers))
	// public
	rg.StaticFile("/targets", "./targets.json")
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
	router.Use(static.ServeRoot("/", "./static"))
	router.GET("/metrics", metricsHandler())

	api := router.Group("/api")
	api.Use(GinMetrics)
	app.addAPIRoutes(api)

	return router.Run(listen)
}
