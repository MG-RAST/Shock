package sam_test

import (
	"bytes"
	"fmt"
	. "github.com/MG-RAST/Shock/store/type/sequence/sam"
	"path/filepath"
	"testing"
)

var (
	sam_files      = []string{"../../../testdata/toy.sam", "../../../testdata/sample1.sam", "../../../testdata/500kb.sam"}
	bam_files      = []string{"../../../testdata/toy.bam", "../../../testdata/sample1.bam", "../../../testdata/500kb.bam"}
	arg_lists      = [][]string{[]string{"-c"}, []string{"-H"}, []string{"-h", "-f", "83"}}
	region_strings = []string{"ref", "ref2", "ref2:2-3"}
)

func TestSamtoolsLookpath(t *testing.T) {
	if SamtoolsLookpath() {
		fmt.Printf("samtools is in PATH\n")
	} else {
		fmt.Printf("samtools is NOT in PATH\n")
	}
}

func TestSamtoolsSamToBam(t *testing.T) {
	if !SamtoolsLookpath() {
		fmt.Printf("samtools is NOT in PATH!!!\n")
		return
	}

	for _, sam_name := range sam_files {
		if bam_name, err := SamtoolsSamToBam(sam_name); err != nil {
			t.Errorf("SamtoolsSamToBam: %v", err)
		} else {
			fmt.Printf("SamtoolsSamToBam success: %s -> %s\n", filepath.Base(sam_name), filepath.Base(bam_name))
		}
	}
}

func TestSamtoolsSortAndIndex(t *testing.T) {
	if !SamtoolsLookpath() {
		fmt.Printf("samtools is NOT in PATH!!!\n")
		return
	}

	for _, bam_name := range bam_files {
		fmt.Printf("Sorting and indexing bam file... %s\n", filepath.Base(bam_name))
		if bai_name, err := SamtoolsSortAndIndex(bam_name); err != nil {
			t.Errorf("SamtoolsSortAndIndex: %v", err)
		} else {

			fmt.Printf("Done: %s sorted, %s genereated\n", filepath.Base(bam_name), filepath.Base(bai_name))
		}

	}
}

func TestSamtoolsViewSam(t *testing.T) {
	if !SamtoolsLookpath() {
		fmt.Printf("samtools is NOT in PATH!!!\n")
		return
	}
	//testing viewing sam file
	for _, sam_name := range sam_files {
		for _, args := range arg_lists {
			fmt.Printf("\ntesting SamtoolsViewSam for %s:\nARGS=%v\n", filepath.Base(sam_name), args)
			if out, err := SamtoolsViewSam(sam_name, args...); err != nil {
				t.Errorf("SamtoolsViewSam: %v", err)
			} else {
				fmt.Printf("RETURN=\n")
				if len(out) > 1024 {
					b := bytes.NewBuffer(out)
					b.Truncate(1024)
					fmt.Printf("%s...[only showing the first 1024 bytes of the output]\n", b)
				} else {
					fmt.Printf("%s\n", out)
				}

			}
		}
	}
}

func TestSamtoolsViewBam(t *testing.T) {

	//testing view bam file without region
	for _, bam_name := range bam_files {
		for _, args := range arg_lists {
			fmt.Printf("\ntesting SamtoolsViewBam for %s:\nARGS=%v\n", filepath.Base(bam_name), args)
			if out, err := SamtoolsViewBam(bam_name, "", args...); err != nil {
				t.Errorf("SamtoolsViewBam: %v", err)
			} else {
				fmt.Printf("RETURN=\n")
				if len(out) > 1024 {
					b := bytes.NewBuffer(out)
					b.Truncate(1024)
					fmt.Printf("%s...[only showing the first 1024 bytes of the output]\n", b)
				} else {
					fmt.Printf("%s\n", out)
				}
			}
		}
	}

	//testing view bam file with region
	bam_name := bam_files[0]
	for _, region := range region_strings {
		fmt.Printf("\ntesting SamtoolsViewBam for %s:\nREGION=%v\n", filepath.Base(bam_name), region)
		if out, err := SamtoolsViewBam(bam_name, region, "-h"); err != nil {
			t.Errorf("SamtoolsViewBam: %v", err)
		} else {
			fmt.Printf("RETURN=\n")
			if len(out) > 1024 {
				b := bytes.NewBuffer(out)
				b.Truncate(1024)
				fmt.Printf("%s...[ouput longer than 1024 bytes truncated]\n", b)
			} else {
				fmt.Printf("%s\n", out)
			}
		}
	}

	//	for _, bam_name := range bam_files {
	//		bai_name := fmt.Sprintf("%s.bai", bam_name)
	//		exec.Command("rm", bam_name, bai_name).Run()
	//	}

}
