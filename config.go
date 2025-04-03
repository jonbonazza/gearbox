package gearbox

import (
	"time"

	"github.com/jonbonazza/gearbox/http"
)

type Config struct {
	HTTP            http.Config   `yaml:"http"`
	Logging         LogConfig     `yaml:"logging"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}
