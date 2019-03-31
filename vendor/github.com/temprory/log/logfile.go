package log

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type FileWriter struct {
	sync.Mutex
	RootDir      string
	DirFormat    string
	FileFormat   string
	TimeBegin    int
	TimePrefix   string
	MaxFileSize  int
	SyncInterval time.Duration
	SaveEach     bool
	EnableBufio  bool

	inited       bool
	currdir      string
	currfile     string
	currFileSize int
	currFileIdx  int

	logfile    *os.File
	filewriter *bufio.Writer
	logticker  *time.Ticker
	inittime   time.Duration
}

func (w *FileWriter) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	w.checkFileWithData(p)
	if w.EnableBufio {
		n, err = w.filewriter.Write(p)
	} else {
		n, err = w.logfile.Write(p)
	}
	w.currFileSize += n
	if err != nil {
		log.Printf("logfile Write failed: %v", err)
	}
	if w.SaveEach {
		w.save()
	}
	return n, err
}

func (w *FileWriter) WriteString(str string) (n int, err error) {
	w.Lock()
	defer w.Unlock()
	w.checkFileWithString(str)
	if w.EnableBufio {
		n, err = w.filewriter.WriteString(str)
	} else {
		n, err = w.logfile.WriteString(str)
	}
	w.currFileSize += n
	if err != nil {
		log.Printf("logfile WriteString failed:: %v", err)
	}
	if w.SaveEach {
		w.save()
	}
	return n, err
}

func (w *FileWriter) Save() {
	w.Lock()
	defer w.Unlock()
	w.save()
}

func (w *FileWriter) newFile(path string) error {
	w.logfile = nil
	w.filewriter = nil
	//file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		w.logfile = file
		if w.EnableBufio {
			if w.filewriter == nil {
				w.filewriter = bufio.NewWriter(file)
			} else {
				w.filewriter.Reset(file)
			}
		}
	} else {
		log.Printf("logfile newFile failed: %s, %s", path, err.Error())
	}
	return err
}

func (w *FileWriter) checkFileWithData(data []byte) bool {
	var (
		err      error = nil
		now      time.Time
		filename = ""
		currfile = ""
	)

	if w.TimePrefix != "" {
		now, err = time.Parse(w.TimePrefix, string(data[w.TimeBegin:w.TimeBegin+len(w.TimePrefix)]))
		if err != nil {
			log.Printf("logfile time.Parse(%s) failed: %s", string(data[:len(w.TimePrefix)]), err.Error())
			return false
		}
	} else {
		now = time.Now()
	}

	filename = now.Format(w.FileFormat)

	if !w.inited {
		w.Init(now)
	}

	currdir := w.RootDir
	if w.DirFormat != "" {
		currdir += now.Format(w.DirFormat)
	}
	if w.currdir != currdir {
		w.currdir = currdir
		err = w.makeDir(currdir)
	}

	if w.currFileIdx == 0 {
		currfile = currdir + filename
	} else {
		currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)
	}

	if w.currfile != currfile {
		w.currFileIdx = 0
		w.currFileSize = 0
		w.currfile = currfile

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	} else if w.MaxFileSize > 0 && w.currFileSize+len(data) > w.MaxFileSize {
		w.currFileIdx++
		w.currFileSize = 0
		w.currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	}

	return err == nil
}

func (w *FileWriter) checkFileWithString(str string) bool {
	var (
		err      error = nil
		now      time.Time
		filename = ""
		currfile = ""
	)

	if w.TimePrefix != "" {
		now, err = time.Parse(w.TimePrefix, str[w.TimeBegin:w.TimeBegin+len(w.TimePrefix)])
		if err != nil {
			log.Printf("logfile time.Parse(%s) failed: %s", str[:len(w.TimePrefix)], err.Error())
			return false
		}
	} else {
		now = time.Now()
	}

	filename = now.Format(w.FileFormat)

	if !w.inited {
		w.Init(now)
	}

	currdir := w.RootDir
	if w.DirFormat != "" {
		currdir += now.Format(w.DirFormat) //path.Join(w.RootDir, now.Format(w.DirFormat)) //
	}
	if w.currdir != currdir {
		w.currdir = currdir
		err = w.makeDir(currdir)
	}

	if w.currFileIdx == 0 {
		currfile = currdir + filename
	} else {
		currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)
	}

	if w.currfile != currfile {
		w.currFileIdx = 0
		w.currFileSize = 0
		w.currfile = currfile

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	} else if w.MaxFileSize > 0 && w.currFileSize+len(str) > w.MaxFileSize {
		w.currFileIdx++
		w.currFileSize = 0
		w.currfile = fmt.Sprintf("%s%s.%04d", currdir, filename, w.currFileIdx)

		// w.save()
		if w.logfile != nil {
			w.logfile.Close()
		}

		err = w.newFile(w.currfile)
	}

	return err == nil
}

func (w *FileWriter) makeDir(path string) error {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		log.Printf("logfile makeDir failed: %s, %s", path, err.Error())
	}
	return err
}

func (w *FileWriter) Init(now time.Time) {
	if !w.inited {
		w.inited = true
		currdir := w.RootDir
		if w.DirFormat != "" {
			currdir += now.Format(w.DirFormat)
		}
		if err := w.makeDir(currdir); err != nil {
			log.Printf("logfile init mkdir(%s) failed: %v", w.RootDir, err)
		}
		if !w.SaveEach && w.EnableBufio {
			go func() {
				defer func() {
					recover()
				}()

				w.logticker = time.NewTicker(w.SyncInterval)
				for {
					_, ok := <-w.logticker.C
					w.Save()
					if !ok {
						return
					}
				}
			}()
		}
	}
}

func (w *FileWriter) save() {
	if w.EnableBufio {
		if w.filewriter != nil {
			w.filewriter.Flush()
		}
	} else {
		if w.logfile != nil {
			w.logfile.Sync()
		}
	}
}

// func NewLogFile() *FileWriter {
// 	return &FileWriter{
// 		RootDir: "./logs/",      //日志根目录
// 		DirFormat:  "",             //日志子目录格式化规则，如果拆分子目录，设置成这种格式"20060102/"
// 		FileFormat: "20060102.log", //日志
// 	}
// }
