# lathos [![Go Reference](https://pkg.go.dev/badge/github.com/theflyingcodr/lathos.svg)](https://pkg.go.dev/github.com/theflyingcodr/lathos) [![example workflow](https://github.com/theflyingcodr/lathos/actions/workflows/go.yml/badge.svg)](https://github.com/theflyingcodr/lathos/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/theflyingcodr/lathos)](https://goreportcard.com/report/github.com/theflyingcodr/lathos) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/theflyingcodr/lathos?style=flat-square)

Greek for error, lathos is an errors package for go.

It utilises checking errors for behaviour, not type or substring. This helps to make error checking elegant and removes dependency on your code from concrete error types.

This library was heavily influenced by a [blog from Dave Cheney (Don’t just check errors, handle them gracefully)](https://tpow.app/2a736d9f) and is my take on the paradigms discussed there.

## Examples

A quick example error check:

```go
err := Do()
if lathos.IsNotFound(err){
	// do something else
}
```

This is much neater than Sentinel error checking such as:

```go
err := Do()
if strings.Contains(err.Error(), "not found"){
	// do something else
}
```

This is brittle code, the error could change its message, it could change the casing, meaning you also need to add a `strings.ToLower` wrapper to handle that and you have to remember the exact text throughout your code to check for specific errors; we all make spelling mistakes...

Lathos is also neater than type checks, type checks are ok, but they tie you to a concrete error implementation:

```go
err := Do()
if ok := err.(lathos.NotFound); ok {
	// do something else
}
```

This reads ok (in my opinion not as nice as a lathos check though), but if you want to change the NotFound type, you need to update this throughout your code base where you check the errors. You may want to implement your own version of the NotFound error for example.

## Usage

Lathos is mostly made up of interfaces that when implemented on an error type give it a particular behaviour, these can be found in the [lathos.go](lathos.go) file.

There are two main error types:

1) client - errors to be returned to a client, these would generally be equivalent to 4XX http errors.
2) internal - server related errors where you will want to record a stack trace, metadata and send it to a logging system

Errors can then derive from these, for example, you could create a NotFound error that embeds a client error, therefor it is both a client error and a notfound error - this will be useful in an error handler where you may want to branch between client and internal errors.

### Error Types

The library also has implementations of each error behaviour for your convenience, you can use these in your code base or implement your own error types.

As long as your errors implement the relevant interface, and you use the lathos.Is{ErrorType} methods to check any error implementing the interface will return true in the checks.

## Error Handlers

The idea with the library is that it will be used in a service of some kind, you will usually just return errors and let them bubble up.

Occasionally you will expect a particular error such as a Duplicate. At this point, return a Duplicate error.

If you then create a global error handler, you can check the errors in one place, convert to a response of your choosing and return. Or you may log them.

There are some examples in the [examples](examples) folder.

## Compatibility

As this uses features introduced in Go1.13 relating to errors and error checks it will only work in projects using Go 1.13 and above.

It can still be used with the excellent [pkg/errors](https://tpow.app/f8efe08c) package as from version 0.9.0 they added support for the Go1.13 error types.

## Contributions

If you have any suggestions or improvements feel free to add an Issue or create a PR and I'll be very grateful!

