package nullabledb
// Thanks to Supid Raval
// https://gist.github.com/rsudip90/45fad7d8959c58bcc91d464873b50013
// to get us started, though that code did not handle unmarshalling nulls
// correctly. ;)

import (
  "bytes"
  "database/sql"
  "encoding/json"
  "fmt"
	"reflect"
  "regexp"
  "strconv"
	"time"

  "github.com/go-sql-driver/mysql"
)

var nullJSON = []byte("null")
var nullTime, _ = time.Parse(time.RFC3339, "0000-00-00T00:00:00Z00:00")

type NullInt64 sql.NullInt64

func (ni *NullInt64) Scan(value interface{}) error {
	var i sql.NullInt64
	if err := i.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ni = NullInt64{i.Int64, false}
	} else {
		*ni = NullInt64{i.Int64, true}
	}
	return nil
}

func (ni *NullInt64) MarshalJSON() ([]byte, error) {
	if !ni.Valid {
		return nullJSON, nil
	}
	return json.Marshal(ni.Int64)
}

func (ni *NullInt64) UnmarshalJSON(b []byte) error {
  var err error = nil
  if bytes.Equal(nullJSON, b) {
    ni.Int64 = 0
    ni.Valid = false
  } else {
  	err = json.Unmarshal(b, &ni.Int64)
  	ni.Valid = (err == nil)
  }
	return err
}

func (ni* NullInt64) Native() *sql.NullInt64 {
  return &sql.NullInt64{Int64: ni.Int64, Valid: ni.Valid}
}
// END: Null64Int handlers

type NullBool sql.NullBool

func (nb *NullBool) Scan(value interface{}) error {
	var b sql.NullBool
	if err := b.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*nb = NullBool{b.Bool, false}
	} else {
		*nb = NullBool{b.Bool, true}
	}

	return nil
}

func (nb *NullBool) MarshalJSON() ([]byte, error) {
	if !nb.Valid {
		return nullJSON, nil
	}
	return json.Marshal(nb.Bool)
}

func (nb *NullBool) UnmarshalJSON(b []byte) error {
  var err error = nil
  if bytes.Equal(nullJSON, b) {
    nb.Bool = false
    nb.Valid = false
  } else {
  	err = json.Unmarshal(b, &nb.Bool)
  	nb.Valid = (err == nil)
  }
	return err
}

func (nb* NullBool) Native() *sql.NullBool {
  return &sql.NullBool{Bool: nb.Bool, Valid: nb.Valid}
}
// END NullBool handlers

type NullFloat64 sql.NullFloat64

func (nf *NullFloat64) Scan(value interface{}) error {
	var f sql.NullFloat64
	if err := f.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*nf = NullFloat64{f.Float64, false}
	} else {
		*nf = NullFloat64{f.Float64, true}
	}

	return nil
}

func (nf *NullFloat64) MarshalJSON() ([]byte, error) {
	if !nf.Valid {
		return nullJSON, nil
	}
	return json.Marshal(nf.Float64)
}

func (nf *NullFloat64) UnmarshalJSON(b []byte) error {
  var err error = nil
  if bytes.Equal(nullJSON, b) {
    nf.Float64 = 0.0
    nf.Valid = false
  } else {
  	err = json.Unmarshal(b, &nf.Float64)
  	nf.Valid = (err == nil)
  }
	return err
}

func (nf* NullFloat64) Native() *sql.NullFloat64 {
  return &sql.NullFloat64{Float64: nf.Float64, Valid: nf.Valid}
}
// END NullFloat64 handlers

// NullString is an alias for sql.NullString data type
type NullString sql.NullString

func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ns = NullString{s.String, false}
	} else {
		*ns = NullString{s.String, true}
	}

	return nil
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return nullJSON, nil
	}
	return json.Marshal(ns.String)
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
  var err error = nil
  if bytes.Equal(nullJSON, b) {
    ns.String = ""
    ns.Valid = false
  } else {
  	err = json.Unmarshal(b, &ns.String)
  	ns.Valid = (err == nil)
  }
	return err
}

func (ns* NullString) Native() *sql.NullString {
  return &sql.NullString{String: ns.String, Valid: ns.Valid}
}
// END NullString handlers

// TODO: This should be NullTimestamp
type NullTime mysql.NullTime

func (nt *NullTime) Scan(value interface{}) error {
	var t mysql.NullTime
	if err := t.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*nt = NullTime{t.Time, false}
	} else {
		*nt = NullTime{t.Time, true}
	}

	return nil
}

func (nt *NullTime) MarshalJSON() ([]byte, error) {
	if !nt.Valid {
		return nullJSON, nil
	}
	val := fmt.Sprintf("\"%s\"", nt.Time.Format(time.RFC3339))
	return []byte(val), nil
}

func (nt *NullTime) UnmarshalJSON(b []byte) error {
  if bytes.Equal(nullJSON, b) {
    nt.Time = nullTime
    nt.Valid = false
  } else {
  	s := string(b)
  	// s = Stripchars(s, "\"")

  	x, err := time.Parse(time.RFC3339, s)
  	if err != nil {
  		nt.Valid = false
  		return err
  	}

  	nt.Time = x
  	nt.Valid = true
  }
	return nil
}

func (nt* NullTime) Native() *mysql.NullTime {
  return &mysql.NullTime{Time: nt.Time, Valid: nt.Valid}
}
// END NullTime handlers

// Date is 'timezone-less' so we base if off string as all we care about is
// YYYY-MM-DD format.
type NullDate sql.NullString
var dateRegexp *regexp.Regexp = regexp.MustCompile(`((\d{4})[\.-](\d\d)[\.-](\d\d))`)

func (nt *NullDate) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*nt = NullDate{s.String, false}
	} else {
    matches := dateRegexp.FindStringSubmatch(s.String)
    // Any invalid date results in an error.
    if matches == nil {
      return fmt.Errorf("'%s' does not parse as a date.", s.String)
    }
    var year, month, day int64
    var err error
    if year, err = strconv.ParseInt(matches[2], 10, 32); err != nil {
      return fmt.Errorf("'%s' does not parse as a date.", s.String)
    }
    if month, err = strconv.ParseInt(matches[3], 10, 32); err != nil {
      return fmt.Errorf("'%s' does not parse as a date.", s.String)
    }
    if day, err = strconv.ParseInt(matches[4], 10, 32); err != nil {
      return fmt.Errorf("'%s' does not parse as a date.", s.String)
    }
    // We use this to test that the string hits a valid day
    testTime := time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, time.UTC)

    // 'Date' normalizes, but we don't.
    if int(year) != testTime.Year() || time.Month(month) != testTime.Month() || int(day) != testTime.Day() {
      return fmt.Errorf("'%s' specifies a non-existent date (e.g., '2000-10-32').", s.String)
    }
    // Pull out just the 'YYYY-MM-DD' part; 'nt.String' will come in from MySQL with 'T00:00:00Z' on the end.
		*nt = NullDate{matches[1], true}
	}

	return nil
}

func (nd *NullDate) MarshalJSON() ([]byte, error) {
  if !nd.Valid {
    return nullJSON, nil
  }
  return json.Marshal(nd.String)
}

func (nd *NullDate) UnmarshalJSON(b []byte) error {
  var err error = nil
  if bytes.Equal(nullJSON, b) {
    nd.String = ""
    nd.Valid = false
  } else {
    err = json.Unmarshal(b, &nd.String)
    nd.Valid = (err == nil)
  }
  return err
}

func (nd* NullDate) Native() *sql.NullString {
  return &sql.NullString{String: nd.String, Valid: nd.Valid}
}
// END NullDate handlers
