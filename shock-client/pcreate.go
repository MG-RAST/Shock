package main

import (
	"fmt"
	"github.com/MG-RAST/Shock/shock-client/conf"
	"github.com/MG-RAST/Shock/shock-client/lib"
	"github.com/MG-RAST/Shock/pool"
	"io"
	"os"
	"strconv"
)

func uploader(args ...interface{}) interface{} {
	n := args[0].(lib.Node)
	part := args[1].(int)
	fh := args[2].(*os.File)
	size := args[3].(int64)
	return n.UploadPart(strconv.Itoa(part), io.NewSectionReader(fh, int64(part-1)*conf.CHUNK_SIZE, size), size)
}

func pcreate(args []string) (err error) {
	n := lib.Node{}
	var filename string
	if ne(conf.Flags["full"]) {
		filename = (*conf.Flags["full"])
	} else {
		helpf("pcreate requires file path: -full=<u>")
	}

	var filesize int64
	fh, err := os.Open(filename)
	if err != nil {
		handleString(fmt.Sprintf("Error open file: %s\n", err.Error()))
	}
	if fi, _ := fh.Stat(); err == nil {
		filesize = fi.Size()
	}

	chunks := int(filesize / (conf.CHUNK_SIZE))
	if filesize%conf.CHUNK_SIZE != 0 {
		chunks += 1
	}

	if chunks == 1 {
		opts := lib.Opts{}
		opts["upload_type"] = "full"
		opts["full"] = filename
		if err := n.Create(opts); err != nil {
			handleString(fmt.Sprintf("Error creating node: %s\n", err.Error()))
		} else {
			n.PP()
		}
	} else {
		threads, _ := strconv.Atoi(*conf.Flags["threads"])
		if threads == 0 {
			threads = 1
		}

		//create node
		opts := lib.Opts{}
		opts["upload_type"] = "parts"
		opts["parts"] = strconv.Itoa(chunks)
		if err := n.Create(opts); err != nil {
			handleString(fmt.Sprintf("Error creating node: %s\n", err.Error()))
		}

		workers := pool.New(threads)
		workers.Run()
		for i := 0; i < chunks; i++ {
			size := int64(conf.CHUNK_SIZE)
			if size*(int64(i)+1) > filesize {
				size = filesize - size*(int64(i))
			}
			workers.Add(uploader, n, (i + 1), fh, size)
		}
		workers.Wait()
		maxRetries := 10
		for i := 1; i <= maxRetries; i++ {
			errCount := 0
			completed_jobs := workers.Results()
			for _, job := range completed_jobs {
				if job.Result != nil {
					err := job.Result.(error)
					println("Chunk", job.Args[1].(int), "error:", err.Error())
					workers.Add(job.F, job.Args...)
					errCount++
				}
			}
			if errCount == 0 {
				println("All chunks successfully upload.")
				break
			} else {
				println("Retry", i, "of", maxRetries)
				workers.Wait()
			}
		}
		workers.Stop()

		n.Get()
		n.PP()
	}
	return
}
