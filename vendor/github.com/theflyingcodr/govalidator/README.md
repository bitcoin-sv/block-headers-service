# Validator [![Go Reference](https://pkg.go.dev/badge/github.com/theflyingcodr/govalidator.svg)](https://pkg.go.dev/github.com/theflyingcodr/govalidator) [![Go Report Card](https://goreportcard.com/badge/github.com/theflyingcodr/govalidator)](https://goreportcard.com/report/github.com/theflyingcodr/govalidator) [![CI Status](https://github.com/theflyingcodr/govalidator/workflows/Go/badge.svg)](https://github.com/theflyingcodr/govalidator/actions?query=workflow%3AGo)

This is a simple input validator for go programs which can be used to test inputs at the boundary of your application.

It uses a fluent API to allow chaining of validators to keep things concise and terse.

There is also a provided Validator interface which can applied to structs and tested in the like of http handlers or error handlers.

## Error structure

Validation functions are ran per field and multiple functions can be evaluated per field.

Any validation errors found are then stored in a map[string][]string. This can be printed using the .String() method or can be encoded to Json and wrapped in an `errors` object for example.

For an idea of the json output see this below example validation error:

```json
{
    "errors": {
        "count": [
            "value 0 should be greater than 0"
        ],
        "isEnabled": [
            "value true does not evaluate to false"
        ]
    }
}
```

In this example I wanted my errors to be wrapped in an errors object, you may just want them output raw and unwrapped or wrapped in something else.

This is up to you to decide how best to handle the presentation of the error list.

## Usage

There are two main ways of using the library, either via inline checks or by implementing the `validator.Validator` interface.

To get started though, call `validator.New()`.

From this you can then chain validators, the idea is you supply the field name and then a series of one or more validator functions.

### Inline Chaining

Below is an inline method shown, this shows the fluent nature of the API allowing chaining of validators.

Each Validate call is for a single field but multiple validator functions can be added per field as can be seen for the "dob" field.

```go
    func(s *svc) Create(ctx context.Context, req Request) error{
        if err := validator.New().
            Validate("name", validator.Length(req.Name, 4, 10)).
            Validate("dob", validator.NotEmpty(req.DOB), validator.DateBefore(req.DOB, time.Now().AddDate(-16, 0, 0))).
            Validate("isEnabled", validator.Bool(req.IsEnabled, false)).
            Validate("count", validator.PositiveInt(req.Count)).Err(); err != nil {
                return err
        }
    }
```

*Note* - the final call here is the `.Err()` method, this will return nil if no errors are found or error if one or more have been found.

### Struct Validation

The second method to validate is by implementing the validator.Validator interface on a struct.

The interface is very simple:

```go
    type Validator interface {
        Validate() ErrValidation
    }
```

Taking an example from the [examples directory](examples), you can apply to a struct as shown:

```go
    type Request struct {
        Name      string    `json:"name"`
        DOB       time.Time `json:"dob"`
        IsEnabled bool      `json:"isEnabled"`
        Count     int       `json:"count"`
    }
    
    // Validate implements validator.Validator and evaluates Request.
    func (r *Request) Validate() validator.ErrValidation {
        return validator.New().
            Validate("name", validator.Length(r.Name, 4, 10)).
            Validate("dob", validator.NotEmpty(r.DOB), validator.DateBefore(r.DOB, time.Now().AddDate(-16, 0, 0))).
            Validate("isEnabled", validator.Bool(r.IsEnabled, false)).
            Validate("count", validator.PositiveInt(r.Count))
    }
```

This is an ideal usecase for handling common errors in a global error handler, you can simply parse your requests, check, if they implement the interface and evaluate the struct. An Example of this is found in the [examples](examples).

## Examples

There are examples in the [examples directory](examples), you can clone the repo and have a play with these to ensure the validator meets your needs.

## Functions

All functions are currently located in the [functions](functions.go) file.

These must return a validator.ValidationFunc function and can be wrapped to allow custom params to be passed.

Here is an example from the functions.go file:

```go
    func Length(val string, min, max int) ValidationFunc {
        return func() error {
            if len(val) >= min && len(val) <= max {
                return nil
            }
            return fmt.Errorf(validateLength, val, min, max)
        }
    }
```

Pretty simple! You can add your own compatible functions in your code base and call them in the Validate(..,...) methods.

You can also apply one time functions to the Validate calls as shown:

```go
    Validate("name", validator.Length(name, 1, 20), func() error{
        if mything == 0{
            return fmt.Errorf("oh no")
        }
        return nil
    })
```

## Contributing

I've so far added a limited set of validation functions, if you have an idea for some useful functions feel free to open an issue and PR.
