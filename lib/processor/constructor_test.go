package processor

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Jeffail/benthos/v3/lib/cache"
	"github.com/Jeffail/benthos/v3/lib/log"
	"github.com/Jeffail/benthos/v3/lib/metrics"
	"github.com/Jeffail/benthos/v3/lib/ratelimit"
	"github.com/Jeffail/benthos/v3/lib/types"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v3"
)

func TestConstructorDescription(t *testing.T) {
	if len(Descriptions()) == 0 {
		t.Error("package descriptions were empty")
	}
}

func TestExamples(t *testing.T) {
	for typeName, ctor := range Constructors {
		for _, example := range ctor.Examples {
			s := struct {
				Pipeline struct {
					Processors []Config `yaml:"processors"`
				} `yaml:"pipeline"`
				Resources struct {
					Caches     map[string]cache.Config     `yaml:"caches"`
					Ratelimits map[string]ratelimit.Config `yaml:"rate_limits"`
				}
			}{}
			dec := yaml.NewDecoder(bytes.NewReader([]byte(example.Config)))
			dec.KnownFields(true)
			require.NoError(t, dec.Decode(&s), "%v:%v", typeName, example.Title)

			type confAlias Config
			sAliased := struct {
				Pipeline struct {
					Processors []confAlias `yaml:"processors"`
				} `yaml:"pipeline"`
				Resources struct {
					Caches     map[string]cache.Config     `yaml:"caches"`
					Ratelimits map[string]ratelimit.Config `yaml:"rate_limits"`
				}
			}{}
			dec = yaml.NewDecoder(bytes.NewReader([]byte(example.Config)))
			dec.KnownFields(true)
			require.NoError(t, dec.Decode(&sAliased), "%v:%v", typeName, example.Title)
		}
	}
}

func TestConstructorBadType(t *testing.T) {
	conf := NewConfig()
	conf.Type = "not_exist"

	logConfig := log.NewConfig()
	logConfig.LogLevel = "NONE"

	if _, err := New(conf, nil, log.New(os.Stdout, logConfig), metrics.DudType{}); err == nil {
		t.Error("Expected error, received nil for invalid type")
	}
}

func TestConstructorBlockType(t *testing.T) {
	Constructors["footype"] = TypeSpec{
		constructor: func(
			conf Config,
			mgr types.Manager,
			log log.Modular,
			stats metrics.Type,
		) (Type, error) {
			return nil, nil
		},
	}

	conf := NewConfig()
	conf.Type = "footype"

	Block("footype", "because testing")

	_, err := New(conf, nil, log.Noop(), metrics.Noop())
	if err == nil {
		t.Fatal("Expected error, received nil for blocked type")
	}
	if !strings.Contains(err.Error(), "because testing") {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestConstructorConfigYAMLInference(t *testing.T) {
	conf := []Config{}

	if err := yaml.Unmarshal([]byte(`[
		{
			"text": {
				"value": "foo"
			},
			"jmespath": {
				"query": "foo"
			}
		}
	]`), &conf); err == nil {
		t.Error("Expected error from multi candidates")
	}

	if err := yaml.Unmarshal([]byte(`[
		{
			"text": {
				"value": "foo"
			}
		}
	]`), &conf); err != nil {
		t.Error(err)
	}

	if exp, act := 1, len(conf); exp != act {
		t.Errorf("Wrong number of config parts: %v != %v", act, exp)
		return
	}
	if exp, act := TypeText, conf[0].Type; exp != act {
		t.Errorf("Wrong inferred type: %v != %v", act, exp)
	}
	if exp, act := "trim_space", conf[0].Text.Operator; exp != act {
		t.Errorf("Wrong default operator: %v != %v", act, exp)
	}
	if exp, act := "foo", conf[0].Text.Value; exp != act {
		t.Errorf("Wrong value: %v != %v", act, exp)
	}
}

func TestConstructorConfigDefaultsYAML(t *testing.T) {
	conf := []Config{}

	if err := yaml.Unmarshal([]byte(`[
		{
			"type": "bounds_check",
			"bounds_check": {
				"max_part_size": 50
			}
		}
	]`), &conf); err != nil {
		t.Error(err)
	}

	if exp, act := 1, len(conf); exp != act {
		t.Errorf("Wrong number of config parts: %v != %v", act, exp)
		return
	}
	if exp, act := 100, conf[0].BoundsCheck.MaxParts; exp != act {
		t.Errorf("Wrong default parts: %v != %v", act, exp)
	}
	if exp, act := 50, conf[0].BoundsCheck.MaxPartSize; exp != act {
		t.Errorf("Wrong overridden part size: %v != %v", act, exp)
	}
}

func TestSanitise(t *testing.T) {
	var actObj interface{}
	var act []byte
	var err error

	exp := `{` +
		`"type":"archive",` +
		`"archive":{` +
		`"format":"binary",` +
		`"path":"nope"` +
		`}` +
		`}`

	conf := NewConfig()
	conf.Type = "archive"
	conf.Archive.Path = "nope"

	if actObj, err = SanitiseConfig(conf); err != nil {
		t.Fatal(err)
	}
	if act, err = json.Marshal(actObj); err != nil {
		t.Fatal(err)
	}
	if string(act) != exp {
		t.Errorf("Wrong sanitised output: %s != %v", act, exp)
	}
}
