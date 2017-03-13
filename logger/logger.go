package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var level int
var flag_log int
var cout io.Writer
var mutex sync.Mutex

const (
	LOG_LEVEL_DISABLE = -1
	LOG_LEVEL_INFO    = 0
	LOG_LEVEL_WARN    = 1
	LOG_LEVEL_DEBUG   = 2
	LOG_LEVEL_ERROR   = 3
	LOG_LEVEL_FATAL   = 4
)

const (
	LOG_NO_FILE    = 0x0
	LOG_LONG_FILE  = 0x1
	LOG_SHORT_FILE = 0x2
	LOG_TIME       = 0x4
)

func SetLogLevel(l int) {
	level = l
}

func SetFlags(flag int) {

	flag_log = flag
}

func SetOutput(w io.Writer) {
	mutex.Lock()
	defer mutex.Unlock()
	cout = w
}

func LOGI(v ...interface{}) {
	if level <= LOG_LEVEL_INFO {
		getLogAppend(LOG_LEVEL_INFO, v)
	}
}

func LOGW(v ...interface{}) {
	if level <= LOG_LEVEL_WARN {
		getLogAppend(LOG_LEVEL_WARN, v)
	}
}

func LOGD(v ...interface{}) {
	if level <= LOG_LEVEL_DEBUG {
		getLogAppend(LOG_LEVEL_DEBUG, v)
	}
}

func LOGE(v ...interface{}) {
	if level <= LOG_LEVEL_ERROR {
		getLogAppend(LOG_LEVEL_ERROR, v)
	}
}

func LOGF(v ...interface{}) {
	if level <= LOG_LEVEL_FATAL {
		getLogAppend(LOG_LEVEL_FATAL, v)
		os.Exit(1)
	}
}

func getLogAppend(lvl int, v ...interface{}) (str string) {
	str = ""
	flag := 0
	//time
	flag = (flag_log & 0x4)
	if flag == 0x4 {
		t := time.Now()
		str += t.Format("[2006/01/02 15:04:05] ")

	}
	//lvl
	switch lvl {
	case LOG_LEVEL_INFO:
		str += "[INFO] "
	case LOG_LEVEL_WARN:
		str += "[WARN] "
	case LOG_LEVEL_DEBUG:
		str += "[DEBUG] "
	case LOG_LEVEL_ERROR:
		str += "[ERROR] "
	case LOG_LEVEL_FATAL:
		str += "[FATAL] "
	}
	//location
	flag = (flag_log & 0x3)
	if LOG_SHORT_FILE == flag || LOG_LONG_FILE == flag {
		_, file, line, ok := runtime.Caller(2)
		if false == ok {
			str += "???:0 "
		} else {
			if LOG_LONG_FILE == flag {
				str = file + " " + strconv.Itoa(line)
			} else if LOG_SHORT_FILE == flag {
				short := file
				for i := len(file) - 1; i > 0; i-- {
					if file[i] == '/' {
						short = file[i+1:]
						break
					}
				}
				str += short + ":" + strconv.Itoa(line) + " "
			}
		}
	}
	strbrackets := fmt.Sprint(v)
	if len(strbrackets) > 0 {
		strbrackets = strings.TrimLeft(strbrackets, "[")
		strbrackets = strings.TrimRight(strbrackets, "]")
		str += strbrackets
	}
	str += "\r\n"
	fmt.Print(str)
	mutex.Lock()
	if cout != nil {
		cout.Write([]byte(str))
	}
	mutex.Unlock()
	return
}
