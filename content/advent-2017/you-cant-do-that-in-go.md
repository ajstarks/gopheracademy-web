+++
author = ["David Crawshaw"]
title = "Method Closures: You Can't Do That In Go"
date = 2017-12-26T00:00:01Z
series = ["Advent 2017"]
+++

[Neugram](https://neugram.io) is a scripting language that sticks very
close to Go.
Go statements are Neugram statements, you can import Go packages,
scripts can be compiled to Go programs, and types look just like the
equivalent Go types at run time (which means packages built on
reflection, like `fmt`, work as expected).
These requirements put a lot of restrictions on the design of Neugram.
This post is about one such restriction on methods that I did not
discover until I tried to use it without thinking.

# Background: Go without declarations

When designing a language for use in a REPL (a read-eval-print-loop
like your shell), you want to be able to dive right in and have code
executed as quickly as possible.
That is, a scripting language should be able to say
`"Hello, World!"` in one reasonable line.

Popular scripting languages like Perl and Python use statements
as the topmost grammatical construction. A simple statement can
consist of a single expression, like the command to print a string.

Go is different. The topmost grammatical construction in Go is a
[declaration](https://golang.org/ref/spec#Declarations_and_scope).
Declarations consist of package-wide constants, variables, functions,
types and methods.
Inside declarations are statements.
The [statement](https://golang.org/ref/spec#Statements) is
in charge of program control flow, and contain some number of
expressions.
An [expression](https://golang.org/ref/spec#Expressions) is an
actual computation, where we do the work of programming.

The concept of having a layer of declarations above statements
is common in programming languages.
Both C and Java have declarations.
Declarations are useful.
The order of top-level declarations in Go does not affect the order
of execution of the program.
This makes it possible to depend on names defined later in the file
(or in an entirely different file in the package) without developing
a system of forward declarations or header files.

One of the key changes that makes Neugram a different language from
Go is we do not have top-level declarations.
Neugram starts with statements.
We lose the advantages of declarations in exchange for
executing statements quickly

Without declarations, packages are restricted to a single file
(to avoid thinking about order of file execution) and referring to
names not yet defined is tricky, but the feel of many programs stays
the same because in Go most declarations also work as statements.
For example:

```
var V = 4
type T int
```

# Method grammar

The one top-level declaration that we miss in Neugram is the
[method declaration](https://golang.org/ref/spec#Method_declarations).
In Go you declare a method by writing:

```
func (t T) String() string {
	return fmt.Sprintf("%d", int(t))
}
```

Critically, this declaration does not stand on its own.
You need another declaration somewhere in your package defining the
type T.
While type declarations can be made as statements, method declarations
cannot.
There are several possible arguments for why not, but given the current
syntax one is that it would introduce the notion of incomplete types
to the run time phase of Go programs. Imagine:

```
func main() {
	type T int
	var t interface{} = T{}

	_, isReader := t.(io.Reader)
	fmt.Println(isReader) // prints false

	if rand {
		func (t T) Read([]byte) (int, error) {
			return 0, io.EOF
		}
	}

	_, isReader = t.(io.Reader)
	fmt.Println(isReader) // prints ... what?
}
```

Method declarations in Go break the complete definition of a type out
over many top-level declarations.
This works in Go because there is no concept of time for declarations,
they all happen simultaneously before a program is run.
This won't work in Neugram where all declarations have to be made
inside statements that happen during program execution.

# Methodik

To resolve this, Neugram introduces a new keyword to define types
with all of its methods in a single statement, `methodik`.

```
methodik T int {
	func (t) Read([]byte) (int, error) {
		return 0, io.EOF
	}
}
```

This statement is evaluated in one step.
The type T does not exist beforehand, and after the statement is
evaluated it exists with all of its methods.

So far so good.

# Method closures: You can't do that in Go

While testing out method declarations, I attempted to reimplement
io.LimitReader. The version I came up with didn't work:

```
func limit(r io.Reader, n int) io.Reader {
	methodik lr struct{} {
		func (*l) Read(p []byte) (int, error) {
			if n <= 0 {
				return 0, io.EOF
			}
			if len(p) > n {
				p = p[:n]
			}
			rn, err := r.Read(p)
			n -= rn
			return rn, err
		}
	}
	return &lr{}
}
```

Why not? Using the values r and n in a closure is normal Go
programming, but this is something unusual:
I am trying to construct a method closure.

An implication of methods only being definable by top-level
declaration in Go is that there is no closure equivalent form.
There is also no way (presently,
[issue #16522](https://golang.org/issues/16522) may make it possible)
to create a method using reflection which would allow closing over
variables.

This is not a particularly problematic limitation, we can move the
free variables of the closure explicitly into the type being defined
to get the same effect:

```
func limit(r io.Reader, n int) io.Reader {
	methodik lr struct{
		R io.Reader
		N int
	} {
		func (*l) Read(p []byte) (int, error) {
			if l.N <= 0 {
				return 0, io.EOF
			}
			if len(p) > n {
				p = p[:l.N]
			}
			rn, err := l.R.Read(p)
			l.N -= rn
			return rn, err
		}
	}
	return &lr{r, n}
}
```

Avoiding method closures also avoids some reflection surprises:
two different lr types, defined as closing over different values,
would probably have to be different types. That means run time
creation of new types without the use of the reflect package, which
is a category of possibilities I'm glad I don't have to imagine.

The restriction itself however could be confusing for someone new
to Neugram who doesn't know about the limits of Go underlying it.
In particular, consider the interaction with global variables.
It is fine for a method defined in Go to refer to globals, and
so too in Neugram:

```
var x = "hello" // a global

methodik obj struct{} {
	func (o) String() string {
		return x // this is fine
	}
}
```

However if we take this code and try to indent it into a block, the
type checker will now have to produce an error, because `x` is no
longer a global variable.
This is unfortunate. In Go there is a clear distinction between
global (defined by top-level declarations) and non-global variables
(defined by statements).
In Neugram they look similar, so this is one more thing the programmer
has to track themselves.

# Surprising expressivity

Accidentally introducing syntax for method closures is a good example
of the kind of problem I have spent a lot of time trying to avoid
in Neugram.
Even the smallest changes to Go result in unexpected ways to write
programs.
I did not find this particular problem until months after creating
the `methodik` syntax.
