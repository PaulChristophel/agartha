package routes

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/PaulChristophel/agartha/server/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	Build       = "" // Set at compile time with -ldflags -X github.com/PaulChristophel/agartha/server/routes.Build=b6902e2846042d9d0c2ce27ca290ef54cc013ba8"
	Version     = "" // Set at compile time with -ldflags -X github.com/PaulChristophel/agartha/server/routes.Version=3006.8.0+b6902e2"
	CommitDate  = ""
	CompileDate = ""
	GoVersion   = ""
)

type VersionInfo struct {
	Build       string     `mapstructure:"build" json:"build" example:"d07d28656cb05812e1f36cb6a476278589d1046a"`
	Version     string     `mapstructure:"version" json:"version" example:"v0.5.7-20240718195141-d07d28656cb0"`
	CompileDate *time.Time `mapstructure:"compile_time" json:"compile_time" example:"2024-07-18T20:14:50Z"`
	CommitDate  *time.Time `mapstructure:"commit_time" json:"commit_time" example:"2024-07-18T19:51:41Z"`
	GoVersion   string     `mapstructure:"go_version" json:"go_version" example:"go1.22.5"`
	Platform    string     `mapstructure:"platform" json:"platform" example:"linux/amd64"`
	Path        string     `mapstructure:"path" json:"path" example:"github.com/PaulChristophel/agartha"`
}

func ReadVersionInfo() (*VersionInfo, error) {
	log := logger.GetLogger()
	compileDate, err := time.Parse(time.RFC3339, CompileDate)
	if err != nil {
		log.Error("Invalid date format", zap.Error(err))
		compileDate = time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	}

	commitDate, err := time.Parse("20060102150405", CommitDate)
	if err != nil {
		log.Error("Invalid date format", zap.Error(err))
		commitDate = time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC)
	}

	info, parsed := debug.ReadBuildInfo()
	if !parsed {
		err = fmt.Errorf("error retrieving BuildInfo")
		log.Error("Failed to retrieve BuildInfo", zap.Error(err))
	}

	parts := strings.Split(GoVersion, " ")

	vers := VersionInfo{
		Path:        info.Path,
		GoVersion:   info.GoVersion,
		Platform:    parts[len(parts)-1],
		Version:     Version,
		Build:       Build,
		CommitDate:  &commitDate,
		CompileDate: &compileDate,
	}

	return &vers, nil
}

func PrintVersion() {
	log, err := logger.InitLogger(gin.ReleaseMode)
	if err != nil {
		log = logger.GetLogger()
	}
	info, err := ReadVersionInfo()
	if err != nil {
		log.Error("Failed to read version info", zap.Error(err))
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Path: %s\n", info.Path)
	fmt.Printf("GoVersion: %s\n", info.GoVersion)
	fmt.Printf("Platform: %s\n", info.Platform)
	fmt.Printf("Version: %s\n", info.Version)
	fmt.Printf("Build: %s\n", info.Build)
	fmt.Printf("Compile Time: %s\n", info.CompileDate.String())
	fmt.Printf("Commit Time: %s\n", info.CommitDate.String())
}

func AddVersionRoutes(rg *gin.RouterGroup) {
	grp := rg.Group("/version")
	grp.GET("", GetVersion)
}

// GetVersion godoc
//
//	@Summary	Get server version
//	@Schemes
//	@Description	Gets the version of the server.
//	@Tags			Version
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	VersionInfo
//	@Failure		500	{object}	httputil.HTTPError500
//	@Router			/version [get]
func GetVersion(c *gin.Context) {
	log := logger.GetLogger()
	vers, err := ReadVersionInfo()
	if err != nil {
		log.Error("Failed to retrieve version info", zap.Error(err))
		httputil.NewError(c, http.StatusInternalServerError, "No version data present.")
		return
	}
	log.Info("Returning version info", zap.Any("version_info", vers))
	c.JSON(http.StatusOK, vers)
}
