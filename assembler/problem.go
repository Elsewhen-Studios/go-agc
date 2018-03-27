package assembler

import "fmt"

type problemKind int

const (
	problemKindError   problemKind = iota
	problemKindWarning problemKind = iota
	problemKindInfo    problemKind = iota
)

type problem struct {
	Kind    problemKind
	File    string
	Line    int
	Message string
}

func (k problemKind) String() string {
	switch k {
	case problemKindError:
		return "Error"
	case problemKindWarning:
		return "Warning"
	case problemKindInfo:
		return "Info"
	default:
		return "Unknown"
	}
}

func (p problem) String() string {
	return fmt.Sprintf("%v: %v (%v, %v)", p.Kind, p.Message, p.File, p.Line)
}

type problemLogger interface {
	LogError(msg string)
	LogErrorf(format string, a ...interface{})
	LogWarning(msg string)
	LogWarningf(format string, a ...interface{})
	LogInfo(msg string)
	LogInfof(format string, a ...interface{})
}

type assemblerLogger struct {
	asm      *assembler
	fileName string
	lineNum  int
}

func (al *assemblerLogger) LogError(msg string) {
	al.asm.problems = append(al.asm.problems, problem{Kind: problemKindError, File: al.fileName, Line: al.lineNum, Message: msg})
	al.asm.errorCount++
}

func (al *assemblerLogger) LogErrorf(format string, a ...interface{}) {
	al.LogError(fmt.Sprintf(format, a...))
}

func (al *assemblerLogger) LogWarning(msg string) {
	al.asm.problems = append(al.asm.problems, problem{Kind: problemKindWarning, File: al.fileName, Line: al.lineNum, Message: msg})
}

func (al *assemblerLogger) LogWarningf(format string, a ...interface{}) {
	al.LogWarning(fmt.Sprintf(format, a...))
}

func (al *assemblerLogger) LogInfo(msg string) {
	al.asm.problems = append(al.asm.problems, problem{Kind: problemKindInfo, File: al.fileName, Line: al.lineNum, Message: msg})
}

func (al *assemblerLogger) LogInfof(format string, a ...interface{}) {
	al.LogInfo(fmt.Sprintf(format, a...))
}
