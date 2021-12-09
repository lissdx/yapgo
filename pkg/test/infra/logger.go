package buffer_logger

import (
	"bytes"
	standartlogger "log"
	"os"
)

type BufferLogger struct {
	loggerBuff *standartlogger.Logger
	logger     *standartlogger.Logger
	Out        *bytes.Buffer
	printToIO  bool
}

func NewBufferLogger(printToIO bool) Logger {
	outBuff := new(bytes.Buffer)
	stdLogg := standartlogger.New(os.Stdout, "", 0)

	return &BufferLogger{
		loggerBuff: standartlogger.New(outBuff, "", 0),
		logger:     stdLogg,
		Out:        outBuff,
		printToIO:  printToIO,
	}
}

func (bl *BufferLogger) Debug(args ...interface{}) {
	bl.loggerBuff.Println("DEBUG ", args)
	if bl.printToIO {
		bl.logger.Println("DEBUG ", args)
	}
}

func (bl *BufferLogger) Info(args ...interface{}) {
	bl.loggerBuff.Println("INFO ", args)
	if bl.printToIO {
		bl.logger.Println("INFO ", args)
	}
}

func (bl *BufferLogger) Warn(args ...interface{}) {
	bl.loggerBuff.Println("WARN ", args)
	if bl.printToIO {
		bl.logger.Println("WARN ", args)
	}
}

func (bl *BufferLogger) Error(args ...interface{}) {
	bl.loggerBuff.Println("ERROR ", args)
	if bl.printToIO {
		bl.logger.Println("ERROR ", args)
	}
}

func (bl *BufferLogger) Panic(args ...interface{}) {
	bl.loggerBuff.Println("PANIC ", args)
	if bl.printToIO {
		bl.logger.Println("PANIC ", args)
	}
}

func (bl *BufferLogger) Fatal(args ...interface{}) {
	bl.loggerBuff.Println("FATAL ", args)
	if bl.printToIO {
		bl.logger.Println("FATAL ", args)
	}
}

func (bl *BufferLogger) GetBufferedString() string {
	return bl.Out.String()
}
