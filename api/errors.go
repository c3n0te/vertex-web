package api

type InvalidPassword struct{}

func (m *InvalidPassword) Error() string {
	return "Passwords do not match"
}
