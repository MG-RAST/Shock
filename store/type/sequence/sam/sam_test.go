package sam_test

import (
	"fmt"
	. "github.com/MG-RAST/Shock/store/type/sequence/sam"
	"io"
	"testing"
)

var (
	sample = "../testdata/sample1.sam"
	Idx    [][]int64 //list of {offset, length} pair
)

func TestRead(t *testing.T) {
	var (
		obtainN [][]byte
		obtainS [][]byte
	)

	if r, err := NewReaderName(sample); err != nil {
		t.Errorf("Failed to open test file %s: %v", sample, err.Error())
	} else {
		for i := 0; i < 2; i++ {
			var linect int
			for {
				if s, err := r.Read(); err != nil {
					if err == io.EOF {
						break
					} else {
						t.Errorf("Failed to read %s: %v", sample, err.Error())
					}
				} else {
					fmt.Println(i + 1)
					obtainN = append(obtainN, s.ID)
					obtainS = append(obtainS, s.Seq)
					linect += 1
					//fmt.Printf("line %d = %s\n", linect, s.Seq)
				}
			}
			obtainN = nil
			obtainS = nil
			if err = r.Rewind(); err != nil {
				t.Errorf("Failed to rewind %s: %v", sample, err.Error())
			}
		}
		r.Close()
	}
}

func TestReadRaw(t *testing.T) {
	if r, err := NewReaderName(sample); err != nil {
		t.Errorf("Failed to open test file %s: %v", sample, err.Error())
	} else {
		for i := 0; i < 2; i++ {
			var linect int
			for {
				buf := make([]byte, 32*1024)

				if n, err := r.ReadRaw(buf); err != nil {
					if err == io.EOF {
						break
					} else {
						t.Errorf("Fail to read in TestReadRaw() %s: %v", sample, err.Error())
					}
				} else {
					fmt.Println(i + 1)
					linect += 1
					fmt.Printf("line=%d, length=%d, line_content=%s\n", linect, n, buf)
				}
			}

			if err = r.Rewind(); err != nil {
				t.Errorf("Failed to rewind %s: %v", sample, err.Error())
			}
		}
		r.Close()
	}
}

func TestCreateIndex(t *testing.T) {
	curr := int64(0)

	if r, err := NewReaderName(sample); err != nil {
		t.Errorf("Failed to open test file %s: %v", sample, err.Error())
	} else {
		for {
			buf := make([]byte, 32*1024)

			if n, err := r.ReadRaw(buf); err != nil {
				if err == io.EOF {
					break
				} else {
					t.Errorf("Fail to read in TestCreatIndex() %s: %v", sample, err.Error())
				}
			} else {
				Idx = append(Idx, []int64{curr, int64(n)})
				curr += int64(n)
			}
		}

		if err = r.Rewind(); err != nil {
			t.Errorf("Failed to rewind %s: %v", sample, err.Error())
		}

		fmt.Printf("indices= %v", Idx)
		r.Close()
	}
	return
}

func TestReadSeqByIndex(t *testing.T) {
	curr := int64(0)

	if r, err := NewReaderName(sample); err != nil {
		t.Errorf("Failed to open test file %s: %v", sample, err.Error())
	} else {
		for {
			buf := make([]byte, 32*1024)

			if n, err := r.ReadRaw(buf); err != nil {
				if err == io.EOF {
					break
				} else {
					t.Errorf("Fail to read in TestCreatIndex() %s: %v", sample, err.Error())
				}
			} else {
				Idx = append(Idx, []int64{curr, int64(n)})
				curr += int64(n)
			}
		}

		if err = r.Rewind(); err != nil {
			t.Errorf("Failed to rewind %s: %v", sample, err.Error())
		}

		fmt.Printf("indices= %v", Idx)
		r.Close()
	}
	return
}
