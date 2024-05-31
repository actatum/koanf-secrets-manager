### Installation

```shell
# Install the core.
go get -u github.com/knadh/koanf/v2

# Install the necessary Provider(s).
# Available: aws, gcp
# eg: go get -u github.com/actatum/koanf-secrets-manager/aws
# eg: go get -u github.com/actatum/koanf-secrets-manager/gcp

go get -u github.com/actatum/koanf-secrets-manager/aws


# Install the necessary Parser(s).
# Available: toml, toml/v2, json, yaml, dotenv, hcl, hjson, nestedtext
# go get -u github.com/knadh/koanf/parsers/$parser

go get -u github.com/knadh/koanf/parsers/json
```

### Reading config from AWS Secrets Manager

```go
package main

import (
	"fmt"
	"log"

	"github.com/knadh/koanf/v2"
	"github.com/knadh/koanf/parsers/json"
	"github.com/actatum/koanf-secrets-manager/aws"
)

// Global koanf instance. Use "." as the key path delimiter. This can be "/" or any character.
var k = koanf.New(".")

func main() {
    // Initialize Provider
    provider, err := aws.Provider(aws.Config{
        Region: "us-east-1",
        Secret: "my-top-secret-info",
        Timeout: 5*time.Second,
    })
    if err != nil {
        log.Fatalf("error initializing provider: %v", err)
    }

    // Load JSON config.
    if err := k.Load(provider, json.Parser()); err != nil {
        log.Fatalf("error loading config: %v", err)
    }

    fmt.Println("parent's name is = ", k.String("parent1.name"))
    fmt.Println("parent's ID is = ", k.Int("parent1.id"))
}

```
