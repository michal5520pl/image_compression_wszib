package main

import (
	"fmt"

	"github.com/tomcraven/goga"
)

func main() {
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
	inputPath := "../szambo_jednokomorowe.png"             //plik wejściowy
	outputPath := "../szambo_jednokomorowe.compressed.jpg" //plik wyjściowy
	quality := 90                                          //zakres kompresji (1-100)

	err := CompressImage(inputPath, outputPath, quality)
	if err != nil {
		fmt.Println("Error:", err.Error())
	} else {
		fmt.Println("Done UwU")
	}

	err = RunGeneticAlgorithm(inputPath, outputPath)
	if err != nil {
		fmt.Println("Error:", err.Error())
	} else {
		fmt.Println("Done UwU")
	}

}
