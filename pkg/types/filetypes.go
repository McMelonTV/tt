package types

type FileType uint

const (
	FileTypePersistent FileType = iota
	FileTypeTemporary
)
