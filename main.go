package main

import (

	. "lifestyle/somewhere"
	. "lifestyle/fav"

	"flag"
	"os"
	"path"
	"runtime"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

func main() {
	// log
	log.AddHook(ContextHook{})
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: time.RFC1123})
	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)

	// config
	flag.String("stage", "localhost", "set working environment")
	port := flag.String("p", "8040", "Port number")
	flag.Parse()
	conf := &Configs{Stage: ParseStage(os.Getenv("STAGE"))}
	conf.InitViper()

	// error messages
	EM.New()

	// load location
	time.LoadLocation("Asia/Bangkok")

	//Map data
	Alphabet = InitData()

	// redis
	// rdPool := CreateRedisPool(conf)
	// rdStore := CreateRedisStore(rdPool, conf)

	// mongo
	CreateMongoPool(conf)

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.Use(RequestMiddleware())

	// cors.Default() setup the middleware with default options being
	// all origins accepted with simple methods (GET, POST). See
	// documentation below for more options.
	corsHandler := cors.Default().Handler(NewRouter())
	n.UseHandler(corsHandler)
	n.Run(":" + *port)
}

type ContextHook struct{}

func (hook ContextHook) Levels() []log.Level {
	return []log.Level{log.ErrorLevel}
}

func (hook ContextHook) Fire(entry *log.Entry) error {
	if pc, file, line, ok := runtime.Caller(8); ok {
		funcName := runtime.FuncForPC(pc).Name()

		entry.Data["file"] = path.Base(file)
		entry.Data["func"] = path.Base(funcName)
		entry.Data["line"] = line
	}
	return nil
}
