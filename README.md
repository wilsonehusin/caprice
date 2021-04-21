# Caprice

Caprice allows you to have better visibility in your time-intensive tests by emitting [CloudEvents](https://cloudevents.io/) and building a statistical model as well as visualizing the distribution of time spent for sections in a test.

## Usage

### Basic

With `CAPRICE_HOST=https://caprice.corp` set, the following snippet will emit CloudEvents to the specified host.

```go
import (
  "testing"
  
  "xargs.dev/caprice/scribe"
)

func TestMyApp(t *testing.T) {
  s := scribe.New("TestMyApp")
  prepareTest()
  s.Begin()
  if err := runTest(); err != nil {
    s.Fail()
  }
  s.Success()
}
```

Upon execution, `STDOUT` should have a report of the following:

```json
{
  "caprice": {
    "host": "https://caprice.corp",
    "scribe": "uuid-uuid-uuid-uuid",
    "view": "https://caprice.corp/view/uuid-uuid-uuid-uuid"
  }
}
```

Throughout test execution, users can see in the linked URL above to follow the progress of testing.

### Advanced

The powerful part comes in when stages are incorporated.

```go
import (
  "testing"
  
  "xargs.dev/caprice/scribe"
)

func TestMyApp(t *testing.T) {
  s := caprice.Scribe("TestMyApp")
  
  s.Stage("prepare", t, prepareFunc)
  
  ts := s.StartStage("things")
  s.Stage("do thing1", t, func() error {
    return thing1("some", "param")
  })
  s.Stage("do thing2", t, func() error {
    return thing2("some", "other", "param")
  })
  ts.Done()
  
  cleanUpPotentiallySlowInfrastructure()
  s.Success()
}
```

This allows Caprice to understand that there are stages and sub-stages.
If the visualization were transformed to ASCII output, it will roughly look like this:

```
00.000s  ┌ TestMyApp
00.000s  │ ┌ prepare
00.002s  │ └ prepare [DONE 0.002s]
00.002s  │ ┌ things
00.003s  │ │ ┌ do thing1
02.497s  │ │ └ do thing1 [DONE 02.494s]
02.498s  │ │ ┌ do thing2
20.012s  │ │ └ do thing2 [DONE 17.514s]
20.013s  │ └ things [DONE 20.011]
40.130s  └ TestMyApp [DONE 40.130s]
```

### Identifiers

There are valid reasons to add more metadata into a given execution.
This can be setup through `init()` or `CAPRICE_METADATA=`,  such as:

```go
func init() {
  scribe.AddMetadata("release", "1.21")  # accepts key-value pair
}
```

```
CAPRICE_METADATA='{"mode"="local"}'
```

## Trivia

The name _Caprice_ was taken from [24 Caprices](https://en.wikipedia.org/wiki/24_Caprices_for_Solo_Violin_(Paganini)) by Niccolò Paganini.
The musical piece was an inspiration to [many other composers](https://en.wikipedia.org/wiki/Niccol%C3%B2_Paganini#Compositions).

Just like the musical piece, this project brings value through repetition of test execution.
