# salmonflake

A concurrency-safe Snowflake ID generator for Go.

## Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/sdual/salmonflake"
	"github.com/sdual/salmonflake/config"
)

func main() {
	generator := salmonflake.New(config.Config{MachineID: "1"})
	id, err := generator.NextID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)
}
```

## ID Layout

IDs are returned as `uint64` values, with the most significant bit always set to 0.

| Field | Bits | Description |
| --- | --- | --- |
| Timestamp | 41 | Milliseconds since the epoch (approximately 69.7 years) |
| Machine ID | 10 | 0–1023 |
| Sequence | 12 | 0–4095 within the same millisecond |

`config.Config` settings:

- `Start`: The shared epoch, with millisecond precision. Defaults to 2014-09-01 00:00:00 UTC when omitted.
- `MachineID`: A decimal string from `"0"` to `"1023"`. Must be specified explicitly.

Generators in the same ID space must use the same epoch, and each active generator must have a distinct machine ID. Goroutines within a process can share a single generator. Do not copy a generator after first use.

When the sequence is exhausted, the generator waits for the next millisecond. `NextID` returns an error if the clock moves backwards or the timestamp exceeds its range. `New` panics on invalid configuration.

Generator state is stored only in memory. When reusing a machine ID, resume generation only after the clock has advanced past the timestamp of the last previously generated ID. Clock rollback across restarts is not detected.

## Tests

```sh
go test -race ./...
```
