package gologger

import (
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"time"

	"github.com/imperiuse/golib/colormap"
	"github.com/imperiuse/golib/concat"
)

// LoggerBean - Bean описывающий настройки логера
type LoggerBean struct {
	Output                               string
	DestFlag, N, CallDepth, SettingFlags int
	Delimiter, ThemeName                 string
}

// LoggerI - Main interface of GoLogger
type LoggerI interface {
	Log(lvl LogLvl, msg ...interface{})
	LogC(lvl LogLvl, cf ColorFlag, msg ...interface{})

	Info(...interface{})
	Debug(...interface{})
	Warning(...interface{})
	Error(...interface{})
	Fatal(...interface{})

	Test(...interface{})
	Print(...interface{})
	P()
	Other(...interface{})

	LoggerController

	Close()
}

// LoggerController - sub interface for LoggerI (control action)
// Interface LoggerI Controller describes base control settings of Logger: color map and I/O mechanism)
// ALL METHOD UNDER MUTEX!
type LoggerController interface {
	SetColorScheme(colormap.CSM)
	SetColorThemeName(string)

	SetDefaultDestinations(io.Writer, DestinationFlag)
	SetNewDestinations(Destinations)

	SetDestinationLvl(LogLvl, []io.Writer)
	GetDestinationLvl(LogLvl) []io.Writer
	SetDestinationLvlColor(LogLvl, ColorFlag, io.Writer)
	GetDestinationLvlColor(LogLvl, ColorFlag) io.Writer

	DisableDestinationLvl(LogLvl)
	DisableDestinationLvlColor(LogLvl, ColorFlag)

	EnableDestinationLvl(LogLvl)
	EnableDestinationLvlColor(LogLvl, ColorFlag)

	SetAndEnableDestinationLvl(LogLvl, []io.Writer)
	SetAndEnableDestinationLvlColor(LogLvl, ColorFlag, io.Writer)
}

// NewLogger - Constructor LoggerI
func NewLogger(defaultOutput io.Writer, destFlag DestinationFlag, n, callDepth, settingFlags int, delimiter string, csm colormap.CSM) LoggerI {

	i := 0
	l := Logger{
		make(chan logMsg, n),
		settingFlags,
		delimiter,
		callDepth,
		0,
		func() int { i++; return i },
		make(Destinations, len(defaultDestinations)),
		make(LogMap, len(defaultDestinations)),
		csm,
	}

	l.SetDefaultDestinations(defaultOutput, destFlag)
	go l.writeChanGoroutine()

	return &l
}

// These flags define which text to prefix to each log entry generated by the LoggerI.
// Bits are or'ed together to control what's printed.
// For example, flags Ldate | Ltime (or LstdFlags) produce,
//	2009/01/23 01:23:23 message
// while flags Ldate | Ltime | Lmicroseconds | Llongfile produce,
//	2009/01/23 01:23:23.123123 /a/b/c/d.go:23: message
const (
	Ldate         = 1 << iota // the date in the local time zone: 2009/01/23
	Ltime                     // the time in the local time zone: 01:23:23
	Lmicroseconds             // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                 // full file name and line number: /a/b/c/d.go:23
	Lshortfile                // final file name element and line number: d.go:23. overrides Llongfile
	Lutc                      // if Ldate or Ltime is set, use UTC rather than the local time zone

	LNoStackTrace // No Run runtime.Caller()  for getting Stack Trace info

	LstdFlags = Ldate | Ltime // initial values for the standard Logger
)

// LogLvl - type of log lvl
type LogLvl int

// const for logger lvl
const (
	Info LogLvl = iota
	Debug
	Warning
	Error
	Fatal
	Test
	Print
	P
	Other
	Db
	Redis
	Memchd
	DbOk
	DbFail
	RedisOk
	RedisFail
	MemchdOk
	MemchdFail
)

// ColorFlag - type of color flag
type ColorFlag int

