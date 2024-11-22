package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tomcraven/goga"
)

func main() {
	var input []string

	flag.Func("input", "Input file(s) of images. Separate them with ','", func(s string) error {
		input = strings.Split(s, ",")

		if len(input) == 0 || input == nil {
			return errors.New("no file was given")
		}

		for _, value := range input {
			if _, err := os.Stat(value); errors.Is(err, os.ErrNotExist) {
				return errors.New(value + " is not existing")
			} else if err != nil {
				return err
			}
		}

		return nil
	})

	flag.Parse()

	// jakieś użycie goga
	bitset := goga.Bitset{}
	bitset.Create(10)

	// test
	bitset.Set(0, 1)
	bitset.Set(1, 0)
	bitset.Set(2, 1)
	bitset.Set(3, 1)

	fmt.Println("Bitset:")
	for i := 0; i < bitset.GetSize(); i++ {
		fmt.Printf("Bit %d: %d\n", i, bitset.Get(i))
	}

	//==============
	const quality = 90 //zakres kompresji (1-100)

	for _, inputFile := range input {
		index := strings.LastIndex(inputFile, ".")

		if index == -1 {
			log.Default().Fatal("Filename doesn't contain a correct extension!")
		}

		err := CompressImage(inputFile, inputFile[:index]+".compressed.jpg", quality)

		if err != nil {
			log.Default().Fatal(err.Error())
		}

		err = RunGeneticAlgorithm(inputFile, inputFile[:index]+".algorithm.jpg")

		if err != nil {
			log.Default().Fatal(err.Error())
		}
	}

	fmt.Println("Done UwU")

}
