package option

type LogOption struct {
	OutputPath string
	MaxSizeMB  int
	MaxAgeDay  int
	MaxBackups int
}
