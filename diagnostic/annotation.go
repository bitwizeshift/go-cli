package diagnostic

// Annotation is an optional annotation that can be connected to [Diagnostic]
// objects.
type Annotation interface {
	annotate(*Diagnostic)
}

type annotationFunc func(*Diagnostic)

func (f annotationFunc) annotate(a *Diagnostic) {
	f(a)
}

// Title returns an annotation that sets the title of the diagnostic message.
func Title(title string) Annotation {
	return annotationFunc(func(a *Diagnostic) {
		a.Title = title
	})
}

// Line returns an annotation that sets the line of the diagnostic message.
func Line(line int) Annotation {
	return annotationFunc(func(a *Diagnostic) {
		if a.Start == nil {
			a.Start = &Position{}
		}
		a.Start.Line = line
	})
}

// LineRange returns an annotation that sets the line range of the diagnostic
// message.
func LineRange(start, end int) Annotation {
	return annotationFunc(func(a *Diagnostic) {
		if a.Start == nil {
			a.Start = &Position{}
		}
		if a.End == nil {
			a.End = &Position{}
		}
		a.Start.Line = start
		a.End.Line = end
	})
}

// Column returns an annotation that sets the column of the diagnostic message.
func Column(column int) Annotation {
	return annotationFunc(func(a *Diagnostic) {
		if a.Start == nil {
			a.Start = &Position{}
		}
		a.Start.Column = column
	})
}

// ColumnRange returns an annotation that sets the column range of the diagnostic
// message.
func ColumnRange(start, end int) Annotation {
	return annotationFunc(func(a *Diagnostic) {
		if a.Start == nil {
			a.Start = &Position{}
		}
		if a.End == nil {
			a.End = &Position{}
		}
		a.Start.Column = start
		a.End.Column = end
	})
}

// File returns an annotation that sets the file of the diagnostic message.
func File(file string) Annotation {
	return annotationFunc(func(a *Diagnostic) {
		a.File = file
	})
}

// Start returns an annotation that sets the start position of the diagnostic
// message.
func Start(line, column int) Annotation {
	return annotationFunc(func(a *Diagnostic) {
		if a.Start == nil {
			a.Start = &Position{}
		}
		a.Start.Line = line
		a.Start.Column = column
	})
}

// End returns an annotation that sets the end position of the diagnostic
// message.
func End(line, column int) Annotation {
	return annotationFunc(func(a *Diagnostic) {
		if a.End == nil {
			a.End = &Position{}
		}
		a.End.Line = line
		a.End.Column = column
	})
}

// Code returns an annotation that sets the code of the diagnostic message.
func Code(code string) Annotation {
	return annotationFunc(func(a *Diagnostic) {
		a.Code = code
	})
}
