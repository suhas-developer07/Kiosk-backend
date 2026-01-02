package subjects

import "strings"

type Subject string

const (
	Kannada         Subject = "kannada"
	English         Subject = "english"
	Maths           Subject = "maths"
	Chemistry       Subject = "chemistry"
	Physics         Subject = "physics"
	Biology         Subject = "biology"
	ComputerScience Subject = "cs"
)

var ValidSubjects = map[Subject]struct{}{
	Kannada:         {},
	English:         {},
	Maths:           {},
	Chemistry:       {},
	Physics:         {},
	Biology:         {},
	ComputerScience: {},
}

func NormalizeSubject(s string)Subject{
	return  Subject(strings.ToLower(strings.TrimSpace(s)))
}

func IsValidSubject(s string)bool{
	_,ok := ValidSubjects[NormalizeSubject(s)]
	return ok
}