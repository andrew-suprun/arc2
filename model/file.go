package model

// import "fmt"

// type File interface {
// 	Meta()
// 	State()
// 	Hashed()
// }

// type Folder struct {
// 	Meta
// 	State
// 	Hashed      uint64
// 	TotalHashed uint64
// 	Counts      []int
// }

// type RegilarFile struct {
// 	Meta
// 	State
// 	Hashed      uint64
// 	TotalHashed uint64
// 	Counts      []int
// }

// func (f *File) String() string {
// 	return fmt.Sprintf("File{FileId: %q, Kind: %s, Size: %d, Hash: %q, State: %s, Hashed: %d, TotalHashed; %d}",
// 		f.Id, f.Kind, f.Size, f.Hash, f.State, f.Hashed, f.TotalHashed)
// }

// type Kind int

// const (
// 	FileRegular Kind = iota
// 	FileFolder
// )

// func (k Kind) String() string {
// 	switch k {
// 	case FileFolder:
// 		return "FileFolder"
// 	case FileRegular:
// 		return "FileRegular"
// 	}
// 	return "UNKNOWN FILE KIND"
// }

// type State int

// const (
// 	Resolved State = iota
// 	Scanned
// 	Hashing
// 	Pending
// 	Divergent
// )

// func (s State) String() string {
// 	switch s {
// 	case Scanned:
// 		return "Scanned"
// 	case Hashing:
// 		return "Hashing"
// 	case Resolved:
// 		return "Resolved"
// 	case Pending:
// 		return "Pending"
// 	case Divergent:
// 		return "Divergent"
// 	}
// 	return "UNKNOWN FILE STATE"
// }

// func (s State) Merge(other State) State {
// 	if other > s {
// 		return other
// 	}
// 	return s
// }
