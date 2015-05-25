package common

import (
  "gopkg.in/yaml.v2"
  //"config"
  "path"
  "os"
  "io/ioutil"
  "github.com/tgulacsi/go-locking"
)

const metadata_limit = 1024*1024

type Metadata struct {
  Id string
  Status Status
  Timestamp int
  Course string
  Assignment string
  User string
}

func (m *Metadata) file_ext(name string) string {
  return name + ".metadata"
}

// TODO: Check if a file already exists when we create it

func (m *Metadata) Get(id string) error {
  dir := ConfigValue("SubmissionDir")
  fname := path.Join(dir, m.file_ext(id))

  lock, err := locking.NewFLock(fname)
  if err != nil {
    return err
  }

  lock.Lock()
  defer lock.Unlock()
  file, err := os.Open(fname)
  if err != nil {
    return err
  }
  defer file.Close()

  // FIXME: Better handling of variable size metadata files
  data := make([]byte, metadata_limit)
  file.Read(data)
  yaml.Unmarshal(data, m)

  return nil

}

func (m *Metadata) Commit() error {
    dir := ConfigValue("SubmissionDir")
  fname := path.Join(dir, m.file_ext(m.Id))

  // TODO: Test locking
  lock, err := locking.NewFLock(fname)
  if !os.IsNotExist(err) && err != nil{
    return err
  } else {
    lock.Lock()
    defer lock.Unlock()
  }

  data, err := yaml.Marshal(m)
  if err != nil {
    return err
  }
  if err := ioutil.WriteFile(fname, data, 0600); err != nil {
    return err
  }
  return nil

}
