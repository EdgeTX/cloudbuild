package server

import (
	"errors"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/auth"
	"github.com/edgetx/cloudbuild/processor"
	"github.com/edgetx/cloudbuild/targets"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ginlogrus "github.com/toorop/gin-logrus"
)

const (
	staticContentDir = "./static"
	defaultFile      = "./static/index.html"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrNotFound       = errors.New("object not found")
)

type Application struct {
	artifactory *artifactory.Artifactory
	auth        *auth.AuthTokenDB
	workers     *processor.WorkerDB
	promReg     *prometheus.Registry
}

func New(art *artifactory.Artifactory,
	auth *auth.AuthTokenDB,
	workers *processor.WorkerDB,
) *Application {
	r := RegisterMetrics()
	go art.RunMetrics(
		metricBuildRequestQueued,
		metricBuildRequestBuilding,
		metricBuildRequestFailed,
	)
	return &Application{
		artifactory: art,
		auth:        auth,
		workers:     workers,
		promReg:     r,
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
	req := artifactory.NewBuildRequest()
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

func (app *Application) metricsHandler() gin.HandlerFunc {
	h := promhttp.HandlerFor(
		app.promReg,
		promhttp.HandlerOpts{},
	)
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

func (app *Application) getBuildJobLogs(c *gin.Context) {
	jobID := c.Param("id")
	if jobID == "" {
		BadRequestResponse(c, ErrInvalidRequest)
		return
	}
	logs, err := app.artifactory.GetLogs(jobID)
	if err != nil {
		ServiceUnavailableResponse(c, err)
		return
	}
	if logs == nil {
		c.AbortWithStatusJSON(
			http.StatusNotFound,
			NewErrorResponse("Failed to find job"),
		)
		return
	}
	c.JSON(http.StatusOK, logs)
}

func (app *Application) listWorkers(c *gin.Context) {
	workers, err := app.workers.List()
	if err != nil {
		ServiceUnavailableResponse(c, err)
		return
	}
	c.JSON(http.StatusOK, processor.WorkersDtoFromModels(workers))
}

func (app *Application) writeTargets(c *gin.Context) {
	var targetsDef targets.TargetsDef
	if err := c.BindJSON(&targetsDef); err != nil {
		UnprocessableEntityResponse(c, err.Error())
		return
	}
	targets.SetTargets(&targetsDef)
	c.JSON(http.StatusOK, gin.H{"message": "ok"})
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

func (app *Application) getTargets(c *gin.Context) {
	c.JSON(http.StatusOK, targets.GetTargets())
}

func (app *Application) authenticated(handler gin.HandlerFunc) gin.HandlerFunc {
	return BearerAuth(app.auth, handler)
}

func (app *Application) addAPIRoutes(rg *gin.RouterGroup) {
	// authenticated endpoints
	rg.GET("/jobs", app.authenticated(app.listBuildJobs))
	rg.DELETE("/job/:id", app.authenticated(app.deleteBuildJob))
	rg.GET("/logs/:id", app.authenticated(app.getBuildJobLogs))
	rg.GET("/workers", app.authenticated(app.listWorkers))
	rg.PUT("/targets", app.authenticated(app.writeTargets))
	// public
	rg.POST("/jobs", app.createBuildJob)
	rg.POST("/status", app.buildJobStatus)
	rg.GET("/targets", app.getTargets)
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
	router.Use(static.ServeRoot("/", staticContentDir))
	router.GET("/metrics", app.metricsHandler())

	api := router.Group("/api")
	api.Use(GinMetrics)
	api.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
	})
	app.addAPIRoutes(api)

	// catch-all route to serve the UI
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.AbortWithStatusJSON(
				http.StatusNotFound,
				NewErrorResponse("route not found"),
			)
		} else {
			c.File(defaultFile)
		}
	})

	return router.Run(listen)
}
