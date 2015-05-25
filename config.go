package common

import (
  "reflect"
  "strconv"
  "code.google.com/p/gcfg"
)

var ConfigCh chan string

type Config struct {
  Default struct {
    Basedir string
    SubmissionDir string
    AssessmentTimeout int
    AssessmentPort string
  }
}

func (c *Config) Parse(file string) error {

  ConfigCh = make(chan string)

  err := gcfg.ReadFileInto(c, file);

  //fmt.Printf("%+v\n", c);
  if err != nil {
    return err
  }

 return nil

}

func (c *Config) Serve() {
  go func() {
    for {
      field := <- ConfigCh
      // FIXME: This is really bad because it circumvents the
      // type safety governing direct access to the struct and
      // introduces a potential for runtime-errors.
      // Do something else later!
      r := reflect.ValueOf(c.Default)
      f := reflect.Indirect(r).FieldByName(field)
      ConfigCh <- f.String()
    }
  }()
}

// FIXME: How do we handle non-existing config strings?
func ConfigValue(v string) string {
  ConfigCh <- v
  r := <- ConfigCh
  if r == "<invalid value>" {
    panic("No such config item")
  }
  return r
}

func ConfigIntValue(v string) int {
  ConfigCh <- v
  r := <- ConfigCh
  if r == "<invalid value>" {
    panic("No such config item")
  }
  i, err := strconv.ParseInt(r, 10, 0)
  if err != nil {
    panic("Something wrong with the intval")
  }
  return int(i)
}
