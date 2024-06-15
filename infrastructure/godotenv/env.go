package godotenv

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Env struct {
	DATABASE_HOST string
	ServerAddr    string
}

func NewEnv() *Env {
	return &Env{}
}

func (e *Env) Load() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err)
	}
	e.DATABASE_HOST = os.Getenv("DATABASE_HOST")
	e.ServerAddr = os.Getenv("ServerAddr")
}
