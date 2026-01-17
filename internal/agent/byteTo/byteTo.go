package byteto

func Gb(b uint64) uint64 {
	if b == 0 {
		return 0
	}
	return b / 1024 / 1024 / 1024
}
func Mb(b uint64) uint64 {
	if b == 0 {
		return 0
	}
	return b / 1024 / 1024
}
