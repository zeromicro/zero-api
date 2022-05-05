package token

import "go/token"

type (
	// Pos is a compact encoding of a source position within a file set.
	// It can be converted into a Position for a more convenient, but much
	// larger, representation.
	//
	// The Pos value for a given file is a number in the range [base, base+size],
	// where base and size are specified when a file is added to the file set.
	// The difference between a Pos value and the corresponding file base
	// corresponds to the byte offset of that position (represented by the Pos value)
	// from the beginning of the file. Thus, the file base offset is the Pos value
	// representing the first byte in the file.
	//
	// To create the Pos value for a specific source offset (measured in bytes),
	// first add the respective file to the current file set using FileSet.AddFile
	// and then call File.Pos(offset) for that file. Given a Pos value p
	// for a specific file set fset, the corresponding Position value is
	// obtained by calling fset.Position(p).
	//
	// Pos values can be compared directly with the usual comparison operators:
	// If two Pos values p and q are in the same file, comparing p and q is
	// equivalent to comparing the respective source file offsets. If p and q
	// are in different files, p < q is true if the file implied by p was added
	// to the respective file set before the file implied by q.
	//
	Pos = token.Pos

	// -----------------------------------------------------------------------------
	// Positions

	// Position describes an arbitrary source position
	// including the file, line, and column location.
	// A Position is valid if the line number is > 0.
	//
	Position = token.Position

	// -----------------------------------------------------------------------------
	// File

	// A File is a handle for a file belonging to a FileSet.
	// A File has a name, size, and line offset table.
	//
	File = token.File

	// -----------------------------------------------------------------------------
	// FileSet

	// A FileSet represents a set of source files.
	// Methods of file sets are synchronized; multiple goroutines
	// may invoke them concurrently.
	//
	// The byte offsets for each file in a file set are mapped into
	// distinct (integer) intervals, one interval [base, base+size]
	// per file. Base represents the first byte in the file, and size
	// is the corresponding file size. A Pos value is a value in such
	// an interval. By determining the interval a Pos value belongs
	// to, the file, its file base, and thus the byte offset (position)
	// the Pos value is representing can be computed.
	//
	// When adding a new file, a file base must be provided. That can
	// be any integer value that is past the end of any interval of any
	// file already in the file set. For convenience, FileSet.Base provides
	// such a value, which is simply the end of the Pos interval of the most
	// recently added file, plus one. Unless there is a need to extend an
	// interval later, using the FileSet.Base should be used as argument
	// for FileSet.AddFile.
	//
	FileSet = token.FileSet
)

// NewFileSet creates a new file set.
func NewFileSet() *FileSet {
	return token.NewFileSet()
}
