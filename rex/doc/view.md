# View

Header should have two lines:

```text
<package path>
<TypeName>
```

The individual items in the column should have two lines:


```text
<Field Name> <Type Name>
<package name>

<First sentence of the docstring>
```

## Docstring expansion

Each docstring that is truncated (e.g. has more than one sentence) should have a
"[+]" symbol at the end that when clicked will expand out to show the rest of
the docstring. The "..." symbol should change to "(close)" which can then be
used to collapse the expanded docstring.
