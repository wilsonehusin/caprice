# Caprice

Caprice allows you to have better visibility in your time-intensive code by emitting [CloudEvents](https://cloudevents.io/) and building a statistical model as well as visualizing the distribution of time spent.

## Usage

### Basic

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

Throughout execution, users can see in the linked URL above to follow the progress of execution. `CAPRICE_HOST=` determines the output values above.

#### Generic

```go
import (
  "github.com/wilsonehusin/caprice/scribe"
)

func thing1() {}

func thing2() error {}

func RunSomething() {
  s := scribe.New("RunSomething")
  s.Run("thing1", thing1)
  if err := s.RunErr("thing2", thing2); err != nil {
    // err here is just passed from the returned value of thing2
    panic(err)
  }
  s.Done(nil)
}
```

#### Testing

```go
import (
  "testing"

  "github.com/wilsonehusin/caprice/scribe"
)

func runStuff() error {}

func TestMyApp(t *testing.T) {
  s := scribe.NewT(t, "TestMyApp")

  s.RunT("runStuff", runStuff) // internally calls t.Fatal() on error

  err := runMoreStuff()
  s.Done(err) // if err != nil, internally call t.Fatal()
}
```

### Advanced

The powerful part comes in when stages are incorporated.

```go
import (
  "testing"

  "github.com/wilsonehusin/caprice/scribe"
)

func TestMyApp(t *testing.T) {
  s := scribe.NewT(t, "TestMyApp")

  s.Run("prepare", prepareFunc)

  thingsDone := s.NewStage("things")
  s.RunT("do thing1", func() error {
    return thing1("some", "param")
  })
  s.RunT("do thing2", func() error {
    return thing2("some", "other", "param")
  })
  thingsDone()

  s.Run("cleanup", cleanUpPotentiallySlowInfrastructure)
  s.Done(nil)
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
20.013s  │ └ things [DONE 20.011s]
40.130s  └ TestMyApp [DONE 40.130s]
```

### Identifiers

[TODO] There are valid reasons to add more metadata into a given execution.
This can be setup through `init()` or `CAPRICE_METADATA=`,  such as:

```go
func init() {
  scribe.AddMetadata("release", "1.21")  // accepts key-value pair
}
```

```
CAPRICE_METADATA='{"mode"="local"}'
```

## Trivia

The name _Caprice_ was taken from [24 Caprices](https://en.wikipedia.org/wiki/24_Caprices_for_Solo_Violin_(Paganini)) by Niccolò Paganini.
The musical piece was an inspiration to [many other composers](https://en.wikipedia.org/wiki/Niccol%C3%B2_Paganini#Compositions).

Just like the musical piece, this project brings significantly more value through repetition of execution.
