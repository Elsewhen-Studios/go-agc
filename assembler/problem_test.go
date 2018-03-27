package assembler

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_problemKind_String_error(t *testing.T) {
	// arrange

	// act
	s := fmt.Sprint(problemKindError)

	// assert
	assert.True(t, strings.EqualFold("ERROR", s))
}

func Test_problemKind_String_warning(t *testing.T) {
	// arrange

	// act
	s := fmt.Sprint(problemKindWarning)

	// assert
	assert.True(t, strings.EqualFold("WARNING", s))
}

func Test_problemKind_String_info(t *testing.T) {
	// arrange

	// act
	s := fmt.Sprint(problemKindInfo)

	// assert
	assert.True(t, strings.EqualFold("INFO", s))
}

func Test_problemKind_String_unknown(t *testing.T) {
	// arrange

	// act
	s := fmt.Sprint(problemKindInfo + 1)

	// assert
	assert.True(t, strings.EqualFold("UNKNOWN", s))
}

func Test_problem_String(t *testing.T) {
	// arrange
	k := problemKindError
	f := "file.asm"
	l := 123
	m := "FOO BAR"
	p := problem{Kind: k, File: f, Line: l, Message: m}

	// act
	s := fmt.Sprint(p)

	// assert
	assert.True(t, strings.Contains(s, fmt.Sprint(k)), "problem kind")
	assert.True(t, strings.Contains(s, f), "file name")
	assert.True(t, strings.Contains(s, fmt.Sprint(l)), "line number")
	assert.True(t, strings.Contains(s, m), "message")
}

func Test_assemblerLogger_LogError(t *testing.T) {
	// arrange
	a := new(assembler)
	startCount := a.errorCount
	al := &assemblerLogger{asm: a, fileName: "file123.asm", lineNum: 456}
	m := "Message 789"

	// act
	al.LogError(m)

	// assert
	assert.Len(t, a.problems, startCount+1, "problem count")
	assert.EqualValues(t, startCount+1, a.errorCount, "errors count")

	assert.EqualValues(t, problemKindError, a.problems[startCount].Kind, "problem type")
	assert.EqualValues(t, al.fileName, a.problems[startCount].File, "line number")
	assert.EqualValues(t, al.lineNum, a.problems[startCount].Line, "line number")
	assert.EqualValues(t, m, a.problems[startCount].Message, "messsge")
}

func Test_assemblerLogger_LogErrorf(t *testing.T) {
	// arrange
	a := new(assembler)
	startCount := a.errorCount
	al := &assemblerLogger{asm: a, fileName: "file123.asm", lineNum: 456}
	f := "Format: %v, %#o"
	arg1 := "Foo"
	arg2 := 0123456

	// act
	al.LogErrorf(f, arg1, arg2)

	// assert
	assert.Len(t, a.problems, startCount+1, "problem count")
	assert.EqualValues(t, startCount+1, a.errorCount, "errors count")

	assert.EqualValues(t, problemKindError, a.problems[startCount].Kind, "problem type")
	assert.EqualValues(t, al.fileName, a.problems[startCount].File, "line number")
	assert.EqualValues(t, al.lineNum, a.problems[startCount].Line, "line number")
	assert.EqualValues(t, fmt.Sprintf(f, arg1, arg2), a.problems[startCount].Message, "messsge")
}

func Test_assemblerLogger_LogWarning(t *testing.T) {
	// arrange
	a := new(assembler)
	startCount := a.errorCount
	al := &assemblerLogger{asm: a, fileName: "file231.asm", lineNum: 564}
	m := "Message 897"

	// act
	al.LogWarning(m)

	// assert
	assert.Len(t, a.problems, startCount+1, "problem count")
	assert.EqualValues(t, startCount, a.errorCount, "errors count")

	assert.EqualValues(t, problemKindWarning, a.problems[startCount].Kind, "problem type")
	assert.EqualValues(t, al.fileName, a.problems[startCount].File, "line number")
	assert.EqualValues(t, al.lineNum, a.problems[startCount].Line, "line number")
	assert.EqualValues(t, m, a.problems[startCount].Message, "messsge")
}

func Test_assemblerLogger_LogWarningf(t *testing.T) {
	// arrange
	a := new(assembler)
	startCount := a.errorCount
	al := &assemblerLogger{asm: a, fileName: "file231.asm", lineNum: 564}
	f := "Format: %v, %#o"
	arg1 := "Bar"
	arg2 := 0345612

	// act
	al.LogWarningf(f, arg1, arg2)

	// assert
	assert.Len(t, a.problems, startCount+1, "problem count")
	assert.EqualValues(t, startCount, a.errorCount, "errors count")

	assert.EqualValues(t, problemKindWarning, a.problems[startCount].Kind, "problem type")
	assert.EqualValues(t, al.fileName, a.problems[startCount].File, "line number")
	assert.EqualValues(t, al.lineNum, a.problems[startCount].Line, "line number")
	assert.EqualValues(t, fmt.Sprintf(f, arg1, arg2), a.problems[startCount].Message, "messsge")
}

func Test_assemblerLogger_LogInfo(t *testing.T) {
	// arrange
	a := new(assembler)
	startCount := a.errorCount
	al := &assemblerLogger{asm: a, fileName: "file312.asm", lineNum: 645}
	m := "Message 978"

	// act
	al.LogInfo(m)

	// assert
	assert.Len(t, a.problems, startCount+1, "problem count")
	assert.EqualValues(t, startCount, a.errorCount, "errors count")

	assert.EqualValues(t, problemKindInfo, a.problems[startCount].Kind, "problem type")
	assert.EqualValues(t, al.fileName, a.problems[startCount].File, "line number")
	assert.EqualValues(t, al.lineNum, a.problems[startCount].Line, "line number")
	assert.EqualValues(t, m, a.problems[startCount].Message, "messsge")
}

func Test_assemblerLogger_LogInfof(t *testing.T) {
	// arrange
	a := new(assembler)
	startCount := a.errorCount
	al := &assemblerLogger{asm: a, fileName: "file312.asm", lineNum: 645}
	f := "Format: %v, %#o"
	arg1 := "Baz"
	arg2 := 0561234

	// act
	al.LogInfof(f, arg1, arg2)

	// assert
	assert.Len(t, a.problems, startCount+1, "problem count")
	assert.EqualValues(t, startCount, a.errorCount, "errors count")

	assert.EqualValues(t, problemKindInfo, a.problems[startCount].Kind, "problem type")
	assert.EqualValues(t, al.fileName, a.problems[startCount].File, "line number")
	assert.EqualValues(t, al.lineNum, a.problems[startCount].Line, "line number")
	assert.EqualValues(t, fmt.Sprintf(f, arg1, arg2), a.problems[startCount].Message, "messsge")
}
