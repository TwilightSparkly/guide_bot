package logging

// Ivan Orshak, 12.07.2023

import (
	"log"
	"os"
	"time"
)

// NewLog creates a logging system that writes to a file,
// the path to which is passed as a parameter.
// Three logging options are implemented:
// useful information output,
// error notification
// and fatal errors notification
func (l *Log) NewLog(fileName string) {
	var err error

	l.logFile, err = os.OpenFile(fileName+time.Now().Format("01-02-2006")+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666) // FIXME Неверная дата в имени файла
	if err != nil {
		log.Println("NewLog(): Unable to open or create file for logs!", err)

		l.LogInfo = log.New(os.Stdout, "INFO:\t", log.Ldate|log.Ltime)
		l.LogErr = log.New(os.Stdout, "ERROR:\t", log.Ldate|log.Ltime)
		l.LogFatal = log.New(os.Stdout, "FATAL ERROR:\t", log.Ldate|log.Ltime)
	} else {
		l.LogInfo = log.New(l.logFile, "INFO:\t", log.Ldate|log.Ltime)
		l.LogErr = log.New(l.logFile, "ERROR:\t", log.Ldate|log.Ltime)
		l.LogFatal = log.New(l.logFile, "FATAL ERROR:\t", log.Ldate|log.Ltime)
	}
}

// Close closes the file for logging
func (l *Log) Close() {
	err := l.logFile.Close()

	if err != nil {
		l.LogErr.Println("Close(): Unable to close log file, error:", err)
	} else {
		l.LogInfo.Println("Close(): Log file has been closed.")
	}
}
