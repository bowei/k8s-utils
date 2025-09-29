package sample

// MyStruct is a struct.
//
// This is some text.
//
// - A ListOfStuff
// - another item
type MyStruct struct {
	TypeMeta   string
	ObjectMeta string

	// Content is content.
	//
	// This is some text.
	// This is some text.
	// This is some text.
	// This is some text.
	// This is some text.
	// This is some text.
	// This is some text.
	// This is some text.
	// This is some text.
	// This is some text.
	//
	// - A ListOfStuff
	// - another item
	//
	// Heading:
	//
	// ## Another heading
	//
	// This is some text.
	// This is some text.
	// This is some text.
	// This is some text.
	// This is some text.
	Content              Content
	ListOfStuff          []SubContent
	MapOfStuff           map[string]SubContent
	PointerToStuff       *SubContent
	PointerToListOfStuff *[]SubContent
}

type Content struct {
	NumField   int
	SubContent SubContent
}

type SubContent struct {
	Name string
}
