package shared

type SuccessKey string

func (s SuccessKey) String() string {
	return string(s)
}

var (
	SuccessFileExported SuccessKey = "COMMON.FILE_EXPORTED"
)
