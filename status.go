package common

type Status int
const (
  STATUS_ACCEPTED Status = iota
  STATUS_QUEUED
  STATUS_PROCESSING
  STATUS_FINISHED
)

var status_text = [...]string {
  "Accepted",
  "Queued",
  "Processing",
  "Finished",
}

var status_descr = [...]string {
  "Your submission has been accepted and is awaiting scheduling",
  "Your submission is in queue to be processed",
  "Your submission is being processed",
  "Your assessment is ready",
}

func (s Status) String() string {
  return status_text[s]
}

func (s Status) Description() string {
  return status_descr[s]
}

type SubmissionType int
const (
  SUBMISSION_NORMAL SubmissionType = iota
  SUBMISSION_ASSESSMENT
)
