/* Handles submission metadata files

A submission metadata file goes through the following lifecycle:
A newly created submission ends up in the incoming directory
where it is picked up by the scheduler.

The status of the submission determines which directory its metadata*/
/* will reside in. Metadata files with status ACCEPTED will reside in*/
/* the incoming directory while submissions with any other status will*/
/* be placed in the submission directory

TODO: Find a folder organization that makes more sense

Appropriate placement of metadata is handled "magically" and
transparently by this module.


*/



package common

import (
  "gopkg.in/yaml.v2"
  //"config"
  "path"
  "os"
  "log"
  "errors"
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

type file_location int
const (
  FILE_NONE file_location = iota
  FILE_INCOMING
  FILE_SUBMISSION
  FILE_BOTH
)

func (f file_location) lookup() string {
  paths := [4]string{
    "",
    ConfigValue("IncomingDir"),
    ConfigValue("SubmissionDir"),
    "",
  }
  return paths[f]
}

func file_ext(name string) string {
  return name + ".metadata"
}

// Implements the file finding logic. If file is not found
// we report its location as being in incoming so that it will
// be found there
// TODO: Reduce WTF??? count
func find_file(id string) file_location {
  // FIXME: We shouldn't duplicate this
  dirs := [2]file_location {
    FILE_INCOMING,
    FILE_SUBMISSION,
  }
  c := FILE_NONE;
  for _, p := range dirs {
    file := path.Join(p.lookup(), file_ext(id))
    if _, err := os.Stat(file); !os.IsNotExist(err) {
      c = c + file_location(p)
    }
  }
  return c
}

// Returns a new Metadata struct with it's ID value set if id
// represents an unknown submission or return a new Metadata struct
// with data from existing metadata file if id refers to a known submission
func Get(id string) (*Metadata, error) {

  m := Metadata{}
  m.Id = id

  loc := find_file(id)

  switch {
  case loc == FILE_BOTH:
    return nil, errors.New("Inconsistent metadata state. Somethingwent badly wrong here")
  case loc == FILE_NONE:
    return &m, nil
  }

  fname := path.Join(loc.lookup(), file_ext(id))

  lock, err := locking.NewFLock(fname)
  if err != nil {
    return nil, err
  }

  lock.Lock()
  defer lock.Unlock()
  log.Print(fname)
  file, err := os.Open(fname)
  if err != nil {
    return nil, err
  }
  defer file.Close()

  // FIXME: What to do if file is too large?
  info, err := os.Stat(fname)
  if info.Size() > metadata_limit {
    return nil, errors.New("Metadata file too large")
  }
  data := make([]byte, info.Size())
  file.Read(data)

  if err := yaml.Unmarshal(data, &m); err != nil {
    return nil, err
  }

  return &m, nil
}

// BUG(truls): Every time this function is called with
// Metadata.Status == STATUS_ACCEPTED the scheduler will treat it as a new
// submission. We probably only want this to happen the first time
func (m *Metadata) Commit() error {
  loc := find_file(m.Id)
  current_file := path.Join(loc.lookup(), file_ext(m.Id))

  if loc == FILE_BOTH {
    return errors.New("Inconsistent metadata state. Something went badly wrong here")
  }

  var file string
  if m.Status == STATUS_ACCEPTED {
    file = path.Join(FILE_INCOMING.lookup(), file_ext(m.Id))
  } else {
    file = path.Join(FILE_SUBMISSION.lookup(), file_ext(m.Id))
  }

  if loc > FILE_NONE {
    // If file is moving location, delete old file
    // BUG(truls) Deletion/creation of file isn't atomic. This may not
    // be a problem.
    // FIXME: Should we acquire the lock here?
    if current_file != file {
      if err := os.Remove(current_file); err != nil {
        return err
      }
    }
    // TODO: Test locking
    lock, err := locking.NewFLock(file)
    if !os.IsNotExist(err) && err != nil{
      return err
    } else {
      lock.Lock()
      defer lock.Unlock()
    }
  }

  data, err := yaml.Marshal(m)
  if err != nil {
    return err
  }
  if err := ioutil.WriteFile(file, data, 0600); err != nil {
    return err
  }
  return nil

}
