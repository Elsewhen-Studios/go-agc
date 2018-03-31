package assembler

import (
	"fmt"
)

//ProblemKind is an enumeration for defining the severity of a Problem.
type ProblemKind int

//Problem severity includes Error, Warning, and Info.
const (
	ProblemKindError ProblemKind = iota
	ProblemKindWarning
	ProblemKindInfo
)

//Problem contains debugging information for issues encountered during assembly.
type Problem struct {
	Kind    ProblemKind
	File    string
	Line    int
	Message string
}

func (k ProblemKind) String() string {
	switch k {
	case ProblemKindError:
		return "Error"
	case ProblemKindWarning:
		return "Warning"
	case ProblemKindInfo:
		return "Info"
	default:
		return "Unknown"
	}
}

func (p Problem) String() string {
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
	asm      *Assembler
	fileName string
	lineNum  int
}

func (al *assemblerLogger) LogError(msg string) {
	al.asm.Problems = append(al.asm.Problems, Problem{Kind: ProblemKindError, File: al.fileName, Line: al.lineNum, Message: msg})
	al.asm.errorCount++
}

func (al *assemblerLogger) LogErrorf(format string, a ...interface{}) {
	al.LogError(fmt.Sprintf(format, a...))
}

func (al *assemblerLogger) LogWarning(msg string) {
	al.asm.Problems = append(al.asm.Problems, Problem{Kind: ProblemKindWarning, File: al.fileName, Line: al.lineNum, Message: msg})
}

func (al *assemblerLogger) LogWarningf(format string, a ...interface{}) {
	al.LogWarning(fmt.Sprintf(format, a...))
}

func (al *assemblerLogger) LogInfo(msg string) {
	al.asm.Problems = append(al.asm.Problems, Problem{Kind: ProblemKindInfo, File: al.fileName, Line: al.lineNum, Message: msg})
}

func (al *assemblerLogger) LogInfof(format string, a ...interface{}) {
	al.LogInfo(fmt.Sprintf(format, a...))
}
