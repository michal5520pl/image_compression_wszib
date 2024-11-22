package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"math/rand"
	"os"

	"github.com/tomcraven/goga"

	"golang.org/x/image/bmp"
)

const (
	populationSize = 50  // µ
	offspringSize  = 100 // λ
	maxIterations  = 1000
)

type ImageGenome struct {
	imageData image.Image
	fitness   float64
	//GetBits   func() *goga.Bitset
}

func RunGeneticAlgorithm(inputImagePath string, outputImagePath string) error {
    //wczytywanie obrazu
    _, err := loadImage(inputImagePath)
    if err != nil {
        return fmt.Errorf("failed to load image: %w", err)
    }

    //inicjalizacja algorytmu genetycznego
    genAlgo := goga.NewGeneticAlgorithm()
    genAlgo.BitsetCreate = &myBitsetCreate{}
    genAlgo.Simulator = &myImageSimulator{}
    genAlgo.Mater = goga.NewMater([]goga.MaterFunctionProbability{
        {P: 0.5, F: goga.OnePointCrossover},
        {P: 0.5, F: goga.Mutate},
    })
    genAlgo.Selector = goga.NewSelector([]goga.SelectorFunctionProbability{
        {P: 1.0, F: goga.Roulette},
    })

    //inicjalizacja populacji i potomków
    genAlgo.Init(populationSize, offspringSize)

    //główna pętla symulacji
    for i := 0; i < maxIterations; i++ {
        for _, genome := range genAlgo.GetPopulation() {//to chyba powinno pnaprawić problem z funkcją simulate
            genAlgo.Simulator.Simulate(genome)
        }
    }

    //zapisz najlepszy genom jako obraz JPEG
    bestGenome := findBestGenome(&genAlgo)
    err = saveImageAsJPEG(bestGenome.imageData, outputImagePath)

    return err
}

func findBestGenome(genAlgo *goga.GeneticAlgorithm) *ImageGenome {
	var bestGenome *ImageGenome
	bestFitness := math.Inf(-1) // Ustaw na najmniejszą możliwą wartość

	for _, genome := range genAlgo.GetPopulation() {
        imgGenome := genome.(*ImageGenome)
        if imgGenome.fitness > bestFitness { //tutaj powinno być float64 do porownania
            bestFitness = imgGenome.fitness
            bestGenome = imgGenome
        }
    }

	return bestGenome
}

func (g ImageGenome) GetFitness() float64 {
	return g.fitness
}

func (g ImageGenome) SetFitness(fitness float64) {
	g.fitness = fitness
}

func (g ImageGenome) GetBits() *goga.Bitset {
	//rozmiar obrazu
	bounds := g.imageData.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	//oblicza rozmiar bitsetu
	//piksel ma 4 kanały (R, G, B, A) (32 bity na piksel)
	bitsetSize := width * height * 32
	bitset := &goga.Bitset{}
	bitset.Create(bitsetSize)

	//indeks do bitsetu
	index := 0

	//iterowanie przez piksele
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			//uzyskiwanie koloru piksela
			r, g, b, a := g.imageData.At(x, y).RGBA()

			//zapisanie wartości kanałów w bitset
			//przekształcenie wartości z 16-bitowych na 8-bitowe
			bitset.Set(index, int(r>>8)) // R
			index++
			bitset.Set(index, int(g>>8)) // G
			index++
			bitset.Set(index, int(b>>8)) // B
			index++
			bitset.Set(index, int(a>>8)) // A
			index++
		}
	}

	return bitset
}

func NewImageGenome(img image.Image) *ImageGenome {
	return &ImageGenome{imageData: img, fitness: 0}
}

// funkcja do oceny (funkcja fitness)
func evaluateFitness(genome *ImageGenome) {
	//kompresja obrazu do formatu JPEG
	compressedImg := compressImage(genome)

	//oblicza PSNR
	psnr := calculatePSNR(genome.imageData, compressedImg)

	//ustawienia fitness na wartość PSNR
	genome.SetFitness(psnr)
}

// funkcja do obliczania PSNR
func calculatePSNR(original image.Image, compressed image.Image) float64 {
	//uzyskanie rozmiaru obrazów
	bounds := original.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	//oblicza sume bedów kwadratowych (MSE)
	var mse float64
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			origR, origG, origB, _ := original.At(x, y).RGBA()
			compR, compG, compB, _ := compressed.At(x, y).RGBA()

			//oblicza różnice dla kazdego kanału
			mse += math.Pow(float64(origR>>8)-float64(compR>>8), 2)
			mse += math.Pow(float64(origG>>8)-float64(compG>>8), 2)
			mse += math.Pow(float64(origB>>8)-float64(compB>>8), 2)
		}
	}

	//jaki jest sredni blad kwadratowy (MSE)
	mse /= float64(width * height * 3) // 3 kanały: R, G, B

	//oblicza PSNR
	if mse == 0 {
		//jesli nie ma bledu to nieskończony PSNR
		return math.Inf(1)
	}
	return 20*math.Log10(255) - 10*math.Log10(mse)
}

