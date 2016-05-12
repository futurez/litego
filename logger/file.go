package logger

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileLogAdapter struct {
}

func (this *FileLogAdapter) newLoggerInstance() LoggerInterface {
	return &FileLogWriter{
		config: FileLogConfig{
			LogFlag:  (log.Ldate | log.Ltime | log.Lmicroseconds),
			FileName: "log",
			MaxSize:  1 << 30, //1024MB(1G)
			MaxDays:  7},
	}
}

type FileLogConfig struct {
	LogFlag  int    `json:"logflag"`
	FileName string `json:"filename"`
	MaxSize  int    `json:"maxsize"`
	MaxDays  int    `json:"maxdays"`
	LogLevel int    `json:"loglevel"`
}

type FileLogWriter struct {
	lg         *log.Logger
	fd         *os.File
	openDate   int
	curFileNum int
	curSize    int
	config     FileLogConfig
}

// Init file logger with json config.
// jsonconfig like:
//	{
//  "logflag"  :log.Ldate|log.Ltime|log.Lmicroseconds
//	"filename":"test.log",
//	"maxsize" :1<<30,
//	"maxdays" :7,
//	}
func (this *FileLogWriter) Init(jsonconfig string) error {
	err := json.Unmarshal([]byte(jsonconfig), &this.config)
	if err != nil {
		return err
	}
	if len(this.config.FileName) == 0 {
		return errors.New("fileconfig must have filename")
	}
	this.lg = log.New(this, "", this.config.LogFlag)
	return this.createLogFile()
}

func (fw *FileLogWriter) SetLogLevel(loglevel int) {
	fw.config.LogLevel = loglevel
}

func (fw *FileLogWriter) Write(b []byte) (int, error) {
	fw.curSize += len(b)
	return fw.fd.Write(b)
}

func (this *FileLogWriter) setFd(fd *os.File) error {
	finfo, err := fd.Stat()
	if err != nil {
		return err
	}

	if this.fd != nil {
		this.fd.Close()
	}
	this.curSize = int(finfo.Size())
	this.fd = fd
	return nil
}

func (this *FileLogWriter) docheck() {
	if (this.config.MaxSize > 0 && this.curSize >= this.config.MaxSize) ||
		(time.Now().Day() != this.openDate) {
		if err := this.backupFile(); err != nil {
			fmt.Fprintf(os.Stderr, "FileLogWriter(%q): %s\n", this.config.FileName, err)
			return
		}
	}
}

// write logger message into file.
func (fw *FileLogWriter) WriteMsg(msg string, level int) error {
	if fw.fd == nil || level < fw.config.LogLevel {
		return nil
	}
	fw.lg.Println(msg)
	fw.docheck()
	return nil
}

func (this *FileLogWriter) createLogFile() error {
	offset := strings.LastIndex(this.config.FileName, "/")
	if offset > 0 {
		pathname := this.config.FileName[0:offset]
		err := os.MkdirAll(pathname, 0660)
		if err != nil {
			log.Printf("%s\n", err.Error())
			return err
		}
	}
	fd, err := os.OpenFile(this.config.FileName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("%s\n", err.Error())
		return err
	}
	this.openDate = time.Now().Day()
	return this.setFd(fd)
}

func (this *FileLogWriter) backupFile() error {
	_, err := os.Lstat(this.config.FileName)
	if err == nil { // file exists
		// Find the next available number
		fname := ""
		if time.Now().Day() != this.openDate {
			this.curFileNum = 0
		}

		for err == nil {
			fname = this.config.FileName + fmt.Sprintf(".%s.%03d", time.Now().Format("2006-01-02"), this.curFileNum)
			_, err = os.Lstat(fname)
			this.curFileNum++
		}

		this.fd.Close()

		err = os.Rename(this.config.FileName, fname)
		if err != nil {
			return fmt.Errorf("backupFile: %s\n", err)
		}

		// re-create logger
		err := this.createLogFile()
		if err != nil {
			return fmt.Errorf("backupFile StartLogger: %s\n", err)
		}

		go this.deleteOldLog()
	}
	return nil
}

func (this *FileLogWriter) deleteOldLog() {
	dir := filepath.Dir(this.config.FileName)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				returnErr = fmt.Errorf("Unable to delete old log '%s', error: %+v", path, r)
				fmt.Println(returnErr)
			}
		}()

		if !info.IsDir() && info.ModTime().Second() <= (time.Now().Second()-60*60*24*int(this.config.MaxDays)) {
			if strings.HasPrefix(filepath.Base(path), filepath.Base(this.config.FileName)) {
				os.Remove(path)
			}
		}
		return
	})
}

func (this *FileLogWriter) Close() {
	this.fd.Sync()
	this.fd.Close()
}

func init() {
	Register(FILE_PROTOCOL, &FileLogAdapter{})
}
