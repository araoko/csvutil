package csvstruct

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

//LoadFromReader creates csvStruct object from a reader
func LoadFromReader(reader io.Reader) (*CsvStruct, error){
	r := csv.NewReader(reader)
	//r.Comma = '\t'
	headers, err := r.Read()
	if err != nil {
		return nil, err
	}

	t := CsvStruct{
		h: copySlice(headers),
		c: make([][]string, 0),
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
	return &t, nil
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
	return LoadFromReader(file)
}

func copySlice(s []string) []string {
	a := make([]string, len(s))
	copy(a, s)
	return a
}

type CsvStruct struct {
	h []string
	c [][]string
}

func (r *CsvStruct) append(row []string) error {
	if len(row) != len(r.h) {
		return fmt.Errorf("Error: header count (%d) is not equal to entry count (%d)", len(r.h), len(row))
	}
	r.c = append(r.c, copySlice(row))
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

func (r *CsvStruct) GetValueAtIndex(rowIndex, headerIndex int) (string, error){
if rowIndex < 0 || headerIndex <0 || rowIndex >= r.RowCount() || headerIndex >= r.HeaderCount(){
	return "", fmt.Errorf("Error: Row Index (%v) and Header Index (%v) out of bounds. Row Count (%v), Header Count (%v)",rowIndex,headerIndex,r.RowCount(),r.HeaderCount())
}
return r.c[rowIndex][headerIndex], nil
}

func (r *CsvStruct) FindEntry(headerName string, value string) ([]string, int) {

	p := r.HeaderIndex(headerName)
	if p == -1 {
		return nil, -1
	}

	for i, v := range r.c {
		if strings.EqualFold(value, v[p]) {
			return v, i
		}
	}
	return nil, -1
}
