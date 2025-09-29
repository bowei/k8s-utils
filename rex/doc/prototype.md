# Prototype

Create a static webpage that has a dynamic multi-column display similar to the
Mac OS Finder but for data types (structs) and their fields. This is also
similar to how Smalltalk UIs allow the user to browser the objects.

For the prototype, we want to take a graph structure like this:


{
  typeName: "MyStruct",

  fields: [
    {
       fieldName: "Content",
       typeName: "Content",
    },
    ...
  ]
}

{
  typeName: "Content",

  fields: [
    {
      fieldName: "NumField",
      typeName: "int",
    },
    {
      fieldName: "SubContent"
      typeName: "SubContent",
    },
    ...
  ]
}

{
  typeName: "SubContent",

  fields: [
    ...
  ]
}

You will then be able to start with MyStruct, browse each field in MyStruct to
open up a column for Content next to the column for MyStruct by clicking on the
field in the column. This will work recursively for fields in Content,
SubContent etc.

For the prototype, create a static demo with 5 levels of nesting. This should
be a single HTML file with associated JS support functions to enable the
dynamic column browsing.
