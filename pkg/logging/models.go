package logging

// Ivan Orshak, 12.07.2023

import (
	"log"
	"os"
)

type Log struct {
	logFile                   *os.File
	LogInfo, LogErr, LogFatal *log.Logger
}