// funkcja kompresująca obraz
func compressImage(genome *ImageGenome) image.Image {
	//rekonstrukcja obrazu z bitsetu
	bounds := genome.imageData.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y

	//tworzenie nowego obrazu
	newImage := image.NewRGBA(bounds)

	//indeks do bitsetu
	index := 0

	//iteracja przez każdy piksel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			//odczytywanie wartosci kanałow z bitset
			r := genome.GetBits().Get(index) // R
			index++
			g := genome.GetBits().Get(index) // G
			index++
			bValue := genome.GetBits().Get(index) // B
			index++
			a := genome.GetBits().Get(index) // A
			index++

			//ustawianie koloru pikseli w nowym obrazie
			newImage.Set(x, y, color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(bValue),
				A: uint8(a),
			})
		}
	}

	//tworzenie pliku JPEG
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, newImage, nil)
	if err != nil {
		//idk czy nie rozbudować bardziej obsługi błedów
		panic(err)
	}

	//zwraca skompresowany obraz
	//mozna zwrócić nowy obraz JPEG jako .Image bo tutaj zwracamy nowy obraz w formacie RGBA
	return newImage
}

// func geneticAlgorithm() {
// 	genAlgo := goga.NewGeneticAlgorithm()

// 	genAlgo.BitsetCreate = &myBitsetCreate{}
// 	genAlgo.Simulator = &myImageSimulator{}
// 	genAlgo.Mater = goga.NewMater([]goga.MaterFunctionProbability{
// 		{P: 0.5, F: goga.OnePointCrossover},
// 		{P: 0.5, F: goga.Mutate},
// 	})
// 	genAlgo.Selector = goga.NewSelector([]goga.SelectorFunctionProbability{
// 		{P: 1.0, F: goga.Roulette},
// 	})

// 	genAlgo.Init(populationSize, offspringSize)

// 	for i := 0; i < maxIterations; i++ {
// 		genAlgo.Simulate()
// 	}
// }

type myBitsetCreate struct{}

func (bc *myBitsetCreate) Go() goga.Bitset {
	b := goga.Bitset{}

	//ustala rozmiar bitsetu na podstawie populacji i rozmiaru obrazu
	// np jeśli obraz ma 100x100 pikseli to :
	width, height := 100, 100         // Można to ustawić dynamicznie w zależności od obrazu
	bitsetSize := width * height * 32 // 32 bity na piksel (4 kanały po 8 bitów)

	b.Create(bitsetSize)

	//indeks do bitsetu
	index := 0

	//losowanie wartości dla każdego piksela
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			//losowanie wartości kanałow (R, G, B, A) zakres (0-255)
			r := int(rand.Intn(256))
			g := int(rand.Intn(256))
			bValue := int(rand.Intn(256))
			a := int(rand.Intn(256))

			//ustawianie wartości w bitsecie
			b.Set(index, r) // R
			index++
			b.Set(index, g) // G
			index++
			b.Set(index, bValue) // B
			index++
			b.Set(index, a) // A
			index++
		}
	}

	return b
}

type myImageSimulator struct{}

func (sim *myImageSimulator) OnBeginSimulation() {}

func (sim *myImageSimulator) OnEndSimulation() {}

func (sim *myImageSimulator) Simulate(g goga.Genome) {
    imgGenome, ok := g.(*ImageGenome)
	// ten kod jest niepoprawny. Najpierw degradujemy obiekt typu ImageGenome do goga.Genome, a potem próbujemy przywrócić go do ImageGenome. Gdyby do funkcji był przekazywany pointer do obiektu, to by się dało, ale nie można zmienić
	// Pozostaje zmiana kodu, aby imageData (btw. jest nieeksportowany, więc możliwe są dalsze błędy) nie znajdował się w genomie
	// fitness jest domyślnie w obiekcie tylko jako int, po co zmiana na float64?
	//odp: fitness jest lepiej użuwac jako float64 i masz rację trzeba to zmienić bo domyślnie w struct jest float64
    if !ok {
        log.Println("Errot: g to nie *ImageGenome")
        return
    }

	//compressedImg := compressImage(imgGenome)
    evaluateFitness(imgGenome)
}

func (sim *myImageSimulator) ExitFunc(g goga.Genome) bool {
	return false
}

// funkcja do wczytywania obrazu z pliku
func loadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var img image.Image
	switch ext := getFileExtension(filePath); ext {
	case ".png":
		img, err = png.Decode(file)
	case ".bmp":
		img, err = bmp.Decode(file)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
	return img, err
}

// funkcja do zapisu obrazu jako plik JPEG
func saveImageAsJPEG(img image.Image, filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return jpeg.Encode(file, img, nil)
}

// funkcja pomocnicza by uzyskać rozszerzenie pliku
func getFileExtension(filePath string) string {
	// idk czy to będzie działać
	if len(filePath) < 4 {
		return ""
	}
	return filePath[len(filePath)-4:]
}
