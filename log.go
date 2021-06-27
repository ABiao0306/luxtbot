package luxtbot

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
)

const (
	DefaultLevel       = logrus.WarnLevel
	DefaultFileNameFmt = "./logs/%Y-%m-%d.log"
	DefaultMaxFiles    = uint(15)
	DefaultRotateDura  = time.Hour * 12
)

type LogConf struct {
	Level    string `yaml:"level"`
	MaxFiles uint   `yaml:"max-files"`
}

func InitLogConf() {
	var (
		level logrus.Level
		err   error
	)

	formatter := &Formatter{
		TimestampFormat: "2006/01/02-15:04:05",
		LineFormat:      "[%lvl%] %fn%-%fln% | %time%: %msg% --- ",
	}
	logConf := &Conf.LogConf
	if logConf.MaxFiles <= 0 {
		fmt.Println("Log Config max files is Worong. Use default config: ", DefaultMaxFiles)
		logConf.MaxFiles = DefaultMaxFiles
	}
	writers := io.MultiWriter(os.Stdout, rotateWriter(DefaultFileNameFmt, logConf.MaxFiles, DefaultRotateDura))
	logrus.SetOutput(writers)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(formatter)
	level, err = logrus.ParseLevel(Conf.LogConf.Level)
	if err != nil {
		logrus.Warnln("Log Config level is Wrong, use default log level: WARN")
		level = DefaultLevel
	} else {
		logrus.SetLevel(level)
		logrus.Infoln("Use Log Level: ", Conf.LogConf.Level)
	}
}

/**
 * @description: logrus formatter
 *   %msg% - messsage, %lvl% - level, %time% - time, %fn% - filename, %fln% - file line number
 */
type Formatter struct {
	TimestampFormat string
	LineFormat      string
}

func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	output := f.LineFormat

	timestampFormat := f.TimestampFormat

	output = strings.Replace(output, "%time%", entry.Time.Format(timestampFormat), 1)

	output = strings.Replace(output, "%msg%", entry.Message, 1)

	level := strings.ToUpper(entry.Level.String())
	output = strings.Replace(output, "%lvl%", level, 1)

	fullName := entry.Caller.File
	fileName := fullName[strings.LastIndex(fullName, "/")+1:]
	output = strings.Replace(output, "%fn%", fileName, 1)

	output = strings.Replace(output, "%fln%", strconv.Itoa(entry.Caller.Line), 1)

	for k, val := range entry.Data {
		switch v := val.(type) {
		case string:
			output = strings.Replace(output, "%"+k+"%", v, 1)
		case int:
			s := strconv.Itoa(v)
			output = strings.Replace(output, "%"+k+"%", s, 1)
		case bool:
			s := strconv.FormatBool(v)
			output = strings.Replace(output, "%"+k+"%", s, 1)
		}
	}

	var fb bytes.Buffer
	fb.WriteString(output)
	fb.WriteString("[ ")

	for k, v := range entry.Data {
		fb.WriteString(k)
		fb.WriteString("=")
		fb.WriteString(fmt.Sprintf("%v", v))
		fb.WriteByte(' ')
	}
	fb.WriteString("]\n")

	return fb.Bytes(), nil
}

func rotateWriter(logNameFmt string, save uint, rotateTime time.Duration) *rotatelogs.RotateLogs {
	logier, err := rotatelogs.New(
		logNameFmt,
		rotatelogs.WithRotationCount(save),      // 文件最大保存份数
		rotatelogs.WithRotationTime(rotateTime), // 日志切割时间间隔
	)
	if err != nil {
		panic(err)
	}
	return logier
}
