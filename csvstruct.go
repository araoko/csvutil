package csvutil

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

func NewCsvStruct(headers []string) (*CsvStruct, error) {
	if len(headers) == 0 {
		return nil, fmt.Errorf("cannot create CsvStruct with nil or zero lenght headers")
	}
	return &CsvStruct{
		h: copySlice(headers),
		c: make([][]string, 0),
	}, nil
}

//LoadFromIOReader creates csvStruct object from a io.reader
func LoadFromIOReader(reader io.Reader) (*CsvStruct, error) {
	r := csv.NewReader(reader)
	//r.Comma = '\t'
	r.FieldsPerRecord = -1
	return LoadFromCSVReader(r)

}

//LoadFromCSVReader creates csvStruct object from a csv.reader
func LoadFromCSVReader(r *csv.Reader) (*CsvStruct, error) {
	headers, err := r.Read()
	if err != nil {
		return nil, err
	}

	t, err := NewCsvStruct(headers)
	if err != nil {
		return nil, err
	}

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		err = t.append(row)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

//LoadFile reads the csv file and creates a csvStruct object
func LoadFile(csvFile string) (*CsvStruct, error) {
	file, err := os.Open(csvFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		file.Close()
	}()
	r := CreateCSVReaderFromIOReadSeeker(file)
	return LoadFromCSVReader(r)
}

func CreateCSVReaderFromIOReadSeeker(file io.ReadSeeker) *csv.Reader {
	SkipBOM(file)
	return CreateCSVReaderFromIOReader(file)
}

func CreateCSVReaderFromIOReader(file io.Reader) *csv.Reader {
	r := csv.NewReader(file)
	//r.Comma = '\t'
	r.FieldsPerRecord = -1
	return r
}

func copySlice(s []string) []string {
	a := make([]string, len(s))
	copy(a, s)
	return a
}

func copySliceL(s []string, l int) []string {
	a := make([]string, l)
	copy(a, s)
	return a
}

func SkipBOM(fd io.ReadSeeker) error {
	var bom [3]byte
	_, err := io.ReadFull(fd, bom[:])
	if err != nil {
		return err
	}
	if bom[0] != 0xef || bom[1] != 0xbb || bom[2] != 0xbf {
		_, err = fd.Seek(0, 0) // Not a BOM -- seek back to the beginning
		if err != nil {
			return err
		}
	}
	return nil
}

type CsvStruct struct {
	h []string
	c [][]string
}

func (r *CsvStruct) Add(s *CsvStruct) error {
	//TODO this works only if the collomns for both csv sfruct are at
	//the same indeces, must fix this
	if r.HeaderCount() != s.HeaderCount() {
		return fmt.Errorf("Error: header count  mis-match. first = (%d)   second = (%d)", len(r.h), len(s.h))
	}
	for i, v := range s.h {
		if r.h[i] != v {
			return fmt.Errorf("Error: header mis-match. at index (%d) : (%s)  and (%s)", i, r.h[i], s.h[i])
		}
	}
	for _, v := range s.c {
		r.c = append(r.c, v)
	}
	return nil
}

func (r *CsvStruct) append(row []string) error {
	if len(row) > len(r.h) {
		return fmt.Errorf("Error: header count (%d) is less than entry count (%d)", len(r.h), len(row))
	}

	r.c = append(r.c, copySliceL(row, len(r.h)))
	return nil
}

func (r *CsvStruct) RowAtIndex(i int) []string {
	if i < 0 || i >= len(r.c) {
		return nil
	}
	return copySlice(r.c[i])
}
func (r *CsvStruct) HeaderAtIndex(i int) (string, error) {
	if i < 0 || i >= len(r.h) {
		return "", fmt.Errorf("Error: Index (%d) is out of range (%d)", i, len(r.h))
	}
	return r.h[i], nil
}
func (r *CsvStruct) HeaderIndex(headerName string) int {
	for p, v := range r.h {
		if strings.EqualFold(v, headerName) {
			return p
		}
	}
	return -1
}

func (r *CsvStruct) HeaderCount() int {
	return len(r.h)
}

func (r *CsvStruct) RowCount() int {
	return len(r.c)
}

func (r *CsvStruct) Headers() []string {
	return copySlice(r.h)
}

func (r *CsvStruct) GetValueAtIndex(rowIndex, headerIndex int) (string, error) {
	if rowIndex < 0 || headerIndex < 0 || rowIndex >= r.RowCount() || headerIndex >= r.HeaderCount() {
		return "", fmt.Errorf("Error: Row Index (%v) and Header Index (%v) out of bounds. Row Count (%v), Header Count (%v)", rowIndex, headerIndex, r.RowCount(), r.HeaderCount())
	}
	return r.c[rowIndex][headerIndex], nil
}

func (r *CsvStruct) SetValueAtIndex(rowIndex, headerIndex int, value string) error {
	if rowIndex < 0 || headerIndex < 0 || rowIndex >= r.RowCount() || headerIndex >= r.HeaderCount() {
		return fmt.Errorf("Error: Row Index (%v) and Header Index (%v) out of bounds. Row Count (%v), Header Count (%v)", rowIndex, headerIndex, r.RowCount(), r.HeaderCount())
	}
	r.c[rowIndex][headerIndex] = value
	return nil
}

func (r *CsvStruct) FindEntryI(headerIndex int, value string) ([]string, int) {

	for i, v := range r.c {
		if strings.EqualFold(value, v[headerIndex]) {
			return v, i
		}
	}
	return nil, -1
}

func (r *CsvStruct) FindEntry(headerName string, value string) ([]string, int) {

	p := r.HeaderIndex(headerName)
	if p == -1 {
		return nil, -1
	}
	return r.FindEntryI(p, value)
}

func (r *CsvStruct) Write2File(f string) error {
	file, err := os.OpenFile(f, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	writer := csv.NewWriter(file)
	err = writer.Write(r.Headers())
	if err != nil {
		return err
	}
	err = writer.WriteAll(r.c)
	if err != nil {
		return err
	}
	return nil
}
