package main

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

// An Encoder that can provide a file query specification to be used for
// getting files from s3, and later encode the results of that
// query in a specific format.
type Encoder interface {
	// Encode writes data to the response of an http request
	// to fetch file from s3.
	//
	// It encodes the query object into a specific format such
	// as XML or JSON and writes to the response.
	Encode(w http.ResponseWriter, r *http.Request, data []byte, fileName string) error
}

// JSONEncoder encodes the results of an IP lookup as JSON.
type JSONEncoder struct {
	Indent bool
}

// Encode implements the Encoder interface.
func (f *JSONEncoder) Encode(w http.ResponseWriter, r *http.Request, data []byte, fileName string) error {
	record := data
	callback := r.FormValue("callback")
	if len(callback) > 0 {
		return f.P(w, r, record, callback)
	}
	w.Header().Set("Content-Type", "application/json")
	if f.Indent {
	}
	return json.NewEncoder(w).Encode(record)
}

// P writes JSONP to an http response.
func (f *JSONEncoder) P(w http.ResponseWriter, r *http.Request, data []byte, callback string) error {
	w.Header().Set("Content-Type", "application/javascript")
	_, err := io.WriteString(w, callback+"(")
	if err != nil {
		return err
	}
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, ");")
	return err
}

// XMLEncoder encodes the results of an IP lookup as XML.
type XMLEncoder struct {
	Indent bool
}

// Encode implements the Encoder interface.
func (f *XMLEncoder) Encode(w http.ResponseWriter, r *http.Request, data []byte, fileName string) error {
	record := ""
	w.Header().Set("Content-Type", "application/xml")
	_, err := io.WriteString(w, xml.Header)
	if err != nil {
		return err
	}
	if f.Indent {
		enc := xml.NewEncoder(w)
		enc.Indent("", "\t")
		err := enc.Encode(record)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte("\n"))
		return err
	}
	return xml.NewEncoder(w).Encode(record)
}

// CSVEncoder encodes the results of an IP lookup as CSV.
type CSVEncoder struct {
	UseCRLF bool
}

// Encode implements the Encoder interface.
func (f *CSVEncoder) Encode(w http.ResponseWriter, r *http.Request, data []byte, fileName string) error {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename="+fileName)
	cw := csv.NewWriter(w)
	cw.UseCRLF = f.UseCRLF
	var err error
	if err != nil {
		return err
	}
	cw.Flush()
	w.Write(data)
	return nil
}
