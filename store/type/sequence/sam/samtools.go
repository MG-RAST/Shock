package sam

//wrapper functions of selected commands from samtools
//(http://samtools.sourceforge.net/samtools.shtml)

import (
	"errors"
	"fmt"
	//	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

//check whether samtools is in PATH
func SamtoolsLookpath() (inPat bool) {
	_, err := exec.LookPath("samtools")
	if err != nil {
		return false
	}
	return true
}

//convert sam to bam
//Inputs: the name of sam_file to be converted (e.g. example.sam)
//Samtools commands: 
//        samtools view -o <example.bam> -Sb <example.sam>
//Ouputs:
//    bam file <example.bam> generated and name returned
func SamtoolsSamToBam(sam_name string) (bam_name string, err error) {
	ext := filepath.Ext(sam_name)
	if ext != ".sam" {
		fmt.Printf("SamtoolsSamToBam(): .sam file required\n")
		return "", errors.New("SamtoolsSamToBam(): .sam file required")
	}

	bam_name = changeExt(sam_name, "bam")

	err = exec.Command("samtools", "view", "-o", bam_name, "-Sb", sam_name).Run()
	if err != nil {
		fmt.Printf("error in viewing sam, %v", err)
		return "", err
	}
	return bam_name, nil
}

//sort and index bam file
//Inputs: the name of bam_file to sort and index (e.g. example.bam)
//Samtools commands:
//        samtools sort <example.bam> <example>
//        samtools index <example.bam>
//Ouputs:
//    1. <example.bam> sorted 
//    2. <example.bam.bai> generated and name returned
func SamtoolsSortAndIndex(bam_file string) (bai_file string, err error) {
	ext := filepath.Ext(bam_file)
	if ext != ".bam" {
		return "", errors.New("SamtoolsSortAndIndex(): .bam file required")
	}

	sorted_name_base := changeExt(bam_file, "")

	err = exec.Command("samtools", "sort", bam_file, sorted_name_base).Run()
	if err != nil {
		return "", err
	}

	sorted_bam_file := fmt.Sprintf("%s.bam", sorted_name_base)
	err = exec.Command("samtools", "index", sorted_bam_file).Run()
	if err != nil {
		return "", err
	}
	bai_file = fmt.Sprintf("%s.bai", sorted_bam_file)
	return bai_file, nil
}

//view sam file
//invoking: samtools view -S <args> <example.sam>
//args can be  [-h] [-H] [-c] [-f INT], and so on
func SamtoolsViewSam(sam_name string, args ...string) (out []byte, err error) {
	ext := filepath.Ext(sam_name)
	if ext != ".sam" {
		fmt.Printf("SamtoolsSamToBam(): .sam file required\n")
		return nil, errors.New("SamtoolsSamToBam(): .sam file required")
	}

	argv := []string{}
	argv = append(argv, "view")
	argv = append(argv, "-S")
	argv = append(argv, args...)
	argv = append(argv, sam_name)

	out, err = exec.Command("samtools", argv...).Output()

	return
}

//view sam file
//invoking: samtools view  <args> <example.bam> [region_string]
//args can be  [-h] [-H] [-c] [-f INT], and so on
func SamtoolsViewBam(bam_file string, region string, args ...string) (out []byte, err error) {
	ext := filepath.Ext(bam_file)
	if ext != ".bam" {
		return nil, errors.New("SamtoolsSortAndIndex(): .bam file required")
	}

	argv := []string{}
	argv = append(argv, "view")
	argv = append(argv, args...)
	argv = append(argv, bam_file)
	if region != "" {
		argv = append(argv, region)
	}

	out, err = exec.Command("samtools", argv...).Output()

	return
}

//change the extention of file_name to new_ext (if new_ext is "", just remove original ext)
func changeExt(file_name string, new_ext string) (new_file_name string) {
	for i := len(file_name) - 1; i >= 0 && !os.IsPathSeparator(file_name[i]); i-- {
		if file_name[i] == '.' {
			name_base := file_name[0:i]
			if new_ext == "" {
				new_file_name = fmt.Sprintf("%s", name_base)
			} else {
				new_file_name = fmt.Sprintf("%s.%s", name_base, new_ext)
			}
			return
		}
	}
	if new_ext == "" {
		new_file_name = fmt.Sprintf("%s", file_name)
	} else {
		new_file_name = fmt.Sprintf("%s.%s", file_name, new_ext)
	}
	return
}