// Const for Color or NoColor Mode
const (
	NoColor ColorFlag = iota
	Color
)

// DestinationFlag - type of destination flag
type DestinationFlag int

// Const for destination
const (
	OffAll DestinationFlag = iota
	OnNoColor
	OnColor
	OnAll
)

// Destinations - map of log_lvl/color destinations
type Destinations map[LogLvl][]io.Writer

// Default io.Writers Destinations
var defaultDestinations = Destinations{
	Info:      {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Debug:     {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Warning:   {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Error:     {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Fatal:     {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Test:      {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Print:     {NoColor: ioutil.Discard, Color: ioutil.Discard},
	P:         {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Other:     {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Db:        {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Redis:     {NoColor: ioutil.Discard, Color: ioutil.Discard},
	Memchd:    {NoColor: ioutil.Discard, Color: ioutil.Discard},
	DbOk:      {NoColor: ioutil.Discard, Color: ioutil.Discard},
	DbFail:    {NoColor: ioutil.Discard, Color: ioutil.Discard},
	RedisOk:   {NoColor: ioutil.Discard, Color: ioutil.Discard},
	RedisFail: {NoColor: ioutil.Discard, Color: ioutil.Discard},
}

// LoggerColorSchemeDetached - Привязка обозначений цветовой схемы colormap к уровням логирования
var LoggerColorSchemeDetached = map[LogLvl]colormap.CSN{
	Info:       colormap.CS_INFO,
	Debug:      colormap.CS_DEBUG,
	Warning:    colormap.CS_WARNING,
	Error:      colormap.CS_ERROR,
	Fatal:      colormap.CS_FATAL_ERROR,
	Test:       colormap.CS_TEST,
	Print:      colormap.CS_PRINT,
	P:          colormap.CS_PRINT,
	Other:      colormap.CS_PRINT,
	Db:         colormap.CS_DB,
	Redis:      colormap.CS_REDIS,
	Memchd:     colormap.CS_MEMCHD,
	DbOk:       colormap.CS_DB_OK,
	DbFail:     colormap.CS_DB_FAIL,
	RedisOk:    colormap.CS_REDIS_OK,
	RedisFail:  colormap.CS_REDIS_FAIL,
	MemchdOk:   colormap.CS_MEMCHD_OK,
	MemchdFail: colormap.CS_MEMCHD_FAIL,
}

// LogHandler - function of handlers
type LogHandler func(*Logger, ...interface{}) // Func which execute in specific Log method (Info, Debug and etc.)
// LogMap - map of LogHandlers
type LogMap map[LogLvl][]LogHandler

// Logger -Main struct of Logger
type Logger struct {
	msgChan       chan logMsg  // Channel ready to write log msg
	settingsFlags int          // Global formatting Logger settings
	delimiter     string       // Delimiters of columns (msg...)       TODO add param to NewLoggerFunc
	callDepth     int          // CallDepth Ignore value
	width         uint         // width pretty print (columns, msg..)  TODO
	pGen          func() int   // Gen sequences
	Destinations               // TwoD-slice of io.Writers
	LogMap                     // Map []LogHandlers
	colorMap      colormap.CSM // Color Map of LoggerI
}

// logMsg - Log msg struct
type logMsg struct {
	LogLvl    // Log LogLvl msg
	ColorFlag // Features msg:  Color or NoColor msg
	string    // Msg for logging
}

// Generator Discard. Return LogHandler func which Discard log msg (Nothing to do)
func genDiscardFunc() LogHandler {
	return func(l *Logger, msg ...interface{}) {
		return
	}
}

// genLogFunc - Generator LogFunc. Return LogHandler func which print non color log
func genLogFunc(lvl LogLvl) LogHandler {
	return func(l *Logger, msg ...interface{}) {
		l.log(lvl, NoColor, concatInterfaces(l.delimiter, msg...))
	}
}

// genColorLogFunc - Generator ColorLogFunc. Return LogHandler func which print color log msg
func genColorLogFunc(lvl LogLvl, csn colormap.CSN) LogHandler {
	return func(l *Logger, msg ...interface{}) {
		l.log(lvl, Color,
			colorConcatInterfaces(l.colorMap[csn], l.colorMap[colormap.CS_RESET][0], l.delimiter, msg...))
	}
}

// writeChanGoroutine - Goroutine func, Writer log msg by io.Writers info from lvlDestinations twoD-slice
func (l *Logger) writeChanGoroutine() {
	for {
		if msg, ok := <-l.msgChan; ok { // канал закрыт
			_, _ = io.WriteString(l.Destinations[msg.LogLvl][msg.ColorFlag], msg.string) // TODO Maybe optimized this by directly use syscall to write
		} else {
			break
		}
	}
}

// SetColorScheme - set color scheme
func (l *Logger) SetColorScheme(cs colormap.CSM) {
	newCS := make(colormap.CSM, len(cs))
	for i, v := range cs {
		newCS[i] = v
	}
	l.colorMap = newCS
}

// SetColorThemeName - set color theme by name
func (l *Logger) SetColorThemeName(name string) {
	l.SetColorScheme(colormap.CSMthemePicker(name))
}

// GetDefaultDestinations - get default destination
func GetDefaultDestinations() (defaultDest Destinations) {
	defaultDest = make(Destinations, len(defaultDestinations))
	for lvl := range defaultDestinations {
		defaultDest[lvl] = make([]io.Writer, 2)
		copy(defaultDest[lvl], defaultDestinations[lvl])
	}
	return
}

// SetDefaultDestinations - set default destination
func (l *Logger) SetDefaultDestinations(defaultWriter io.Writer, flag DestinationFlag) {
	for lvl := range defaultDestinations {
		switch flag {
		case OffAll:
			l.Destinations[lvl] = []io.Writer{NoColor: ioutil.Discard, Color: ioutil.Discard}
			l.LogMap[lvl] = []LogHandler{genDiscardFunc(), genDiscardFunc()}
		case OnNoColor:
			l.Destinations[lvl] = []io.Writer{NoColor: defaultWriter, Color: ioutil.Discard}
			l.LogMap[lvl] = []LogHandler{genLogFunc(lvl), genDiscardFunc()}
		case OnColor:
			l.Destinations[lvl] = []io.Writer{NoColor: ioutil.Discard, Color: defaultWriter}
			l.LogMap[lvl] = []LogHandler{genDiscardFunc(), genColorLogFunc(lvl, LoggerColorSchemeDetached[lvl])}
		case OnAll:
			l.Destinations[lvl] = []io.Writer{NoColor: defaultWriter, Color: defaultWriter}
			l.LogMap[lvl] = []LogHandler{genLogFunc(lvl), genColorLogFunc(lvl, LoggerColorSchemeDetached[lvl])}
		}
	}
}

// SetDestinationLvl - set destination to lvl
func (l *Logger) SetDestinationLvl(lvl LogLvl, sWriters []io.Writer) {
	l.Destinations[lvl] = sWriters
}

// GetDestinationLvl - get destination by lvl
func (l *Logger) GetDestinationLvl(lvl LogLvl) []io.Writer {
	return l.Destinations[lvl]
}

// SetDestinationLvlColor - set destination to lvl & color
func (l *Logger) SetDestinationLvlColor(lvl LogLvl, color ColorFlag, writer io.Writer) {
	l.Destinations[lvl][color] = writer
}

// GetDestinationLvlColor - get destination  by lvl & color
func (l *Logger) GetDestinationLvlColor(lvl LogLvl, color ColorFlag) io.Writer {
	return l.Destinations[lvl][color]
}

// SetNewDestinations - set new destination
func (l *Logger) SetNewDestinations(destinations Destinations) {
	for lvl := range destinations {
		l.Destinations[lvl] = make([]io.Writer, 2)
		l.LogMap[lvl] = make([]LogHandler, 2)
		for color, writer := range destinations[lvl] {
			l.SetAndEnableDestinationLvlColor(lvl, ColorFlag(color), writer)
		}
	}
}

// DisableDestinationLvl - disable destination lvl
func (l *Logger) DisableDestinationLvl(lvl LogLvl) {
	l.LogMap[lvl][Color] = genDiscardFunc()
	l.LogMap[lvl][NoColor] = genDiscardFunc()
}

// DisableDestinationLvlColor - disable destination lvl color
func (l *Logger) DisableDestinationLvlColor(lvl LogLvl, color ColorFlag) {
	l.LogMap[lvl][color] = genDiscardFunc()
}

// EnableDestinationLvl - enable destination lvl
func (l *Logger) EnableDestinationLvl(lvl LogLvl) {
	l.LogMap[lvl][Color] = genLogFunc(lvl)
	l.LogMap[lvl][NoColor] = genColorLogFunc(lvl, LoggerColorSchemeDetached[lvl]) // TODO CUSTOMIZE LoggerColorSchemeDetached
}

// EnableDestinationLvlColor - enable destination lvl color
func (l *Logger) EnableDestinationLvlColor(lvl LogLvl, color ColorFlag) {
	switch color {
	case NoColor:
		l.LogMap[lvl][color] = genLogFunc(lvl)
	case Color:
		l.LogMap[lvl][color] = genColorLogFunc(lvl, LoggerColorSchemeDetached[lvl]) // TODO CUSTOMIZE LoggerColorSchemeDetached
	}
}

// SetAndEnableDestinationLvl - set and enable destination lvl
func (l *Logger) SetAndEnableDestinationLvl(lvl LogLvl, d []io.Writer) {
	l.SetDestinationLvl(lvl, d)
	l.EnableDestinationLvl(lvl)
}

// SetAndEnableDestinationLvlColor - set and enable destination lvl color
func (l *Logger) SetAndEnableDestinationLvlColor(lvl LogLvl, color ColorFlag, d io.Writer) {
	l.SetDestinationLvlColor(lvl, color, d)
	if d == ioutil.Discard {
		l.DisableDestinationLvlColor(lvl, color)
	} else {
		l.EnableDestinationLvlColor(lvl, color)
	}
}

// Log - func log with custom LogLvl
func (l *Logger) Log(lvl LogLvl, msg ...interface{}) {
	l.LogMap[lvl][NoColor](l, msg...)
	l.LogMap[lvl][Color](l, msg...)
}

// LogC - color func log with custom LogLvl
func (l *Logger) LogC(lvl LogLvl, cf ColorFlag, msg ...interface{}) {
	l.LogMap[lvl][cf](l, msg...)
}

// Info - log with Info lvl
func (l *Logger) Info(msg ...interface{}) {
	l.LogMap[Info][NoColor](l, msg...)
	l.LogMap[Info][Color](l, msg...)
}

// Debug - log with Debug lvl
func (l *Logger) Debug(msg ...interface{}) {
	l.LogMap[Debug][NoColor](l, msg...)
	l.LogMap[Debug][Color](l, msg...)
}

// Warning - log with Warning lvl
func (l *Logger) Warning(msg ...interface{}) {
	l.LogMap[Warning][NoColor](l, msg...)
	l.LogMap[Warning][Color](l, msg...)
}

// Error - log with Error lvl
func (l *Logger) Error(msg ...interface{}) {
	l.LogMap[Error][NoColor](l, msg...)
	l.LogMap[Error][Color](l, msg...)
}

// Fatal - log with Fatal lvl
func (l *Logger) Fatal(msg ...interface{}) {
	l.LogMap[Fatal][NoColor](l, msg...)
	l.LogMap[Fatal][Color](l, msg...)
}

// Test - log with Test lvl
func (l *Logger) Test(msg ...interface{}) {
	l.LogMap[Test][NoColor](l, msg...)
	l.LogMap[Test][Color](l, msg...)
}

// Print - log with Print lvl
func (l *Logger) Print(msg ...interface{}) {
	l.LogMap[Print][NoColor](l, msg...)
	l.LogMap[Print][Color](l, msg...)
}

// P - log with P lvl
func (l *Logger) P() {
	l.LogMap[Info][NoColor](l, fmt.Sprint(l.pGen()))
	l.LogMap[Info][Color](l, fmt.Sprint(l.pGen()))
}

// Other - log with Other lvl
func (l *Logger) Other(msg ...interface{}) {
	l.LogMap[Other][NoColor](l, msg...)
	l.LogMap[Other][Color](l, msg...)
}

func concatInterfaces(delimiter string, msg ...interface{}) (result string) {
	for _, v := range msg {
		result = concat.Strings(result, fmt.Sprintf("%v", v))
		result = concat.Strings(result, delimiter)
	}
	return
}

func colorConcatInterfaces(cs colormap.ColorSheme, reset, delimiter string, msg ...interface{}) (s string) {
	lenCS := len(cs)
	for i, v := range msg {
		if lenCS > 0 {
			s = concat.Strings(s, cs[i])
			lenCS--
		}
		s = concat.Strings(s, fmt.Sprintf("%v", v))
		s = concat.Strings(s, delimiter)
	}
	s = concat.Strings(s, reset)
	return s

}

func (l Logger) log(lvl LogLvl, cf ColorFlag, msg string) {
	s := string(getSystemInfo(l.settingsFlags, l.callDepth+5)) // 5 - magic number - cnt LogLvl to func runtimeCaller()
	s = concat.Strings(s, msg)
	s = concat.Strings(s, "\n")
	l.msgChan <- logMsg{lvl, cf, s}
}

// Close - close inner chan for transfer msg into Logger
func (l *Logger) Close() {
	if len(l.msgChan) == 0 {
		close(l.msgChan)
		return
	}
	t := time.NewTimer(time.Second)
	select {
	case <-t.C:
		close(l.msgChan)
		return
	default:
		if len(l.msgChan) == 0 {
			close(l.msgChan)
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
}

func getSystemInfo(settingsFlags int, callDepth int) (result []byte) {
	now := time.Now()
	file, line := "", 0
	if settingsFlags&LNoStackTrace == 0 {
		file, line = getRuntimeInfo(callDepth)
	}
	formatHeader(settingsFlags, &result, now, file, line)

	return result
}

func getRuntimeInfo(callDepth int) (string, int) {
	var ok bool
	_, file, line, ok := runtime.Caller(callDepth)
	if !ok {
		file = "???"
		line = 0
	}
	return file, line
}

// formatHeader writes log header to buf in following order:
//   * date and/or time (if corresponding flags are provided),
//   * file and line number (if corresponding flags are provided).
// nolint
func formatHeader(settingsFlags int, buf *[]byte, t time.Time, file string, line int) {
	if settingsFlags&(Ldate|Ltime|Lmicroseconds) != 0 {
		if settingsFlags&Lutc != 0 {
			t = t.UTC()
		}
		if settingsFlags&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if settingsFlags&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if settingsFlags&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if settingsFlags&LNoStackTrace == 0 {
		if settingsFlags&(Lshortfile|Llongfile) != 0 {
			if settingsFlags&Lshortfile != 0 {
				short := file
				for i := len(file) - 1; i > 0; i-- {
					if file[i] == '/' {
						short = file[i+1:]
						break
					}
				}
				file = short
			}
			*buf = append(*buf, file...)
			*buf = append(*buf, ':')
			itoa(buf, line, -1)
			*buf = append(*buf, ": "...)
		}
	}
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}
