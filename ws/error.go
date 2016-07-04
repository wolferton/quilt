package ws

type ServiceErrorCategory int

const (
	Unexpected = iota
	Client
	Logic
	Security
)

type CategorisedError struct {
	Category ServiceErrorCategory
	Label    string
	Message  string
}

type ServiceErrors struct {
	Errors     []CategorisedError
	HttpStatus int
}

func (se *ServiceErrors) AddError(category ServiceErrorCategory, label string, message string) {

	error := CategorisedError{category, label, message}

	se.Errors = append(se.Errors, error)

}

func (se *ServiceErrors) AddPredefinedError(error CategorisedError) {
	se.Errors = append(se.Errors, error)
}

func (se *ServiceErrors) HasErrors() bool {
	return len(se.Errors) != 0
}
