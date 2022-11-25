package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func watch(cCtx *cli.Context) error {
	inputDir := string(cCtx.String("dir"))
	s3Dir := string(cCtx.String("s3_url"))
	dir := ""

	if len(s3Dir) == 0 {
		return errors.New("s3-sync: watch(): No s3 url provided. Exiting.")
	}
	if len(inputDir) == 0 {
		log.Printf("s3-sync: watch(): No directory provided. Using ./ as default.")
		dir = "./"
	} else {
		dir = inputDir
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Wrap(err, "s3-sync: watch(): Error while creating Watcher: ")
	}
	defer watcher.Close()

	// Launch go routine to sync with s3
	cJobs := make(chan string, 5)
	cErrors := make(chan error, 5)
	wgJobs := sync.WaitGroup{}
	wgErrors := sync.WaitGroup{}

	wgJobs.Add(1)
	go s3Sync(s3Dir, cJobs, cErrors, &wgJobs)
	wgErrors.Add(1)
	go logErrors(cErrors, &wgErrors)

	done := make(chan bool)
	go func() {
		defer close(done)

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					err = errors.New("s3-sync: watch(): Event watching routine exception. Exiting.")
					cErrors <- err
					return
				}
				if event.Op.String() == "WRITE" || event.Op.String() == "CREATE" {
					log.Printf("s3-sync: Syncing file: %s | Op: %s", event.Name, event.Op)
					cJobs <- event.Name
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					err = errors.Wrap(err, "s3-sync: watch(): Event watching routine Error:")
					cErrors <- err
					return
				}
			}
		}

	}()

	err = watcher.Add(dir)
	if err != nil {
		return errors.Wrap(err, "s3-sync: watch(): Add failed:")
	}
	wgErrors.Wait()
	wgJobs.Wait()
	<-done
	close(cJobs)
	close(cErrors)

	return nil
}

func s3Sync(s3Dir string, cFileNames <-chan string, cErrors chan<- error, wgJobs *sync.WaitGroup) {
	defer wgJobs.Done()

	sep := "/"
	s3Path := ""
	for fileName := range cFileNames {
		if string(s3Dir[len(s3Dir)-1]) == sep {
			s3Path = fmt.Sprintf("%s%s", s3Dir, filepath.Base(fileName))
		} else {
			s3Path = fmt.Sprintf("%s/%s", s3Dir, filepath.Base(fileName))
		}

		log.Printf("s3-sync: syncing with s3 path: %s", s3Path)
		cmd := fmt.Sprintf("aws s3 cp '%s' '%s'", fileName, s3Path)
		s3CpCmd := exec.Command("bash", "-c", cmd)

		_, err := s3CpCmd.Output()
		if err != nil {
			cErrors <- errors.Wrap(err, "s3Sync: Command exec error: ")
			return
		}
	}

	cErrors <- nil
}

func logErrors(cErrors <-chan error, wgErrors *sync.WaitGroup) {
	defer wgErrors.Done()

	for err := range cErrors {
		log.Println(err)
	}
}

func main() {
	app := &cli.App{
		Action: watch,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "dir",
				Usage:    "Directory to sync with s3",
				Value:    "",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "s3_url",
				Usage:    "S3 url to sync with",
				Value:    "",
				Required: true,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
