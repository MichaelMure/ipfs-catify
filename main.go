package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/pkg/errors"
	"gopkg.in/auyer/steganography.v2"
)

func main() {
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintln(os.Stderr, "input path to an image required")
		return
	}

	if len(os.Args) < 3 {
		_, _ = fmt.Fprintln(os.Stderr, "output image path required")
		return
	}

	if !strings.HasSuffix(os.Args[2], ".png") {
		_, _ = fmt.Fprintln(os.Stderr, "output image must be a png")
		return
	}

	inFile, err := os.Open(os.Args[1])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errors.Wrap(err, "errors opening file"))
		return
	}


	reader := bufio.NewReader(inFile)
	img, _, err := image.Decode(reader)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}

	err = inFile.Close()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errors.Wrap(err, "error closing file"))
		return
	}

	s := shell.NewLocalShell()

	w := new(bytes.Buffer)
	var data []byte
	for i := 0; ; i++{
		w.Reset()

		err := steganography.Encode(w, img, []byte(strconv.Itoa(i)))
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, errors.Wrap(err, "error catifying file"))
			return
		}

		// duplicate buffer so it's not depleted when sending to the IPFS daemon
		data, _ =   ioutil.ReadAll(w)

		c, err := s.Add(bytes.NewBuffer(data), shell.OnlyHash(true))
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, errors.Wrap(err, "error requesting local IPFS node"))
			return
		}

		if strings.HasPrefix(c,"Qmcat") {
			fmt.Println("Found one !")
			fmt.Println(c)
			break
		}

		fmt.Printf("Attempt %d: %s\n", i, c)
	}

	outFile, err := os.Create(os.Args[2])
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errors.Wrap(err, "error opening file for writing"))
		return
	}

	_, err = bytes.NewBuffer(data).WriteTo(outFile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, errors.Wrap(err, "error writing file"))
		return
	}

	_ = outFile.Close()
}
